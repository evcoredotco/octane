octane.story(5)

# NAME

octane.story - Gherkin-flavored DSL for OCPP conformance scenarios

# SYNOPSIS

Files with the *.story* extension hold a single conformance scenario
each. *octane*(1) discovers and executes them via *octane-run*(1).

# DESCRIPTION

A story file declares an OCPP conformance scenario in three required
sections that appear in fixed order: *Meta*, *Background*, and
*Scenario*. Optional *Setup* and *Teardown* sections may also appear.

Stories are executed by the keyword library. CSMS connection details
(URL, port, subprotocol) live in connection profiles (see
*octane.yml*(5)). Stories must not embed CSMS-specific assumptions.

# GRAMMAR

```
story        := meta_section background? setup? scenario+ teardown?
meta_section := "Meta" NEWLINE meta_entry+
meta_entry   := INDENT IDENT ":" value NEWLINE
background   := "Background" NEWLINE step+
setup        := "Setup" NEWLINE step+
scenario     := "Scenario" ":" text NEWLINE step+
teardown     := "Teardown" NEWLINE step+
step         := INDENT step_keyword text NEWLINE
step_keyword := "Given" | "When" | "Then" | "And" | "But"
```

Indentation is whitespace-significant: bodies are indented by exactly
four spaces. Tabs are forbidden.

# META KEYS

*Name* (required)
	Free text human-readable name.

*Id* (required)
	A snake_case slug identifying the story. Used in *Depends:*
	references from other stories.

*Spec-Ref* (required for conformance tests, forbidden for helpers)
	Format: *OCPP-<version> §<section> <message>*. Identifies the
	OCPP specification section the test derives from. Helper stories
	tagged *helper* must omit this key; conformance stories must
	include it.

*Tags* (required)
	Comma-separated list. At least one of *wire-only*,
	*multi-station*, *operator-assisted*, *helper*. *pure-protocol*
	may be added to declare independence from any fixture.

*Stations* (required)
	Integer >= 1. Pre-allocated station handles.

*Timeout* (optional)
	Default per-step timeout. Format: *30s*, *2m*. Overrides the
	*defaults.timeout* in *octane.yml*(5).

*Auto-Boot* (optional, boolean, default true)
	When true, the runner BootNotifies all declared stations during
	*Background*. Set to *false* in scenarios that test the
	BootNotification handshake itself.

# STEPS

Step text is matched against keyword patterns from the layered keyword
library (profile, then domain, then primitive). A step that does not
match any keyword causes the runner to exit with code 4 and print the
file, line, and the layers searched.

Patterns use *{name:type}* placeholders. Supported types:

. *string*    - double-quoted text
. *int*       - integer
. *float*     - floating-point number
. *bool*      - *true* or *false*
. *duration*  - *30 seconds*, *5 minutes*
. *station*   - alias of string, must be a registered station handle
. *any*       - any of the above; the keyword coerces

# PARALLEL BLOCKS

Steps inside a *Parallel ... End-Parallel* block run concurrently:

```
Parallel
    When  station "CP01" sends StartTransaction
    When  station "CP02" sends StartTransaction
End-Parallel
```

Order on the wire is non-deterministic; the report records both
observed orders.

# COMMENTS

Lines starting with *#* are comments. Trailing comments are not
permitted.

# EXAMPLE

```
Meta
    Spec-Ref:    OCPP-2.0.1 / TC_B_01_CS
    Title:       BootNotification - accepted registration
    Tags:        core, boot, wire-only
    Stations:    1
    Timeout:     30s

Background
    Given the CSMS is reachable
    And   the profile defines station "CP01"

Scenario: Accepted registration on first boot
    When  station "CP01" sends BootNotification with reason "PowerUp"
    Then  the CSMS responds with status "Accepted" within 30 seconds
    And   the response includes a heartbeatInterval between 30 and 86400

Teardown
    Disconnect station "CP01"
```

# DIAGNOSTICS

Parser errors include the file path, line number, column, and a
suggestion for the closest registered keyword (Levenshtein distance
<= 4) when a keyword does not resolve.

# SEE ALSO

*octane*(1), *octane-run*(1), *octane-validate*(1), *octane.yml*(5),
*octane*(7)

Authoring guide: <https://octane.dev/docs/authoring/first-story>.

The DSL grammar is defined in ADR 0006 of the OCTANE source tree.
