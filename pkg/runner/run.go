// Package runner — T-005-42: runner.Run main loop.
// Package runner — T-005-46: --lock-timeout / --no-wait flag plumbing.
//
// Run is the single public entry point for story execution. It
// builds the dependency DAG, computes the topological order, and
// drives the scheduler + worker pool until all stories reach a
// terminal status. The lock timeout and no-wait flag from Config
// are plumbed through to the cache lock subsystem on every story
// execution (ADR 0016 §"Lock timeout and fast-fail", ADR 0019
// §"Lock timeout and fast-fail").

package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/evcoreco/octane/pkg/cache"
	"github.com/evcoreco/octane/pkg/engine/clock"
	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/registry"
	"github.com/evcoreco/octane/pkg/runner/internal/dag"
	"github.com/evcoreco/octane/pkg/story"
	"github.com/evcoreco/octane/pkg/story/ast"
)

// defaultLockTimeout is the lock acquisition timeout used when
// Config.LockTimeout is zero (spec 005 G6, ADR 0019 §"Lock timeout
// and fast-fail").
const defaultLockTimeout = 60 * time.Second

// ocppVersion16 is the only supported OCPP version string.
const ocppVersion16 = "1.6"

// emptyCause is the zero value for the Cause field in StoryResult.
const emptyCause = ""

// placeholderSHA is the eight-zero-digit placeholder used for SHA
// components of the cache key that are not yet implemented.
const placeholderSHA = "00000000"

// errStationNotRegistered is returned when a station handle is not registered.
var errStationNotRegistered = errors.New("runner: station not registered")

// ocppVersionEmpty is used when the story does not declare an OCPP version.
const ocppVersionEmpty = ""

// runnerState implements api.State for use by keyword functions
// during story execution. It wraps the station registry and the
// deterministic clock injected by the runner.
type runnerState struct {
	// mu protects stations, stash, and logLines from concurrent
	// modification. Keywords may be invoked from multiple goroutines.
	mu sync.Mutex

	// stations maps station handle → api.Station instance.
	stations map[string]api.Station

	// clk is the deterministic clock injected by the runner.
	clk clock.Clock

	// stash is the per-story key–value scratch space consumed by
	// api.State.Stash and api.State.Pop.
	stash map[string]any

	// logLines accumulates log output from api.State.Logf.
	logLines []string
}

// newRunnerState creates a fresh runnerState backed by the given
// clock. The clock must be non-nil; use clock.Real() in production.
func newRunnerState(clk clock.Clock) *runnerState {
	return &runnerState{
		mu:       sync.Mutex{},
		stations: make(map[string]api.Station),
		clk:      clk,
		stash:    make(map[string]any),
		logLines: nil,
	}
}

// Station implements api.State.
func (rs *runnerState) Station(handle string) (api.Station, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	station, ok := rs.stations[handle]
	if !ok {
		return nil, fmt.Errorf("runner: station %q: %w", handle, errStationNotRegistered)
	}

	return station, nil
}

// RegisterStation implements api.State.
func (rs *runnerState) RegisterStation(handle string, station api.Station) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.stations[handle] = station
}

// Now implements api.State. It delegates to the injected clock so
// that the runner satisfies constitution principle IV (no direct
// calls to time.Now in keyword/runner code).
func (rs *runnerState) Now() time.Time {
	return rs.clk.Now()
}

// Sleep implements api.State. It delegates to the injected clock.
func (rs *runnerState) Sleep(
	ctx context.Context,
	duration time.Duration,
) error {
	err := rs.clk.Sleep(ctx, duration)
	if err != nil {
		return fmt.Errorf("runner: sleep: %w", err)
	}

	return nil
}

// Stash implements api.State.
func (rs *runnerState) Stash(key string, value any) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.stash[key] = value
}

// Pop implements api.State.
func (rs *runnerState) Pop(key string) (any, bool) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	value, ok := rs.stash[key]
	if ok {
		delete(rs.stash, key)
	}

	return value, ok
}

