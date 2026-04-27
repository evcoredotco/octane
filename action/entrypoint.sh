#!/usr/bin/env bash
# entrypoint.sh — Docker ENTRYPOINT for the octane-action.
#
# Translates GitHub Actions INPUT_* environment variables into
# octane CLI flags and runs the conformance suite.
#
# CLI shape:
#   octane [global-flags] run [run-flags] <story-paths...>
#
# Outputs written to $GITHUB_OUTPUT:
#   exit-code   — numeric exit code from octane run
#   report-path — the report directory (INPUT_REPORT_DIR or "reports/")
#
# Task: T-006-31

set -euo pipefail

# ---------------------------------------------------------------------------
# Build global (persistent) flags
# ---------------------------------------------------------------------------

GLOBAL_ARGS=()

if [[ -n "${INPUT_CONFIG:-}" ]]; then
    GLOBAL_ARGS+=(--config "${INPUT_CONFIG}")
fi

if [[ "${INPUT_NO_CACHE:-false}" == "true" ]]; then
    GLOBAL_ARGS+=(--no-cache)
fi

if [[ -n "${INPUT_CACHE_DIR:-}" ]]; then
    GLOBAL_ARGS+=(--cache-dir "${INPUT_CACHE_DIR}")
fi

# ---------------------------------------------------------------------------
# Build run-subcommand flags
# ---------------------------------------------------------------------------

RUN_ARGS=()

if [[ -n "${INPUT_FAIL_ON:-}" ]]; then
    RUN_ARGS+=(--fail-on "${INPUT_FAIL_ON}")
fi

if [[ -n "${INPUT_OCPP_VERSION:-}" ]]; then
    RUN_ARGS+=(--ocpp-version "${INPUT_OCPP_VERSION}")
fi

if [[ -n "${INPUT_SHARD:-}" ]]; then
    RUN_ARGS+=(--shard "${INPUT_SHARD}")
fi

if [[ -n "${INPUT_MAX_PARALLEL:-}" && "${INPUT_MAX_PARALLEL}" != "1" ]]; then
    RUN_ARGS+=(--max-parallel "${INPUT_MAX_PARALLEL}")
fi

if [[ "${INPUT_INSECURE_SKIP_VERIFY:-false}" == "true" ]]; then
    RUN_ARGS+=(--insecure-skip-verify)
fi

# ---------------------------------------------------------------------------
# Resolve story paths — split on whitespace/newlines for glob expansion
# ---------------------------------------------------------------------------

read -ra STORY_PATHS <<< "${INPUT_STORIES:-scenarios/}"

# ---------------------------------------------------------------------------
# Execute octane — temporarily disable errexit so we can capture exit code
# ---------------------------------------------------------------------------

set +e
/usr/local/bin/octane "${GLOBAL_ARGS[@]}" run "${RUN_ARGS[@]}" "${STORY_PATHS[@]}"
OCTANE_EXIT=$?
set -e

# ---------------------------------------------------------------------------
# Write outputs to GITHUB_OUTPUT
# ---------------------------------------------------------------------------

REPORT_DIR="${INPUT_REPORT_DIR:-reports/}"

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
    printf 'exit-code=%d\n' "${OCTANE_EXIT}" >> "${GITHUB_OUTPUT}"
    printf 'report-path=%s\n' "${REPORT_DIR}"   >> "${GITHUB_OUTPUT}"
fi

exit "${OCTANE_EXIT}"
