---
sidebar_position: 1
---

# CI Integration

OCTANE is built to gate pull requests on OCPP conformance. It runs as the
`octane` CLI in any CI system, or through the `octane-action` GitHub
Action — both wrap the same engine. A run exits `0` when every story
passed and `1` when any failed, which is exactly what a required check
needs.

## The shape of a conformance job

Every CI integration follows the same five steps:

1. **Check out** the repository.
2. **Stand up the CSMS** under test (or point at an existing endpoint).
3. **Restore the OCTANE cache** so unchanged stories are skipped.
4. **Run** the conformance suite.
5. **Upload the reports** as artifacts (always, even on failure).

## GitHub Actions

### Using the CLI directly

Use the CLI when you need to set the endpoint explicitly with
`--csms-endpoint`:

```yaml
name: ocpp-conformance
on:
  push: { branches: [main] }
  pull_request: { branches: [main] }

permissions:
  contents: read

concurrency:
  group: octane-${{ github.ref }}
  cancel-in-progress: true

jobs:
  conformance:
    runs-on: ubuntu-latest
    env:
      OCTANE_CACHE_DIR: ${{ github.workspace }}/.octane-cache
    steps:
      - uses: actions/checkout@v4

      - name: Start CSMS
        run: |
          cd test/reference
          docker compose up -d --wait     # CitrineOS on ws://localhost:9210

      - name: Set up Go
        uses: actions/setup-go@v5
        with: { go-version: '1.26' }

      - name: Restore OCTANE cache
        uses: actions/cache@v4
        with:
          path: ${{ env.OCTANE_CACHE_DIR }}
          key: octane-${{ runner.os }}-${{ hashFiles('scenarios/**', 'connections/**', 'octane.yml') }}
          restore-keys: |
            octane-${{ runner.os }}-

      - name: Run conformance
        run: go run ./cmd/octane run scenarios/v16 --csms-endpoint ws://localhost:9210

      - name: Upload reports
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: octane-reports
          path: reports/
          retention-days: 14
```

### Using the GitHub Action

The published action wraps the binary in a container:

```yaml
      - name: Run OCTANE
        uses: evcoreco/octane-action@v0
        with:
          stories: scenarios/v16/
          fail-on: major
          report-dir: reports/
```

#### Action inputs

| Input | Default | Description |
|---|---|---|
| `stories` | `scenarios/` | Path or glob of `.story` files to run. |
| `fail-on` | `major` | Severity threshold that fails the action. |
| `config` | `octane.yml` | Path to the config file. |
| `cache-dir` | `` | Cache directory override. |
| `report-dir` | `reports/` | Output directory for reports. |
| `ocpp-version` | `` | Restrict to stories declaring this version. |
| `shard` | `` | Shard index in `N/M` format. |
| `max-parallel` | `1` | Max stories run concurrently. |
| `no-cache` | `false` | Bypass the result cache. |
| `insecure-skip-verify` | `false` | Disable TLS verification (never in production). |

#### Action outputs

| Output | Description |
|---|---|
| `report-path` | Directory containing the generated reports. |
| `exit-code` | Numeric exit code from `octane run`. |

The action runs the image `ghcr.io/evcoreco/octane:v0`.

## GitLab CI

```yaml
ocpp-conformance:
  image: ghcr.io/evcoreco/octane:v0
  services:
    - name: citrineos/citrineos:latest
      alias: csms
  variables:
    OCTANE_CACHE_DIR: "$CI_PROJECT_DIR/.octane-cache"
  cache:
    key:
      files: [scenarios/**, octane.yml]
    paths: [.octane-cache/]
  script:
    - octane run scenarios/v16 --csms-endpoint ws://csms:9210
  artifacts:
    when: always
    paths: [reports/]
```

## Sharding for fan-out

Large suites split cleanly across parallel workers. Each shard runs a
disjoint subset selected by `sha256(test_id) % M`:

```bash
octane run scenarios/v16 --csms-endpoint ws://localhost:9210 --shard 1/4
```

In a GitHub Actions matrix:

```yaml
    strategy:
      matrix:
        shard: ['1/4', '2/4', '3/4', '4/4']
    steps:
      - run: go run ./cmd/octane run scenarios/v16 --csms-endpoint ws://localhost:9210 --shard ${{ matrix.shard }}
```

## Cache strategy

The OCTANE [cache](../concepts/dependency-graph.md) is a content-addressed
file tree designed for CI restoration:

- Key the CI cache on the files that affect results — typically
  `scenarios/**`, `connections/**`, and `octane.yml`.
- Use `restore-keys` (or a `files:` cache key in GitLab) for **partial**
  hits: a commit that touches one scenario invalidates only that
  scenario's entry and reuses the rest.
- Disjoint shards write disjoint key-hash prefixes and merge cleanly.
- Pass `--no-cache` to force a full clean run (for example, a nightly
  validation against a fresh CSMS build).

:::tip Make conformance a required check
Add an aggregate gate job that depends on the conformance job(s) and fail
it unless they all succeeded; mark that job required in branch protection.
:::

## Next

- **[Reports](./reports.md)** — what the artifacts contain.
- **[Troubleshooting](./troubleshooting.md)** — diagnosing CI failures.
- **[CLI reference](../reference/cli.md)** — every flag the job can pass.
