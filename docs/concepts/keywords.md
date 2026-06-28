# Keywords

A **keyword** is a Go function that maps one sentence in a `.story` file to
one unit of wire behavior. The function receives the step's bound arguments,
drives the OCPP-J wire, and returns `nil` on success or an error that becomes
a finding in the run report.

This document describes the two-layer model, how the resolver matches step
text to keywords, the `{name:type}` placeholder grammar, and the error types
the resolver emits.

---

## Two-layer model

OCTANE organizes keywords into exactly two layers (ADR 0007; constitution
principle XII):

```txt
Layer 2  Domain keywords    pkg/keywords/domain/v16/
                            pkg/keywords/domain/
                            pkg/keywords/domain/v21/

Layer 1  Primitive keywords pkg/keywords/primitive/
```

### Domain layer

Domain keywords encode OCPP semantics. They are scoped to a specific OCPP
version and are invisible to stories that declare a different version. A story
declaring `OCPP 1.6` only sees domain keywords registered with
`api.OCPP16`; keywords registered with a different version are not
candidates.

Domain keywords do not vary by CSMS. There is no per-CSMS override layer. If
a CSMS deviates from the OCPP specification, the keyword fails against it and
the deviation surfaces as a finding.

### Primitive layer

Primitive keywords are transport-level operations: opening a WebSocket,
sending a raw frame, asserting a frame field. They are OCPP-version-agnostic
and are eligible for every story regardless of the declared OCPP version.
Primitive keywords are escape hatches; domain keywords compose them.

---

## How the resolver works

The resolver is invoked with a step string and the story's declared OCPP
version. It returns a `Match` (the matched `api.Keyword` and bound `api.Args`)
or a typed error.

### Eligibility

A keyword is a candidate for resolution when:

- It is a primitive-layer keyword (always eligible regardless of
  `OCPPVersion`), or
- It is a domain-layer keyword whose `OCPPVersion` equals the story's declared
  version, or
- It is a domain-layer keyword with a zero `OCPPVersion` value (treated as
  version-agnostic).

### Resolution order

Among eligible candidates:

1. Domain-layer keywords are tried before primitive-layer keywords.
2. Within each layer, longer patterns (by character count) are tried before
   shorter ones, so more-specific patterns win over less-specific ones.
3. The first pattern that matches is returned; remaining patterns are not
   consulted.

This ordering is deterministic for a given set of registered keywords
regardless of registration order.

### Pattern matching

A pattern match succeeds when every literal segment of the pattern appears in
the step text (case-insensitively, with flexible whitespace) and every
`{name:type}` placeholder captures exactly one whitespace-delimited token from
the step text that can be coerced to the declared type.

---

## The `{name:type}` placeholder grammar

Patterns are strings containing literal text and typed placeholders:

```txt
station {station:station} sends BootNotification with reason {reason:string}
```

### Supported types

| Type       | Go type         | Accepts in step text                                                                             |
|------------|-----------------|--------------------------------------------------------------------------------------------------|
| `string`   | `string`        | any whitespace-delimited token                                                                   |
| `int`      | `int`           | base-10 integer (no fraction)                                                                    |
| `float`    | `float64`       | decimal number                                                                                   |
| `bool`     | `bool`          | `true` or `false` (case-insensitive)                                                             |
| `duration` | `time.Duration` | Go duration string: `30s`, `1m30s`, `500ms`                                                      |
| `station`  | `string`        | station handle; semantically distinct from `string` — the resolver can validate handle existence |
| `any`      | `string`        | any token, stored as a raw string without further coercion                                       |

### Rules

- Placeholder names must be unique within a pattern.
- The colon separator is required: `{name:type}`, not `{name}`.
- An unrecognized type causes a registration-time error, not a runtime error.
- Optional parameters are not supported. Use two separate patterns pointing to
  the same `Func` when both a short form and a long form are needed.

---

## Resolver error types

### `*registry.ErrNoMatch`

Returned when no registered keyword pattern matches the step text. The error
carries:

- `StepText` — the full step string that failed to match.
- `Closest` — the nearest registered pattern by Levenshtein distance, or an
  empty string when no pattern is within edit distance 5.

```go
var noMatch *registry.ErrNoMatch
if errors.As(err, &noMatch) {
    fmt.Println("unmatched step:", noMatch.StepText)
    if noMatch.Closest != "" {
        fmt.Println("did you mean:", noMatch.Closest)
    }
}
```

The `Closest` field is populated only when the nearest pattern's Levenshtein
distance is 5 or less. Patterns farther than this are not surfaced.

### `*registry.ErrTypeMismatch`

Returned when a pattern matches structurally but a placeholder value cannot be
coerced to its declared type. For example, a step supplying `"abc"` for a
`{count:int}` placeholder.

The error carries:

- `ArgName` — the placeholder name (e.g., `"count"`).
- `Expected` — the declared type (e.g., `"int"`).
- `Got` — the raw string token from the step text.

```go
var mismatch *registry.ErrTypeMismatch
if errors.As(err, &mismatch) {
    fmt.Printf(
        "argument %q: expected %s, got %q\n",
        mismatch.ArgName,
        mismatch.Expected,
        mismatch.Got,
    )
}
```

---

## Inspecting the registered keyword set

`registry.All()` returns a stable-sorted copy of every registered keyword in
`(Layer ascending, OCPPVersion ascending, Pattern lexicographic)` order. The
order is deterministic and byte-identical across calls with the same set of
registered keywords (constitution principle IV).

The CLI commands `octane keywords list` and `octane keywords resolve` (defined
in spec 006) expose this data for interactive inspection.

---

## Related

- [ADR 0007](../adr/0007-keyword-library-layering.md) — design rationale for
  the two-layer model and resolution rules.
- [Keyword author guide](../../CONTRIBUTING.md#keyword-author-guide) —
  step-by-step instructions for writing and registering a keyword.
- `pkg/keywords/api` — the public Go interface contract.
- `pkg/keywords/registry` — the registration and resolver implementation.
