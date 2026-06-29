package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/evcoreco/octane/cmd/octane/internal/config"
	"github.com/evcoreco/octane/cmd/octane/internal/exitcode"
	reportpkg "github.com/evcoreco/octane/pkg/report"
	reportjson "github.com/evcoreco/octane/pkg/report/json"
	"github.com/evcoreco/octane/pkg/report/robotxml"
	"github.com/evcoreco/octane/pkg/runner"
)

// Sentinel errors for parseShard. These are package-level so callers can
// use errors.Is to distinguish error kinds.
var (
	errShardBadFormat  = errors.New("expected format N/M (e.g. \"1/4\")")
	errShardTotalRange = errors.New("shard total must be >= 1")
	errShardIndexRange = errors.New("shard index out of range")
)

// runFlagsT holds the parsed values of the flags specific to the
// "octane run" subcommand.
type runFlagsT struct {
	maxParallel        int
	shard              string
	ocppVersion        string
	lockTimeout        time.Duration
	noWait             bool
	insecureSkipVerify bool
	failOn             string
	reportDir          string
	noTraceOnPass      bool
	csmsEndpoint       string
	params             []string
}

const (
	// shardParts is the expected number of parts in a "N/M" shard value.
	shardParts = 2

	// shardMinValue is the minimum valid value for shard index or total.
	shardMinValue = 1

	// defaultReportDir is the default value for the --report-dir flag.
	defaultReportDir = "reports/"

	// emptyFlagValue is the zero string used to detect unset string flags.
	emptyFlagValue = ""

	// zeroIntDefault is the zero default for integer flags (meaning
	// "inherit from config" for max-parallel and lock-timeout).
	zeroIntDefault = 0

	// exactlyOneArg is the expected positional argument count for
	// subcommands that require exactly one argument.
	exactlyOneArg = 1

	// firstArgIndex is the index of the first positional argument in
	// the args slice passed to RunE.
	firstArgIndex = 0
)

