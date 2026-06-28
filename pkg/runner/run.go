package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

// defaultMaxParallel is the fallback parallelism when Config.MaxParallel
// is zero or negative (sequential execution mode).
const defaultMaxParallel = 1

// emptyTestID is the sentinel used when a TestID is not yet populated
// (for example, a lock-failure StoryResult before executeStory runs).
const emptyTestID = ""

// ocppVersionEmpty is used when the story does not declare an OCPP version.
const ocppVersionEmpty = ""

// emptyString is the zero-value sentinel for string-typed guards and
// return values where a named domain constant does not apply.
const emptyString = ""

// emptyLen is the zero-length sentinel used in collection size checks
// (e.g., len(x) == emptyLen) to satisfy the add-constant linter rule.
const emptyLen = 0

// noOverride is passed to buildDAG when no station-count override is
// needed; the function then uses the value declared in the story AST.
const noOverride = 0

// noLockTimeout is the zero Duration sentinel for the lock-timeout
// guard in resolveLockTimeout.
const noLockTimeout = 0

// noCapacity is the capacity sentinel used in dispatcher guards
// (capacity <= noCapacity) meaning the pool has no remaining slots.
const noCapacity = 0

// noPendingCount is the lower bound checked in the completion loop;
// a value of 0 means all nodes have been dispatched or completed.
const noPendingCount = 0

// noParallel is the sentinel for an unset or non-positive MaxParallel
// config value, meaning sequential execution.
const noParallel = 0

// nodeIDParts holds the two components decoded from a DAG node ID by
// splitNodeID. Using a struct avoids the confusing-results linter
// violation that arises when a function returns two unnamed strings.
type nodeIDParts struct {
	storyID  string
	scopeKey string
}

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

	// csmsBaseURL is the base WebSocket URL of the CSMS under test,
	// sourced from Config.CSMSEndpoint. Exposed via CSMSBaseURL() to
	// lifecycle domain keywords.
	csmsBaseURL string
}

// newRunnerState creates a fresh runnerState backed by the given
// clock and CSMS base URL. The clock must be non-nil; use clock.Real()
// in production. csmsBaseURL may be empty if no CSMS endpoint is
// configured (lifecycle keywords will return a descriptive error).
func newRunnerState(clk clock.Clock, csmsBaseURL string) *runnerState {
	return &runnerState{
		mu:          sync.Mutex{},
		stations:    make(map[string]api.Station),
		clk:         clk,
		stash:       make(map[string]any),
		logLines:    nil,
		csmsBaseURL: csmsBaseURL,
	}
}

