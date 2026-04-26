#!/usr/bin/env bash
#
# gen-completions.sh — produce shell completion scripts.
#
# v1 ships bash and zsh per ADR 0022. fish and PowerShell remain
# available via `octane completion <shell>` but are not packaged.
#
# Output: build/completions/{octane.bash, _octane}
#
# Used by: make completions, packaging pipeline.

set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
out_dir="${repo_root}/build/completions"
bin="${repo_root}/bin/octane"

if [[ ! -x "${bin}" ]]; then
  echo "error: ${bin} not found; run 'make build' first" >&2
  exit 1
fi

mkdir -p "${out_dir}"

"${bin}" completion bash > "${out_dir}/octane.bash"
"${bin}" completion zsh  > "${out_dir}/_octane"

echo "Generated completions under ${out_dir}:"
ls -l "${out_dir}"
