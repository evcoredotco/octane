package pattern

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CoercionError is returned by [Coerce] when a captured string
// token cannot be converted to the Go type declared by its
// placeholder. It carries the three fields that the registry layer
// needs to construct a [registry.ErrTypeMismatch] value for the
// caller.
//
// Callers outside the internal package (i.e., the registry itself)
// should inspect this error with [errors.As] and then wrap the
// fields in a registry.ErrTypeMismatch before returning to
// consumers. The split avoids a circular import: the registry
// package imports internal/pattern, so internal/pattern must not
// import the registry package.
type CoercionError struct {
	// ArgName is the placeholder name as declared in the pattern
	// (e.g., "n" in {n:int}).
	ArgName string

	// Expected is the placeholder type name as declared in the
	// pattern (e.g., "int", "bool", "duration").
	Expected string

	// Got is the raw string token supplied by the step text that
	// could not be converted to the expected type.
	Got string
}

// Error returns a human-readable description of the failed
// coercion in the same format used by registry.ErrTypeMismatch so
// that test assertions can compare message strings without
// importing the registry package.
func (e *CoercionError) Error() string {
	return fmt.Sprintf(
		"argument %q: expected type %s, got %q",
		e.ArgName,
		e.Expected,
		e.Got,
	)
}

// Coerce converts the raw string captures produced by [Match] into
// typed Go values according to the declared [PlaceholderType] of
// each [KindPlaceholder] token in tokens.
//
// The captures map must be the map returned by a successful [Match]
// call: keys are placeholder names, values are the raw step-text
// words. The tokens slice must be the slice returned by [Parse] for
// the same pattern. Only KindPlaceholder tokens are processed; all
// KindLiteral tokens are ignored.
//
// Coerce returns a new map[string]any whose keys are placeholder
// names and whose values have the following Go types:
//
//   - TypeString   → string (no conversion required)
//   - TypeAny      → string (stored as a raw string)
//   - TypeStation  → string (station handles are strings at the wire
//     level; semantic validation is the registry's job)
//   - TypeInt      → int   (base-10; strconv.Atoi)
//   - TypeFloat    → float64 (strconv.ParseFloat, bit-size 64)
//   - TypeBool     → bool  ("true"/"false", case-insensitive)
//   - TypeDuration → time.Duration (time.ParseDuration)
//
// If any conversion fails, Coerce returns nil and a *[CoercionError]
// identifying the argument name, declared type, and raw value that
// triggered the failure. The registry layer is expected to wrap the
// *CoercionError fields into a registry.ErrTypeMismatch before
// surfacing it to callers.
func Coerce(
	captures map[string]string,
	tokens []Token,
) (map[string]any, error) {
	result := make(map[string]any, len(captures))

	for i := range tokens {
		tok := tokens[i]
		if tok.Kind != KindPlaceholder {
			continue
		}

		raw, ok := captures[tok.Name]
		if !ok {
			continue
		}

		coerced, err := coerceOne(tok.Name, tok.Type, raw)
		if err != nil {
			return nil, err
		}

		result[tok.Name] = coerced
	}

	return result, nil
}

// coerceOne converts a single raw string value to the Go type
// declared by pType. It returns a *CoercionError when the
// conversion fails.
func coerceOne(
	name string,
	pType PlaceholderType,
	raw string,
) (any, error) {
	switch pType {
	case TypeString, TypeAny, TypeStation:
		return raw, nil

	case TypeInt:
		intVal, err := strconv.Atoi(raw)
		if err != nil {
			return nil, &CoercionError{
				ArgName:  name,
				Expected: string(TypeInt),
				Got:      raw,
			}
		}

		return intVal, nil

	case TypeFloat:
		floatVal, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, &CoercionError{
				ArgName:  name,
				Expected: string(TypeFloat),
				Got:      raw,
			}
		}

		return floatVal, nil

	case TypeBool:
		switch strings.ToLower(raw) {
		case "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return nil, &CoercionError{
				ArgName:  name,
				Expected: string(TypeBool),
				Got:      raw,
			}
		}

	case TypeDuration:
		durVal, err := time.ParseDuration(raw)
		if err != nil {
			return nil, &CoercionError{
				ArgName:  name,
				Expected: string(TypeDuration),
				Got:      raw,
			}
		}

		return durVal, nil

	default:
		// Parse rejects unknown types at registration time, so
		// reaching this branch indicates an internal invariant
		// violation rather than a user-facing authoring error.
		return nil, fmt.Errorf(
			"internal: unknown placeholder type %q for argument %q",
			pType,
			name,
		)
	}
}
