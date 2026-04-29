package main

import (
	"context"
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

// runFlags holds the parsed values of the flags specific to the
// "octane run" subcommand.
var runFlags struct {
	maxParallel        int
	shard              string
	ocppVersion        string
	lockTimeout        time.Duration
	noWait             bool
	insecureSkipVerify bool
	failOn             string
	reportDir          string
	noTraceOnPass      bool
}

//nolint:exhaustruct // cobra.Command has many optional fields
var runCmd = &cobra.Command{
	Use:   "run [story-paths...]",
	Short: "Run .story conformance test suites",
	Long: `Run discovers and executes .story files against a CSMS endpoint.

Story paths may be files or directories. When no paths are given,
octane searches the stories_dir configured in octane.yml (default:
"scenarios").

Sharding splits the story set across CI workers:
  --shard 1/4   run the first quarter of stories
  --shard 2/4   run the second quarter, etc.`,
	RunE: runStories,
}

func init() {
	flags := runCmd.Flags()

	flags.IntVar(
		&runFlags.maxParallel,
		"max-parallel",
		0,
		"maximum number of stories to run concurrently",
	)

	flags.StringVar(
		&runFlags.shard,
		"shard",
		"",
		`shard index in "N/M" format (e.g. "1/4" for the first of four shards)`,
	)

	flags.StringVar(
		&runFlags.ocppVersion,
		"ocpp-version",
		"",
		"restrict run to stories declaring this OCPP version (e.g. \"1.6\")",
	)

	flags.DurationVar(
		&runFlags.lockTimeout,
		"lock-timeout",
		0,
		"maximum wait time to acquire a cache lock (default: 60s)",
	)

	flags.BoolVar(
		&runFlags.noWait,
		"no-wait",
		false,
		"fail immediately when a cache lock is busy instead of waiting",
	)

	flags.BoolVar(
		&runFlags.insecureSkipVerify,
		"insecure-skip-verify",
		false,
		"disable TLS certificate verification (WARNING: insecure)",
	)

	flags.StringVar(
		&runFlags.failOn,
		"fail-on",
		"",
		`exit with failure when threshold is reached: "any" (default) or "major"`,
	)

	flags.StringVar(
		&runFlags.reportDir,
		"report-dir",
		"reports/",
		"directory in which per-run report subdirectories are written",
	)

	flags.BoolVar(
		&runFlags.noTraceOnPass,
		"no-trace-on-pass",
		false,
		"omit wire-trace data from reports for stories that passed",
	)

	rootCmd.AddCommand(runCmd)
}

// runStories is the RunE function for "octane run". It loads
// configuration, resolves flags over the config, builds a
// runner.Config, and delegates to runner.Run.
func runStories(_ *cobra.Command, storyPaths []string) error {
	cfg, err := config.Load(globalFlags.configPath)
	if err != nil {
		dieErrf(
			exitcode.ConfigError,
			"octane: %q: config error: %v\n",
			globalFlags.configPath,
			err,
		)

		return nil
	}

	cfg = config.ApplyEnv(cfg)
	cfg = applyRunFlagOverrides(cfg)

	if cfg.InsecureSkipVerify {
		_, _ = fmt.Fprintln(
			os.Stderr,
			"WARNING: --insecure-skip-verify is set;"+
				" TLS certificate verification is disabled",
		)
	}

	shardIndex, shardTotal, shardErr := parseShard(runFlags.shard)
	if shardErr != nil {
		dieErrf(
			exitcode.ConfigError,
			"octane: invalid --shard value: %v\n",
			shardErr,
		)

		return nil
	}

	if len(storyPaths) == 0 {
		storyPaths = []string{cfg.StoriesDir}
	}

	runnerCfg := runner.Config{
		StoryPaths:         storyPaths,
		MaxParallel:        cfg.MaxParallel,
		LockTimeout:        cfg.LockTimeout,
		NoWait:             runFlags.noWait,
		ShardIndex:         shardIndex,
		ShardTotal:         shardTotal,
		CacheDir:           cfg.CacheDir,
		NoCache:            globalFlags.noCache,
		NoTraceOnPass:      runFlags.noTraceOnPass,
		OCPPVersion:        cfg.OCPPVersion,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
	}

	result, runErr := runner.Run(context.Background(), runnerCfg)
	if runErr != nil {
		dieErrf(exitcode.ToolError, "octane: run error: %v\n", runErr)

		return nil
	}

	_, _ = fmt.Fprintf(
		os.Stdout,
		"passed=%d failed=%d skipped=%d cache-hits=%d\n",
		result.Summary.Passed,
		result.Summary.Failed,
		result.Summary.Skipped,
		result.Summary.CacheHits,
	)

	if runFlags.reportDir != "" {
		reportPath := filepath.Join(runFlags.reportDir, result.RunID)

		jsonOpts := reportpkg.JSONOptions{
			NoTraceOnPass: runFlags.noTraceOnPass,
			OctaneVersion: version,
		}

		writeErr := reportjson.WriteJSON(result, reportPath, jsonOpts)
		if writeErr != nil {
			_, _ = fmt.Fprintf(
				os.Stderr,
				"octane: warning: JSON report write failed: %v\n",
				writeErr,
			)
		}

		xmlOpts := reportpkg.RobotXMLOptions{ //nolint:exhaustruct // SuiteName defaults to "OCTANE Conformance"
		}

		writeErr = robotxml.WriteRobotXML(result, reportPath, xmlOpts)
		if writeErr != nil {
			_, _ = fmt.Fprintf(
				os.Stderr,
				"octane: warning: Robot XML report write failed: %v\n",
				writeErr,
			)
		}

		_, _ = fmt.Fprintf(os.Stdout, "report-dir=%s\n", reportPath)
	}

	if result.Summary.Failed > 0 {
		exitcode.Exec(exitcode.TestFailed)
	}

	return nil
}

// applyRunFlagOverrides builds a FlagOverrides from the run-specific
// flags and applies them to cfg. Only non-zero/non-empty flag values
// are treated as explicit overrides; the zero value of each flag type
// means the operator did not set that flag.
func applyRunFlagOverrides(cfg config.Config) config.Config {
	var overrides config.FlagOverrides

	if runFlags.maxParallel != 0 {
		maxParallel := runFlags.maxParallel
		overrides.MaxParallel = &maxParallel
	}

	if runFlags.ocppVersion != "" {
		ocppVersion := runFlags.ocppVersion
		overrides.OCPPVersion = &ocppVersion
	}

	if runFlags.lockTimeout != 0 {
		lockTimeout := runFlags.lockTimeout
		overrides.LockTimeout = &lockTimeout
	}

	if runFlags.failOn != "" {
		failOn := runFlags.failOn
		overrides.FailOn = &failOn
	}

	if runFlags.insecureSkipVerify {
		skip := true
		overrides.InsecureSkipVerify = &skip
	}

	if globalFlags.cacheDir != "" {
		cacheDir := globalFlags.cacheDir
		overrides.CacheDir = &cacheDir
	}

	return config.Resolve(cfg, overrides)
}

// parseShard parses the "--shard N/M" flag value. An empty string
// is valid and returns (0, 0, nil) meaning sharding is disabled.
// Returns an error when the format is invalid or the values are out
// of range (N < 1, N > M, M < 1).
func parseShard(value string) (int, int, error) {
	if value == "" {
		return 0, 0, nil
	}

	parts := strings.SplitN(value, "/", 2) //nolint:mnd // 2 parts: N and M
	if len(
		parts,
	) != 2 { //nolint:mnd // exactly 2 parts required
		return 0, 0, fmt.Errorf(
			"expected format N/M (e.g. \"1/4\"), got %q",
			value,
		)
	}

	numerator, parseErr := strconv.Atoi(parts[0])
	if parseErr != nil {
		return 0, 0, fmt.Errorf("shard index %q is not an integer", parts[0])
	}

	denominator, parseErr := strconv.Atoi(parts[1])
	if parseErr != nil {
		return 0, 0, fmt.Errorf("shard total %q is not an integer", parts[1])
	}

	if denominator < 1 {
		return 0, 0, fmt.Errorf(
			"shard total must be >= 1, got %d",
			denominator,
		)
	}

	if numerator < 1 || numerator > denominator {
		return 0, 0, fmt.Errorf(
			"shard index must be between 1 and %d, got %d",
			denominator,
			numerator,
		)
	}

	// runner.Config uses zero-based ShardIndex.
	return numerator - 1, denominator, nil
}