// Logf implements api.State. Log lines are collected in-memory and
// surfaced as Findings on the StoryResult.
func (rs *runnerState) Logf(format string, args ...any) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.logLines = append(rs.logLines, fmt.Sprintf(format, args...))
}

// Run is the single public entry point for the runner.
//
// Run discovers all .story files under cfg.StoryPaths, applies the
// shard filter (when cfg.ShardTotal > 0), builds the dependency DAG,
// detects cycles (returns ErrCycle), computes the stable topological
// order, and drives the scheduler + worker pool.
//
// The returned RunResult is always non-nil on a nil error. On a
// non-nil error (ErrCycle, discovery failure) RunResult is nil.
//
// Run respects ctx cancellation: in-flight stories abort and
// remaining pending stories are marked StatusSkipped.
func Run(ctx context.Context, cfg Config) (*RunResult, error) {
	clk := clock.Real()
	startedAt := clk.Now()

	runID := generateRunID(clk)

	cacheDir, err := resolveCacheDir(cfg.CacheDir)
	if err != nil {
		return nil, fmt.Errorf("runner: resolve cache dir: %w", err)
	}

	var storyCache cache.Cache

	if !cfg.NoCache {
		storyCache, err = cache.Open(cacheDir)
		if err != nil {
			return nil, fmt.Errorf("runner: open cache: %w", err)
		}
	}

	stories, err := discoverStories(cfg)
	if err != nil {
		return nil, fmt.Errorf("runner: discover stories: %w", err)
	}

	if len(stories) == 0 {
		return &RunResult{
			RunID:      runID,
			StartedAt:  startedAt,
			FinishedAt: clk.Now(),
			Stories:    nil,
			Summary: Summary{
				Total:     0,
				Passed:    0,
				Failed:    0,
				Skipped:   0,
				CacheHits: 0,
			},
		}, nil
	}

	dagResult, err := buildDAG(stories, runID, 0)
	if err != nil {
		return nil, err
	}

	topoNodes, err := dag.TopologicalOrder(dagResult.graph)
	if err != nil {
		var errCycle *dag.CycleError
		if errors.As(err, &errCycle) {
			return nil, fmt.Errorf("%w: %w", ErrCycle, errCycle)
		}

		return nil, fmt.Errorf("runner: topological order: %w", err)
	}

	topoOrder := make([]string, len(topoNodes))

	for idx, topoNode := range topoNodes {
		topoOrder[idx] = topoNode.ID
	}

	schedState := newSchedulerState(dagResult, topoOrder)

	lockTimeout := cfg.LockTimeout
	if lockTimeout == 0 && !cfg.NoWait {
		lockTimeout = defaultLockTimeout
	}

	lockDir := filepath.Join(cacheDir, "locks")

	results := runScheduler(
		ctx,
		cfg,
		schedState,
		dagResult,
		storyCache,
		lockDir,
		lockTimeout,
		clk,
		runID,
	)

	finishedAt := clk.Now()

	return buildRunResult(
		runID,
		startedAt,
		finishedAt,
		topoOrder,
		results,
	), nil
}

