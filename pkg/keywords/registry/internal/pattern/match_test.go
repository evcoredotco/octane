// Package pattern_test exercises the Parse, Match, and Coerce
// functions from the pattern package in a black-box manner.
//
// Task: T-003-13
// AC3: resolver returns correct Func and bound Args for a matching step.
// AC5: resolver returns ErrTypeMismatch when a placeholder type is violated.
package pattern_test

import (
	"errors"
	"testing"
	"time"

	"github.com/evcoreco/octane/pkg/keywords/registry/internal/pattern"
)

// ── Named test-value constants ────────────────────────────────────────────────

const (
	// Parse happy-path inputs.
	patternLiteralOnly     = "the CSMS connects"
	patternPlaceholderOnly = "{n:int}"
	patternMixed           = "send {count:int} frames to {target:station} within {timeout:duration}"

	// Parse error inputs.
	patternMissingColon = "{name}"
	patternEmptyName    = "{:int}"
	patternUnknownType  = "{n:uuid}"
	patternUnclosed     = "{n:int"
	patternBareClose    = "hello } world"
	patternEmptyBody    = "{}"

	// Match step strings.
	stepExactMatch       = "the CSMS connects"
	stepCaseDifferent    = "THE CSMS CONNECTS"
	stepExtraWords       = "the CSMS connects now unexpectedly"
	stepTooFewWords      = "the CSMS"
	stepMultiPlaceholder = "send 3 frames to CP01 within 30s"
	stepWrongWordOrder   = "CSMS the connects" //nolint:gosec // not a credential

	// Coerce raw values.
	valueValidString     = "hello"
	valueValidInt        = "42"
	valueInvalidInt      = "notAnInt"
	valueValidFloat      = "3.14"
	valueInvalidFloat    = "pi"
	valueValidBoolTrue   = "true"
	valueValidBoolFalse  = "false"
	valueValidBoolMixed  = "TRUE"
	valueInvalidBool     = "yes"
	valueValidDuration   = "30s"
	valueInvalidDuration = "thirtyseconds"
	valueValidStation    = "CP01"
	valueValidAny        = "anything_goes"
)

// ── Parse tests ───────────────────────────────────────────────────────────────

// Test_pattern_Parse_literalOnly verifies that a pattern with no
// placeholders parses into a single KindLiteral token.
func Test_pattern_Parse_literalOnly(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patternLiteralOnly)
	if err != nil {
		t.Fatalf("Parse(%q) unexpected error: %v", patternLiteralOnly, err)
	}

	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}

	tok := tokens[0]
	if tok.Kind != pattern.KindLiteral {
		t.Errorf("token Kind: want KindLiteral, got %v", tok.Kind)
	}

	if tok.Text != patternLiteralOnly {
		t.Errorf("token Text: want %q, got %q", patternLiteralOnly, tok.Text)
	}

	if tok.Name != "" {
		t.Errorf("token Name: want empty, got %q", tok.Name)
	}
}

// Test_pattern_Parse_placeholderOnly verifies that a pattern that is
// solely a placeholder produces a single KindPlaceholder token with
// the correct Name and Type.
func Test_pattern_Parse_placeholderOnly(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patternPlaceholderOnly)
	if err != nil {
		t.Fatalf("Parse(%q) unexpected error: %v", patternPlaceholderOnly, err)
	}

	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}

	tok := tokens[0]
	if tok.Kind != pattern.KindPlaceholder {
		t.Errorf("token Kind: want KindPlaceholder, got %v", tok.Kind)
	}

	if tok.Name != "n" {
		t.Errorf("token Name: want %q, got %q", "n", tok.Name)
	}

	if tok.Type != pattern.TypeInt {
		t.Errorf("token Type: want %q, got %q", pattern.TypeInt, tok.Type)
	}
}

