# Changelog

All notable changes to OCTANE are documented here.
This project adheres to [Keep a Changelog 1.1.0][kac] and
[Semantic Versioning 2.0.0][semver].

[kac]: https://keepachangelog.com/en/1.1.0/
[semver]: https://semver.org/spec/v2.0.0.html

## [Unreleased]

### Added — Spec 005: Dependency Cache

- `pkg/runner/`: story runner with DAG traversal, worker pool, shard
  filtering, and cache integration. `runner.Run(ctx, Config) (*RunResult,
  error)` is the single public entry point consumed by the CLI (spec 006) and
  GitHub Action (spec 006) (spec 005 G1, constitution principle II).
- Dependency graph built from `Depends:` YAML blocks in `.story` Meta
  sections. Three scope types supported: `per-station` (one prerequisite
  instance per station handle), `per-run` (one instance per `octane run`
  invocation), and `global` (shared across runs via cache) (ADR 0015, spec
  005 §10).
- Cycle detection via topological sort; `runner.ErrCycle` is returned when
  a cycle is found. The error wraps `*dag.ErrCycle` which names the offending
  edges (spec 005 AC3).
- Failure propagation: a failed story marks all of its dependents
  `StatusSkipped`. The `Cause` and `CauseChain` fields on `StoryResult`
  identify the root failure (spec 005 AC4).
- Worker pool with configurable parallelism via `Config.MaxParallel`
  (`--max-parallel` flag in spec 006). Stories within a topological level
  execute concurrently up to the pool size; ordering within a level is
  lexicographic for determinism (ADR 0019, constitution principle IV).
- CI sharding via `Config.ShardIndex` / `Config.ShardTotal` (`--shard N/M`
  flag). Distributes stories by `sha256(test_id) % M`; prerequisites outside
  the shard are always included (spec 005 OQ1).