// Station implements api.State.
func (rs *runnerState) Station(handle string) (api.StationValue, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	station, ok := rs.stations[handle]
	if !ok {
		return api.StationValue{}, fmt.Errorf(
			"runner: station %q: %w",
			handle,
			errStationNotRegistered,
		)
	}

	return api.StationValue{Station: station}, nil
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

// CSMSBaseURL implements api.State. It returns the base WebSocket URL of
// the CSMS under test, as sourced from Config.CSMSEndpoint at run time.
func (rs *runnerState) CSMSBaseURL() string {
	return rs.csmsBaseURL
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

	cacheDir, storyCache, err := openRunCache(cfg)
	if err != nil {
		return nil, err
	}

	stories, err := discoverStories(cfg)
	if err != nil {
		return nil, fmt.Errorf("runner: discover stories: %w", err)
	}

	if len(stories) == emptyLen {
		return emptyRunResult(runID, startedAt, clk.Now()), nil
	}

	dagResult, topoOrder, err := buildDAGAndSort(stories, runID)
	if err != nil {
		return nil, err
	}

	lockTimeout := resolveLockTimeout(cfg)
	lockDir := filepath.Join(cacheDir, "locks")
	schedState := newSchedulerState(dagResult, topoOrder)

	results := runScheduler(ctx, runSchedulerArgs{
		cfg:         cfg,
		schedState:  schedState,
		dagResult:   dagResult,
		storyCache:  storyCache,
		lockDir:     lockDir,
		lockTimeout: lockTimeout,
		clk:         clk,
		runID:       runID,
	})

	return buildRunResult(runID, startedAt, clk.Now(), topoOrder, results), nil
}

// openRunCache resolves the cache directory and opens the cache when enabled.
// It returns the cache dir and the open cache (nil when cfg.NoCache is set).
func openRunCache(cfg Config) (string, *cache.FileCache, error) {
	cacheDir, err := resolveCacheDir(cfg.CacheDir)
	if err != nil {
		return emptyString, nil, fmt.Errorf(
			"runner: resolve cache dir: %w", err,
		)
	}

	if cfg.NoCache {
		return cacheDir, nil, nil
	}

	storyCache, err := cache.Open(cacheDir)
	if err != nil {
		return emptyString, nil, fmt.Errorf(
			"runner: open cache: %w", err,
		)
	}

	return cacheDir, storyCache, nil
}

// emptyRunResult returns a RunResult with no stories executed.
func emptyRunResult(
	runID string,
	startedAt,
	finishedAt time.Time,
) *RunResult {
	return &RunResult{
		RunID:      runID,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
		Stories:    nil,
		Summary: Summary{
			Total:     emptyLen,
			Passed:    emptyLen,
			Failed:    emptyLen,
			Skipped:   emptyLen,
			CacheHits: emptyLen,
		},
	}
}

// buildDAGAndSort builds the dependency DAG from stories and returns the
// buildDAGResult plus the stable topological node order.
func buildDAGAndSort(
	stories []*ast.Story,
	runID string,
) (*buildDAGResult, []string, error) {
	dagResult, err := buildDAG(stories, runID, noOverride)
	if err != nil {
		return nil, nil, err
	}

	topoNodes, err := dag.TopologicalOrder(dagResult.graph)
	if err != nil {
		var errCycle *dag.CycleError
		if errors.As(err, &errCycle) {
			return nil, nil, fmt.Errorf("%w: %w", ErrCycle, errCycle)
		}

		return nil, nil, fmt.Errorf("runner: topological order: %w", err)
	}

	topoOrder := make([]string, len(topoNodes))

	for idx, topoNode := range topoNodes {
		topoOrder[idx] = topoNode.ID
	}

	return dagResult, topoOrder, nil
}

// resolveLockTimeout returns the effective lock timeout from cfg. When
// LockTimeout is zero and NoWait is false it defaults to defaultLockTimeout.
func resolveLockTimeout(cfg Config) time.Duration {
	if cfg.LockTimeout != noLockTimeout || cfg.NoWait {
		return cfg.LockTimeout
	}

	return defaultLockTimeout
}

// schedRunner groups the shared state for one runScheduler invocation so that
// dispatchBatch and the completion loop can be extracted as methods instead of
// capturing closures.
type schedRunner struct {
	cfg         Config
	schedState  *schedulerState
	dagResult   *buildDAGResult
	storyCache  cache.Cache
	lockDir     string
	lockTimeout time.Duration
	clk         clock.Clock
	runID       string
	maxParallel int
	pool        *workerPool
	inProcess   sync.Map
	// storyIdx maps story ID → *ast.Story used when inlining prerequisites.
	storyIdx storyIndex
}

// runSchedulerArgs bundles the arguments for [runScheduler] that
// would otherwise exceed the 8-parameter limit.
type runSchedulerArgs struct {
	cfg         Config
	schedState  *schedulerState
	dagResult   *buildDAGResult
	storyCache  cache.Cache
	lockDir     string
	lockTimeout time.Duration
	clk         clock.Clock
	runID       string
}

// runScheduler drives the scheduler + worker pool until all nodes
// are in nodeDone. It returns a map of nodeID → StoryResult.
func runScheduler(
	ctx context.Context,
	args runSchedulerArgs,
) map[string]StoryResult {
	maxParallel := args.cfg.MaxParallel
	if maxParallel <= noParallel {
		maxParallel = defaultMaxParallel
	}

	sched := &schedRunner{
		cfg:         args.cfg,
		schedState:  args.schedState,
		dagResult:   args.dagResult,
		storyCache:  args.storyCache,
		lockDir:     args.lockDir,
		lockTimeout: args.lockTimeout,
		clk:         args.clk,
		runID:       args.runID,
		maxParallel: maxParallel,
		pool:        newWorkerPool(ctx, maxParallel),
		inProcess:   sync.Map{},
		storyIdx:    buildStoryIdx(args.dagResult),
	}

	var workGroup sync.WaitGroup

	workGroup.Go(func() {
		sched.completionLoop(ctx)
	})

	workGroup.Wait()

	return args.schedState.result
}

// buildStoryIdx builds a storyIndex from all nodes in result.
// Multiple nodes with the same story ID (different scopes) map to the
// same *ast.Story pointer, so last-write-wins is correct.
func buildStoryIdx(result *buildDAGResult) storyIndex {
	idx := make(storyIndex, len(result.nodes))

	for _, node := range result.nodes {
		idx[node.story.Meta.ID] = node.story
	}

	return idx
}

// dispatchBatch sends all currently eligible nodes to the pool up to the
// pool's remaining capacity.
func (sr *schedRunner) dispatchBatch(ctx context.Context) {
	eligible := sr.schedState.eligibleNodes()
	capacity := sr.maxParallel - sr.schedState.running

	for _, nodeID := range eligible {
		if capacity <= noCapacity {
			break
		}

		if sr.schedState.status[nodeID] != nodePending {
			continue
		}

		if !sr.submitNode(ctx, nodeID, capacity) {
			break
		}

		capacity--
	}
}

// submitNode builds the exec function for nodeID and submits it to the pool.
// It returns false when the pool rejected the item (pool is full or closed).
func (sr *schedRunner) submitNode(
	ctx context.Context,
	nodeID string,
	_ int,
) bool {
	sn := sr.dagResult.nodes[sr.dagResult.index[nodeID]]

	execFn := makeExecFunc(execParams{
		nodeID:      nodeID,
		storyNode:   sn,
		cfg:         sr.cfg,
		storyCache:  sr.storyCache,
		lockDir:     sr.lockDir,
		lockTimeout: sr.lockTimeout,
		clk:         sr.clk,
		runID:       sr.runID,
		inProcess:   &sr.inProcess,
		storyIdx:    sr.storyIdx,
	})

	item := workItem{nodeID: nodeID, execute: execFn}

	if !sr.pool.submit(ctx, item) {
		return false
	}

	sr.schedState.status[nodeID] = nodeRunning
	sr.schedState.running++

	return true
}

// completionLoop drains completion events and drives the scheduler until no
// nodes remain pending or running. pool.close() is deferred here so that
// the channel is closed after the last result is delivered.
func (sr *schedRunner) completionLoop(ctx context.Context) {
	defer sr.pool.close()

	sr.dispatchBatch(ctx)

	for sr.schedState.pendingCount() > noPendingCount {
		if done := sr.drainOneCompletion(ctx); done {
			return
		}
	}
}

// drainOneCompletion blocks until one completion event arrives or the context
// is cancelled. It returns true when the loop should terminate.
func (sr *schedRunner) drainOneCompletion(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		cancelPendingNodes(ctx, sr.schedState)

		return true

	case done, ok := <-sr.pool.completionCh:
		if !ok {
			return true
		}

		sr.schedState.status[done.nodeID] = nodeDone
		sr.schedState.result[done.nodeID] = done.result
		sr.schedState.running--

		if done.result.Status == StatusFailed {
			sr.schedState.propagateFailures(done.nodeID, done.result)
		}

		sr.dispatchBatch(ctx)
	}

	return false
}