// runScheduler drives the scheduler + worker pool until all nodes
// are in nodeDone. It returns a map of nodeID → StoryResult.
func runScheduler(
	ctx context.Context,
	cfg Config,
	schedState *schedulerState,
	dagResult *buildDAGResult,
	storyCache cache.Cache,
	lockDir string,
	lockTimeout time.Duration,
	clk clock.Clock,
	runID string,
) map[string]StoryResult {
	maxParallel := cfg.MaxParallel
	if maxParallel <= 0 {
		maxParallel = 1
	}

	pool := newWorkerPool(ctx, maxParallel)

	// inProcess is an in-process sync.Map[nodeID → *sync.Once]
	// that provides the fast-path double-check before the filesystem
	// lock is consulted (ADR 0019 §"Interaction with cache lock").
	var inProcess sync.Map

	// dispatchBatch sends all currently eligible nodes to the pool
	// up to the pool's remaining capacity.
	dispatchBatch := func() {
		eligible := schedState.eligibleNodes()
		capacity := maxParallel - schedState.running

		for _, nodeID := range eligible {
			if capacity <= 0 {
				break
			}

			// Guard against double-dispatch (defensive).
			if schedState.status[nodeID] != nodePending {
				continue
			}

			sn := dagResult.nodes[dagResult.index[nodeID]]

			execFn := makeExecFunc(
				nodeID,
				sn,
				cfg,
				storyCache,
				lockDir,
				lockTimeout,
				clk,
				runID,
				&inProcess,
			)

			item := workItem{
				nodeID:  nodeID,
				execute: execFn,
			}

			if !pool.submit(ctx, item) {
				break
			}

			schedState.status[nodeID] = nodeRunning
			schedState.running++
			capacity--
		}
	}

	// completionLoop drains completion events and drives the
	// scheduler until no nodes remain pending or running. It runs
	// in a goroutine so that pool.close() can be deferred safely.
	completionLoop := func() {
		defer pool.close()

		dispatchBatch()

		for schedState.pendingCount() > 0 {
			select {
			case <-ctx.Done():
				cancelPendingNodes(ctx, schedState)

				return

			case done, ok := <-pool.completionCh:
				if !ok {
					return
				}

				schedState.status[done.nodeID] = nodeDone
				schedState.result[done.nodeID] = done.result
				schedState.running--

				if done.result.Status == StatusFailed {
					schedState.propagateFailures(
						done.nodeID,
						done.result,
					)
				}

				dispatchBatch()
			}
		}
	}

	// Run completionLoop in a goroutine and block until it finishes.
	// completionLoop owns pool.completionCh exclusively; the main
	// goroutine must NOT read from completionCh while completionLoop is
	// running (doing so races and steals completion items, causing the
	// scheduler to block forever). A sync.WaitGroup is the simplest
	// correct synchronization here.
	var workGroup sync.WaitGroup

	workGroup.Go(func() {
		completionLoop()
	})

	workGroup.Wait()

	return schedState.result
}

// makeExecFunc constructs the execution function for a single story
// node. The returned function follows the double-checked acquire
// pattern from ADR 0016 §"Acquire pattern".
func makeExecFunc(
	nodeID string,
	storyNodeVal storyNode,
	cfg Config,
	storyCache cache.Cache,
	lockDir string,
	lockTimeout time.Duration,
	clk clock.Clock,
	runID string,
	inProcess *sync.Map,
) func(ctx context.Context) StoryResult {
	return func(ctx context.Context) StoryResult {
		startedAt := clk.Now()

		cacheKey := buildCacheKey(storyNodeVal, cfg, runID)

		// Step 1: fast path — check cache without lock.
		if !cfg.NoCache {
			entry, entryErr := storyCache.Get(ctx, cacheKey)
			if entryErr == nil {
				return cacheHitResult(storyNodeVal, entry, startedAt, clk)
			}
		}

		if cfg.NoCache {
			result := executeStory(ctx, storyNodeVal, cfg, clk)
			result.StartedAt = startedAt
			result.FinishedAt = clk.Now()
			result.CacheStatus = CacheBypassed

			return result
		}

		// In-process deduplication via sync.Once (ADR 0019 fast path).
		rawOnce, _ := inProcess.LoadOrStore(nodeID, &sync.Once{})
		once, ok := rawOnce.(*sync.Once)

		if !ok {
			panic("internal: inProcess value is not *sync.Once")
		}

		var lockedResult StoryResult

		once.Do(func() {
			lockedResult = executeWithLock(
				ctx,
				nodeID,
				storyNodeVal,
				cfg,
				storyCache,
				cacheKey,
				lockDir,
				lockTimeout,
				clk,
				startedAt,
			)
		})

		// If another goroutine ran the Once, re-read from cache.
		if lockedResult.TestID == "" {
			entry, entryErr := storyCache.Get(ctx, cacheKey)
			if entryErr == nil {
				return cacheHitResult(storyNodeVal, entry, startedAt, clk)
			}

			// Cache miss even after the Once: execute fresh.
			result := executeStory(ctx, storyNodeVal, cfg, clk)
			result.StartedAt = startedAt
			result.FinishedAt = clk.Now()

			return result
		}

		return lockedResult
	}
}

