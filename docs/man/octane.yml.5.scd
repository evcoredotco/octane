octane.yml(5)

# NAME

octane.yml - configuration file for the *octane*(1) conformance harness

# SYNOPSIS

OCTANE reads configuration from the first matching source in this order:

. *--config* command-line flag
. *OCTANE_CONFIG* environment variable
. *./octane.yml* in the current working directory
. *./.octane/config.yml* in the current working directory
. *$XDG_CONFIG_HOME/octane/config.yml* (default *~/.config/octane/config.yml*)
. */etc/octane/config.yml*

User and project configurations are *merged*, not replaced. A project
file overrides only the keys it sets.

# FORMAT

Configuration is YAML. The schema is versioned via a top-level
*schema_version* key. Schema 1 is the current version.

# SCHEMA

```
schema_version: "1"
csms:
  url: <wss-url>
  ocpp_version: "1.6" | "2.0.1" | "2.1"
connection: <profile-name>
auth:
  mode: none | basic | mtls | bearer
  basic:
    username_env: <env-var-name>
    password_env: <env-var-name>
  mtls:
    cert_path: <path>
    key_path: <path>
  bearer:
    token_env: <env-var-name>
defaults:
  station_prefix: <string>
  timeout: <duration>
  seed: <integer>
runtime:
  max_stations: <integer>
  parallel_max: <integer>
  insecure: <bool>
report:
  output_dir: <path>
  formats: [json, robot-xml]
  redact_credentials: <bool>
```

# KEYS

*csms.url*
	Required. The WebSocket URL of the target CSMS.

*csms.ocpp_version*
	Required. One of *1.6*, *2.0.1*, *2.1*.

*connection*
	Required. The connection profile name. Resolved per the chain
	defined in *connection*(5) (project → user → system → bundled
	samples).

*auth.mode*
	Required. Selects the authentication strategy. *none* is permitted
	only for development against an unauthenticated CSMS and emits a
	banner in the report.

*defaults.timeout*
	Default per-step timeout. Story-level *Timeout* meta keys override.
	Format: integer with unit (*30s*, *5m*).

*runtime.insecure*
	Disables TLS verification for the CSMS connection. Setting this to
	*true* adds an *insecure* banner to the report and marks the run
	as *not-conformant-for-certification*.

*report.formats*
	List of report formats to emit. Always includes *json*. *robot-xml*
	is recommended for CI dashboards.

# EXAMPLES

A minimal user configuration:

```
schema_version: "1"
csms:
  url: wss://localhost:8081/ocpp
  ocpp_version: "2.0.1"
connection: citrineos
auth:
  mode: none
```

A CI configuration with mTLS:

```
schema_version: "1"
csms:
  url: wss://csms.example.org:8443/ocpp
  ocpp_version: "2.0.1"
connection: citrineos
auth:
  mode: mtls
  mtls:
    cert_path: /run/secrets/octane-client.pem
    key_path: /run/secrets/octane-client.key
defaults:
  timeout: 60s
report:
  formats: [json, robot-xml]
```

# ENVIRONMENT

*OCTANE_CONFIG*
	Path to the configuration file, overriding all default search paths.

*OCTANE_*BASIC_USER*, *OCTANE_BASIC_PASS*, *OCTANE_BEARER_TOKEN*
	Common credential variables referenced from *auth.basic.\*_env* and
	*auth.bearer.token_env*.

# SEE ALSO

*octane*(1), *octane-run*(1), *octane-validate*(1), *octane.story*(5),
*octane*(7)

The full schema and validation rules: <https://octane.dev/docs/reference/config-schema>.
