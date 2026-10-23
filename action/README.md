# octane-action

GitHub Action wrapper for the [OCTANE](../README.md) OCPP conformance
harness. Runs the same engine as the `octane` CLI; outputs are byte-identical
for identical inputs (constitutional principle IV).

The action lives in the `action/` directory of the OCTANE repository and runs
the `ghcr.io/evcoredotco/octane` container image.

## Usage

The stories and `octane.yml` come from your repository's workspace.

```yaml
- uses: actions/checkout@v4

- name: OCPP conformance
  uses: evcoredotco/octane/action@v0
  with:
    csms-endpoint: wss://csms.example.org/ocpp
    stories: scenarios/v16/
    fail-on: major

- name: Upload reports
  if: always()
  uses: actions/upload-artifact@v4
  with:
    name: octane-reports
    path: reports/
```

## Inputs

| Input                  | Required | Default      | Description                                                    |
|------------------------|----------|--------------|----------------------------------------------------------------|
| `csms-endpoint`        | yes      | —            | Base WebSocket URL of the CSMS under test.                     |
| `stories`              | yes      | `scenarios/` | Paths of `.story` files or directories, separated by spaces.   |
| `fail-on`              | no       | `major`      | Threshold that fails the step: `any` or `major`.               |
| `config`               | no       | `octane.yml` | Path to the configuration file in the workspace.               |
| `cache-dir`            | no       | `""`         | Cache directory override.                                      |
| `report-dir`           | no       | `reports/`   | Directory for the generated reports.                           |
| `ocpp-version`         | no       | `""`         | Restrict the run to stories declaring this OCPP version.       |
| `shard`                | no       | `""`         | Shard in `N/M` format, to split stories across parallel jobs.  |
| `max-parallel`         | no       | `1`          | Maximum number of stories run concurrently.                    |
| `no-cache`             | no       | `false`      | Set to `true` to bypass the result cache.                      |
| `insecure-skip-verify` | no       | `false`      | Set to `true` to disable TLS verification (never in production). |

Story parameters come from the `parameters` map in `octane.yml`.

## Outputs

| Output        | Description                                  |
|---------------|----------------------------------------------|
| `report-path` | Directory containing the generated reports.  |
| `exit-code`   | Numeric exit code returned by `octane run`.  |

## Pinning

The action follows the GitHub convention of a movable major-version tag:

- `evcoredotco/octane/action@v0` — latest in the v0.x line (recommended)
- `evcoredotco/octane/action@v0.1.1` — pins the action definition

Both forms run the `v0` tag of the container image, which follows the
latest stable v0.x release. For a fully pinned run, use the CLI image
directly at an exact version (see `docs/distribution.md`).

## Credentials

CSMS authentication (HTTP Basic auth, client certificates, bearer tokens) is
not supported yet. It is planned with connection profiles (ADR 0010).

## License

Apache-2.0. See [`LICENSE`](../LICENSE).