// executeWithLock implements steps 2–7 of the ADR 0016 acquire
// pattern: acquire flock, re-read cache, execute, write cache,
// release flock.
func executeWithLock(
	ctx context.Context,
	_ string,
	storyNodeVal storyNode,
	cfg Config,
	storyCache cache.Cache,
	cacheKey cache.Key,
	lockDir string,
	lockTimeout time.Duration,
	clk clock.Clock,
	startedAt time.Time,
) StoryResult {
	lockPath := filepath.Join(lockDir, cacheKey.Hash()+".lock")

	lockCloser, lockErr := cache.AcquireLock(
		ctx,
		lockPath,
		lockTimeout,
		cfg.NoWait,
	)
	if lockErr != nil {
		return StoryResult{
			Order:       0,
			TestID:      storyNodeVal.story.Meta.ID,
			ScopeKey:    storyNodeVal.scopeKey,
			OCPPVersion: ocppVersionEmpty,
			Status:      StatusFailed,
			CacheStatus: CacheMiss,
			StartedAt:   startedAt,
			FinishedAt:  clk.Now(),
			Findings: []Finding{
				{
					Message:  "lock acquisition failed: " + lockErr.Error(),
					Severity: "error",
				},
			},
			Trace:      nil,
			Cause:      emptyCause,
			CauseChain: nil,
		}
	}

	defer func() {
		_ = lockCloser.Close()
	}()

	// Step 3: re-read after acquiring lock (double-checked locking).
	entry, entryErr := storyCache.Get(ctx, cacheKey)
	if entryErr == nil {
		return cacheHitResult(storyNodeVal, entry, startedAt, clk)
	}

	// Step 4: execute the story.
	result := executeStory(ctx, storyNodeVal, cfg, clk)
	result.StartedAt = startedAt
	result.FinishedAt = clk.Now()

	// Steps 5–6: write result and trace to the cache.
	if result.Status == StatusPassed || result.Status == StatusFailed {
		writeToCache(
			ctx,
			storyCache,
			cacheKey,
			result,
			storyNodeVal.story.Meta.CacheTTL,
		)
	}

	return result
}

// cacheHitResult builds a StoryResult from a cache entry, mapping
// the recorded status to the appropriate CacheStatus value.
func cacheHitResult(
	storyNodeVal storyNode,
	entry *cache.Entry,
	startedAt time.Time,
	clk clock.Clock,
) StoryResult {
	cacheStatus := CacheHitPass

	// Determine whether the cached result was a pass or skip.
	var recorded struct {
		Status string `json:"status"`
	}

	err := json.Unmarshal(entry.Result, &recorded)
	if err == nil {
		if recorded.Status == "skipped" {
			cacheStatus = CacheHitSkip
		}
	}

	resultStatus := StatusPassed
	if cacheStatus == CacheHitSkip {
		resultStatus = StatusSkipped
	}

	return StoryResult{
		Order:       0,
		TestID:      storyNodeVal.story.Meta.ID,
		ScopeKey:    storyNodeVal.scopeKey,
		OCPPVersion: "",
		Status:      resultStatus,
		CacheStatus: cacheStatus,
		StartedAt:   startedAt,
		FinishedAt:  clk.Now(),
		Findings:    nil,
		Trace:       nil,
		Cause:       emptyCause,
		CauseChain:  nil,
	}
}

// executeStory invokes the resolved keyword functions for each step
// in the story's Background, Setup, Scenarios, and Teardown sections.
// It returns a StoryResult whose Status is StatusPassed or StatusFailed.
//
// The timing fields (StartedAt, FinishedAt) are set by the caller.
func executeStory(
	ctx context.Context,
	storyNodeVal storyNode,
	cfg Config,
	clk clock.Clock,
) StoryResult {
	state := newRunnerState(clk)

	ocppVer := resolveOCPPVersion(storyNodeVal.story.Meta.Tags, cfg.OCPPVersion)

	findings := make([]Finding, 0, len(state.logLines))

	result := executeAllSections(
		ctx,
		storyNodeVal.story,
		state,
		ocppVer,
		&findings,
	)

	// Merge any Logf lines from the runnerState as info findings.
	for _, line := range state.logLines {
		findings = append(findings, Finding{
			Message:  line,
			Severity: "info",
		})
	}

	result.TestID = storyNodeVal.story.Meta.ID
	result.ScopeKey = storyNodeVal.scopeKey
	result.OCPPVersion = ocppVer
	result.Findings = findings
	result.CacheStatus = CacheMiss

	return result
}

