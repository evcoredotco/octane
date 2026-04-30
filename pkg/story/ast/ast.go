// Package ast defines the typed abstract syntax tree produced by parsing
// .story files. Every exported type uses ordered slices rather than maps
// to guarantee deterministic serialization (constitution principle IV).
package ast

import "time"

// Position records a source location within a .story file.
type Position struct {
	// Line is the 1-based line number.
	Line int `json:"line"`
	// Column is the 1-based column number (byte offset from line start).
	Column int `json:"column"`
}

// StepKind identifies the keyword that introduces a step.
type StepKind int

const (
	// StepGiven introduces a precondition.
	StepGiven StepKind = iota + 1
	// StepWhen introduces an action.
	StepWhen
	// StepThen introduces an expected outcome.
	StepThen
	// StepAnd continues the preceding step kind.
	StepAnd
	// StepBut introduces a negative continuation of the preceding step kind.
	StepBut
	// StepAction represents a bare action line with no Gherkin keyword
	// prefix. Used in Teardown sections and Parallel blocks where bare
	// command verbs are conventional (e.g. "Disconnect station X").
	StepAction
)

// String returns the keyword text for a StepKind.
func (k StepKind) String() string {
	switch k {
	case StepGiven:
		return "Given"
	case StepWhen:
		return "When"
	case StepThen:
		return "Then"
	case StepAnd:
		return "And"
	case StepBut:
		return "But"
	case StepAction:
		return "Action"
	default:
		return "Unknown"
	}
}

// Step represents a single step line within a Background, Setup, Scenario,
// or Teardown section. The Text field preserves {placeholder} tokens
// verbatim; parameter resolution happens at runtime, not at parse time.
type Step struct {
	// Kind is the step keyword (Given, When, Then, And, But).
	Kind StepKind `json:"kind"`
	// Text is the verbatim step text after the keyword, with
	// {placeholder} tokens preserved.
	Text string `json:"text"`
	// Position is the source location of the step keyword.
	Position Position `json:"position"`
}

// Meta holds the structured header of a .story file. All collection fields
// use slices to preserve insertion order and guarantee deterministic
// serialization.
type Meta struct {
	// Name is the human-readable test name (required).
	Name string `json:"name"`
	// ID is the stable snake_case identifier used in Depends references
	// (required).
	ID string `json:"id"`
	// SpecRef is the OCPP specification section reference. It is nil for
	// helper stories (tagged "helper") and required for conformance stories.
	SpecRef *string `json:"specRef,omitempty"`
	// Tags classifies the story. At least one tag is required. The
	// "helper" tag is structural: it toggles the SpecRef requirement.
	Tags []string `json:"tags"`
	// Stations is the declared station count for preflight resource
	// allocation (required, >= 1).
	Stations int `json:"stations"`
	// Timeout is the default per-step timeout. Zero means use the
	// global default from configuration.
	Timeout time.Duration `json:"timeout"`
	// Parameters declares story inputs that resolve from octane.yml.
	// Empty when the story takes no parameters.
	Parameters []string `json:"parameters"`
	// CacheTTL overrides the default cache time-to-live for this story's
	// results. Nil means use the default (1h for helpers, infinite for
	// conformance tests).
	CacheTTL *time.Duration `json:"cacheTtl,omitempty"`
	// Depends lists the prerequisite stories that must run before this
	// story. Populated from the Depends: YAML block in the Meta section.
	Depends []Dependency `json:"depends"`
	// Position is the source location of the Meta section keyword.
	Position Position `json:"position"`
}

// Scenario is a named step container within a story. A story has one or
// more scenarios.
type Scenario struct {
	// Name is the scenario title text after the "Scenario:" keyword.
	Name string `json:"name"`
	// Steps are the ordered steps within this scenario.
	Steps []Step `json:"steps"`
	// Position is the source location of the "Scenario:" keyword.
	Position Position `json:"position"`
}

// Story is the root AST node produced by parsing a single .story file.
// Field order mirrors the required section order in the grammar:
// Meta, Background, Setup, Scenarios, Teardown.
type Story struct {
	// Path is the filesystem path of the source .story file.
	Path string `json:"path"`
	// Meta is the structured header section.
	Meta Meta `json:"meta"`
	// Background holds the optional Background steps executed before
	// every scenario.
	Background []Step `json:"background"`
	// Setup holds the optional Setup steps executed once before the
	// first scenario.
	Setup []Step `json:"setup"`
	// Scenarios holds the one or more named scenarios in declaration
	// order.
	Scenarios []Scenario `json:"scenarios"`
	// Teardown holds the optional Teardown steps executed after the
	// last scenario.
	Teardown []Step `json:"teardown"`
	// Position is the source location of the start of the file.
	Position Position `json:"position"`
}