// Test_pattern_Parse_mixed verifies that a pattern interleaving
// literals and placeholders produces tokens in the correct order with
// the correct kinds.
func Test_pattern_Parse_mixed(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patternMixed)
	if err != nil {
		t.Fatalf("Parse(%q) unexpected error: %v", patternMixed, err)
	}

	// Expect: literal, placeholder(count:int), literal,
	// placeholder(target:station), literal, placeholder(timeout:duration).
	const wantTokenCount = 6
	if len(tokens) != wantTokenCount {
		t.Fatalf(
			"expected %d tokens, got %d: %v",
			wantTokenCount,
			len(tokens),
			tokens,
		)
	}

	wantKinds := []pattern.Kind{
		pattern.KindLiteral,
		pattern.KindPlaceholder,
		pattern.KindLiteral,
		pattern.KindPlaceholder,
		pattern.KindLiteral,
		pattern.KindPlaceholder,
	}

	for idx, wantKind := range wantKinds {
		if tokens[idx].Kind != wantKind {
			t.Errorf(
				"token[%d] Kind: want %v, got %v",
				idx,
				wantKind,
				tokens[idx].Kind,
			)
		}
	}

	if tokens[1].Name != "count" {
		t.Errorf("token[1].Name: want %q, got %q", "count", tokens[1].Name)
	}

	if tokens[1].Type != pattern.TypeInt {
		t.Errorf(
			"token[1].Type: want %q, got %q",
			pattern.TypeInt,
			tokens[1].Type,
		)
	}

	if tokens[3].Name != "target" {
		t.Errorf("token[3].Name: want %q, got %q", "target", tokens[3].Name)
	}

	if tokens[3].Type != pattern.TypeStation {
		t.Errorf(
			"token[3].Type: want %q, got %q",
			pattern.TypeStation,
			tokens[3].Type,
		)
	}

	if tokens[5].Name != "timeout" {
		t.Errorf("token[5].Name: want %q, got %q", "timeout", tokens[5].Name)
	}

	if tokens[5].Type != pattern.TypeDuration {
		t.Errorf(
			"token[5].Type: want %q, got %q",
			pattern.TypeDuration,
			tokens[5].Type,
		)
	}
}

// Test_pattern_Parse_allSupportedTypes verifies that all seven
// placeholder types are accepted without error.
func Test_pattern_Parse_allSupportedTypes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		placeholder string
		wantType    pattern.PlaceholderType
	}{
		{
			name:        "string",
			placeholder: "{v:string}",
			wantType:    pattern.TypeString,
		},
		{name: "int", placeholder: "{v:int}", wantType: pattern.TypeInt},
		{name: "float", placeholder: "{v:float}", wantType: pattern.TypeFloat},
		{name: "bool", placeholder: "{v:bool}", wantType: pattern.TypeBool},
		{
			name:        "duration",
			placeholder: "{v:duration}",
			wantType:    pattern.TypeDuration,
		},
		{
			name:        "station",
			placeholder: "{v:station}",
			wantType:    pattern.TypeStation,
		},
		{name: "any", placeholder: "{v:any}", wantType: pattern.TypeAny},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tokens, err := pattern.Parse(testCase.placeholder)
			if err != nil {
				t.Fatalf(
					"Parse(%q) unexpected error: %v",
					testCase.placeholder,
					err,
				)
			}

			if len(tokens) != 1 {
				t.Fatalf("expected 1 token, got %d", len(tokens))
			}

			if tokens[0].Type != testCase.wantType {
				t.Errorf(
					"Type: want %q, got %q",
					testCase.wantType,
					tokens[0].Type,
				)
			}
		})
	}
}

// Test_pattern_Parse_malformedPlaceholders verifies that all
// documented malformed placeholder forms return a non-nil error.
func Test_pattern_Parse_malformedPlaceholders(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		patternText string
	}{
		{name: "missing colon", patternText: patternMissingColon},
		{name: "empty name", patternText: patternEmptyName},
		{name: "unknown type", patternText: patternUnknownType},
		{name: "unclosed brace", patternText: patternUnclosed},
		{name: "bare close brace", patternText: patternBareClose},
		{name: "empty body", patternText: patternEmptyBody},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tokens, err := pattern.Parse(testCase.patternText)
			if err == nil {
				t.Errorf(
					"Parse(%q): expected error, got tokens %v",
					testCase.patternText,
					tokens,
				)
			}
		})
	}
}

// Test_pattern_Parse_emptyString verifies that an empty pattern
// string returns an error.
func Test_pattern_Parse_emptyString(t *testing.T) {
	t.Parallel()

	_, err := pattern.Parse("")
	if err == nil {
		t.Fatal("Parse(\"\"): expected error, got nil")
	}
}

// ── Match tests ───────────────────────────────────────────────────────────────

