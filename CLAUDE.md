# CLAUDE.md — OCTANE

> **What this file is:** Claude Code-specific rules and domain refusals
> that extend the cross-tool agent contract in `AGENTS.md`.
>
> **What this file is not:** a substitute for `AGENTS.md`. Every rule in
> `AGENTS.md` is binding here. Where this file and `AGENTS.md`
> contradict, `AGENTS.md` wins. When in doubt about agent scope,
> responsibility, or workflow, read `AGENTS.md` first.

## Loading order

1. `ARCHITECTURE.md` section 2 — the project principles, cited elsewhere
   as "constitution principle N"
2. `AGENTS.md` — agent contract (always authoritative)
3. This file — Claude Code-specific rules
4. The ADRs in `docs/adr/` that cover the area you are changing

## Things Claude should refuse to do

- Change the principles in `ARCHITECTURE.md` section 2 without an ADR
  that records the change.
- Disable TLS verification anywhere in `pkg/transport/`.
- Add new third-party Go dependencies without a draft ADR.
- Mark a test case stable before it has passed against the pinned CitrineOS.
- Push directly to `master`. Work on a branch and open a pull request.
- Declare any OCPP 1.6 data type (struct, enum, sub-object) locally in OCTANE.
  All OCPP 1.6 types **must** come from `github.com/evcoreco/ocpp16types`
  (ADR 0020). If the required type is missing from that module, stop and
  instruct the user to contribute it upstream first.
- Construct any OCPP 1.6 request or response message outside of the
  constructors exported by `github.com/evcoreco/ocpp16messages` (ADR 0020).
  Raw struct literals for OCPP 1.6 messages are forbidden. If the required
  message sub-package is missing, stop and instruct the user to contribute
  it upstream first.
- Parse, construct, or marshal any OCPP-J JSON message (Call, CallResult,
  CallError) outside of `github.com/evcoreco/ocpp16j` (ADR 0020).
  Specifically forbidden: calling `json.Unmarshal` directly against OCPP-J
  arrays, hand-assembling `[2,"id","Action",{...}]` literals, and building
  `UniqueId` values via raw string casts. All JSON framing **must** go through
  `ocpp16json.Parse()`, `ocpp16json.NewCall()`, `ocpp16json.NewCallResult()`,
  `ocpp16json.NewCallError()`, and `ocpp16json.NewUniqueId()`. Payload decode
  **must** use `ocpp16json.Registry` + `ocpp16json.JSONDecoder[Input, Output]`.
  If the required feature is missing from `ocpp16j`, stop and instruct the
  user to contribute it upstream first.

## EVCore OCPP 1.6 three-layer standard

> **Why this is here:** The three rules in "Things Claude should refuse"
> above all reference this standard. Every agent working in OCTANE must
> know the layer boundaries. Canonical reference: ADR 0020.

```
JSON bytes  →  ocpp16j (Parse/Marshal/Validate)  →  ocpp16messages (Req/Conf)
                                                          ↓
                                                     ocpp16types (field types, enums)
```

| Layer | Module | Scope |
|-------|--------|-------|
| JSON framing | `github.com/evcoreco/ocpp16j` | Call / CallResult / CallError envelopes, UniqueId, ErrorCode, Registry |
| Message construction | `github.com/evcoreco/ocpp16messages` | `Req()` / `Conf()` constructor pairs per OCPP action |
| Data types | `github.com/evcoreco/ocpp16types` | CiString, DateTime, enums, sub-objects |

All three are first-party EVCore infrastructure and must be pinned as direct
`go.mod` dependencies (never `// indirect`).

## Quick references

- OCPP 1.6J spec: https://www.openchargealliance.org/protocols/ocpp-16/
- CitrineOS: https://citrineos.github.io/