// executeAllSections runs Background, Setup, Scenarios, and Teardown
// in order. Teardown always runs even when a prior section fails
// (best-effort cleanup).
func executeAllSections(
	ctx context.Context,
	storyAST *ast.Story,
	state api.State,
	ocppVer string,
	findings *[]Finding,
) StoryResult {
	failed := false

	failed = runSteps(ctx, storyAST.Background, state, ocppVer, findings) ||
		failed
	failed = runSteps(ctx, storyAST.Setup, state, ocppVer, findings) || failed

	for _, scenario := range storyAST.Scenarios {
		failed = runSteps(
			ctx, scenario.Steps, state, ocppVer, findings,
		) || failed
	}

	// Teardown always runs (best-effort).
	_ = runSteps(ctx, storyAST.Teardown, state, ocppVer, findings)

	resultStatus := StatusPassed
	if failed {
		resultStatus = StatusFailed
	}

	return StoryResult{
		Order:       0,
		TestID:      "",
		ScopeKey:    "",
		OCPPVersion: "",
		Status:      resultStatus,
		CacheStatus: CacheMiss,
		StartedAt:   time.Time{},
		FinishedAt:  time.Time{},
		Findings:    nil,
		Trace:       nil,
		Cause:       emptyCause,
		CauseChain:  nil,
	}
}

// runSteps invokes the keyword function for each step in steps.
// It returns true when any step fails.
func runSteps(
	ctx context.Context,
	steps []ast.Step,
	state api.State,
	ocppVer string,
	findings *[]Finding,
) bool {
	failed := false

	for _, step := range steps {
		err := runStep(ctx, step, state, ocppVer, findings)
		if err != nil {
			failed = true
		}
	}

	return failed
}

// runStep resolves the keyword for step and invokes it.
func runStep(
	ctx context.Context,
	step ast.Step,
	state api.State,
	ocppVer string,
	findings *[]Finding,
) error {
	ver := parseOCPPVersion(ocppVer)

	match, err := registry.Resolve(step.Text, ver)
	if err != nil {
		*findings = append(*findings, Finding{
			Message:  fmt.Sprintf("step %q: %v", step.Text, err),
			Severity: "error",
		})

		return fmt.Errorf("runner: resolve keyword: %w", err)
	}

	err = match.Keyword.Func(ctx, state, match.Args)
	if err != nil {
		*findings = append(*findings, Finding{
			Message:  fmt.Sprintf("step %q: %v", step.Text, err),
			Severity: "error",
		})

		return fmt.Errorf("runner: execute keyword: %w", err)
	}

	return nil
}

// writeToCache serialises the StoryResult and writes it to the
// cache. Errors are silently dropped (a write failure is not fatal;
// the next run will re-execute).
//
// cacheTTL is the override from the story's Meta.CacheTTL field.
// When nil the entry uses TTL=0 (no expiry, valid indefinitely).
//
// Constitution principle X: no credentials must appear in the cache.
// At this layer we trust that keyword functions do not expose
// credentials via the StoryResult fields.
func writeToCache(
	ctx context.Context,
	storyCache cache.Cache,
	key cache.Key,
	result StoryResult,
	cacheTTL *time.Duration,
) {
	encoded, err := json.Marshal(struct {
		Status string `json:"status"`
	}{Status: result.Status.String()})
	if err != nil {
		return
	}

	var ttl time.Duration
	if cacheTTL != nil {
		ttl = *cacheTTL
	}

	entry := cache.Entry{
		Result:    encoded,
		Trace:     nil,
		WrittenAt: time.Time{},
		TTL:       ttl,
	}

	_ = storyCache.Put(ctx, key, entry)
}

