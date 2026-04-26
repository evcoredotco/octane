// Package diag defines typed diagnostic errors returned by the .story file
// parser. Every error type carries source location (file, line, column) and
// a human-readable suggestion so tooling can render actionable messages.
//
// This package is a leaf: it has zero imports from other octane packages
// to avoid import cycles.
package diag

import (
	"fmt"
	"sort"
	"strings"
)

// ErrMissingKey indicates that a required Meta key is absent from a
// .story file.
type ErrMissingKey struct {
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
func (e *ErrMissingKey) Error() string {
	return fmt.Sprintf("%s:%d:%d: missing required key %q; %s",
		e.File, e.Line, e.Column, e.Key, e.Suggestion)
}

// ErrMissingSpecRef indicates that a conformance story (one without the
// "helper" tag) is missing its Spec-Ref Meta key.
type ErrMissingSpecRef struct {
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
func (e *ErrMissingSpecRef) Error() string {
	return fmt.Sprintf("%s:%d:%d: missing Spec-Ref; %s",
		e.File, e.Line, e.Column, e.Suggestion)
}

// ErrSpecRefOnHelper indicates that a helper story (tagged "helper") has a
// Spec-Ref key, which is not allowed because helpers are not conformance
// tests.
type ErrSpecRefOnHelper struct {
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
func (e *ErrSpecRefOnHelper) Error() string {
	return fmt.Sprintf("%s:%d:%d: helper story must not have Spec-Ref %q; %s",
		e.File, e.Line, e.Column, e.SpecRef, e.Suggestion)
}

// ErrMalformedDepends indicates that a single entry in a Depends: block is
// malformed and cannot be parsed.
type ErrMalformedDepends struct {
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
func (e *ErrMalformedDepends) Error() string {
	return fmt.Sprintf("%s:%d:%d: malformed Depends entry [%d]: %s; %s",
		e.File, e.Line, e.Column, e.EntryIndex, e.Reason, e.Suggestion)
}

// ErrUnboundParameter indicates that one or more {placeholder} tokens in a
// step's text reference parameters not declared in the Parameters: block.
type ErrUnboundParameter struct {
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
func (e *ErrUnboundParameter) Error() string {
	sorted := make([]string, len(e.Parameters))
	copy(sorted, e.Parameters)
	sort.Strings(sorted)

	return fmt.Sprintf("%s:%d:%d: unbound parameter(s) {%s} in step %q; %s",
		e.File, e.Line, e.Column,
		strings.Join(sorted, "}, {"),
		e.StepText, e.Suggestion)
}
