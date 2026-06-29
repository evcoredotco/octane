---
sidebar_position: 2
---

# Configuration Schema

OCTANE assembles its effective configuration from four sources. Later
sources override earlier ones.

```text
  lowest  ┌─ 1. built-in defaults
          ├─ 2. octane.yml
          ├─ 3. OCTANE_* environment variables
  highest └─ 4. CLI flags
```

A field absent from `octane.yml` falls back to the default. A non-empty
environment variable overrides the file and the default. An explicitly set
CLI flag wins over everything.

## `octane.yml`

Place `octane.yml` in the directory you run OCTANE from, or pass
`--config <path>`. **Keys are camelCase.**

```yaml
# Root directory searched for .story files when none are given on the CLI.
storiesDir: scenarios

# Override the cache directory. Empty uses the XDG default.
cacheDir: ""

# Maximum number of stories that may execute concurrently.
maxParallel: 1

# Restrict execution to stories declaring this OCPP version.
# Valid value: "1.6". Empty means all versions.
ocppVersion: "1.6"

# Maximum wait when acquiring a per-cache-key lock (Go duration).
lockTimeout: 60s

# Disable TLS certificate verification. Never use in production.
insecureSkipVerify: false

# Exit threshold for the run command: "any" (default) or "major".
failOn: any

# Runtime values for placeholders declared in story Meta Parameters.
parameters:
  connectorId: "1"
  valid_idTag: "AABBCC"
  meterStart: "0"
```

### Field reference

| Field                | Type     | Default          | Environment override          |
|----------------------|----------|------------------|-------------------------------|
| `storiesDir`         | string   | `scenarios`      | —                             |
| `cacheDir`           | string   | `` (XDG default) | `OCTANE_CACHE_DIR`            |
| `maxParallel`        | int      | `1`              | `OCTANE_MAX_PARALLEL`         |
| `ocppVersion`        | string   | ``               | `OCTANE_OCPP_VERSION`         |
| `lockTimeout`        | duration | `60s`            | `OCTANE_LOCK_TIMEOUT`         |
| `insecureSkipVerify` | bool     | `false`          | `OCTANE_INSECURE_SKIP_VERIFY` |
| `failOn`             | string   | `any`            | `OCTANE_FAIL_ON`              |
| `parameters`         | map      | `{}`             | `--param name=value`          |

:::note The CSMS endpoint is not a config field
There is no endpoint key in `octane.yml`. The endpoint typically differs
between environments, so it is supplied with the `--csms-endpoint` flag at
run time. See [connection profiles](../concepts/profiles.md).
:::

## Environment variables

| Variable                      | Type     | Maps to                                         | Description                           |
|-------------------------------|----------|-------------------------------------------------|---------------------------------------|
| `OCTANE_CACHE_DIR`            | string   | `cacheDir` / `--cache-dir`                      | Cache directory override.             |
| `OCTANE_MAX_PARALLEL`         | int      | `maxParallel` / `--max-parallel`                | Max concurrent stories.               |
| `OCTANE_OCPP_VERSION`         | string   | `ocppVersion` / `--ocpp-version`                | Restrict to this OCPP version.        |
| `OCTANE_LOCK_TIMEOUT`         | duration | `lockTimeout` / `--lock-timeout`                | Cache-lock acquisition timeout.       |
| `OCTANE_FAIL_ON`              | string   | `failOn` / `--fail-on`                          | Failure threshold (`any` or `major`). |
| `OCTANE_INSECURE_SKIP_VERIFY` | bool     | `insecureSkipVerify` / `--insecure-skip-verify` | Accepts `true` or `1`.                |

Durations use Go's syntax: `60s`, `5m`, `1h30m`. If a numeric or duration
variable is present but unparseable, OCTANE silently drops the invalid
value and the lower-priority source takes effect.

## Cache directory resolution

The cache location is resolved in this order (highest priority first):

1. `--cache-dir` flag
2. `OCTANE_CACHE_DIR`
3. `$XDG_CACHE_HOME/octane/cache`
4. `$HOME/.cache/octane/cache`

OCTANE follows the XDG convention on every platform (matching Go's own
build cache), so on macOS the default is `~/.cache/octane/cache` rather
than `~/Library/Caches`.

## Next

- **[CLI reference](./cli.md)** — flags that override these fields.
- **[Dependency graph & caching](../concepts/dependency-graph.md)** — what
  the cache settings control.
- **[Exit codes](./exit-codes.md)** — a malformed config exits `64`.
