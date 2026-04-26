# OCTANE consumer examples

This directory holds copy-pasteable artifacts for projects *using*
OCTANE to test their CSMS. They are not used by OCTANE itself —
they're examples for downstream consumers.

## Layout

| Path | Purpose |
|------|---------|
| `ci/github-actions/ocpp-conformance.yml` | Drop into a project's `.github/workflows/` directory. Demonstrates cache integration, matrix strategy, and report artifact upload. |
| `ci/gitlab-ci/.gitlab-ci.yml` | Drop into a project's repository root. Demonstrates equivalent behavior on GitLab. |

Both examples follow the cache-integration pattern documented in
ADR 0016 (Cache and Lock Subsystem) and use OCTANE's content-
addressed file tree cache to avoid redundant test execution across
CI runs.

## What to adjust

Both example workflows assume:

- Your CSMS can be brought up via a docker image or `docker compose`.
  Adapt the relevant service / step to match your deployment
  (Helm chart, external endpoint, in-cluster deployment, etc.).
- You have an `octane.yml` at the repository root and your
  `.story` files under `scenarios/v<ocpp-version>/`.
- You want to fan out by OCPP version. Remove or simplify the
  matrix / parallel jobs if you only target one version.

The cache key includes a hash of every scenario file, connection
profile, and `octane.yml`. Edit any of those, and the cache is
invalidated for that combination at the GitHub/GitLab level —
OCTANE's per-entry SHA check then handles fine-grained per-test
invalidation within the restored cache.

## Why split GitHub and GitLab examples

The two CI platforms differ in how cache keys, restore-keys, and
parallel-job sharing work. A single "generic" example would obscure
the platform-specific patterns that actually matter. Keep the two
files independent.
