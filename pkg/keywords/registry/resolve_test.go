// Package registry — white-box resolver unit tests (T-003-33).
//
// Covers AC3, AC4, AC6, AC7 and the related resolution paths:
//   - Happy path: primitive keyword resolved by step text.
//   - Happy path: domain keyword resolved for matching OCPPVersion.
//   - Domain wins over primitive for same pattern (AC6).
//   - Domain keyword with OCPPVersion=0 (version-agnostic) matches all
//     versions.
//   - No match → *NoMatchError returned; Closest populated when near
//     pattern exists (AC4).
//   - Type coercion failure → *TypeMismatchError returned (AC5).
//   - Multiple placeholders resolved correctly into api.Args (AC3).
//
// Tests use package registry (white-box) to access the unexported reset().
// Tests that mutate the global registry must NOT call t.Parallel() to
// prevent interference. Tests that construct local state only may parallel.

package registry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// ── Named test-value constants ──────────────────────────────────────────

const (
	// patternPrimitive is the step pattern for the primitive-layer keyword
	// used across resolution tests.
	patternPrimitive = "wait {d:duration}"

	// patternDomain16 is the step pattern shared by domain (OCPP 1.6) and
	// primitive keywords to exercise layer precedence (AC6, AC7).
	patternDomain16 = "station {s:station} sends BootNotification"

	// patternMultiPlaceholder is the step pattern for the multi-placeholder
	// happy-path test (AC3).
	patternMultiPlaceholder = "the CSMS sends ReserveNow with connectorId" +
		" {connectorId:int} and idTag {idTag:string}" +
		" to station {station:station} within {timeout:duration}"

	// patternIntType is a pattern with a single {n:int} placeholder used
	// to trigger coercion failure (AC5).
	patternIntType = "count is {n:int}"

	// patternNearMiss is a literal-only pattern whose Levenshtein distance
	// from stepNearMissInput is within the suggestion threshold (AC4).
	// Using a literal-only pattern avoids accidental structural matches due
	// to valid placeholder token values.
	patternNearMiss = "open connection"

	// stepPrimitive is the step text that resolves against patternPrimitive.
	stepPrimitive = "wait 30s"

	// stepDomain16 is the step text that resolves against patternDomain16.
	stepDomain16 = "station CP01 sends BootNotification"

	// stepMultiPlaceholderUnquoted is the step text for the multi-placeholder
	// test (AC3). Values are unquoted so the whitespace-delimited matcher
	// captures each token without embedded quote characters.
	stepMultiPlaceholderUnquoted = "the CSMS sends ReserveNow with connectorId 1" +
		" and idTag X to station CP01 within 30s"

	// stepIntTypeGood is a step text that satisfies the int placeholder.
	stepIntTypeGood = "count is 7"

	// stepIntTypeBad is a step text that supplies a non-integer to
	// an int placeholder.
	stepIntTypeBad = "count is abc"

	// stepUnregistered is a step text with no registered pattern.
	stepUnregistered = "this step text matches nothing at all ever"

	// stepNearMissInput is a step text that does not structurally match any
	// registered pattern but is within Levenshtein distance 5 of
	// patternNearMiss ("open connexion" vs "open connection" = distance 2).
	stepNearMissInput = "open connexion"

	// valueDuration30s is the expected time.Duration value for "30s".
	valueDuration30s = 30 * time.Second

	// valueConnectorIDOne is the expected int for connectorId in the
	// multi-placeholder test.
	valueConnectorIDOne = 1

	// valueSeven is the expected int value for stepIntTypeGood.
	valueSeven = 7

	// argNameN is the placeholder name used in patternIntType.
	argNameN = "n"

	// typeNameInt is the expected type string for int placeholders.
	typeNameInt = "int"

	// emptyClosest is the empty string used for NoMatchError.Closest.
	emptyClosest = ""

	// fmtResolveUnexpectedErr is the format string for unexpected Resolve errors.
	fmtResolveUnexpectedErr = "Resolve: unexpected error: %v"

	// fmtResolveErrType is the format string for unexpected error types from
	// Resolve when a NoMatchError is expected.
	fmtResolveErrType = "Resolve error type: want *NoMatchError, got %T: %v"
)

// ── helpers ─────────────────────────────────────────────────────────────

// resolveNoopFunc is a minimal keyword Func used wherever the function body
// is irrelevant to the invariant under test.
//
// Note: collision_test.go defines noopFunc in this package. We use a distinct
// name to avoid a redeclaration compile error.
func resolveNoopFunc(_ context.Context, _ api.State, _ api.Args) error {
	return nil
}