// execParams bundles the parameters needed to execute a single story node.
// The struct is used to avoid exceeding the 8-argument function limit.
type execParams struct {
	nodeID      string
	storyNode   storyNode
	cfg         Config
	storyCache  cache.Cache
	lockDir     string
	lockTimeout time.Duration
	clk         clock.Clock
	runID       string
	inProcess   *sync.Map
	// storyIdx maps story ID → *ast.Story for inline prerequisite lookup.
	storyIdx storyIndex
}

// makeExecFunc constructs the execution function for a single story
// node. The returned function follows the double-checked acquire
// pattern from ADR 0016 §"Acquire pattern".
func makeExecFunc(params execParams) func(ctx context.Context) StoryResult {
	return func(ctx context.Context) StoryResult {
		return execStoryNode(ctx, params)
	}
}

// execStoryNode implements the execution logic for one story node following
// the double-checked acquire pattern from ADR 0016.
func execStoryNode(ctx context.Context, params execParams) StoryResult {
	startedAt := params.clk.Now()

	cacheKey := buildCacheKey(params.storyNode, params.cfg, params.runID)

	if params.cfg.NoCache {
		result := executeStory(ctx, params.storyNode, params.cfg, params.clk, params.storyIdx)
		result.StartedAt = startedAt
		result.FinishedAt = params.clk.Now()
		result.CacheStatus = CacheBypassed

		return result
	}

	// Step 1: fast path — check cache without lock.
	entry, entryErr := params.storyCache.Get(ctx, cacheKey)
	if entryErr == nil {
		return cacheHitResult(params.storyNode, entry, startedAt, params.clk)
	}

	return execWithDedup(ctx, params, cacheKey, startedAt)
}