// buildCacheKey constructs the cache.Key for a story node. Several
// fields (CSMSEndpointSHA, StoryContentSHA, ParameterSHA) require
// information not yet available in this phase (spec 002, spec 003);
// they are set to placeholders so the key derivation is structurally
// correct.
func buildCacheKey(
	storyNodeVal storyNode,
	cfg Config,
	runID string,
) cache.Key {
	scopeKey := storyNodeVal.scopeKey

	for _, dep := range storyNodeVal.story.Meta.Depends {
		if dep.Scope == ast.ScopePerRun {
			scopeKey = runID

			break
		}

		if dep.Scope == ast.ScopeGlobal {
			scopeKey = ""

			break
		}
	}

	ocppVer := cfg.OCPPVersion
	if ocppVer == "" {
		ocppVer = statusUnknown
	}

	return cache.Key{
		TestID:          storyNodeVal.story.Meta.ID,
		ScopeKey:        scopeKey,
		CSMSEndpointSHA: placeholderSHA, // spec 002 placeholder
		OctaneVersion:   "dev",          // spec 006 injects real version
		OCPPVersion:     ocppVer,
		StoryContentSHA: placeholderSHA, // spec 001 content hash
		ParameterSHA:    placeholderSHA, // spec 003 parameter hash
	}
}

// buildRunResult assembles the final RunResult from the scheduler
// state, preserving the stable topological order (ADR 0019 §
// "Dispatch order and determinism").
func buildRunResult(
	runID string,
	startedAt,
	finishedAt time.Time,
	topoOrder []string,
	results map[string]StoryResult,
) *RunResult {
	stories := make([]StoryResult, 0, len(topoOrder))

	summary := Summary{
		Total:     len(topoOrder),
		Passed:    0,
		Failed:    0,
		Skipped:   0,
		CacheHits: 0,
	}

	for orderIdx, nodeID := range topoOrder {
		result, ok := results[nodeID]
		if !ok {
			// Node never executed; mark as skipped.
			storyID, scopeKey := splitNodeID(nodeID)
			result = StoryResult{
				Order:       orderIdx,
				TestID:      storyID,
				ScopeKey:    scopeKey,
				OCPPVersion: ocppVersionEmpty,
				Status:      StatusSkipped,
				CacheStatus: CacheMiss,
				StartedAt:   time.Time{},
				FinishedAt:  time.Time{},
				Findings: []Finding{
					{Message: "not executed", Severity: "info"},
				},
				Trace:      nil,
				Cause:      emptyCause,
				CauseChain: nil,
			}
		}

		result.Order = orderIdx
		stories = append(stories, result)

		switch result.Status {
		case StatusPassed:
			summary.Passed++
		case StatusFailed:
			summary.Failed++
		case StatusSkipped:
			summary.Skipped++
		default:
			// no-op: unrecognised status does not affect summary counts
		}

		if result.CacheStatus == CacheHitPass ||
			result.CacheStatus == CacheHitSkip {
			summary.CacheHits++
		}
	}

	slices.SortFunc(stories, func(a, b StoryResult) int {
		return a.Order - b.Order
	})

	return &RunResult{
		RunID:      runID,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
		Stories:    stories,
		Summary:    summary,
	}
}

// discoverStories reads cfg.StoryPaths, recursively discovers .story
// files, parses them, applies OCPPVersion and shard filters, and
// collects prerequisites.
func discoverStories(cfg Config) ([]*ast.Story, error) {
	var allParsed []*ast.Story

	for _, root := range cfg.StoryPaths {
		found, err := walkStoryFiles(root)
		if err != nil {
			return nil, fmt.Errorf("walk %q: %w", root, err)
		}

		allParsed = append(allParsed, found...)
	}

	// Build a lookup index of all parsed stories.
	allIndex := make(map[string]*ast.Story, len(allParsed))

	for _, storyAST := range allParsed {
		allIndex[storyAST.Meta.ID] = storyAST
	}

	filtered := allParsed

	// Apply OCPP version filter.
	if cfg.OCPPVersion != "" {
		filtered = filterByOCPPVersion(filtered, cfg.OCPPVersion)
	}

	// Apply shard filter.
	sharded := applyShardFilter(filtered, cfg.ShardIndex, cfg.ShardTotal)

	// Collect prerequisites from the full index so that stories
	// outside the shard that are prerequisites of sharded stories
	// are included (ADR 0019 §"Prerequisite inclusion").
	return collectPrerequisites(sharded, allIndex), nil
}