// registerPrimitive registers a primitive-layer, version-agnostic keyword
// with the given pattern and returns the registered api.Keyword for assertions.
func registerPrimitive(pattern string) api.Keyword {
	registered := api.Keyword{
		Pattern:     pattern,
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        resolveNoopFunc,
	}

	Register(registered)

	return registered
}

// registerDomain16 registers a domain-layer keyword using patternDomain16
// for the given OCPP version and returns the registered api.Keyword.
// All tests that exercise domain-layer precedence use this pattern.
func registerDomain16(version api.OCPPVersion) api.Keyword {
	registered := api.Keyword{
		Pattern:     patternDomain16,
		Layer:       api.LayerDomain,
		OCPPVersion: version,
		Func:        resolveNoopFunc,
	}

	Register(registered)

	return registered
}

// ── Happy path: primitive keyword ─────────────────────────────────────────────

// Test_registry_Resolve_primitiveKeywordResolvesByStepText verifies that a
// primitive-layer keyword is matched when the step text satisfies its pattern
// and a Duration argument is correctly bound (AC3 primitive path).
func Test_registry_Resolve_primitiveKeywordResolvesByStepText(t *testing.T) {
	t.Parallel()

	// Invariant: Resolve returns the primitive keyword and correct Args for a
	// step text that matches its pattern.
	reset()

	registered := registerPrimitive(patternPrimitive)

	match, err := Resolve(stepPrimitive, api.OCPP16)
	if err != nil {
		t.Fatalf(fmtResolveUnexpectedErr, err)
	}

	if match.Keyword.Pattern != registered.Pattern {
		t.Errorf(
			"Match.Keyword.Pattern: want %q, got %q",
			registered.Pattern,
			match.Keyword.Pattern,
		)
	}

	if match.Keyword.Layer != api.LayerPrimitive {
		t.Errorf(
			"Match.Keyword.Layer: want LayerPrimitive, got %v",
			match.Keyword.Layer,
		)
	}

	// Invariant: the duration placeholder is coerced to 30 * time.Second.
	gotDuration := match.Args.Duration("d")
	if gotDuration != valueDuration30s {
		t.Errorf(
			"Args.Duration(%q): want %v, got %v",
			"d",
			valueDuration30s,
			gotDuration,
		)
	}
}

// ── Happy path: domain keyword for matching OCPPVersion ───────────────────────

