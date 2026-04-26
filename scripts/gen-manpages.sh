#!/usr/bin/env bash
#
# gen-manpages.sh — produce roff man pages for OCTANE.
#
# Section 1 (subcommands) is generated from cobra by invoking the
# OCTANE binary's hidden gen-manpages subcommand.
# Sections 5 and 7 are generated from scdoc sources under docs/man/.
#
# Output: build/man/man{1,5,7}/*.{1,5,7}
#
# Prerequisites:
#   - ./bin/octane built (run `make build` first)
#   - scdoc on PATH (apt install scdoc / brew install scdoc)
#
# Used by: make man, packaging pipeline, CI docs check.

set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
out_root="${repo_root}/build/man"
bin="${repo_root}/bin/octane"

if [[ ! -x "${bin}" ]]; then
  echo "error: ${bin} not found; run 'make build' first" >&2
  exit 1
fi

if ! command -v scdoc >/dev/null 2>&1; then
  echo "error: scdoc not on PATH (apt install scdoc / brew install scdoc)" >&2
  exit 1
fi

rm -rf "${out_root}"
mkdir -p "${out_root}/man1" "${out_root}/man5" "${out_root}/man7"

# Section 1 — generated from cobra.
"${bin}" gen-manpages --section 1 --out "${out_root}/man1"

# Sections 5 and 7 — scdoc.
for src in "${repo_root}"/docs/man/*.5.scd; do
  base="$(basename "${src}" .scd)"
  scdoc < "${src}" > "${out_root}/man5/${base}"
done

for src in "${repo_root}"/docs/man/*.7.scd; do
  base="$(basename "${src}" .scd)"
  scdoc < "${src}" > "${out_root}/man7/${base}"
done

# Verify section 1 actually produced output.
shopt -s nullglob
section1_files=( "${out_root}/man1"/*.1 )
if [[ "${#section1_files[@]}" -eq 0 ]]; then
  echo "error: gen-manpages produced no Section 1 output" >&2
  exit 1
fi

echo "Generated man pages under ${out_root}:"
find "${out_root}" -type f | sort
