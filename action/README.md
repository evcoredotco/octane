# octane-action

GitHub Action wrapper for the [OCTANE](../README.md) OCPP conformance
harness. Runs the same engine as the `octane` CLI; outputs are byte-identical
for identical inputs (constitutional principle IV).

## Usage

```yaml
- name: OCPP conformance
  uses: octane-project/octane-action@v0
  with:
    csms: wss://csms.example.org/ocpp/CP01
    ocpp-version: "2.0.1"
    scenario: all
    seed: "42"
    fail-on: major
  env:
    OCTANE_BASIC_AUTH_USER: ${{ secrets.OCTANE_BASIC_AUTH_USER }}
    OCTANE_BASIC_AUTH_PASS: ${{ secrets.OCTANE_BASIC_AUTH_PASS }}
```

## Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `csms` | yes | — | Target CSMS WSS endpoint. |
| `ocpp-version` | yes | `2.0.1` | One of `1.6`, `2.0.1`, `2.1`. |
| `scenario` | yes | `all` | Scenario id, comma-list, or `all`. |
| `config` | no | `""` | Path to a YAML config file in the workspace. |
| `seed` | no | `0` | Deterministic seed for reproducible runs. |
| `insecure` | no | `false` | Disable TLS verification (banner added to report). |
| `report-path` | no | `report.json` | Output path inside `$GITHUB_WORKSPACE`. |
| `fail-on` | no | `major` | Severity threshold: `block`, `major`, `minor`, `never`. |

## Outputs

| Output | Description |
|--------|-------------|
| `report-path` | Path to the produced JSON report. |
| `passed` | `true` if findings did not exceed `fail-on`. |
| `summary` | One-line summary suitable for `$GITHUB_STEP_SUMMARY`. |

## Pinning

This Action follows the GitHub convention of a movable major-version tag:

- `octane-project/octane-action@v0` — latest in the v0.x line (recommended)
- `octane-project/octane-action@v0.1.0` — exact version (most reproducible)

Both forms are signed and attested via SLSA provenance.

## Secrets and credentials

Credentials are sourced exclusively from environment variables. Never
embed them in the workflow inputs. Supported variables:

| Env var | Purpose |
|---------|---------|
| `OCTANE_BASIC_AUTH_USER` / `OCTANE_BASIC_AUTH_PASS` | OCPP-J Basic auth |
| `OCTANE_MTLS_CERT` / `OCTANE_MTLS_KEY` | mTLS client cert/key (PEM) |
| `OCTANE_BEARER_TOKEN` | OAuth2 / JWT bearer token |

All credential material is redacted from the report.

## License

Apache-2.0. See [`LICENSE`](../LICENSE).