- `pkg/cache/`: content-addressed file cache with TTL, atomic writes, and
  flock-based in-machine locking (ADR 0016):
  - `cache.Key` — seven-field tuple whose SHA-256 becomes the filesystem path.
  - `cache.Entry` — result JSON + optional trace JSON + TTL metadata.
  - `cache.Cache` — minimal interface (`Get`, `Put`, `Prune`) consumed by the
    runner; in-memory test doubles satisfy it without touching the filesystem.
  - `cache.FileCache` — file tree implementation using two-character fanout
    directories matching Bazel / ccache / Go's build cache layout.
  - `cache.Open(dir)` — creates or verifies the cache directory structure and
    returns a `Cache` implementation.
  - Atomic write protocol: temp file → `fsync` → rename → directory `fsync`
    (spec 005 §10 step 4).
  - `cache.AcquireLock` — exclusive `flock` on `<hash>.lock`; implements the
    double-checked acquire pattern to prevent two concurrent `octane run`
    processes from executing the same story twice (ADR 0016 §"Acquire
    pattern", ADR 0019).
  - `Cache-TTL:` Meta key support: `Entry.TTL == 0` means never expires;
    helper stories default to no TTL (spec 005 AC10).
  - `cache.ErrCacheMiss` — typed sentinel returned by `Get` on miss or
    expiry; callers use `errors.Is` to detect and re-execute.
- `Config.NoCache` (`--no-cache`): bypasses all cache reads and writes;
  `CacheStatus` is `bypassed` for every story in the report (spec 005 G4).
- `docs/concepts/dependency-graph.md` — authoring guide for `Depends:`
  declarations, scope types, DAG construction, and cycle detection (T-005-61).
- `docs/concepts/cache.md` — operational guide covering cache location,
  entry lifetime, atomic write protocol, flock locking, `--no-cache`, and CI
  sharding (T-005-62).

### Added — Spec 004: Primitive Keywords

- `pkg/keywords/primitive/`: ten transport-level primitive keywords covering
  WebSocket open (with and without subprotocol negotiation), close, send raw
  frame, send raw bytes, expect any frame, expect frame of type, wait, assert
  connection open, and assert connection closed (spec 004 §10).
- Self-registration at `init()` time under `api.LayerPrimitive` with a zero
  `OCPPVersion`; importing the package is sufficient to activate all primitives
  (spec 004 G2).
- `primitive.ErrTimeout`: typed error returned by the expect keywords when no
  matching frame arrives within the configured deadline. Carries `Station`,
  `Timeout`, and `Deadline` fields; `Deadline` is derived from
  `state.Now().Add(timeout)` so the deterministic clock governs it
  (spec 004 AC4, constitution principle IV).
- `docs/keywords/primitives.md`: full primitive catalog with pattern syntax,
  argument types, error behavior, and example story steps (T-004-41).
- Unit tests in `pkg/keywords/primitive/` exercising every keyword via
  `mock.MockState` and `mock.MockStation` with no network dependency
  (spec 004 G3).

### Added — Spec 003: Keyword API and Registry

- `pkg/keywords/api`: public contract package for keyword authors — `Func`,
  `Args`, `State`, `Station`, `Keyword`, `Layer`, `OCPPVersion` types and
  interfaces. No implementation logic; stdlib-only.
- `pkg/keywords/api/mock`: in-memory test doubles `mock.State` and
  `mock.Station` (`NewMockState`, `NewMockStation`) enabling keyword unit
  tests with no dependency on `pkg/runner/`, `pkg/transport/`, or any network
  library (spec 003 AC8).
- `pkg/keywords/registry`: global keyword registry with `Register`, `All`,
  and `Resolve`. Self-registration at `init()` time per ADR 0007. Collision
  detection panics at startup naming both conflicting registration sites (AC2).
- `pkg/keywords/registry/internal/pattern`: `{name:type}` placeholder parser
  (`Parse`), pattern matcher (`Match`), and type coercer (`Coerce`) covering
  all seven placeholder types: `string`, `int`, `float`, `bool`, `duration`,
  `station`, `any`.
- `pkg/keywords/registry/internal/levenshtein`: edit-distance helpers for
  "did you mean?" suggestions on unmatched steps.
- Layered resolution per ADR 0007: domain keywords for the active OCPP version
  take precedence over primitive keywords; longer patterns win within a layer
  (AC3, AC4, AC5, AC6, AC7).
- `ErrNoMatch` (with optional `Closest` suggestion at Levenshtein ≤ 5) and
  `ErrTypeMismatch` typed resolver errors.
- `registry.All()` returns keywords in stable `(Layer, OCPPVersion, Pattern)`
  sort order, satisfying constitution principle IV (determinism) (AC1).

### Added — Spec 002: Wire Engine

- `pkg/transport`: WebSocket client with TLS, subprotocol negotiation, and configurable timeouts (`Station`, `Dial`, `DialOptions`)
- `pkg/wire`: OCPP-J frame parsing and serialization for CALL (type 2), CALLRESULT (type 3), CALLERROR (type 4)
- `pkg/engine/clock`: `Clock` interface with real-clock and deterministic test-double implementations
- `pkg/engine/rand`: `Rand` interface with crypto-seeded and fixed-seed (PCG) implementations
- ADR 0018: Clock/Rand injection model for reproducible test execution

### Added
- **Story DSL parser** (`pkg/story`): recursive-descent parser for `.story`
  files implementing the grammar from ADR 0006. Produces a typed AST
  (`pkg/story/ast`), typed diagnostic errors (`pkg/story/diag`), and
  validates Spec-Ref/helper constraints, parameter bindings, and Depends
  block integrity. Covers spec 001 tasks T-001-00 through T-001-52.

### Changed
- **Spec set decomposed from 2 specs to 7 specs.** The previous
  `001-bootstrap-engine` and `002-story-framework` specs were
  rewritten in place into a 7-spec set that maps cleanly to Go
  package boundaries:
  `001-story-parser` (`pkg/story/`),
  `002-wire-engine` (`pkg/transport/`, `pkg/wire/`,
  `pkg/engine/clock`, `pkg/engine/rand`),
  `003-keyword-api` (`pkg/keywords/api/`,
  `pkg/keywords/registry/`),
  `004-primitive-keywords` (`pkg/keywords/primitive/`),
  `005-dependency-cache` (`pkg/runner/`, `pkg/cache/`),
  `006-cli-action` (`cmd/octane/`, GitHub Action, GitLab),
  `007-reports` (`pkg/report/`).
  Each new spec includes full `spec.md`, `plan.md`, and
  `tasks.md`. Specs 006 and 007 are marked Approved Provisional
  pending spec 005 implementation. The directories
  `specs/001-bootstrap-engine/` and `specs/002-story-framework/`
  were renamed and their content rewritten; implementation notes
  from the old specs were lifted into the appropriate new specs
  (JSON decoding quirk → spec 002; per-station scratch space →
  spec 003 §10; primitive keyword catalog → spec 004 §10).
- **Cache backend reversed from SQLite to a content-addressed file
  tree.** A previous draft of ADR 0016 specified SQLite as the
  cache backend; this was reconsidered after re-examining OCTANE's
  CI deployment scenarios. SQLite's strengths (in-database
  concurrency, ACID transactions on a single file) work *against*
  the CI cache model: parallel jobs writing to the same cache key
  race and overwrite, the SQLite file compresses poorly compared
  to JSON text inflating CI bandwidth, and partial cache
  restoration corrupts the database. A content-addressed file tree
  (the pattern used by Bazel, ccache, Go's build cache, and Cargo)
  fits CI cache restoration natively. ADR 0017 (the SQLite
  dependency justification) is removed entirely; OCTANE returns to
  a single non-stdlib runtime dependency (`nhooyr.io/websocket`),
  honoring constitution principle V more cleanly.

### Removed
- **All Go code.** The exploratory `pkg/keywords/`, `pkg/wire/`,
  `go.mod`, and associated tests have been removed. The project is
  now in a deliberate **design-complete, code-empty** state. All
  design intent that was load-bearing in the deleted code has been
  lifted into the appropriate ADR or spec before deletion:
  - ADR 0007 gained sections on the keyword-author surface
    (`Args`, `State`, `Station` interfaces; `Func` signature;
    determinism rule; mock-friendliness contract; authoring
    patterns; resolver inspection commands).
  - Spec 001 gained an Implementation Notes section covering the
    JSON decoding quirk (arrays decode to `[]any`, numbers to
    `float64`), the per-station scratch space contract for
    request/response correlation, and the OCPP-J MessageType
    constants.
  - Spec 002 gained a Starter Keyword Catalog section listing the
    eleven OCPP 1.6 keyword patterns (9 lifecycle + 2 reservation)
    that the example stories reference.
- The `Makefile` and `.github/workflows/ci.yml` carry STATUS
  comments noting that Go-related targets are reserved and will
  activate once code lands.

### Added
- Three new OCPP 1.6 conformance stories independently authored from
  the OCPP-J 1.6 specification:
  - `boot_sequence_accepted` — exercises the full cold-boot
    sequence (BootNotification → StatusNotification per connector
    → first Heartbeat) and validates that the CSMS honors the
    interval it advertised in BootNotification.conf.
  - `transaction_pluginfirst_accepted` — exercises the plugin-
    first transaction-start flow (Preparing → Authorize →
    StartTransaction → Charging).
  - `transaction_identificationfirst_accepted` — exercises the
    identification-first variant where Authorize precedes the
    Preparing status.
- Spec 002 §10 starter keyword catalog extended with 8 new patterns
  covering Heartbeat (3 patterns), Authorize (2), and
  StartTransaction (3).
- **ADR 0014** — Intellectual Property and Authoring Guidelines.
  All conformance stories derive from the public OCPP specifications
  rather than from any third-party test catalog. Naming follows
  `resource_function_desire` (snake_case). No third-party tooling
  references appear in published artefacts.
- **ADR 0015** — Test Dependency Graph. Every `.story` can be a
  prerequisite for others via the `Depends:` Meta key. Helper
  stories (no `Spec-Ref`, tagged `helper`) bring the system to
  known states; conformance stories assert specification compliance.
  Failure propagation is skip-with-finding.
- **ADR 0016** — Cache and Lock Subsystem. **Content-addressed file
  tree** at `$XDG_CACHE_HOME/octane/cache/`, with one
  `result.json` and optional `trace.json` per cache key (SHA-256
  of the (test_id, scope_key, csms_endpoint_sha, octane_version,
  ocpp_version, story_content_sha, parameter_sha) tuple). Two-
  character fanout matches the layout used by Bazel, ccache, and
  Go's build cache. Atomic writes via temp-file-and-rename.
  Cross-process safety on a single machine via POSIX advisory
  locks (`flock`); cross-machine concurrency delegated to the
  CI cache layer.
- **`CONTRIBUTING.md`** — operational authoring guidelines aligned
  with ADR 0014.
- **`docs/conformance-claim.md`** — public statement of what OCTANE
  does and does not assert.
- **Helper stories** (`station_connection_established`,
  `station_boot_accepted`, `connector_status_available`) under
  `scenarios/v16/`.
- **Conformance story** `connector_reservation_faulted.story`
  demonstrating the dependency chain in practice.
- **Lifecycle keyword catalog** specified in spec 002 §10 with 9
  OCPP 1.6 patterns covering the helper-story dependency chain
  (`station_connection_established` → `station_boot_accepted` →
  `connector_status_available`).

### Changed
- **Constitution bumped to v1.4.0.** Principle I rewritten to trace
  conformance tests to OCPP specification sections (via `Spec-Ref`)
  rather than to any third-party catalog scenario ID. Helper stories
  may omit `Spec-Ref`.
- **Story Meta keys.** `Title:` renamed to `Name:`; new required
  `Id:` for slug-based dependency references; `Octt-Ref:` removed
  in favor of `Spec-Ref:` pointing to OCPP specification sections.
- **Example story files renamed** to the
  `resource_function_desire` snake_case schema:
  - `TC_B_01_CS.story` → `boot_notification_accepted.story`
  - `TC_E_07_CS.story` → `authorize_concurrent_rejected.story`
  - `TC_PR_01.story` → `boot_notification_malformed.story`
- **Sweep across the project** removed every reference to
  third-party CSMS testing tooling from published artefacts (ADRs,
  ARCHITECTURE.md, README.md, AGENTS.md, CLAUDE.md, man pages,
  spec files, story files, Go code comments, agent files,
  website prose).

### Added
- Spec-driven development scaffolding: constitution, templates,
  `.specify/` scripts, `AGENTS.md`, `CLAUDE.md`.
- Claude Code subagents: architect, backend, **keyword-author**,
  devops, qa, security, reviewer, docs.
- Slash commands: `/specify`, `/plan`, `/tasks`, `/implement`, `/adr`,
  `/check`.
- ADR-0001 through ADR-0004 (license, language, WebSocket, reference CSMS).
- ADR-0005 (story-driven framework), ADR-0006 (`.story` DSL grammar),
  ADR-0007 (layered keyword library), ADR-0008 (multi-station
  orchestration), ADR-0009 (Robot Framework `output.xml` compatibility),
  ADR-0010 (connection profiles).
- ADR-0011 (manual pages — cobra §1, scdoc §5/§7), ADR-0012 (shell
  completion — bash/zsh, dynamic, read-only rule),
  ADR-0013 (Docusaurus website, separate from man pages).
- Constitution principles XI (wire conformance) and XII (no CSMS-
  specific adaptation).
- Spec `001-bootstrap-engine` (revised) covering the wire path foundation.
- Spec `002-story-framework` covering the DSL parser, keyword library,
  multi-station orchestration, and Robot XML emission.
- Example stories `scenarios/v201/TC_B_01_CS.story`,
  `TC_E_07_CS.story`, `TC_PR_01.story`.
- Man-page sources under `docs/man/` for §5 (config, story) and §7
  (concepts).
- Packaging via goreleaser + nfpm (`packaging/nfpm.yaml`,
  `.goreleaser.yaml`) producing `.deb`, `.rpm`, Homebrew formula,
  and SBOM-attested static binaries.
- Generation scripts: `scripts/gen-manpages.sh`,
  `scripts/gen-completions.sh`.
- Docusaurus website skeleton under `website/`.
- CI workflow `docs.yml` (man pages, completions, website build).

### Changed
- Constitution bumped from 1.0.0 → 1.2.0.
- README and AGENTS.md updated for the story-driven architecture.

### Deprecated
- *(none)*

### Removed
- An earlier exploration of vendor-implemented test harness adapters
  (no surviving ADRs).
- `api/octane/v1/` OpenAPI artifacts and `reference/citrineos-tha/`
  TypeScript scaffolding (the THA mechanism was retired).
- Spec `002-test-harness-adapter` — replaced by `002-story-framework`.

### Fixed
- *(none)*

### Security
- Constitution principle X unchanged; the story-framework pivot does
  not introduce any new privileged-access surface (no THA = no
  vendor-side admin API).