// walkStoryFiles recursively finds and parses all .story files
// under root.
func walkStoryFiles(root string) ([]*ast.Story, error) {
	var stories []*ast.Story

	err := filepath.WalkDir(
		root,
		func(
			path string,
			dirEntry fs.DirEntry,
			walkErr error,
		) error {
			if walkErr != nil {
				return walkErr
			}

			if dirEntry.IsDir() || filepath.Ext(path) != ".story" {
				return nil
			}

			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return fmt.Errorf("read %q: %w", path, readErr)
			}

			storyAST, parseErr := story.Parse(path, data)
			if parseErr != nil {
				return fmt.Errorf("parse %q: %w", path, parseErr)
			}

			stories = append(stories, storyAST)

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("runner: walk stories: %w", err)
	}

	return stories, nil
}

// filterByOCPPVersion removes stories that do not declare the
// requested OCPP version via their tags.
func filterByOCPPVersion(
	stories []*ast.Story,
	version string,
) []*ast.Story {
	out := make([]*ast.Story, 0, len(stories))

	for _, storyAST := range stories {
		for _, tag := range storyAST.Meta.Tags {
			if tag == "ocpp"+version || tag == version {
				out = append(out, storyAST)

				break
			}
		}
	}

	if len(out) == 0 {
		// No story declared the version via tags; return all and
		// let keyword resolution handle version scoping.
		return stories
	}

	return out
}

// resolveOCPPVersion determines the OCPP version string to use for
// keyword resolution. It checks the story's tags for the ocpp1.6
// marker and falls back to the config's OCPPVersion.
func resolveOCPPVersion(tags []string, cfgVersion string) string {
	for _, tag := range tags {
		if tag == "ocpp1.6" || tag == ocppVersion16 {
			return ocppVersion16
		}
	}

	if cfgVersion != "" {
		return cfgVersion
	}

	return ocppVersion16
}

// parseOCPPVersion converts a version string to api.OCPPVersion.
// Only OCPP 1.6 is supported.
func parseOCPPVersion(_ string) api.OCPPVersion {
	return api.OCPP16
}

// resolveCacheDir returns the effective cache directory, applying
// the XDG_CACHE_HOME / HOME fallback when dir is empty.
func resolveCacheDir(dir string) (string, error) {
	if dir != "" {
		return dir, nil
	}

	if envDir := os.Getenv("OCTANE_CACHE_DIR"); envDir != "" {
		return envDir, nil
	}

	if xdgHome := os.Getenv("XDG_CACHE_HOME"); xdgHome != "" {
		return filepath.Join(xdgHome, "octane", "cache"), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}

	return filepath.Join(homeDir, ".cache", "octane", "cache"), nil
}

// generateRunID produces a time-ordered unique identifier for the
// run. Format is a 16-character hex string encoding the nanosecond
// Unix timestamp. A ULID library is intentionally avoided to honour
// the no-new-dependency rule (constitution principle V).
func generateRunID(clk clock.Clock) string {
	now := clk.Now()

	return fmt.Sprintf("%016x", now.UnixNano())
}

// splitNodeID parses a DAG node ID of the form "story_id/scope_key"
// or just "story_id" (when scope is global). Story IDs are
// snake_case (no "/"); scope keys are alphanumeric station handles
// ("CP01") or run IDs (hex strings). The split is safe because
// story IDs never contain a "/".
func splitNodeID(nodeID string) (string, string) {
	if before, after, ok := strings.Cut(nodeID, "/"); ok {
		return before, after
	}

	return nodeID, ""
}