// execWithDedup performs in-process deduplication via sync.Once and then
// the double-checked flock acquire.
func execWithDedup(
	ctx context.Context,
	params execParams,
	cacheKey cache.Key,
	startedAt time.Time,
) StoryResult {
	rawOnce, _ := params.inProcess.LoadOrStore(params.nodeID, &sync.Once{})
	once, ok := rawOnce.(*sync.Once)

	if !ok {
		panic("internal: inProcess value is not *sync.Once")
	}

	var lockedResult StoryResult

	once.Do(func() {
		lockedResult = executeWithLock(ctx, params, cacheKey, startedAt)
	})

	if lockedResult.TestID != emptyTestID {
		return lockedResult
	}

	// Another goroutine ran the Once — re-read from cache.
	entry, entryErr := params.storyCache.Get(ctx, cacheKey)
	if entryErr == nil {
		return cacheHitResult(params.storyNode, entry, startedAt, params.clk)
	}

	// Cache miss even after the Once: execute fresh.
	result := executeStory(ctx, params.storyNode, params.cfg, params.clk, params.storyIdx)
	result.StartedAt = startedAt
	result.FinishedAt = params.clk.Now()

	return result
}

// openLock acquires the cache lock for the given lockPath, choosing
// between blocking ([cache.AcquireLock]) and non-blocking
// ([cache.TryLock]) based on cfg.NoWait.
func openLock(
	ctx context.Context,
	lockPath string,
	lockTimeout time.Duration,
	cfg Config,
) (io.Closer, error) {
	if cfg.NoWait {
		closer, err := cache.TryLock(ctx, lockPath)
		if err != nil {
			return nil, fmt.Errorf("runner: try lock: %w", err)
		}

		return closer, nil
	}

	closer, err := cache.AcquireLock(ctx, lockPath, lockTimeout)
	if err != nil {
		return nil, fmt.Errorf("runner: acquire lock: %w", err)
	}

	return closer, nil
}

