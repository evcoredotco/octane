# ADR 0011: Manual Pages — Cobra for Section 1, scdoc for Sections 5 and 7

- **Status:** Accepted
- **Date:** 2026-04-26
- **Deciders:** Project maintainer, DevOps, Docs
- **Constitution principles touched:** II (Two Distribution Surfaces),
  V (Stdlib-Heavy)

## Context

OCTANE ships as a system binary distributed via `.deb`, `.rpm`,
Homebrew, Scoop, and Docker (per ADR 0020). Users reasonably expect
`man octane` to work after installation. Beyond the operator command
reference, OCTANE has two stable artifacts that deserve dedicated
file-format references — the `octane.yml` config (ADR 0019) and the
`.story` DSL (ADR 0006) — and one concepts overview useful as a
landing page.

Three categories of man pages are required, mapping to the standard
Linux sections:

- **Section 1** — one page per subcommand (`octane.1`,
  `octane-run.1`, `octane-validate.1`, etc.). Operator-facing,
  changes with every CLI revision.
- **Section 5** — file format references (`octane.yml.5`,
  `octane.story.5`). Stable, citable across binary versions.
- **Section 7** — concepts overview (`octane.7`). Explains the
  wire-conformance model and the OCPP specification scope.

The toolchain choice is consequential: Section 1 churns with the CLI
and benefits from generation; Sections 5 and 7 are hand-maintained
prose where generation buys nothing.

## Decision

Adopt a **hybrid toolchain**:

| Section | Tool                           | Rationale                                                                          |
|---------|--------------------------------|------------------------------------------------------------------------------------|
| -1      | `spf13/cobra` `doc.GenManTree` | Auto-generated from CLI definitions; zero drift between `--help` and the man page. |
| -5      | `scdoc`                        | Hand-written; minimal mdoc-flavored input; no runtime dependency.                  |
| -7      | `scdoc`                        | Same.                                                                              |

### Section 1 — generated

- A hidden subcommand `octane gen-manpages --section 1 --out <dir>`
  emits one roff file per cobra command into the target directory.
- Build step: `make man` invokes this against `./bin/octane` after
  `make build`.
- Output names follow Linux convention: `octane.1`, `octane-run.1`,
  `octane-validate.1`, `octane-completion.1`, etc.
- Headers carry `OCTANE_VERSION` and the build date so distros'
  packaging tools can cite version provenance.
- Generated man pages are **not** committed to the repository. They
  are produced at packaging time. This keeps PR diffs free of
  generated noise.

### Sections 5 and 7 — hand-written

- Sources live at `docs/man/<name>.<section>.scd`.
- Build step: `make man` runs `scdoc < src > out` for each source.
- Required pages in v1:
  - `docs/man/octane.yml.5.scd` — config file reference
  - `docs/man/octane.story.5.scd` — story DSL reference
  - `docs/man/octane.7.scd` — concepts overview
- Sources are committed; generated `.5` and `.7` files are not.
- Sources are markdown-flavored enough that the same file is
  copy-fitted into the website's "Reference" section by an HTML
  build step (ADR 0013 covers the website; the man source remains
  the single source of truth for these specific pages).

### Cross-references and SEE ALSO

Each man page ends with a `SEE ALSO` section that lists related
pages. Cobra populates this for -1 from the cobra command tree;
hand-written pages cite back to -1 commands and forward to -7
concepts.

### Packaging

- `.deb` and `.rpm` install:
  - `/usr/share/man/man1/octane.1.gz`
  - `/usr/share/man/man1/octane-<sub>.1.gz`
  - `/usr/share/man/man5/octane.yml.5.gz`
  - `/usr/share/man/man5/octane.story.5.gz`
  - `/usr/share/man/man7/octane.7.gz`
- Homebrew formula installs into `man1.install`, `man5.install`,
  `man7.install`.
- The Docker image carries the `.gz`-compressed man pages at
  `/usr/local/share/man/...` so `docker run --rm octane man octane`
  works (with `man-db` installed in the image).

### Build dependencies

- `scdoc` is a small C program (~1000 LoC) widely packaged in distros
  (`apt install scdoc`, `brew install scdoc`). It is a build-time
  dependency only; the OCTANE binary does not depend on it at runtime.
- `gzip` is universally available; man pages are compressed at
  packaging time.

### Versioning policy

- Man page format is stable: changes to -1 mirror CLI changes (semver
  via the binary). Changes to -5 mirror file format changes (config
  schema and story grammar). Changes to -7 are editorial and do not
  affect contracts.
- A breaking change to either file format requires a `.5` page
  amendment and a `CHANGELOG.md` entry under the version that
  introduces it.

## Consequences

### Positive

- Section 1 cannot drift from `--help`, eliminating a class of
  documentation bugs common in CLI tools.
- Sections 5 and 7 are diffable, reviewable prose; no XML, no roff
  authored by hand.
- `scdoc` is an unobtrusive build-time dependency, not a runtime one.
- Generated man pages do not pollute git history.

### Negative

- Two toolchains to maintain (cobra-doc + scdoc). Mitigated by the
  fact that both are stable and OCTANE owns their integration.
- Distros that lack `scdoc` (rare) need to install it before building
  packages locally. Documented in `docs/installation.md` build prereqs.

### Neutral

- The hidden `gen-manpages` subcommand has no value to end users; it
  is documented in `docs/internals/`, not in the user-facing man pages.

## Alternatives considered

- **Pandoc for everything.** Heavier dependency, more flexible than
  scdoc but indistinguishable in output for this use case. Rejected
  on dependency-weight grounds.
- **All sections hand-written.** Rejected: -1 drift is the dominant
  failure mode in long-lived CLIs.
- **All sections generated.** Cobra cannot produce -5 or -7 without
  significant scaffolding; the result would be uglier than scdoc and
  no easier to maintain.
- **AsciiDoc (git's choice).** Considered. Rejected because scdoc's
  output is comparable for our needs and AsciiDoc's tooling
  (`asciidoctor`) is a heavier dependency.

## References

- Constitution: principles II, V
- ADR 0006 (story DSL grammar — referenced by `octane.story.5.scd`)
- ADR 0019 (config schema — referenced by `octane.yml.5.scd`)
- ADR 0020 (distribution channels)
- ADR 0012 (shell completion)
- ADR 0013 (website)
- scdoc: <https://git.sr.ht/~sircmpwn/scdoc>
- cobra/doc.GenManTree:
  <https://pkg.go.dev/github.com/spf13/cobra/doc#GenManTree>
