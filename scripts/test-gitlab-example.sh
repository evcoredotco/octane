#!/usr/bin/env bash
#
# test-gitlab-example.sh — Local smoke test for the GitLab CI example.
# T-006-61
#
# This script is a DOCUMENTATION AID, not a CI gate. Its purpose is to
# give contributors a fast local sanity check without needing a live CSMS
# or a running GitLab instance.
#
# Two modes of operation:
#
#   Fallback (no gitlab-runner on PATH):
#     Performs structural validation of the YAML file: checks that required
#     keys and values are present. Exits 0 on success.
#
#   Full (gitlab-runner available):
#     Prints a notice that live execution requires a CSMS, then exits 0.
#     Actual end-to-end testing happens in the reference.yml CI workflow.
#
# Usage:
#   bash scripts/test-gitlab-example.sh

set -euo pipefail

GITLAB_CI_EXAMPLE="examples/ci/gitlab-ci/.gitlab-ci.yml"

# Resolve paths relative to the repository root regardless of where the
# script is invoked from.
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXAMPLE_FILE="${REPO_ROOT}/${GITLAB_CI_EXAMPLE}"

echo "--- GitLab CI example smoke test (T-006-61) ---"
echo "File: ${EXAMPLE_FILE}"

if [[ ! -f "${EXAMPLE_FILE}" ]]; then
  echo "ERROR: example file not found: ${EXAMPLE_FILE}" >&2
  exit 1
fi

# -----------------------------------------------------------------------
# Structural validation (always runs)
# -----------------------------------------------------------------------
echo ""
echo "Running structural validation..."

CONTENT="$(cat "${EXAMPLE_FILE}")"

check_key() {
  local key="$1"
  if ! echo "${CONTENT}" | grep -q "${key}"; then
    echo "ERROR: required key/value not found: ${key}" >&2
    exit 1
  fi
  echo "  OK: found '${key}'"
}

check_key "stages:"
check_key "variables:"
check_key "OCTANE_IMAGE"
check_key "conformance:1.6"
check_key "conformance:2.0.1"

echo "Structural validation passed."

# -----------------------------------------------------------------------
# Full mode: gitlab-runner exec docker
# -----------------------------------------------------------------------
if command -v gitlab-runner &>/dev/null; then
  echo ""
  echo "NOTE: gitlab-runner is available, but live job execution requires"
  echo "a running CSMS endpoint. Full testing happens in the reference.yml"
  echo "CI workflow against the pinned CitrineOS instance."
  echo "Skipping live execution."
else
  echo ""
  echo "NOTE: gitlab-runner not found on PATH. Structural validation only."
  echo "Install gitlab-runner and a CSMS to run end-to-end locally."
fi

echo ""
echo "Done. All checks passed."