// lockFailureResult builds the StoryResult returned when lock
// acquisition fails. It is extracted to keep executeWithLock under
// the funlen limit.
func lockFailureResult(
	params execParams,
	startedAt time.Time,
	lockErr error,
) StoryResult {
	return StoryResult{
		Order:       emptyLen,
		TestID:      params.storyNode.story.Meta.ID,
		ScopeKey:    params.storyNode.scopeKey,
		OCPPVersion: ocppVersionEmpty,
		Status:      StatusFailed,
		CacheStatus: CacheMiss,
		StartedAt:   startedAt,
		FinishedAt:  params.clk.Now(),
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

// executeWithLock implements steps 2–7 of the ADR 0016 acquire
// pattern: acquire flock, re-read cache, execute, write cache,
// release flock.
func executeWithLock(
	ctx context.Context,
	params execParams,
	cacheKey cache.Key,
	startedAt time.Time,
) StoryResult {
	lockPath := filepath.Join(
		params.lockDir, cacheKey.Hash()+".lock",
	)

	lockCloser, lockErr := openLock(
		ctx, lockPath, params.lockTimeout, params.cfg,
	)
	if lockErr != nil {
		return lockFailureResult(params, startedAt, lockErr)
	}

	defer func() {
		_ = lockCloser.Close()
	}()

	// Step 3: re-read after acquiring lock (double-checked locking).
	entry, entryErr := params.storyCache.Get(ctx, cacheKey)
	if entryErr == nil {
		return cacheHitResult(
			params.storyNode, entry, startedAt, params.clk,
		)
	}

	// Step 4: execute the story.
	result := executeStory(
		ctx, params.storyNode, params.cfg, params.clk, params.storyIdx,
	)
	result.StartedAt = startedAt
	result.FinishedAt = params.clk.Now()

	// Steps 5–6: write result and trace to the cache.
	if result.Status == StatusPassed || result.Status == StatusFailed {
		writeToCache(
			ctx,
			params.storyCache,
			cacheKey,
			result,
			params.storyNode.story.Meta.CacheTTL,
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
		Order:       emptyLen,
		TestID:      storyNodeVal.story.Meta.ID,
		ScopeKey:    storyNodeVal.scopeKey,
		OCPPVersion: ocppVersionEmpty,
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
	storyIdx storyIndex,
) StoryResult {
	state := newRunnerState(clk, cfg.CSMSEndpoint)

	ocppVer := resolveOCPPVersion(storyNodeVal.story.Meta.Tags, cfg.OCPPVersion)

	findings := make([]Finding, emptyLen, len(state.logLines))

	result := executeAllSections(
		ctx,
		storyNodeVal.story,
		storyIdx,
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
	storyIdx storyIndex,
	state api.State,
	ocppVer string,
	findings *[]Finding,
) StoryResult {
	// Run prerequisite stories' non-teardown steps inline so that dependent
	// stories inherit runtime state (e.g., open WebSocket connections).
	if prereqFailed := runPrereqSections(ctx, storyAST, storyIdx, state, ocppVer, findings); prereqFailed {
		// Teardown always runs.
		_ = runSteps(ctx, storyAST.Teardown, state, ocppVer, findings)

		return StoryResult{
			Order:       emptyLen,
			TestID:      emptyTestID,
			ScopeKey:    emptyString,
			OCPPVersion: ocppVersionEmpty,
			Status:      StatusFailed,
			CacheStatus: CacheMiss,
			StartedAt:   time.Time{},
			FinishedAt:  time.Time{},
			Findings:    nil,
			Trace:       nil,
			Cause:       emptyCause,
			CauseChain:  nil,
		}
	}

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
		Order:       emptyLen,
		TestID:      emptyTestID,
		ScopeKey:    emptyString,
		OCPPVersion: ocppVersionEmpty,
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

// runPrereqSections runs the non-teardown steps of each prerequisite story
// inline within the current story's execution context. This allows dependent
// stories to inherit runtime state established by their prerequisites (e.g.,
// open WebSocket connections from a "connect to CSMS" step).
//
// Prerequisites are processed depth-first: if A depends on B which depends on
// C, running A inlines C then B then A (post-order). Teardown sections are NOT
// run for prerequisites — only the outermost story's Teardown runs.
//
// Returns true if any prerequisite step failed, in which case the caller
// should not execute the main story's own steps.
func runPrereqSections(
	ctx context.Context,
	storyAST *ast.Story,
	storyIdx storyIndex,
	state api.State,
	ocppVer string,
	findings *[]Finding,
) bool {
	for _, dep := range storyAST.Meta.Depends {
		prereqStory, ok := storyIdx[dep.ID]
		if !ok {
			continue
		}

		// Recurse into this prerequisite's own prerequisites first.
		if failed := runPrereqSections(ctx, prereqStory, storyIdx, state, ocppVer, findings); failed {
			return true
		}

		// Run the prerequisite's Background, Setup, and Scenario steps.
		if failed := runSteps(ctx, prereqStory.Background, state, ocppVer, findings); failed {
			return true
		}

		if failed := runSteps(ctx, prereqStory.Setup, state, ocppVer, findings); failed {
			return true
		}

		for _, scenario := range prereqStory.Scenarios {
			if failed := runSteps(ctx, scenario.Steps, state, ocppVer, findings); failed {
				return true
			}
		}
	}

	return false
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
			scopeKey = emptyCause

			break
		}
	}

	ocppVer := cfg.OCPPVersion
	if ocppVer == ocppVersionEmpty {
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
	stories := make([]StoryResult, emptyLen, len(topoOrder))

	summary := Summary{
		Total:     len(topoOrder),
		Passed:    emptyLen,
		Failed:    emptyLen,
		Skipped:   emptyLen,
		CacheHits: emptyLen,
	}

	for orderIdx, nodeID := range topoOrder {
		result := resolveNodeResult(nodeID, orderIdx, results)
		stories = append(stories, result)
		accumulateSummary(&summary, result)
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

// resolveNodeResult returns the recorded result for nodeID, or a synthetic
// skipped result when the node never executed.
func resolveNodeResult(
	nodeID string,
	orderIdx int,
	results map[string]StoryResult,
) StoryResult {
	result, ok := results[nodeID]
	if !ok {
		parts := splitNodeID(nodeID)
		result = StoryResult{
			Order:       orderIdx,
			TestID:      parts.storyID,
			ScopeKey:    parts.scopeKey,
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

	return result
}

// accumulateSummary increments the appropriate counter in summary based on
// the result's status and cache status.
func accumulateSummary(summary *Summary, result StoryResult) {
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
	if cfg.OCPPVersion != ocppVersionEmpty {
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

	err := filepath.WalkDir(root, func(
		path string,
		dirEntry fs.DirEntry,
		walkErr error,
	) error {
		return visitStoryFile(path, dirEntry, walkErr, &stories)
	})
	if err != nil {
		return nil, fmt.Errorf("runner: walk stories: %w", err)
	}

	return stories, nil
}

// visitStoryFile is the WalkDir callback for walkStoryFiles. It parses
// .story files and appends them to stories; directories and non-.story
// files are skipped.
func visitStoryFile(
	path string,
	dirEntry fs.DirEntry,
	walkErr error,
	stories *[]*ast.Story,
) error {
	if walkErr != nil {
		return walkErr
	}

	if dirEntry.IsDir() || filepath.Ext(path) != ".story" {
		return nil
	}

	cleanPath := filepath.Clean(path)

	data, readErr := os.ReadFile(cleanPath)
	if readErr != nil {
		return fmt.Errorf("read %q: %w", cleanPath, readErr)
	}

	storyAST, parseErr := story.Parse(path, data)
	if parseErr != nil {
		return fmt.Errorf("parse %q: %w", path, parseErr)
	}

	*stories = append(*stories, storyAST)

	return nil
}

// filterByOCPPVersion removes stories that do not declare the
// requested OCPP version via their tags.
func filterByOCPPVersion(
	stories []*ast.Story,
	version string,
) []*ast.Story {
	out := make([]*ast.Story, emptyLen, len(stories))

	for _, storyAST := range stories {
		if storyMatchesVersion(storyAST.Meta.Tags, version) {
			out = append(out, storyAST)
		}
	}

	if len(out) == emptyLen {
		// No story declared the version via tags; return all and
		// let keyword resolution handle version scoping.
		return stories
	}

	return out
}

// storyMatchesVersion reports whether any tag in tags matches the requested
// OCPP version string.
func storyMatchesVersion(tags []string, version string) bool {
	for _, tag := range tags {
		if tag == "ocpp"+version || tag == version {
			return true
		}
	}

	return false
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

	if cfgVersion != ocppVersionEmpty {
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
	if dir != emptyString {
		return dir, nil
	}

	if envDir := os.Getenv("OCTANE_CACHE_DIR"); envDir != emptyString {
		return envDir, nil
	}

	if xdgHome := os.Getenv("XDG_CACHE_HOME"); xdgHome != emptyString {
		return filepath.Join(xdgHome, "octane", "cache"), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return emptyString, fmt.Errorf("resolve home dir: %w", err)
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
func splitNodeID(nodeID string) nodeIDParts {
	if before, after, ok := strings.Cut(nodeID, "/"); ok {
		return nodeIDParts{storyID: before, scopeKey: after}
	}

	return nodeIDParts{storyID: nodeID, scopeKey: emptyCause}
}