// Test_pattern_Match_exactLiteral verifies that a literal-only pattern
// matches a step string word-for-word.
func Test_pattern_Match_exactLiteral(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patternLiteralOnly)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	captures, matched := pattern.Match(tokens, stepExactMatch)
	if !matched {
		t.Fatalf("Match: expected true, got false")
	}

	if len(captures) != 0 {
		t.Errorf("captures: want empty map, got %v", captures)
	}
}

// Test_pattern_Match_caseInsensitive verifies that matching is
// case-insensitive for literal tokens.
func Test_pattern_Match_caseInsensitive(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patternLiteralOnly)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	_, matched := pattern.Match(tokens, stepCaseDifferent)
	if !matched {
		t.Fatalf(
			"Match(%q): expected case-insensitive match, got false",
			stepCaseDifferent,
		)
	}
}

// Test_pattern_Match_extraStepWordsNoMatch verifies that a step with
// more words than the pattern does not match.
func Test_pattern_Match_extraStepWordsNoMatch(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patternLiteralOnly)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	_, matched := pattern.Match(tokens, stepExtraWords)
	if matched {
		t.Fatalf(
			"Match(%q): expected false for extra words, got true",
			stepExtraWords,
		)
	}
}

// Test_pattern_Match_tooFewStepWordsNoMatch verifies that a step with
// fewer words than the pattern does not match.
func Test_pattern_Match_tooFewStepWordsNoMatch(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patternLiteralOnly)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	_, matched := pattern.Match(tokens, stepTooFewWords)
	if matched {
		t.Fatalf(
			"Match(%q): expected false for too few words, got true",
			stepTooFewWords,
		)
	}
}

// Test_pattern_Match_emptyStepNoMatch verifies that an empty step
// string never matches.
func Test_pattern_Match_emptyStepNoMatch(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patternLiteralOnly)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	captures, matched := pattern.Match(tokens, "")
	if matched {
		t.Fatalf("Match(\"\"): expected false for empty step, got true")
	}

	if captures != nil {
		t.Errorf("captures: want nil for failed match, got %v", captures)
	}
}

// Test_pattern_Match_multiPlaceholderCapture verifies that a pattern
// with multiple placeholders captures each word under the correct name.
func Test_pattern_Match_multiPlaceholderCapture(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patternMixed)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	captures, matched := pattern.Match(tokens, stepMultiPlaceholder)
	if !matched {
		t.Fatalf("Match(%q): expected true, got false", stepMultiPlaceholder)
	}

	if captures["count"] != "3" {
		t.Errorf("captures[count]: want %q, got %q", "3", captures["count"])
	}

	if captures["target"] != "CP01" {
		t.Errorf(
			"captures[target]: want %q, got %q",
			"CP01",
			captures["target"],
		)
	}

	if captures["timeout"] != "30s" {
		t.Errorf(
			"captures[timeout]: want %q, got %q",
			"30s",
			captures["timeout"],
		)
	}
}

// Test_pattern_Match_wordOrderMismatch verifies that mismatched word
// order in the step yields no match even if the right words are present.
func Test_pattern_Match_wordOrderMismatch(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patternLiteralOnly)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	_, matched := pattern.Match(tokens, stepWrongWordOrder)
	if matched {
		t.Fatalf(
			"Match(%q): expected false for wrong word order, got true",
			stepWrongWordOrder,
		)
	}
}

// Test_pattern_Match_successfulMatchReturnsNonNilMap verifies that a
// successful match with no placeholders still returns a non-nil map.
func Test_pattern_Match_successfulMatchReturnsNonNilMap(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patternLiteralOnly)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	captures, matched := pattern.Match(tokens, stepExactMatch)
	if !matched {
		t.Fatalf("Match: expected true, got false")
	}

	// Contract: non-nil even when empty.
	if captures == nil {
		t.Error(
			"captures: want non-nil map on successful match with no placeholders",
		)
	}
}

// Test_pattern_Match_extraInternalWhitespaceNormalized verifies that
// runs of whitespace in the step are treated as single separators.
func Test_pattern_Match_extraInternalWhitespaceNormalized(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patternLiteralOnly)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	// Multiple spaces between words must still match.
	_, matched := pattern.Match(tokens, "the  CSMS   connects")
	if !matched {
		t.Fatal(
			"Match: expected true for step with extra internal whitespace, got false",
		)
	}
}

// ── Coerce tests ──────────────────────────────────────────────────────────────

