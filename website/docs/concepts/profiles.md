---
sidebar_position: 5
---

# Connection Profiles

OCTANE has no surface for adapting its conformance logic to a particular
CSMS — that is a deliberate constraint (see
[wire conformance](./wire-conformance.md)). What *is* legitimately
CSMS-specific is **connection metadata**: how to reach the system on the
network. That information lives in a small, operator-owned YAML file
called a **connection profile**.

## What a profile is — and is not

A connection profile describes **how to connect**. It is owned by the
operator running the tests, not by the CSMS vendor.

A connection profile **never** contains:

- keyword overrides or alternative step behavior,
- behavioral tolerances or "expected deviations",
- anything that could let a non-conformant CSMS pass a conformance test.

It is network metadata, full stop. Conformance logic lives in
[stories](./stories.md) and the keyword library; profiles stay out of it.

## The bundled sample

OCTANE ships a sample profile for the reference CSMS, CitrineOS, at
`connections/citrineos.yaml`. It is intentionally minimal:

```yaml
# Connection profile for a locally running CitrineOS instance.
csmsEndpoint: ws://localhost:9210
ocppVersion: "1.6"
```

| Field | Meaning |
|---|---|
| `csmsEndpoint` | Base WebSocket URL of the CSMS, **without** a station path. |
| `ocppVersion` | The OCPP version stories should run under (`1.6`). |

## Endpoints and station paths

The endpoint is the base WebSocket URL. OCTANE appends a path segment per
simulated station, so `ws://localhost:9210` becomes `ws://localhost:9210/CP01`,
`ws://localhost:9210/CP02`, and so on for each handle the story allocates.

In day-to-day use you supply the endpoint with the `--csms-endpoint` flag,
because it usually differs between environments and should not be baked
into version-controlled config:

```bash
octane run scenarios/v16 --csms-endpoint ws://localhost:9210
```

## CitrineOS reference ports

| Port | Purpose | Used by OCTANE |
|---|---|---|
| `9210` | OCPP-J WebSocket | yes — this is the endpoint |
| `8080` | REST / admin API | no |

:::info Richer connection metadata
Connection profiles are designed to grow into a fuller description of
network reach (host/port templates, subprotocol selection, auth mode) as
the project matures. Today the operative inputs are the `--csms-endpoint`
flag and `ocppVersion`; the sample above reflects the current shape.
:::

## Next

- **[Getting started](../getting-started.md)** — point OCTANE at a CSMS.
- **[Configuration schema](../reference/config-schema.md)** — `octane.yml`
  fields and environment variables.
