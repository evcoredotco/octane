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
	return get[string](a.values, name, "string")
}

// Int returns the int value bound to name. It panics if name
// is absent or if the stored value is not an int.
func (a Args) Int(name string) int {
	return get[int](a.values, name, "int")
}

// Float returns the float64 value bound to name. It panics if
// name is absent or if the stored value is not a float64.
func (a Args) Float(name string) float64 {
	return get[float64](a.values, name, "float64")
}

// Bool returns the bool value bound to name. It panics if name
// is absent or if the stored value is not a bool.
func (a Args) Bool(name string) bool {
	return get[bool](a.values, name, "bool")
}

// Duration returns the time.Duration value bound to name. It
// panics if name is absent or if the stored value is not a
// time.Duration.
func (a Args) Duration(name string) time.Duration {
	return get[time.Duration](a.values, name, "duration")
}

// Station returns the station handle string bound to name. The
// station type is semantically distinct from a bare string in
// the pattern grammar ({name:station}), but it is stored and
// retrieved as a Go string. It panics if name is absent or if
// the stored value is not a string.
func (a Args) Station(name string) string {
	return get[string](a.values, name, "station")
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

// get is the generic typed accessor. It panics with a message
// that names the key and the expected type when the key is
// absent or the stored value has an incompatible type.
func get[T any](
	m map[string]any,
	name string,
	typeName string,
) T {
	rawValue, found := m[name]
	if !found {
		panic(fmt.Sprintf(
			"api.Args: key %q not found; "+
				"expected type %s — this is a registry "+
				"bug, not an authoring bug",
			name, typeName,
		))
	}

	typed, typeOK := rawValue.(T)
	if !typeOK {
		panic(fmt.Sprintf(
			"api.Args: key %q has type %T, "+
				"want %s — the pattern type and the "+
				"accessor do not agree; fix the keyword "+
				"registration",
			name, rawValue, typeName,
		))
	}

	return typed
}