// Test_registry_Resolve_domainKeywordResolvesForMatchingVersion verifies that
// a domain-layer keyword registered for OCPP 1.6 is matched when the resolver
// is called with OCPP 1.6.
func Test_registry_Resolve_domainKeywordResolvesForMatchingVersion(
	t *testing.T,
) {
	t.Parallel()

	// Invariant: Resolve returns the domain keyword when OCPPVersion matches.
	reset()

	registered := registerDomain16(api.OCPP16)

	match, err := Resolve(stepDomain16, api.OCPP16)
	if err != nil {
		t.Fatalf(fmtResolveUnexpectedErr, err)
	}

	if match.Keyword.Pattern != registered.Pattern {
		t.Errorf(
			"Match.Keyword.Pattern: want %q, got %q",
			registered.Pattern,
			match.Keyword.Pattern,
		)
	}

	if match.Keyword.Layer != api.LayerDomain {
		t.Errorf(
			"Match.Keyword.Layer: want LayerDomain, got %v",
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

// ── AC6: domain wins over primitive for same step text ────────────────────────

// Test_registry_Resolve_domainWinsOverPrimitiveForSamePattern verifies that
// when both a domain-layer and a primitive-layer keyword share the same
// pattern, the domain keyword wins for the matching OCPP version (AC6).
func Test_registry_Resolve_domainWinsOverPrimitiveForSamePattern(t *testing.T) {
	t.Parallel()

	// Invariant: domain layer keyword takes precedence over primitive layer
	// keyword for the same pattern when OCPP version matches (AC6).
	reset()

	registerPrimitive(patternDomain16)
	registerDomain16(api.OCPP16)

	match, err := Resolve(stepDomain16, api.OCPP16)
	if err != nil {
		t.Fatalf(fmtResolveUnexpectedErr, err)
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

// ── Domain keyword with OCPPVersion=0 matches all versions ───────────────────

// Test_registry_Resolve_domainVersionAgnosticMatchesOCPP16 verifies that a
// domain-layer keyword registered with OCPPVersion=0 (version-agnostic) is
// matched when the resolver runs with OCPP 1.6.
func Test_registry_Resolve_domainVersionAgnosticMatchesOCPP16(t *testing.T) {
	t.Parallel()

	// Invariant: a domain keyword with OCPPVersion=0 is eligible for any
	// OCPP version — it is version-agnostic.
	reset()

	const versionAgnostic api.OCPPVersion = 0

	versionAgnosticKeyword := api.Keyword{
		Pattern:     patternDomain16,
		Layer:       api.LayerDomain,
		OCPPVersion: versionAgnostic,
		Func:        resolveNoopFunc,
	}

	Register(versionAgnosticKeyword)

	match, err := Resolve(stepDomain16, api.OCPP16)
	if err != nil {
		t.Fatalf(fmtResolveUnexpectedErr, err)
	}

	if match.Keyword.Layer != api.LayerDomain {
		t.Errorf(
			"Match.Keyword.Layer: want LayerDomain, got %v",
			match.Keyword.Layer,
		)
	}

	if match.Keyword.OCPPVersion != versionAgnostic {
		t.Errorf(
			"Match.Keyword.OCPPVersion: want 0 (version-agnostic), got %v",
			match.Keyword.OCPPVersion,
		)
	}
}

// ── AC4: no match → NoMatchError; Closest populated when near pattern exists ───

// Test_registry_Resolve_noMatchReturnsNoMatchError verifies that Resolve returns
// *NoMatchError when the step text matches no registered pattern (AC4).
func Test_registry_Resolve_noMatchReturnsNoMatchError(t *testing.T) {
	t.Parallel()

	// Invariant: Resolve wraps the unmatched step in *NoMatchError (AC4).
	reset()

	registerPrimitive(patternPrimitive)

	_, err := Resolve(stepUnregistered, api.OCPP16)
	if err == nil {
		t.Fatal("Resolve: expected *NoMatchError, got nil")
	}

	var noMatch *NoMatchError
	if !errors.As(err, &noMatch) {
		t.Fatalf(fmtResolveErrType, err, err)
	}
}

// Test_registry_Resolve_noMatchStepTextPreserved verifies that NoMatchError
// carries the original unmatched step text verbatim (AC4).
func Test_registry_Resolve_noMatchStepTextPreserved(t *testing.T) {
	t.Parallel()

	// Invariant: NoMatchError.StepText equals the input step string.
	reset()

	registerPrimitive(patternPrimitive)

	_, err := Resolve(stepUnregistered, api.OCPP16)

	var noMatch *NoMatchError
	if !errors.As(err, &noMatch) {
		t.Fatalf(fmtResolveErrType, err, err)
	}

	if noMatch.StepText != stepUnregistered {
		t.Errorf(
			"NoMatchError.StepText: want %q, got %q",
			stepUnregistered,
			noMatch.StepText,
		)
	}
}

// Test_registry_Resolve_noMatchClosestPopulatedWhenNearPatternExists verifies
// that NoMatchError.Closest is non-empty when a registered pattern is within
// Levenshtein distance 5 of the failed step text (AC4).
func Test_registry_Resolve_noMatchClosestPopulatedWhenNearPatternExists(
	t *testing.T,
) {
	t.Parallel()

	// Invariant: NoMatchError.Closest carries the near pattern when within
	// edit distance 5 (AC4).
	reset()

	// Register the near-miss pattern. "open connexion" vs "open connection"
	// has a Levenshtein distance of 2 — well within the threshold of 5.
	// The step text does not structurally match because "connexion" != "connection".
	registerPrimitive(patternNearMiss)

	_, err := Resolve(stepNearMissInput, api.OCPP16)

	var noMatch *NoMatchError
	if !errors.As(err, &noMatch) {
		t.Fatalf(fmtResolveErrType, err, err)
	}

	if noMatch.Closest == "" {
		t.Errorf(
			"NoMatchError.Closest: want non-empty suggestion for near-miss step %q against pattern %q",
			stepNearMissInput,
			patternNearMiss,
		)
	}
}

// Test_registry_Resolve_noMatchClosestEmptyWhenNoNearPattern verifies that
// NoMatchError.Closest is empty when no registered pattern is within
// Levenshtein distance 5 of the failed step text (AC4).
func Test_registry_Resolve_noMatchClosestEmptyWhenNoNearPattern(t *testing.T) {
	t.Parallel()

	// Invariant: NoMatchError.Closest is empty when no pattern is within
	// edit distance 5 (AC4).
	reset()

	// Register a pattern that is very far from stepUnregistered.
	registerPrimitive(patternPrimitive)

	_, err := Resolve(stepUnregistered, api.OCPP16)

	var noMatch *NoMatchError
	if !errors.As(err, &noMatch) {
		t.Fatalf(fmtResolveErrType, err, err)
	}

	if noMatch.Closest != "" {
		t.Errorf(
			"NoMatchError.Closest: want empty (no near pattern), got %q",
			noMatch.Closest,
		)
	}
}

// ── AC5: type coercion failure → TypeMismatchError ──────────────────────────────

// Test_registry_Resolve_typeMismatchReturnedForBadIntToken verifies that
// Resolve returns *TypeMismatchError when the step text supplies a non-integer
// token for an {n:int} placeholder (AC5).
func Test_registry_Resolve_typeMismatchReturnedForBadIntToken(t *testing.T) {
	t.Parallel()

	// Invariant: a non-integer token for an int placeholder causes
	// *TypeMismatchError with the correct ArgName, Expected, and Got (AC5).
	reset()

	registerPrimitive(patternIntType)

	_, err := Resolve(stepIntTypeBad, api.OCPP16)
	if err == nil {
		t.Fatal("Resolve: expected *TypeMismatchError, got nil")
	}

	var mismatch *TypeMismatchError
	if !errors.As(err, &mismatch) {
		t.Fatalf(
			"Resolve error type: want *TypeMismatchError, got %T: %v",
			err,
			err,
		)
	}

	if mismatch.ArgName != argNameN {
		t.Errorf(
			"TypeMismatchError.ArgName: want %q, got %q",
			argNameN,
			mismatch.ArgName,
		)
	}

	if mismatch.Expected != typeNameInt {
		t.Errorf(
			"TypeMismatchError.Expected: want %q, got %q",
			typeNameInt,
			mismatch.Expected,
		)
	}

	if mismatch.Got != "abc" {
		t.Errorf("TypeMismatchError.Got: want %q, got %q", "abc", mismatch.Got)
	}
}

// Test_registry_Resolve_goodIntTokenResolves verifies that the same int
// placeholder resolves correctly when the step text supplies a valid integer.
func Test_registry_Resolve_goodIntTokenResolves(t *testing.T) {
	t.Parallel()

	// Invariant: a valid integer token for an int placeholder resolves
	// without error and the bound Args value is correct.
	reset()

	registerPrimitive(patternIntType)

	match, err := Resolve(stepIntTypeGood, api.OCPP16)
	if err != nil {
		t.Fatalf(fmtResolveUnexpectedErr, err)
	}

	gotInt := match.Args.Int(argNameN)
	if gotInt != valueSeven {
		t.Errorf("Args.Int(%q): want %d, got %d", argNameN, valueSeven, gotInt)
	}
}

// ── AC3: multi-placeholder step resolves into correctly-bound Args ────────────

// Test_registry_Resolve_multiPlaceholderStepBindsAllArgs verifies that a
// step with four placeholders (int, string, station, duration) is correctly
// resolved and each placeholder value is accessible by name from Args (AC3).
func Test_registry_Resolve_multiPlaceholderStepBindsAllArgs(t *testing.T) {
	t.Parallel()

	// Invariant: Resolve correctly binds all four named placeholders from the
	// step text into the returned Args (AC3).
	reset()

	multiKeyword := api.Keyword{
		Pattern:     patternMultiPlaceholder,
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        resolveNoopFunc,
	}

	Register(multiKeyword)

	match, err := Resolve(stepMultiPlaceholderUnquoted, api.OCPP16)
	if err != nil {
		t.Fatalf(fmtResolveUnexpectedErr, err)
	}

	// Invariant: connectorId is bound as int 1.
	gotConnectorID := match.Args.Int("connectorId")
	if gotConnectorID != valueConnectorIDOne {
		t.Errorf(
			"Args.Int(%q): want %d, got %d",
			"connectorId",
			valueConnectorIDOne,
			gotConnectorID,
		)
	}

	// Invariant: idTag is bound as string "X".
	gotIDTag := match.Args.String("idTag")
	if gotIDTag != "X" {
		t.Errorf("Args.String(%q): want %q, got %q", "idTag", "X", gotIDTag)
	}

	// Invariant: station is bound as string "CP01" (station type stores as string).
	gotStation := match.Args.Station("station")
	if gotStation != "CP01" {
		t.Errorf(
			"Args.Station(%q): want %q, got %q",
			"station",
			"CP01",
			gotStation,
		)
	}

	// Invariant: timeout is bound as 30s duration.
	gotTimeout := match.Args.Duration("timeout")
	if gotTimeout != valueDuration30s {
		t.Errorf(
			"Args.Duration(%q): want %v, got %v",
			"timeout",
			valueDuration30s,
			gotTimeout,
		)
	}

	// Invariant: all four placeholders are present.
	const wantArgCount = 4
	if match.Args.Len() != wantArgCount {
		t.Errorf("Args.Len(): want %d, got %d", wantArgCount, match.Args.Len())
	}
}

// ── Eligibility: empty registry ───────────────────────────────────────────────

// Test_registry_Resolve_emptyRegistryReturnsNoMatchError verifies that Resolve
// against an empty registry always returns NoMatchError.
func Test_registry_Resolve_emptyRegistryReturnsNoMatchError(t *testing.T) {
	t.Parallel()

	// Invariant: an empty registry produces NoMatchError for any step text.
	reset()

	_, err := Resolve(stepPrimitive, api.OCPP16)
	if err == nil {
		t.Fatal("Resolve on empty registry: expected *NoMatchError, got nil")
	}

	var noMatch *NoMatchError
	if !errors.As(err, &noMatch) {
		t.Fatalf(fmtResolveErrType, err, err)
	}
}

// ── Resolution: longer domain pattern beats shorter one ───────────────────────

// Test_registry_Resolve_longerPatternPreferredWithinSameLayer verifies that
// within the same layer, a longer (more specific) pattern is tried before a
// shorter one and wins when both could structurally match.
func Test_registry_Resolve_longerPatternPreferredWithinSameLayer(t *testing.T) {
	t.Parallel()

	// Invariant: among eligible patterns of the same layer, the longer pattern
	// (by character count) is tried first and wins on a match.
	reset()

	// Register the shorter pattern first to ensure registration order does
	// not affect resolution order.
	const (
		shortPattern = "station {s:station} sends BootNotification"
		longPattern  = "station {s:station} sends BootNotification with reason {r:string}"
	)

	registerPrimitive(shortPattern)
	registerPrimitive(longPattern)

	const stepLong = "station CP01 sends BootNotification with reason PoweredUp"

	match, err := Resolve(stepLong, api.OCPP16)
	if err != nil {
		t.Fatalf(fmtResolveUnexpectedErr, err)
	}

	if match.Keyword.Pattern != longPattern {
		t.Errorf(
			"Match.Keyword.Pattern: want longer pattern %q, got %q",
			longPattern,
			match.Keyword.Pattern,
		)
	}
}

// ── NoMatchError.Error() format ─────────────────────────────────────────────────

// Test_registry_NoMatchError_errorStringWithoutClosest verifies the error
// message format when no Closest suggestion is available.
func Test_registry_NoMatchError_errorStringWithoutClosest(t *testing.T) {
	t.Parallel()

	// Invariant: NoMatchError.Error() omits the "did you mean" clause when
	// Closest is empty.
	noMatchErr := &NoMatchError{StepText: "some step", Closest: emptyClosest}

	const wantMsg = `no keyword matches step "some step"`

	gotMsg := noMatchErr.Error()
	if gotMsg != wantMsg {
		t.Errorf("NoMatchError.Error(): want %q, got %q", wantMsg, gotMsg)
	}
}

// Test_registry_NoMatchError_errorStringWithClosest verifies the error message
// format when a Closest suggestion is available.
func Test_registry_NoMatchError_errorStringWithClosest(t *testing.T) {
	t.Parallel()

	// Invariant: NoMatchError.Error() includes "did you mean" clause when
	// Closest is non-empty.
	noMatchErr := &NoMatchError{
		StepText: "some step",
		Closest:  "some {s:string} step",
	}

	gotMsg := noMatchErr.Error()

	const wantFragment = "did you mean"

	if len(gotMsg) == 0 {
		t.Fatal("NoMatchError.Error(): got empty string")
	}

	found := false

	for idx := range len(gotMsg) - len(wantFragment) + 1 {
		if gotMsg[idx:idx+len(wantFragment)] == wantFragment {
			found = true

			break
		}
	}

	if !found {
		t.Errorf(
			"NoMatchError.Error(): want fragment %q in %q",
			wantFragment,
			gotMsg,
		)
	}
}

// ── TypeMismatchError.Error() format ────────────────────────────────────────────

// Test_registry_TypeMismatchError_errorStringFormat verifies the error message
// format of TypeMismatchError.
func Test_registry_TypeMismatchError_errorStringFormat(t *testing.T) {
	t.Parallel()

	// Invariant: TypeMismatchError.Error() identifies the argument, expected
	// type, and the raw value that failed coercion.
	mismatchErr := &TypeMismatchError{
		ArgName:  "count",
		Expected: typeNameInt,
		Got:      "notanint",
	}

	gotMsg := mismatchErr.Error()

	const wantMsg = `argument "count": expected type int, got "notanint"`

	if gotMsg != wantMsg {
		t.Errorf("TypeMismatchError.Error(): want %q, got %q", wantMsg, gotMsg)
	}
}
