package registry

import "fmt"

// NoMatchError is returned by the resolver when no registered
// keyword pattern matches the given step text. It carries the
// unmatched step text and, when available, a suggested closest
// pattern (determined by Levenshtein distance, capped at edit
// distance 5).
//
// Callers should use [errors.As] to extract the typed error:
//
//	var noMatch *NoMatchError
//	if errors.As(err, &noMatch) {
//	    fmt.Println("unmatched step:", noMatch.StepText)
//	    if noMatch.Closest != "" {
//	        fmt.Println("did you mean:", noMatch.Closest)
//	    }
//	}
type NoMatchError struct {
	// StepText is the full step text that failed to match any
	// registered keyword pattern.
	StepText string

	// Closest is the nearest registered pattern by Levenshtein
	// distance, provided only when the edit distance is within 5.
	// An empty string means no sufficiently close pattern was
	// found.
	Closest string
}

// Error returns a human-readable message describing the
// unmatched step. When a close pattern suggestion is available,
// it is appended as a "did you mean" hint.
func (e *NoMatchError) Error() string {
	if e.Closest != "" {
		return fmt.Sprintf(
			"no keyword matches step %q (did you mean: %q?)",
			e.StepText,
			e.Closest,
		)
	}

	return fmt.Sprintf(
		"no keyword matches step %q",
		e.StepText,
	)
}

// TypeMismatchError is returned by the resolver when a
// placeholder capture in a matched pattern cannot be coerced to
// the type declared in the {name:type} placeholder. For example,
// if the pattern declares {n:int} and the step text supplies
// "abc", the resolver returns an TypeMismatchError with ArgName
// "n", Expected "int", and Got "abc".
//
// Callers should use [errors.As] to extract the typed error:
//
//	var mismatch *TypeMismatchError
//	if errors.As(err, &mismatch) {
//	    fmt.Printf(
//	        "argument %q: expected %s, got %q\n",
//	        mismatch.ArgName,
//	        mismatch.Expected,
//	        mismatch.Got,
//	    )
//	}
type TypeMismatchError struct {
	// ArgName is the placeholder name from the keyword pattern
	// (e.g., "n" in {n:int}).
	ArgName string

	// Expected is the declared type name from the keyword
	// pattern (e.g., "int", "duration", "bool").
	Expected string

	// Got is the raw string token from the step text that could
	// not be coerced to the expected type.
	Got string
}

// Error returns a human-readable message identifying the
// argument, its expected type, and the raw value that failed
// coercion.
func (e *TypeMismatchError) Error() string {
	return fmt.Sprintf(
		"argument %q: expected type %s, got %q",
		e.ArgName,
		e.Expected,
		e.Got,
	)
}
