// Package diag defines typed diagnostic errors returned by the .story file
// parser. Every error type carries source location (file, line, column) and
// a human-readable suggestion so tooling can render actionable messages.
//
// This package is a leaf: it has zero imports from other octane packages
// to avoid import cycles.
package diag

import (
	"fmt"
	"strings"
)

// MissingKeyError indicates that a required Meta key is absent from a
// .story file.
type MissingKeyError struct {
	// File is the filesystem path of the .story file.
	File string
	// Line is the 1-based line number where the error was detected.
	Line int
	// Column is the 1-based column number where the error was detected.
	Column int
	// Key is the name of the missing Meta key (e.g. "Name", "Id", "Stations").
	Key string
	// Suggestion is a human-readable hint for fixing the problem.
	Suggestion string
}

// Error returns a formatted diagnostic string.
func (e *MissingKeyError) Error() string {
	return fmt.Sprintf("%s:%d:%d: missing required key %q; %s",
		e.File, e.Line, e.Column, e.Key, e.Suggestion)
}

// MissingSpecRefError indicates that a conformance story (one without the
// "helper" tag) is missing its Spec-Ref Meta key.
type MissingSpecRefError struct {
	// File is the filesystem path of the .story file.
	File string
	// Line is the 1-based line number where the error was detected.
	Line int
	// Column is the 1-based column number where the error was detected.
	Column int
	// Suggestion is a human-readable hint for fixing the problem.
	Suggestion string
}

// Error returns a formatted diagnostic string.
func (e *MissingSpecRefError) Error() string {
	return fmt.Sprintf("%s:%d:%d: missing Spec-Ref; %s",
		e.File, e.Line, e.Column, e.Suggestion)
}

// SpecRefOnHelperError indicates that a helper story (tagged "helper") has a
// Spec-Ref key, which is not allowed because helpers are not conformance
// tests.
type SpecRefOnHelperError struct {
	// File is the filesystem path of the .story file.
	File string
	// Line is the 1-based line number where the error was detected.
	Line int
	// Column is the 1-based column number where the error was detected.
	Column int
	// SpecRef is the Spec-Ref value that was found, helping the author
	// see what to remove.
	SpecRef string
	// Suggestion is a human-readable hint for fixing the problem.
	Suggestion string
}

// Error returns a formatted diagnostic string.
func (e *SpecRefOnHelperError) Error() string {
	return fmt.Sprintf("%s:%d:%d: helper story must not have Spec-Ref %q; %s",
		e.File, e.Line, e.Column, e.SpecRef, e.Suggestion)
}

// MalformedDependsError indicates that a single entry in a Depends: block is
// malformed and cannot be parsed.
type MalformedDependsError struct {
	// File is the filesystem path of the .story file.
	File string
	// Line is the 1-based line number where the error was detected.
	Line int
	// Column is the 1-based column number where the error was detected.
	Column int
	// EntryIndex is the 0-based index of the offending entry in the
	// Depends list.
	EntryIndex int
	// Reason is a brief description of what is wrong with the entry
	// (e.g. "missing id field", "unknown scope value \"foo\"").
	Reason string
	// Suggestion is a human-readable hint for fixing the problem.
	Suggestion string
}

// Error returns a formatted diagnostic string.
func (e *MalformedDependsError) Error() string {
	return fmt.Sprintf("%s:%d:%d: malformed Depends entry [%d]: %s; %s",
		e.File, e.Line, e.Column, e.EntryIndex, e.Reason, e.Suggestion)
}

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

// MissingSectionError indicates that a required top-level section is absent.
// Currently used when no Scenario section is present.
type MissingSectionError struct {
	// File is the filesystem path of the .story file.
	File string
	// Line is the 1-based line number where the error was detected.
	Line int
	// Column is the 1-based column number where the error was detected.
	Column int
	// Section is the name of the missing section (e.g. "Scenario").
	Section string
	// Suggestion is a human-readable hint for fixing the problem.
	Suggestion string
}

// Error returns a formatted diagnostic string.
func (e *MissingSectionError) Error() string {
	return fmt.Sprintf("%s:%d:%d: at least one %s section is required; %s",
		e.File, e.Line, e.Column, e.Section, e.Suggestion)
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
