# OCTANE's Conformance Claim

This document defines, precisely, what OCTANE asserts when it
reports a passing run. It exists so that operators, auditors, and
contributors can rely on a single canonical statement of OCTANE's
scope and limitations.

## What OCTANE asserts

When `octane run` completes with exit code 0 and the report shows
all conformance scenarios passing, OCTANE asserts:

> *Under the conditions described by the executed `.story` files,
> the CSMS under test exhibited wire-level behavior that matches
> what the cited sections of the published OCPP specifications
> require.*

The supporting evidence is in the report:

- Every conformance story carries a `Spec-Ref` citing the OCPP
  specification section it derives from.
- Every step's wire trace (request bytes, response bytes,
  timestamps, observed status) is captured.
- The report is byte-deterministic given the same inputs (modulo
  timestamps).

## What OCTANE does not assert

Equally important is what OCTANE explicitly does **not** claim.

### OCTANE is not a certification

OCTANE is an open-source conformance harness developed
independently. It is not affiliated with, endorsed by, or operated
on behalf of the Open Charge Alliance (OCA), and a passing OCTANE
run does not constitute formal OCA certification.

Formal certification is a separate process operated by the OCA
under its own scope, scope, and authority. Operators who require
formal certification should engage with the OCA directly through
its certification program. OCTANE is intended as a development and
CI-time tool that helps operators build confidence their CSMS will
behave correctly during such certification — not as a substitute
for it.

### OCTANE verifies the wire, not the implementation

OCTANE asserts conformance based on observable wire behavior. It
cannot verify:

- CSMS-internal state transitions invisible on the wire.
- Audit log content.
- Billing pipeline correctness.
- Background reconciliation jobs.
- Database schema validity.
- Operational concerns (uptime, performance, capacity).

Scenarios requiring privileged CSMS observability are documented as
**operator-assisted** and are skipped by default in CI mode.

### OCTANE asserts conformance to the executed scenarios

A passing OCTANE run says nothing about scenarios OCTANE did not
execute. Coverage claims are bounded by the story library; an
operator running 87 of OCTANE's stories has evidence of
conformance for those 87 cases, not for OCPP as a whole.

The report's summary section makes this explicit:

```
Wire conformance:        87/120 scenarios (PASS: 87, FAIL: 0)
Operator-assisted:       12 scenarios skipped (CI mode)
Overall conformance:     wire-tier, partial coverage
```

### OCTANE does not warrant the test suite's correctness

OCTANE's stories are independently authored from the published OCPP
specifications. We aim for accuracy, but the project makes no
warranty that every story correctly captures every specification
nuance. Errata in the OCPP specification, ambiguities in the
normative text, and authoring mistakes in OCTANE itself can all
cause an OCTANE story to assert something the specification does
not actually require, or to miss something it does.

The Apache-2.0 license under which OCTANE is distributed disclaims
all warranties. This document does not narrow that disclaimer.

## How to read an OCTANE report

A passing report is evidence, not proof. Treat it as you would
treat a green build from a unit-test suite: it is meaningful
positive signal, not a guarantee of correctness.

The right reading is:

> "Our CSMS passed every wire-level OCPP scenario in OCTANE's
> library that we executed. We have a high-quality CI gate against
> regressions in those scenarios. We expect to do well in formal
> conformance review."

The wrong readings are:

- "Our CSMS is OCPP-conformant." (Bounded by what OCTANE tests.)
- "Our CSMS is certified." (OCTANE does not certify.)
- "Our CSMS is bug-free." (Unit-test-level claim, not appropriate
  here.)

## OCPP specification source

OCTANE references OCPP specifications published by the Open Charge
Alliance:

- OCPP 1.6J: <https://www.openchargealliance.org/protocols/ocpp-16/>
- OCPP 1.6: <https://www.openchargealliance.org/protocols/ocpp-201/>
- OCPP 1.6: <https://www.openchargealliance.org/protocols/ocpp-21/>

These specifications are the source of truth for every conformance
assertion OCTANE makes. Where the OCPP specification is silent or
ambiguous, OCTANE either declines to test the relevant behavior or
documents the interpretation explicitly in the relevant story's
prose narrative.

## Questions

If you have questions about OCTANE's scope or the conformance
claim, file an issue. If you suspect a story misinterprets the
OCPP specification, file an issue with the specific section
reference and the proposed correction.
