octane(7)

# NAME

octane - OCPP Conformance Testing & Network Evaluation: concepts

# DESCRIPTION

OCTANE is an open-source conformance harness for OCPP 1.6J, 2.0.1,
and 2.1 Charging Station Management Systems (CSMS). It runs against an
unmodified CSMS and verifies wire-level conformance to the publicly
published OCPP specifications, automated and CI-friendly.

This page explains the concepts. *octane*(1) is the command-line
reference; *octane.story*(5) and *octane.yml*(5) document the file
formats; the website at <https://octane.dev/> hosts tutorials and the
keyword catalog.

# WIRE CONFORMANCE

OCTANE drives the CSMS by impersonating one or more charging stations
over OCPP-J WebSocket connections. It sends OCPP messages, observes
responses, and matches them against the OCPP specification.

OCTANE *never* requires changes to the CSMS, *never* uses privileged
admin APIs, and *never* depends on a vendor-implemented adapter. A
CSMS team runs OCTANE against an existing instance with a single
command.

The honest scope of this approach is *wire conformance to OCPP*.
Scenarios that require privileged CSMS introspection (audit log
content, billing pipeline behavior, internal state transitions
invisible on the wire) are either out of scope or marked
*operator-assisted* — runnable interactively, skipped automatically
in CI mode.

# STORIES

Conformance scenarios are written as *.story* files in a Gherkin-
flavored DSL. Each conformance story carries its OCPP specification
reference in a *Meta* section, declares how many simulated charging
stations it needs, and specifies steps in *Given/When/Then* form.

Stories are the executable specification. Certification reviewers
read them as a description of what conformance looks like; the same
file is what OCTANE executes.

See *octane.story*(5) for the grammar.

# KEYWORDS

A story step (*the CSMS responds with status "Accepted"*) is matched
against patterns registered in a layered keyword library:

. Domain keywords (OCPP semantics, scoped per OCPP version)
. Primitive keywords (transport-level escape hatch)

The runner resolves a step against domain first, then primitive. The
first match wins. Stories use domain keywords almost exclusively;
primitive keywords are escape hatches for unusual scenarios.

Domain keywords are identical for every CSMS implementing a given
OCPP version. There is no per-CSMS override layer (constitution
principle XII).

# CONNECTION PROFILES

A connection profile is a small YAML file that describes how to reach
the CSMS at the network level: URL template, port, subprotocol
mapping, supported auth modes. Connection profiles are user
configuration, not vendor adaptation. They contain *no* conformance
logic and *no* keyword overrides — by constitutional rule they cannot
let a non-conformant CSMS pass a conformance test.

OCTANE ships sample connection profiles for well-known open-source
CSMSes (CitrineOS, SteVe, MaEVe). The profile is selected via the
*connection* key in *octane.yml*(5) or the *--connection* flag.

# MULTI-STATION ORCHESTRATION

Many stateful OCPP scenarios are reachable on the wire by
coordinating two or more simulated stations. A story declares the
station count in its *Meta* section; the runner allocates handles
*"CP01"*, *"CP02"*, and so on. Steps reference stations by handle.

Parallel execution is opt-in via *Parallel ... End-Parallel* blocks.
Outside parallel blocks, steps run sequentially in declared order.

# REPORTS

Every run produces *report.json* in OCTANE's native, byte-deterministic
format and (when configured) *output.xml* in Robot Framework's output
schema. The native JSON is the source of truth for certification
submissions; the XML is for ecosystem reporting tools (Allure,
ReportPortal, Jenkins, GitLab, GitHub Actions).

Both files are produced from the same in-memory report tree, so they
agree by construction. Identical inputs produce identical reports
modulo timestamps.

# EXIT CODES

. *0*  success
. *1*  one or more findings exceeded the configured *fail-on* threshold
. *2*  transport error (CSMS unreachable, TLS failure)
. *3*  preflight failed (no profile, missing capabilities, conflicting state)
. *4*  story parse or keyword resolution failed
. *5*  configuration missing or invalid
. *6*  internal error

# CONFORMANCE SCOPE

OCTANE asserts wire-level conformance to the publicly published OCPP
specifications. Operators using OCTANE can be confident their CSMS
implements the OCPP wire protocol correctly under the scenarios
exercised by the test suite.

OCTANE neither replaces nor implies any formal certification process
operated by the Open Charge Alliance. Formal OCA certification is a
separate undertaking with its own process, scope, and authority.

OCTANE's scenarios trace to specific sections of the OCPP
specifications via the *Spec-Ref* meta key in each story file.

# SEE ALSO

*octane*(1), *octane.story*(5), *octane.yml*(5)

Project home: <https://octane.dev/>
Source code:  <https://github.com/octane-project/octane>