// Test_pattern_Coerce_stringType verifies that TypeString captures are
// stored as Go strings without modification.
func Test_pattern_Coerce_stringType(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse("{v:string}")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	captures := map[string]string{"v": valueValidString}

	result, err := pattern.Coerce(captures, tokens)
	if err != nil {
		t.Fatalf("Coerce: unexpected error: %v", err)
	}

	if result["v"] != valueValidString {
		t.Errorf(
			"coerced value: want %q, got %v",
			valueValidString,
			result["v"],
		)
	}
}

// Test_pattern_Coerce_anyType verifies that TypeAny captures are
// stored as raw strings without coercion.
func Test_pattern_Coerce_anyType(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse("{v:any}")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	captures := map[string]string{"v": valueValidAny}

	result, err := pattern.Coerce(captures, tokens)
	if err != nil {
		t.Fatalf("Coerce: unexpected error: %v", err)
	}

	if result["v"] != valueValidAny {
		t.Errorf("coerced value: want %q, got %v", valueValidAny, result["v"])
	}
}

// Test_pattern_Coerce_stationType verifies that TypeStation captures
// are stored as strings (semantic validation is the registry's job).
func Test_pattern_Coerce_stationType(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse("{v:station}")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	captures := map[string]string{"v": valueValidStation}

	result, err := pattern.Coerce(captures, tokens)
	if err != nil {
		t.Fatalf("Coerce: unexpected error: %v", err)
	}

	if result["v"] != valueValidStation {
		t.Errorf(
			"coerced value: want %q, got %v",
			valueValidStation,
			result["v"],
		)
	}
}

// Test_pattern_Coerce_intTypeSuccess verifies that a valid integer
// string is coerced to an int value.
func Test_pattern_Coerce_intTypeSuccess(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse("{v:int}")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	captures := map[string]string{"v": valueValidInt}

	result, err := pattern.Coerce(captures, tokens)
	if err != nil {
		t.Fatalf("Coerce: unexpected error: %v", err)
	}

	got, typeOk := result["v"].(int)
	if !typeOk {
		t.Fatalf("coerced value type: want int, got %T", result["v"])
	}

	const wantInt = 42

	if got != wantInt {
		t.Errorf("coerced value: want %d, got %d", wantInt, got)
	}
}

// Test_pattern_Coerce_intTypeFailure verifies that a non-integer token
// yields a *CoercionError with the correct fields.
func Test_pattern_Coerce_intTypeFailure(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse("{v:int}")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	captures := map[string]string{"v": valueInvalidInt}

	_, err = pattern.Coerce(captures, tokens)
	if err == nil {
		t.Fatal("Coerce: expected CoercionError, got nil")
	}

	var coercErr *pattern.CoercionError
	if !errors.As(err, &coercErr) {
		t.Fatalf(
			"error type: want *pattern.CoercionError, got %T: %v",
			err,
			err,
		)
	}

	if coercErr.ArgName != "v" {
		t.Errorf(
			"CoercionError.ArgName: want %q, got %q",
			"v",
			coercErr.ArgName,
		)
	}

	if coercErr.Expected != "int" {
		t.Errorf(
			"CoercionError.Expected: want %q, got %q",
			"int",
			coercErr.Expected,
		)
	}

	if coercErr.Got != valueInvalidInt {
		t.Errorf(
			"CoercionError.Got: want %q, got %q",
			valueInvalidInt,
			coercErr.Got,
		)
	}
}

// Test_pattern_Coerce_floatTypeSuccess verifies that a valid float
// string is coerced to a float64 value.
func Test_pattern_Coerce_floatTypeSuccess(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse("{v:float}")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	captures := map[string]string{"v": valueValidFloat}

	result, err := pattern.Coerce(captures, tokens)
	if err != nil {
		t.Fatalf("Coerce: unexpected error: %v", err)
	}

	got, typeOk := result["v"].(float64)
	if !typeOk {
		t.Fatalf("coerced value type: want float64, got %T", result["v"])
	}

	const wantFloat = 3.14

	const epsilon = 1e-9

	if got < wantFloat-epsilon || got > wantFloat+epsilon {
		t.Errorf("coerced value: want ~%f, got %f", wantFloat, got)
	}
}