// newRunCmd constructs and returns the "octane run" subcommand.
// globalFlags is the parent global-flags struct populated by cobra before
// RunE executes.
//
//nolint:exhaustruct // cobra.Command has many optional fields
func newRunCmd(globalFlags *globalFlagsT) *cobra.Command {
	flags := &runFlagsT{}

	cmd := &cobra.Command{
		Use:   "run [story-paths...]",
		Short: "Run .story conformance test suites",
		Long: `Run discovers and executes .story files against a CSMS endpoint.

Story paths may be files or directories. When no paths are given,
octane searches the stories_dir configured in octane.yml (default:
"scenarios").

Sharding splits the story set across CI workers:
  --shard 1/4   run the first quarter of stories
  --shard 2/4   run the second quarter, etc.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStories(cmd, args, globalFlags, flags)
		},
	}

	registerRunFlags(cmd, flags)

	return cmd
}

// registerRunFlags adds all "octane run" subcommand flags to cmd and
// binds their values into flags.
func registerRunFlags(cmd *cobra.Command, flags *runFlagsT) {
	registerRunExecutionFlags(cmd, flags)
	registerRunOutputFlags(cmd, flags)
}

// registerRunExecutionFlags registers the execution-control flags for
// "octane run": parallelism, sharding, OCPP version, lock timeout, and
// TLS settings.
func registerRunExecutionFlags(cmd *cobra.Command, flags *runFlagsT) {
	cmdFlags := cmd.Flags()

	cmdFlags.IntVar(
		&flags.maxParallel,
		"max-parallel",
		zeroIntDefault,
		"maximum number of stories to run concurrently",
	)

	cmdFlags.StringVar(
		&flags.shard,
		"shard",
		emptyFlagValue,
		`shard index in "N/M" format (e.g. "1/4" for the first of four shards)`,
	)

	cmdFlags.StringVar(
		&flags.ocppVersion,
		"ocpp-version",
		emptyFlagValue,
		`restrict run to stories declaring this OCPP version (e.g. "1.6")`,
	)

	cmdFlags.DurationVar(
		&flags.lockTimeout,
		"lock-timeout",
		zeroIntDefault,
		"maximum wait time to acquire a cache lock (default: 60s)",
	)

	cmdFlags.BoolVar(
		&flags.noWait,
		"no-wait",
		false,
		"fail immediately when a cache lock is busy instead of waiting",
	)

	cmdFlags.BoolVar(
		&flags.insecureSkipVerify,
		"insecure-skip-verify",
		false,
		"disable TLS certificate verification (WARNING: insecure)",
	)

	cmdFlags.StringVar(
		&flags.csmsEndpoint,
		"csms-endpoint",
		emptyFlagValue,
		`base WebSocket URL of the CSMS under test (e.g. "ws://localhost:9210")`,
	)

	cmdFlags.StringArrayVar(
		&flags.params,
		"param",
		nil,
		`story parameter override in name=value form; may be repeated`,
	)
}

// registerRunOutputFlags registers the output and reporting flags for
// "octane run": fail-on threshold, report directory, and trace settings.
func registerRunOutputFlags(cmd *cobra.Command, flags *runFlagsT) {
	cmdFlags := cmd.Flags()

	cmdFlags.StringVar(
		&flags.failOn,
		"fail-on",
		emptyFlagValue,
		`exit with failure when reached: "any" (default) or "major"`,
	)

	cmdFlags.StringVar(
		&flags.reportDir,
		"report-dir",
		defaultReportDir,
		"directory in which per-run report subdirectories are written",
	)

	cmdFlags.BoolVar(
		&flags.noTraceOnPass,
		"no-trace-on-pass",
		false,
		"omit wire-trace data from reports for stories that passed",
	)
}

// runStories is the RunE function for "octane run". It loads
// configuration, resolves flags over the config, builds a
// runner.Config, and delegates to runner.Run.
func runStories(
	_ *cobra.Command,
	storyPaths []string,
	globalFlags *globalFlagsT,
	flags *runFlagsT,
) error {
	cfg, shard := loadRunConfig(globalFlags, flags)

	if len(storyPaths) == zeroIntDefault {
		storyPaths = []string{cfg.StoriesDir}
	}

	runnerCfg := runner.Config{
		StoryPaths:         storyPaths,
		MaxParallel:        cfg.MaxParallel,
		LockTimeout:        cfg.LockTimeout,
		NoWait:             flags.noWait,
		ShardIndex:         shard.index,
		ShardTotal:         shard.total,
		CacheDir:           cfg.CacheDir,
		NoCache:            globalFlags.noCache,
		NoTraceOnPass:      flags.noTraceOnPass,
		OCPPVersion:        cfg.OCPPVersion,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		CSMSEndpoint:       flags.csmsEndpoint,
		Parameters:         cfg.Parameters,
	}

	result, runErr := runner.Run(context.Background(), runnerCfg)
	if runErr != nil {
		dieErrf(exitcode.ToolError, "octane: run error: %v\n", runErr)

		return nil
	}

	printRunSummary(result.Summary)
	writeReports(result, flags)

	if result.Summary.Failed > zeroIntDefault {
		dieErrf(exitcode.TestFailed, emptyFlagValue)
	}

	return nil
}

// loadRunConfig loads, applies env vars, and overrides the config, then
// parses the shard flag. On any error it calls dieErrf (which panics) so
// the caller never receives invalid values. It returns the effective config
// and a zero shardSpec when sharding is disabled.
func loadRunConfig(
	globalFlags *globalFlagsT,
	flags *runFlagsT,
) (config.Config, shardSpec) {
	cfg, err := config.Load(globalFlags.configPath)
	if err != nil {
		dieErrf(
			exitcode.ConfigError,
			"octane: %q: config error: %v\n",
			globalFlags.configPath,
			err,
		)
	}

	cfg = config.ApplyEnv(cfg)
	cfg = applyRunFlagOverrides(cfg, globalFlags, flags)

	if cfg.InsecureSkipVerify {
		_, _ = fmt.Fprintln(
			os.Stderr,
			"WARNING: --insecure-skip-verify is set;"+
				" TLS certificate verification is disabled",
		)
	}

	shard, shardErr := parseShard(flags.shard)
	if shardErr != nil {
		dieErrf(
			exitcode.ConfigError,
			"octane: invalid --shard value: %v\n",
			shardErr,
		)
	}

	return cfg, shard
}

// printRunSummary writes the one-line result summary to stdout.
func printRunSummary(summary runner.Summary) {
	_, _ = fmt.Fprintf(
		os.Stdout,
		"passed=%d failed=%d skipped=%d cache-hits=%d\n",
		summary.Passed,
		summary.Failed,
		summary.Skipped,
		summary.CacheHits,
	)
}

// writeReports writes JSON and Robot XML reports when a report dir is
// configured.
func writeReports(result *runner.RunResult, flags *runFlagsT) {
	if flags.reportDir == emptyFlagValue {
		return
	}

	reportPath := filepath.Join(flags.reportDir, result.RunID)

	jsonOpts := reportpkg.JSONOptions{
		NoTraceOnPass: flags.noTraceOnPass,
		OctaneVersion: version,
	}

	jsonErr := reportjson.WriteJSON(result, reportPath, jsonOpts)
	if jsonErr != nil {
		_, _ = fmt.Fprintf(
			os.Stderr,
			"octane: warning: JSON report write failed: %v\n",
			jsonErr,
		)
	}

	// RobotXMLOptions: SuiteName defaults to "OCTANE Conformance".
	xmlOpts := reportpkg.RobotXMLOptions{} //nolint:exhaustruct

	xmlErr := robotxml.WriteRobotXML(result, reportPath, xmlOpts)
	if xmlErr != nil {
		_, _ = fmt.Fprintf(
			os.Stderr,
			"octane: warning: Robot XML report write failed: %v\n",
			xmlErr,
		)
	}

	_, _ = fmt.Fprintf(os.Stdout, "report-dir=%s\n", reportPath)
}

// applyRunFlagOverrides builds a FlagOverrides from the run-specific
// flags and applies them to cfg. Only non-zero/non-empty flag values
// are treated as explicit overrides; the zero value of each flag type
// means the operator did not set that flag.
func applyRunFlagOverrides(
	cfg config.Config,
	globalFlags *globalFlagsT,
	flags *runFlagsT,
) config.Config {
	var overrides config.FlagOverrides

	if flags.maxParallel != zeroIntDefault {
		maxParallel := flags.maxParallel
		overrides.MaxParallel = &maxParallel
	}

	if flags.ocppVersion != emptyFlagValue {
		ocppVersion := flags.ocppVersion
		overrides.OCPPVersion = &ocppVersion
	}

	if flags.lockTimeout != zeroIntDefault {
		lockTimeout := flags.lockTimeout
		overrides.LockTimeout = &lockTimeout
	}

	if flags.failOn != emptyFlagValue {
		failOn := flags.failOn
		overrides.FailOn = &failOn
	}

	if flags.insecureSkipVerify {
		skip := true
		overrides.InsecureSkipVerify = &skip
	}

	if globalFlags.cacheDir != emptyFlagValue {
		cacheDir := globalFlags.cacheDir
		overrides.CacheDir = &cacheDir
	}

	params, err := parseParameterOverrides(flags.params)
	if err != nil {
		dieErrf(exitcode.ConfigError, "octane: invalid --param value: %v\n", err)
	}
	overrides.Parameters = params

	return config.Resolve(cfg, overrides)
}

func parseParameterOverrides(values []string) (map[string]string, error) {
	if len(values) == zeroIntDefault {
		return nil, nil
	}

	params := make(map[string]string, len(values))

	for _, raw := range values {
		name, value, ok := strings.Cut(raw, "=")
		if !ok || name == emptyFlagValue {
			return nil, fmt.Errorf("got %q, expected name=value", raw)
		}

		params[name] = value
	}

	return params, nil
}

// shardSpec holds the parsed shard index (zero-based) and total from
// a "--shard N/M" flag value.
type shardSpec struct {
	// index is the zero-based shard index for runner.Config.ShardIndex.
	index int
	// total is the shard count for runner.Config.ShardTotal.
	total int
}

// parseShard parses the "--shard N/M" flag value. An empty string
// is valid and returns a zero shardSpec (sharding disabled). Returns
// an error when the format is invalid or values are out of range
// (N < 1, N > M, M < 1).
func parseShard(value string) (shardSpec, error) {
	zeroShard := shardSpec{index: zeroIntDefault, total: zeroIntDefault}

	if value == emptyFlagValue {
		return zeroShard, nil
	}

	parts := strings.SplitN(value, "/", shardParts)
	if len(parts) != shardParts {
		return zeroShard, errShardFormat(value)
	}

	numerator, parseErr := strconv.Atoi(parts[0])
	if parseErr != nil {
		return zeroShard, fmt.Errorf("shard index %q: %w", parts[0], parseErr)
	}

	denominator, parseErr := strconv.Atoi(parts[1])
	if parseErr != nil {
		return zeroShard, fmt.Errorf("shard total %q: %w", parts[1], parseErr)
	}

	if denominator < shardMinValue {
		return zeroShard, errShardTotal(denominator)
	}

	if numerator < shardMinValue || numerator > denominator {
		return zeroShard, errShardIndex(numerator, denominator)
	}

	// runner.Config uses zero-based ShardIndex.
	return shardSpec{index: numerator - shardMinValue, total: denominator}, nil
}

// errShardFormat wraps errShardBadFormat with the offending value.
func errShardFormat(value string) error {
	return fmt.Errorf("got %q: %w", value, errShardBadFormat)
}

// errShardTotal wraps errShardTotalRange with the actual denominator.
func errShardTotal(denominator int) error {
	return fmt.Errorf("got %d: %w", denominator, errShardTotalRange)
}

// errShardIndex wraps errShardIndexRange with the actual numerator and range.
func errShardIndex(numerator, denominator int) error {
	return fmt.Errorf(
		"index %d not in [1, %d]: %w",
		numerator,
		denominator,
		errShardIndexRange,
	)
}
