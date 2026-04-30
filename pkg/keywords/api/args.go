package api

import (
	"fmt"
	"time"
)

// Args holds the named parameter values extracted by the pattern
// matcher when a step text is resolved against a keyword pattern.
//
// Every typed accessor (String, Int, Float, Bool, Duration,
// Station, Any) panics if the requested key is absent or if the
// stored value is not the expected Go type. This is intentional:
// the registry validates that every {name:type} placeholder in a
// keyword pattern has a corresponding accessor call in the
// keyword body. The check runs at init() time, so reaching a
// runtime panic indicates a registry bug, not an authoring bug.
//
// Keyword authors should never guard accessor calls with
// recover(); instead, fix the pattern or the registration.
type Args struct {
	values map[string]any
}

// NewArgs creates an Args from the given key-value pairs. This
// constructor is called by the resolver after a successful
// pattern match; keyword authors do not call it directly.
func NewArgs(vals map[string]any) Args {
	if vals == nil {
		vals = make(map[string]any)
	}

	return Args{values: vals}
}

// String returns the string value bound to name. It panics if
// name is absent or if the stored value is not a string.
func (a Args) String(name string) string {
	raw := lookupArg(a.values, name, "string")

	v, ok := raw.(string)
	if !ok {
		panicTypeMismatch(name, raw, "string")
	}

	return v
}

// Int returns the int value bound to name. It panics if name
// is absent or if the stored value is not an int.
func (a Args) Int(name string) int {
	raw := lookupArg(a.values, name, "int")

	v, ok := raw.(int)
	if !ok {
		panicTypeMismatch(name, raw, "int")
	}

	return v
}

// Float returns the float64 value bound to name. It panics if
// name is absent or if the stored value is not a float64.
func (a Args) Float(name string) float64 {
	raw := lookupArg(a.values, name, "float64")

	v, ok := raw.(float64)
	if !ok {
		panicTypeMismatch(name, raw, "float64")
	}

	return v
}

// Bool returns the bool value bound to name. It panics if name
// is absent or if the stored value is not a bool.
func (a Args) Bool(name string) bool {
	raw := lookupArg(a.values, name, "bool")

	v, ok := raw.(bool)
	if !ok {
		panicTypeMismatch(name, raw, "bool")
	}

	return v
}

// Duration returns the time.Duration value bound to name. It
// panics if name is absent or if the stored value is not a
// time.Duration.
func (a Args) Duration(name string) time.Duration {
	raw := lookupArg(a.values, name, "duration")

	v, ok := raw.(time.Duration)
	if !ok {
		panicTypeMismatch(name, raw, "duration")
	}

	return v
}

// Station returns the station handle string bound to name. The
// station type is semantically distinct from a bare string in
// the pattern grammar ({name:station}), but it is stored and
// retrieved as a Go string. It panics if name is absent or if
// the stored value is not a string.
func (a Args) Station(name string) string {
	raw := lookupArg(a.values, name, "station")

	v, ok := raw.(string)
	if !ok {
		panicTypeMismatch(name, raw, "station")
	}

	return v
}

// Any returns the raw value bound to name without a type
// assertion. It panics if name is absent.
func (a Args) Any(name string) any {
	rawValue, ok := a.values[name]
	if !ok {
		panic(fmt.Sprintf(
			"api.Args: key %q not found; "+
				"this is a registry bug — the pattern "+
				"should declare {%s:<type>}",
			name, name,
		))
	}

	return rawValue
}

// Has reports whether name exists in the argument set. This is
// a non-panicking probe intended for debugging and test
// assertions, not for conditional logic in keyword bodies.
func (a Args) Has(name string) bool {
	_, ok := a.values[name]

	return ok
}

// Len returns the number of bound arguments.
func (a Args) Len() int {
	return len(a.values)
}

// lookupArg retrieves the raw value for name from m. It panics with
// a message that names the key and expected type when name is absent.
// Callers perform their own type assertion after this call.
func lookupArg(m map[string]any, name, typeName string) any {
	rawValue, found := m[name]
	if !found {
		panic(fmt.Sprintf(
			"api.Args: key %q not found; "+
				"expected type %s — this is a registry "+
				"bug, not an authoring bug",
			name, typeName,
		))
	}

	return rawValue
}

// panicTypeMismatch panics with a message describing a type mismatch
// for the given argument name and raw value.
func panicTypeMismatch(name string, rawValue any, typeName string) {
	panic(fmt.Sprintf(
		"api.Args: key %q has type %T, "+
			"want %s — the pattern type and the "+
			"accessor do not agree; fix the keyword "+
			"registration",
		name, rawValue, typeName,
	))
}