// Test_pattern_Coerce_floatTypeFailure verifies that a non-float token
// yields a *CoercionError.
func Test_pattern_Coerce_floatTypeFailure(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse("{v:float}")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	captures := map[string]string{"v": valueInvalidFloat}

	_, err = pattern.Coerce(captures, tokens)
	if err == nil {
		t.Fatal("Coerce: expected CoercionError, got nil")
	}

	var coercErr *pattern.CoercionError
	if !errors.As(err, &coercErr) {
		t.Fatalf(
			"error type: want *pattern.CoercionError, got %T: %v",
			err,
			err,
		)
	}

	if coercErr.Expected != "float" {
		t.Errorf(
			"CoercionError.Expected: want %q, got %q",
			"float",
			coercErr.Expected,
		)
	}
}

// Test_pattern_Coerce_boolTypeSuccess verifies that "true", "false",
// and mixed-case variants are all coerced to Go bool values.
func Test_pattern_Coerce_boolTypeSuccess(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  bool
	}{
		{input: valueValidBoolTrue, want: true},
		{input: valueValidBoolFalse, want: false},
		{input: valueValidBoolMixed, want: true},
		{input: "FALSE", want: false},
	}

	for _, testCase := range cases {
		t.Run(testCase.input, func(t *testing.T) {
			t.Parallel()

			tokens, err := pattern.Parse("{v:bool}")
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}

			captures := map[string]string{"v": testCase.input}

			result, err := pattern.Coerce(captures, tokens)
			if err != nil {
				t.Fatalf(
					"Coerce(%q): unexpected error: %v",
					testCase.input,
					err,
				)
			}

			got, typeOk := result["v"].(bool)
			if !typeOk {
				t.Fatalf("coerced value type: want bool, got %T", result["v"])
			}

			if got != testCase.want {
				t.Errorf("coerced value: want %v, got %v", testCase.want, got)
			}
		})
	}
}

// Test_pattern_Coerce_boolTypeFailure verifies that an invalid bool
// token yields a *CoercionError.
func Test_pattern_Coerce_boolTypeFailure(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse("{v:bool}")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	captures := map[string]string{"v": valueInvalidBool}

	_, err = pattern.Coerce(captures, tokens)
	if err == nil {
		t.Fatal("Coerce: expected CoercionError, got nil")
	}

	var coercErr *pattern.CoercionError
	if !errors.As(err, &coercErr) {
		t.Fatalf(
			"error type: want *pattern.CoercionError, got %T: %v",
			err,
			err,
		)
	}

	if coercErr.Expected != "bool" {
		t.Errorf(
			"CoercionError.Expected: want %q, got %q",
			"bool",
			coercErr.Expected,
		)
	}

	if coercErr.Got != valueInvalidBool {
		t.Errorf(
			"CoercionError.Got: want %q, got %q",
			valueInvalidBool,
			coercErr.Got,
		)
	}
}

// Test_pattern_Coerce_durationTypeSuccess verifies that a valid
// duration string is coerced to a time.Duration value.
func Test_pattern_Coerce_durationTypeSuccess(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse("{v:duration}")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	captures := map[string]string{"v": valueValidDuration}

	result, err := pattern.Coerce(captures, tokens)
	if err != nil {
		t.Fatalf("Coerce: unexpected error: %v", err)
	}

	got, typeOk := result["v"].(time.Duration)
	if !typeOk {
		t.Fatalf("coerced value type: want time.Duration, got %T", result["v"])
	}

	const wantDuration = 30 * time.Second

	if got != wantDuration {
		t.Errorf("coerced value: want %v, got %v", wantDuration, got)
	}
}

// Test_pattern_Coerce_durationTypeFailure verifies that an
// unparseable duration token yields a *CoercionError.
func Test_pattern_Coerce_durationTypeFailure(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse("{v:duration}")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	captures := map[string]string{"v": valueInvalidDuration}

	_, err = pattern.Coerce(captures, tokens)
	if err == nil {
		t.Fatal("Coerce: expected CoercionError, got nil")
	}

	var coercErr *pattern.CoercionError
	if !errors.As(err, &coercErr) {
		t.Fatalf(
			"error type: want *pattern.CoercionError, got %T: %v",
			err,
			err,
		)
	}

	if coercErr.Expected != "duration" {
		t.Errorf(
			"CoercionError.Expected: want %q, got %q",
			"duration",
			coercErr.Expected,
		)
	}
}

