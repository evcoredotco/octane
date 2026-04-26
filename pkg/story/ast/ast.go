// Package ast defines the typed abstract syntax tree produced by parsing
// .story files. Every exported type uses ordered slices rather than maps
// to guarantee deterministic serialization (constitution principle IV).
package ast

import "time"

// Position records a source location within a .story file.
type Position struct {
	// Line is the 1-based line number.
	Line int
	// Column is the 1-based column number (byte offset from line start).
	Column int
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
	default:
		return "Unknown"
	}
}

// Step represents a single step line within a Background, Setup, Scenario,
// or Teardown section. The Text field preserves {placeholder} tokens
// verbatim; parameter resolution happens at runtime, not at parse time.
type Step struct {
	// Kind is the step keyword (Given, When, Then, And, But).
	Kind StepKind
	// Text is the verbatim step text after the keyword, with
	// {placeholder} tokens preserved.
	Text string
	// Position is the source location of the step keyword.
	Position Position
}

// Meta holds the structured header of a .story file. All collection fields
// use slices to preserve insertion order and guarantee deterministic
// serialization.
type Meta struct {
	// Name is the human-readable test name (required).
	Name string
	// Id is the stable snake_case identifier used in Depends references
	// (required).
	Id string
	// SpecRef is the OCPP specification section reference. It is nil for
	// helper stories (tagged "helper") and required for conformance stories.
	SpecRef *string
	// Tags classifies the story. At least one tag is required. The
	// "helper" tag is structural: it toggles the SpecRef requirement.
	Tags []string
	// Stations is the declared station count for preflight resource
	// allocation (required, >= 1).
	Stations int
	// Timeout is the default per-step timeout. Zero means use the
	// global default from configuration.
	Timeout time.Duration
	// Parameters declares story inputs that resolve from octane.yml.
	// Empty when the story takes no parameters.
	Parameters []string
	// CacheTTL overrides the default cache time-to-live for this story's
	// results. Nil means use the default (1h for helpers, infinite for
	// conformance tests).
	CacheTTL *time.Duration
	// Depends lists the prerequisite stories that must run before this
	// story. Populated from the Depends: YAML block in the Meta section.
	Depends []Dependency
	// Position is the source location of the Meta section keyword.
	Position Position
}

// Scenario is a named step container within a story. A story has one or
// more scenarios.
type Scenario struct {
	// Name is the scenario title text after the "Scenario:" keyword.
	Name string
	// Steps are the ordered steps within this scenario.
	Steps []Step
	// Position is the source location of the "Scenario:" keyword.
	Position Position
}

// Story is the root AST node produced by parsing a single .story file.
// Field order mirrors the required section order in the grammar:
// Meta, Background, Setup, Scenarios, Teardown.
type Story struct {
	// Path is the filesystem path of the source .story file.
	Path string
	// Meta is the structured header section.
	Meta Meta
	// Background holds the optional Background steps executed before
	// every scenario.
	Background []Step
	// Setup holds the optional Setup steps executed once before the
	// first scenario.
	Setup []Step
	// Scenarios holds the one or more named scenarios in declaration
	// order.
	Scenarios []Scenario
	// Teardown holds the optional Teardown steps executed after the
	// last scenario.
	Teardown []Step
	// Position is the source location of the start of the file.
	Position Position
}
