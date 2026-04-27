// Package primitive_test — domain-vs-primitive precedence tests (T-004-32).
//
// Validates spec 004 AC7: when a domain keyword and a primitive keyword share
// the same step pattern in a story declaring OCPP 1.6, the domain keyword
// wins.  When the resolver runs for OCPP 2.0.1 and the domain keyword is
// OCPP-1.6-scoped, the primitive wins as a fallback.
//
// Strategy: register a fixture domain keyword whose pattern is unique enough
// that it cannot collide with any production primitive (the pattern literal
// "fixture: domain keyword beats primitive for ocpp16 step {n:int}" is not
// used by any production keyword).  Alongside it, register a primitive keyword
// with the same pattern.  This avoids calling the unexported reset() function
// from the registry package while keeping the test deterministic and parallel.
package primitive_test

import (
	"context"
	"testing"

	"github.com/octane-project/octane/pkg/keywords/api"
	// Blank import ensures all production primitives are registered at
	// init() time before the test-local keywords are registered.
	_ "github.com/octane-project/octane/pkg/keywords/primitive"
	"github.com/octane-project/octane/pkg/keywords/registry"
)

// ── Named constants ───────────────────────────────────────────────────────────

const (
	// patternPrecedenceFixture is the shared step pattern used for both the
	// fixture domain keyword and the fixture primitive keyword.  The prefix
	// "fixture:" makes it globally unique so it cannot collide with any
	// production-registered keyword and avoids needing registry.reset().
	patternPrecedenceFixture = "fixture: domain keyword beats primitive for ocpp16 step {n:int}"

	// valueFixtureN is the int value bound to the {n:int} placeholder in the
	// step text exercised by the precedence tests.
	valueFixtureN = 42

	// stepPrecedenceFixture is the concrete step text that resolves against
	// patternPrecedenceFixture.
	stepPrecedenceFixture = "fixture: domain keyword beats primitive for ocpp16 step 42"
)

// ── init: register fixture keywords ──────────────────────────────────────────

// fixtureNoopFunc is the Func shared by all fixture keywords.  Its body is
// intentionally empty; the test only cares about which keyword the resolver
// selects, not what the keyword does.
func fixtureNoopFunc(_ context.Context, _ api.State, _ api.Args) error {
	return nil
}

func init() {
	// Register the fixture primitive keyword.  This runs once at package
	// init time, before any test function executes.
	registry.Register(api.Keyword{
		Pattern:     patternPrecedenceFixture,
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        fixtureNoopFunc,
	})

	// Register the fixture domain keyword scoped to OCPP 1.6.  By ADR 0007,
	// the domain layer (value=2) beats the primitive layer (value=1) for the
	// same pattern when the story declares OCPP 1.6.
	registry.Register(api.Keyword{
		Pattern:     patternPrecedenceFixture,
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        fixtureNoopFunc,
	})
}

// ── Tests ─────────────────────────────────────────────────────────────────────

// Test_primitive_precedence_domainWinsForOCPP16 verifies that when the
// resolver runs for OCPP 1.6 and both a domain keyword (OCPP16-scoped) and a
// primitive keyword share the same step pattern, the domain keyword wins
// (spec 004 AC7, ADR 0007).
func Test_primitive_precedence_domainWinsForOCPP16(t *testing.T) {
	t.Parallel()

	// Invariant: Resolve(OCPP16) selects the domain-layer keyword when both a
	// domain keyword and a primitive keyword match the same step text (AC7).
	match, err := registry.Resolve(stepPrecedenceFixture, api.OCPP16)
	if err != nil {
		t.Fatalf("registry.Resolve: unexpected error: %v", err)
	}

	if match.Keyword.Layer != api.LayerDomain {
		t.Errorf(
			"Match.Keyword.Layer: want LayerDomain (domain wins over primitive), got %v",
			match.Keyword.Layer,
		)
	}

	if match.Keyword.OCPPVersion != api.OCPP16 {
		t.Errorf(
			"Match.Keyword.OCPPVersion: want OCPP16, got %v",
			match.Keyword.OCPPVersion,
		)
	}
}

// Test_primitive_precedence_primitiveWinsForOCPP201 verifies that when the
// resolver runs for OCPP 2.0.1 and the only domain keyword sharing the
// pattern is OCPP-1.6-scoped, that domain keyword is excluded and the
// primitive-layer fallback wins (spec 004 AC7, ADR 0007).
func Test_primitive_precedence_primitiveWinsForOCPP201(t *testing.T) {
	t.Parallel()

	// Invariant: Resolve(OCPP201) excludes the OCPP16-scoped domain keyword
	// and falls back to the primitive keyword (AC7).
	match, err := registry.Resolve(stepPrecedenceFixture, api.OCPP201)
	if err != nil {
		t.Fatalf("registry.Resolve: unexpected error: %v", err)
	}

	if match.Keyword.Layer != api.LayerPrimitive {
		t.Errorf(
			"Match.Keyword.Layer: want LayerPrimitive (OCPP16 domain excluded for OCPP201), got %v",
			match.Keyword.Layer,
		)
	}
}

// Test_primitive_precedence_argsCorrectlyBound verifies that the matched
// keyword (regardless of layer) correctly binds the {n:int} placeholder from
// the step text, confirming the pattern match and coercion path is healthy.
func Test_primitive_precedence_argsCorrectlyBound(t *testing.T) {
	t.Parallel()

	// Invariant: the {n:int} placeholder is bound to 42 in both resolution
	// paths (domain for OCPP16, primitive for OCPP201).
	match, err := registry.Resolve(stepPrecedenceFixture, api.OCPP16)
	if err != nil {
		t.Fatalf("registry.Resolve: unexpected error: %v", err)
	}

	gotN := match.Args.Int("n")
	if gotN != valueFixtureN {
		t.Errorf(
			"Args.Int(%q): want %d, got %d",
			"n",
			valueFixtureN,
			gotN,
		)
	}
}

// Test_primitive_precedence_domainPatternString verifies that the domain
// keyword's registered pattern string is preserved verbatim through the
// resolution path — ensuring no silent truncation or modification occurs.
func Test_primitive_precedence_domainPatternString(t *testing.T) {
	t.Parallel()

	// Invariant: the returned Match carries the exact registered pattern string.
	match, err := registry.Resolve(stepPrecedenceFixture, api.OCPP16)
	if err != nil {
		t.Fatalf("registry.Resolve: unexpected error: %v", err)
	}

	if match.Keyword.Pattern != patternPrecedenceFixture {
		t.Errorf(
			"Match.Keyword.Pattern: want %q, got %q",
			patternPrecedenceFixture,
			match.Keyword.Pattern,
		)
	}
}