// Test_pattern_Coerce_missingCaptureKeyIgnored verifies that a
// KindPlaceholder token whose name is absent from the captures map is
// silently skipped rather than panicking or returning an error.
// Per AC5 and the implementation comment in coerce.go, this path
// guards an internal invariant; the registry must never trigger it.
func Test_pattern_Coerce_missingCaptureKeyIgnored(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse("{v:int}")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	// Pass an empty captures map — the key "v" is absent.
	result, err := pattern.Coerce(map[string]string{}, tokens)
	if err != nil {
		t.Fatalf("Coerce with missing key: unexpected error: %v", err)
	}

	// The result map exists but the placeholder key must be absent.
	if _, present := result["v"]; present {
		t.Error("result[v]: key should be absent when capture was missing")
	}
}

// Test_pattern_Coerce_CoercionErrorMessage verifies that the
// CoercionError.Error() string contains the argument name, expected
// type, and got value in the documented format.
func Test_pattern_Coerce_CoercionErrorMessage(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse("{n:int}")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	_, err = pattern.Coerce(map[string]string{"n": valueInvalidInt}, tokens)
	if err == nil {
		t.Fatal("Coerce: expected error, got nil")
	}

	msg := err.Error()

	// The error message must reference the arg name.
	if !containsSubstring(msg, "n") {
		t.Errorf("error message %q does not contain arg name %q", msg, "n")
	}

	// The error message must reference the expected type.
	if !containsSubstring(msg, "int") {
		t.Errorf(
			"error message %q does not contain expected type %q",
			msg,
			"int",
		)
	}

	// The error message must reference the bad value.
	if !containsSubstring(msg, valueInvalidInt) {
		t.Errorf(
			"error message %q does not contain got value %q",
			msg,
			valueInvalidInt,
		)
	}
}

// Test_pattern_Coerce_multipleTokensRoundTrip verifies the full
// pipeline: Parse -> Match -> Coerce against a multi-placeholder
// pattern drawn from AC3 of the spec.
func Test_pattern_Coerce_multipleTokensRoundTrip(t *testing.T) {
	t.Parallel()

	const ac3Pattern = "the CSMS sends ReserveNow with connectorId" +
		" {connectorId:int} and idTag {idTag:string}" +
		" to station {station:station} within {timeout:duration}"

	const ac3Step = "the CSMS sends ReserveNow with connectorId 1" +
		" and idTag X to station CP01 within 30s"

	tokens, err := pattern.Parse(ac3Pattern)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	captures, matched := pattern.Match(tokens, ac3Step)
	if !matched {
		t.Fatalf("Match: expected true for AC3 step, got false")
	}

	result, err := pattern.Coerce(captures, tokens)
	if err != nil {
		t.Fatalf("Coerce: unexpected error: %v", err)
	}

	// connectorId=1 (int)
	connectorID, typeOk := result["connectorId"].(int)
	if !typeOk {
		t.Fatalf("connectorId type: want int, got %T", result["connectorId"])
	}

	const wantConnectorID = 1

	if connectorID != wantConnectorID {
		t.Errorf("connectorId: want %d, got %d", wantConnectorID, connectorID)
	}

	// idTag="X" (string)
	idTag, typeOk := result["idTag"].(string)
	if !typeOk {
		t.Fatalf("idTag type: want string, got %T", result["idTag"])
	}

	if idTag != "X" {
		t.Errorf("idTag: want %q, got %q", "X", idTag)
	}

	// station="CP01" (station stored as string)
	station, typeOk := result["station"].(string)
	if !typeOk {
		t.Fatalf("station type: want string, got %T", result["station"])
	}

	if station != "CP01" {
		t.Errorf("station: want %q, got %q", "CP01", station)
	}

	// timeout=30s (duration)
	timeout, typeOk := result["timeout"].(time.Duration)
	if !typeOk {
		t.Fatalf("timeout type: want time.Duration, got %T", result["timeout"])
	}

	const wantTimeout = 30 * time.Second

	if timeout != wantTimeout {
		t.Errorf("timeout: want %v, got %v", wantTimeout, timeout)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// containsSubstring reports whether s contains sub. Kept as an
// inline helper to avoid importing "strings" solely for test assertions.
func containsSubstring(str, sub string) bool {
	if len(str) < len(sub) {
		return false
	}

	if len(sub) == 0 {
		return true
	}

	for idx := 0; idx <= len(str)-len(sub); idx++ {
		if str[idx:idx+len(sub)] == sub {
			return true
		}
	}

	return false
}
