// Package registry exercises collision detection in the global keyword
// registry. Tests run in the same package (white-box) so that reset()
// can be called to isolate each scenario.
//
// Task: T-003-23
// AC2: Given two keyword registrations with the same (Layer, OCPPVersion,
// Pattern) tuple, when the second Register call executes, then the program
// panics with a message naming both registration sites.

package registry

import (
	"context"
	"strings"
	"testing"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// ── Named test-value constants ───────────────────────────────────────────────

const (
	// patternCollision is the shared pattern used in collision scenarios.
	patternCollision = "station {s:station} sends BootNotification"

	// patternAlt is a distinct pattern that must not collide.
	patternAlt = "station {s:station} sends Heartbeat"
)

// ── helpers ──────────────────────────────────────────────────────────────────

// noopFunc is a minimal keyword Func that satisfies the api.Func signature
// without performing any action. It is used wherever a real implementation
// is not relevant to the invariant under test.
func noopFunc(_ context.Context, _ api.State, _ api.Args) error {
	return nil
}

// mustPanic calls callFunc and returns the recovered panic value as a string.
// If callFunc does not panic, the test is failed immediately.
func mustPanic(t *testing.T, callFunc func()) string {
	t.Helper()

	var recovered any

	var panicked bool

	func() {
		defer func() {
			if r := recover(); r != nil {
				recovered = r
				panicked = true
			}
		}()

		callFunc()
	}()

	didPanic := panicked

	if !didPanic {
		t.Fatal("expected Register to panic on collision, but it did not")
	}

	msg, ok := recovered.(string)
	if !ok {
		t.Fatalf(
			"panic value type: want string, got %T (%v)",
			recovered,
			recovered,
		)
	}

	return msg
}

// ── Collision tests ──────────────────────────────────────────────────────────

// Test_registry_Register_collisionPanics verifies that registering two
// keywords with identical (Layer, OCPPVersion, Pattern) causes a panic.
// Tests in this file mutate the global registry and must NOT call
// t.Parallel() to prevent interference between test cases.
func Test_registry_Register_collisionPanics(t *testing.T) {
	// Invariant: second Register on same (Pattern, Layer, OCPPVersion) panics.
	reset()

	Register(api.Keyword{
		Pattern:     patternCollision,
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        noopFunc,
	})

	_ = mustPanic(t, func() {
		Register(api.Keyword{
			Pattern:     patternCollision,
			Layer:       api.LayerDomain,
			OCPPVersion: api.OCPP16,
			Func:        noopFunc,
		})
	})
}

// Test_registry_Register_collisionPanicNamesOriginalSite verifies that the
// panic message contains a non-empty "existing registrant at" location,
// which corresponds to the first Register call site (the original registrant).
func Test_registry_Register_collisionPanicNamesOriginalSite(t *testing.T) {
	// Invariant: panic message references the original registrant's call site.
	reset()

	Register(api.Keyword{
		Pattern:     patternCollision,
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        noopFunc,
	})

	msg := mustPanic(t, func() {
		Register(api.Keyword{
			Pattern:     patternCollision,
			Layer:       api.LayerDomain,
			OCPPVersion: api.OCPP16,
			Func:        noopFunc,
		})
	})

	const wantFragment = "existing registrant at "

	if !strings.Contains(msg, wantFragment) {
		t.Errorf(
			"panic message %q: want substring %q",
			msg,
			wantFragment,
		)
	}
}

// Test_registry_Register_collisionPanicNamesNewSite verifies that the
// panic message contains a non-empty "new registrant at" location,
// which corresponds to the second (colliding) Register call site.
func Test_registry_Register_collisionPanicNamesNewSite(t *testing.T) {
	// Invariant: panic message references the new (duplicate) registrant's
	// call site.
	reset()

	Register(api.Keyword{
		Pattern:     patternCollision,
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        noopFunc,
	})

	msg := mustPanic(t, func() {
		Register(api.Keyword{
			Pattern:     patternCollision,
			Layer:       api.LayerDomain,
			OCPPVersion: api.OCPP16,
			Func:        noopFunc,
		})
	})

	const wantFragment = "new registrant at "

	if !strings.Contains(msg, wantFragment) {
		t.Errorf(
			"panic message %q: want substring %q",
			msg,
			wantFragment,
		)
	}
}

// Test_registry_Register_collisionPanicMessageContainsBothSites verifies
// that the panic message produced by a duplicate registration names BOTH
// the original registrant's location and the new registrant's location,
// satisfying AC2's "naming both registration sites" requirement.
func Test_registry_Register_collisionPanicMessageContainsBothSites(
	t *testing.T,
) {
	// Invariant: panic message carries both call-site strings simultaneously.
	reset()

	Register(api.Keyword{
		Pattern:     patternCollision,
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        noopFunc,
	})

	msg := mustPanic(t, func() {
		Register(api.Keyword{
			Pattern:     patternCollision,
			Layer:       api.LayerDomain,
			OCPPVersion: api.OCPP16,
			Func:        noopFunc,
		})
	})

	if !strings.Contains(msg, "existing registrant at ") {
		t.Errorf("panic message missing original registrant location: %q", msg)
	}

	if !strings.Contains(msg, "new registrant at ") {
		t.Errorf("panic message missing new registrant location: %q", msg)
	}

	// Both "existing registrant at" and "new registrant at" must appear;
	// confirm the overall format anchors on "registry: keyword collision".
	if !strings.HasPrefix(msg, "registry: keyword collision") {
		t.Errorf(
			"panic message %q: want prefix %q",
			msg,
			"registry: keyword collision",
		)
	}
}

// Test_registry_Register_differentLayerSamePatternDoesNotPanic verifies
// that two keywords sharing the same Pattern and OCPPVersion but different
// Layers are NOT a collision and both register successfully.
func Test_registry_Register_differentLayerSamePatternDoesNotPanic(
	t *testing.T,
) {
	// Invariant: (Pattern, Layer=Primitive, OCPPVersion) and
	// (Pattern, Layer=Domain, OCPPVersion) are distinct keys — no panic.
	reset()

	// First registration: primitive layer.
	Register(api.Keyword{
		Pattern:     patternAlt,
		Layer:       api.LayerPrimitive,
		OCPPVersion: api.OCPP16,
		Func:        noopFunc,
	})

	// Second registration: domain layer — must NOT panic.
	Register(api.Keyword{
		Pattern:     patternAlt,
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        noopFunc,
	})

	keywords := All()

	const wantCount = 2

	if len(keywords) != wantCount {
		t.Errorf(
			"All() count: want %d (both layers registered), got %d",
			wantCount,
			len(keywords),
		)
	}
}
