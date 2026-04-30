package diag

import (
	"fmt"
	"strings"
)

// UnboundParameterError indicates that one or more {placeholder} tokens in a
// step's text reference parameters not declared in the Parameters: block.
type UnboundParameterError struct {
	// File is the filesystem path of the .story file.
	File string
	// Line is the 1-based line number where the error was detected.
	Line int
	// Column is the 1-based column number where the error was detected.
	Column int
	// Parameters lists the unbound parameter names, sorted and
	// deduplicated.
	Parameters []string
	// StepText is the full step text that contains the unbound
	// references.
	StepText string
	// Suggestion is a human-readable hint for fixing the problem.
	Suggestion string
}

// Error returns a formatted diagnostic string.
// Parameters is guaranteed sorted-and-deduplicated at construction time
// (by validateParameters), so no re-sort is needed here.
func (e *UnboundParameterError) Error() string {
	return fmt.Sprintf("%s:%d:%d: unbound parameter(s) {%s} in step %q; %s",
		e.File, e.Line, e.Column,
		strings.Join(e.Parameters, "}, {"),
		e.StepText, e.Suggestion)
}

// UnexpectedTokenError indicates that the parser encountered a token it did
// not expect at that position in the grammar.
type UnexpectedTokenError struct {
	// File is the filesystem path of the .story file.
	File string
	// Line is the 1-based line number where the error was detected.
	Line int
	// Column is the 1-based column number where the error was detected.
	Column int
	// Got is the string representation of the token that was found.
	Got string
	// Expected describes what the parser was expecting at this position.
	Expected string
	// Suggestion is a human-readable hint for fixing the problem.
	Suggestion string
}

// Error returns a formatted diagnostic string.
func (e *UnexpectedTokenError) Error() string {
	return fmt.Sprintf("%s:%d:%d: unexpected %s, expected %s; %s",
		e.File, e.Line, e.Column, e.Got, e.Expected, e.Suggestion)
}
