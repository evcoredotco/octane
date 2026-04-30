// Package pattern_test exercises the Parse, Match, and Coerce
// functions from the pattern package in a black-box manner.
//
// Task: T-003-13
// AC3: resolver returns correct Func and bound Args for a matching step.
// AC5: resolver returns TypeMismatchError when a placeholder type is violated.
package pattern_test

import (
	"errors"
	"testing"
	"time"

	"github.com/evcoreco/octane/pkg/keywords/registry/internal/pattern"
)

// ── Named test-value constants ───────────────────────────────────────────

const (
	// Parse happy-path inputs.
	patternLiteralOnly     = "the CSMS connects"
	patternPlaceholderOnly = "{n:int}"
	patternMixed           = "send {count:int} frames to {target:station}" +
		" within {timeout:duration}"

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
	stepWrongWordOrder   = "CSMS the connects" //nolint:gosec // not cred

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

	// Pattern templates used across multiple Coerce tests.
	patVInt      = "{v:int}"
	patVFloat    = "{v:float}"
	patVBool     = "{v:bool}"
	patVDuration = "{v:duration}"
	patVStation  = "{v:station}"
	patVString   = "{v:string}"
	patVAny      = "{v:any}"
	patNInt      = "{n:int}"

	// Argument-name constants used as map keys.
	argNameV         = "v"
	argNameN         = "n"
	captureKeyCount  = "count"
	captureKeyTarget = "target"
	captureKeyTimeo  = "timeout"

	// Type-name strings compared to CoercionError.Expected.
	typeNameInt      = "int"
	typeNameFloat    = "float"
	typeNameBool     = "bool"
	typeNameDuration = "duration"

	// Repeated format strings.
	fmtParseUnexpectedErr  = "Parse(%q) unexpected error: %v"
	fmtParseErr            = "Parse: %v"
	fmtCoerceUnexpectedErr = "Coerce: unexpected error: %v"
	fmtCoercedValue        = "coerced value: want %q, got %v"
	fmtCoerceExpNilErr     = "Coerce: expected CoercionError, got nil"
	fmtCoerceErrType       = "error type: want *pattern.CoercionError, " +
		"got %T: %v"
	fmtCoerceErrExpected = "CoercionError.Expected: want %q, got %q"

	// Empty string used as a step argument.
	emptyStepStr = ""

	// Magic-number sentinels.
	wantOneToken = 1
	wantZeroLen  = 0
	tokenIdx0    = 0
	tokenIdx1    = 1
	tokenIdx3    = 3
	tokenIdx5    = 5

	// fmtExpectedOneToken is logged when Parse returns a token count != 1.
	fmtExpectedOneToken = "expected 1 token, got %d"
)

// ── Parse tests ────────────────────────────────────────────────────────

// Test_pattern_Parse_literalOnly verifies that a pattern with no
// placeholders parses into a single KindLiteral token.
func Test_pattern_Parse_literalOnly(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patternLiteralOnly)
	if err != nil {
		t.Fatalf(fmtParseUnexpectedErr, patternLiteralOnly, err)
	}

	if len(tokens) != wantOneToken {
		t.Fatalf(fmtExpectedOneToken, len(tokens))
	}

	tok := tokens[tokenIdx0]
	if tok.Kind != pattern.KindLiteral {
		t.Errorf("token Kind: want KindLiteral, got %v", tok.Kind)
	}

	if tok.Text != patternLiteralOnly {
		t.Errorf("token Text: want %q, got %q", patternLiteralOnly, tok.Text)
	}

	if tok.Name != emptyStepStr {
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
		t.Fatalf(fmtParseUnexpectedErr, patternPlaceholderOnly, err)
	}

	if len(tokens) != wantOneToken {
		t.Fatalf(fmtExpectedOneToken, len(tokens))
	}

	tok := tokens[tokenIdx0]
	if tok.Kind != pattern.KindPlaceholder {
		t.Errorf("token Kind: want KindPlaceholder, got %v", tok.Kind)
	}

	if tok.Name != argNameN {
		t.Errorf("token Name: want %q, got %q", argNameN, tok.Name)
	}

	if tok.Type != pattern.TypeInt {
		t.Errorf("token Type: want %q, got %q", pattern.TypeInt, tok.Type)
	}
}

// assertPlaceholderToken verifies that tokens[idx] is a KindPlaceholder
// with the given name and type.
func assertPlaceholderToken(
	t *testing.T,
	tokens []pattern.Token,
	idx int,
	wantName string,
	wantType pattern.PlaceholderType,
) {
	t.Helper()

	if tokens[idx].Name != wantName {
		t.Errorf(
			"token[%d].Name: want %q, got %q",
			idx,
			wantName,
			tokens[idx].Name,
		)
	}

	if tokens[idx].Type != wantType {
		t.Errorf(
			"token[%d].Type: want %q, got %q",
			idx,
			wantType,
			tokens[idx].Type,
		)
	}
}

// Test_pattern_Parse_mixed verifies that a pattern interleaving
// literals and placeholders produces tokens in the correct order with
// the correct kinds.
func Test_pattern_Parse_mixed(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patternMixed)
	if err != nil {
		t.Fatalf(fmtParseUnexpectedErr, patternMixed, err)
	}

	// Expect: literal, placeholder(count:int), literal,
	// placeholder(target:station), literal,
	// placeholder(timeout:duration).
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

	assertPlaceholderToken(
		t, tokens, tokenIdx1, captureKeyCount, pattern.TypeInt,
	)
	assertPlaceholderToken(
		t, tokens, tokenIdx3, captureKeyTarget, pattern.TypeStation,
	)
	assertPlaceholderToken(
		t, tokens, tokenIdx5, captureKeyTimeo, pattern.TypeDuration,
	)
}

// assertSinglePlaceholderType parses placeholder and verifies that the
// single token has the expected PlaceholderType.
func assertSinglePlaceholderType(
	t *testing.T,
	placeholder string,
	wantType pattern.PlaceholderType,
) {
	t.Helper()

	tokens, err := pattern.Parse(placeholder)
	if err != nil {
		t.Fatalf(fmtParseUnexpectedErr, placeholder, err)
	}

	if len(tokens) != wantOneToken {
		t.Fatalf(fmtExpectedOneToken, len(tokens))
	}

	if tokens[tokenIdx0].Type != wantType {
		t.Errorf("Type: want %q, got %q", wantType, tokens[tokenIdx0].Type)
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
		{name: typeNameInt, placeholder: patVInt, wantType: pattern.TypeInt},
		{
			name:        typeNameFloat,
			placeholder: patVFloat,
			wantType:    pattern.TypeFloat,
		},
		{name: typeNameBool, placeholder: patVBool, wantType: pattern.TypeBool},
		{
			name:        typeNameDuration,
			placeholder: patVDuration,
			wantType:    pattern.TypeDuration,
		},
		{
			name:        "station",
			placeholder: patVStation,
			wantType:    pattern.TypeStation,
		},
		{name: "any", placeholder: patVAny, wantType: pattern.TypeAny},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			assertSinglePlaceholderType(
				t, testCase.placeholder, testCase.wantType,
			)
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

	_, err := pattern.Parse(emptyStepStr)
	if err == nil {
		t.Fatal("Parse(\"\"): expected error, got nil")
	}
}

// ── Match tests ────────────────────────────────────────────────────────

// Test_pattern_Match_exactLiteral verifies that a literal-only pattern
// matches a step string word-for-word.
func Test_pattern_Match_exactLiteral(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patternLiteralOnly)
	if err != nil {
		t.Fatalf(fmtParseErr, err)
	}

	captures, matched := pattern.Match(tokens, stepExactMatch)
	if !matched {
		t.Fatal("Match: expected true, got false")
	}

	if len(captures) != wantZeroLen {
		t.Errorf("captures: want empty map, got %v", captures)
	}
}

// Test_pattern_Match_caseInsensitive verifies that matching is
// case-insensitive for literal tokens.
func Test_pattern_Match_caseInsensitive(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patternLiteralOnly)
	if err != nil {
		t.Fatalf(fmtParseErr, err)
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
		t.Fatalf(fmtParseErr, err)
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
		t.Fatalf(fmtParseErr, err)
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
		t.Fatalf(fmtParseErr, err)
	}

	captures, matched := pattern.Match(tokens, emptyStepStr)
	if matched {
		t.Fatal("Match(\"\"): expected false for empty step, got true")
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
		t.Fatalf(fmtParseErr, err)
	}

	captures, matched := pattern.Match(tokens, stepMultiPlaceholder)
	if !matched {
		t.Fatalf("Match(%q): expected true, got false", stepMultiPlaceholder)
	}

	if captures[captureKeyCount] != "3" {
		t.Errorf(
			"captures[count]: want %q, got %q",
			"3",
			captures[captureKeyCount],
		)
	}

	if captures[captureKeyTarget] != valueValidStation {
		t.Errorf(
			"captures[target]: want %q, got %q",
			valueValidStation,
			captures[captureKeyTarget],
		)
	}

	if captures[captureKeyTimeo] != "30s" {
		t.Errorf(
			"captures[timeout]: want %q, got %q",
			"30s",
			captures[captureKeyTimeo],
		)
	}
}

// Test_pattern_Match_wordOrderMismatch verifies that mismatched word
// order in the step yields no match even if the right words are present.
func Test_pattern_Match_wordOrderMismatch(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patternLiteralOnly)
	if err != nil {
		t.Fatalf(fmtParseErr, err)
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
		t.Fatalf(fmtParseErr, err)
	}

	captures, matched := pattern.Match(tokens, stepExactMatch)
	if !matched {
		t.Fatal("Match: expected true, got false")
	}

	// Contract: non-nil even when empty.
	if captures == nil {
		t.Error(
			"captures: want non-nil map on successful match with no " +
				"placeholders",
		)
	}
}

// Test_pattern_Match_extraInternalWhitespaceNormalized verifies that
// runs of whitespace in the step are treated as single separators.
func Test_pattern_Match_extraInternalWhitespaceNormalized(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patternLiteralOnly)
	if err != nil {
		t.Fatalf(fmtParseErr, err)
	}

	// Multiple spaces between words must still match.
	_, matched := pattern.Match(tokens, "the  CSMS   connects")
	if !matched {
		t.Fatal(
			"Match: expected true for step with extra internal " +
				"whitespace, got false",
		)
	}
}

// ── Coerce tests ───────────────────────────────────────────────────────

// Test_pattern_Coerce_stringType verifies that TypeString captures are
// stored as Go strings without modification.
func Test_pattern_Coerce_stringType(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patVString)
	if err != nil {
		t.Fatalf(fmtParseErr, err)
	}

	captures := map[string]string{argNameV: valueValidString}

	result, err := pattern.Coerce(captures, tokens)
	if err != nil {
		t.Fatalf(fmtCoerceUnexpectedErr, err)
	}

	if result[argNameV] != valueValidString {
		t.Errorf(
			fmtCoercedValue,
			valueValidString,
			result[argNameV],
		)
	}
}

// Test_pattern_Coerce_anyType verifies that TypeAny captures are
// stored as raw strings without coercion.
func Test_pattern_Coerce_anyType(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patVAny)
	if err != nil {
		t.Fatalf(fmtParseErr, err)
	}

	captures := map[string]string{argNameV: valueValidAny}

	result, err := pattern.Coerce(captures, tokens)
	if err != nil {
		t.Fatalf(fmtCoerceUnexpectedErr, err)
	}

	if result[argNameV] != valueValidAny {
		t.Errorf(fmtCoercedValue, valueValidAny, result[argNameV])
	}
}

// Test_pattern_Coerce_stationType verifies that TypeStation captures
// are stored as strings (semantic validation is the registry's job).
func Test_pattern_Coerce_stationType(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patVStation)
	if err != nil {
		t.Fatalf(fmtParseErr, err)
	}

	captures := map[string]string{argNameV: valueValidStation}

	result, err := pattern.Coerce(captures, tokens)
	if err != nil {
		t.Fatalf(fmtCoerceUnexpectedErr, err)
	}

	if result[argNameV] != valueValidStation {
		t.Errorf(
			fmtCoercedValue,
			valueValidStation,
			result[argNameV],
		)
	}
}

// Test_pattern_Coerce_intTypeSuccess verifies that a valid integer
// string is coerced to an int value.
func Test_pattern_Coerce_intTypeSuccess(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patVInt)
	if err != nil {
		t.Fatalf(fmtParseErr, err)
	}

	captures := map[string]string{argNameV: valueValidInt}

	result, err := pattern.Coerce(captures, tokens)
	if err != nil {
		t.Fatalf(fmtCoerceUnexpectedErr, err)
	}

	got, typeOk := result[argNameV].(int)
	if !typeOk {
		t.Fatalf("coerced value type: want int, got %T", result[argNameV])
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

	tokens, err := pattern.Parse(patVInt)
	if err != nil {
		t.Fatalf(fmtParseErr, err)
	}

	captures := map[string]string{argNameV: valueInvalidInt}

	_, err = pattern.Coerce(captures, tokens)
	if err == nil {
		t.Fatal(fmtCoerceExpNilErr)
	}

	var coercErr *pattern.CoercionError
	if !errors.As(err, &coercErr) {
		t.Fatalf(fmtCoerceErrType, err, err)
	}

	if coercErr.ArgName != argNameV {
		t.Errorf(
			"CoercionError.ArgName: want %q, got %q",
			argNameV,
			coercErr.ArgName,
		)
	}

	if coercErr.Expected != typeNameInt {
		t.Errorf(fmtCoerceErrExpected, typeNameInt, coercErr.Expected)
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

	tokens, err := pattern.Parse(patVFloat)
	if err != nil {
		t.Fatalf(fmtParseErr, err)
	}

	captures := map[string]string{argNameV: valueValidFloat}

	result, err := pattern.Coerce(captures, tokens)
	if err != nil {
		t.Fatalf(fmtCoerceUnexpectedErr, err)
	}

	got, typeOk := result[argNameV].(float64)
	if !typeOk {
		t.Fatalf("coerced value type: want float64, got %T", result[argNameV])
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

	tokens, err := pattern.Parse(patVFloat)
	if err != nil {
		t.Fatalf(fmtParseErr, err)
	}

	captures := map[string]string{argNameV: valueInvalidFloat}

	_, err = pattern.Coerce(captures, tokens)
	if err == nil {
		t.Fatal(fmtCoerceExpNilErr)
	}

	var coercErr *pattern.CoercionError
	if !errors.As(err, &coercErr) {
		t.Fatalf(fmtCoerceErrType, err, err)
	}

	if coercErr.Expected != typeNameFloat {
		t.Errorf(fmtCoerceErrExpected, typeNameFloat, coercErr.Expected)
	}
}

// assertBoolCoercedTrue parses patVBool, coerces input, and verifies
// the resulting bool value is true.
func assertBoolCoercedTrue(t *testing.T, input string) {
	t.Helper()

	tokens, err := pattern.Parse(patVBool)
	if err != nil {
		t.Fatalf(fmtParseErr, err)
	}

	result, err := pattern.Coerce(
		map[string]string{argNameV: input}, tokens,
	)
	if err != nil {
		t.Fatalf("Coerce(%q): unexpected error: %v", input, err)
	}

	got, typeOk := result[argNameV].(bool)
	if !typeOk {
		t.Fatalf(
			"coerced value type: want bool, got %T",
			result[argNameV],
		)
	}

	if !got {
		t.Error("coerced value: want true, got false")
	}
}

// assertBoolCoercedFalse parses patVBool, coerces input, and verifies
// the resulting bool value is false.
func assertBoolCoercedFalse(t *testing.T, input string) {
	t.Helper()

	tokens, err := pattern.Parse(patVBool)
	if err != nil {
		t.Fatalf(fmtParseErr, err)
	}

	result, err := pattern.Coerce(
		map[string]string{argNameV: input}, tokens,
	)
	if err != nil {
		t.Fatalf("Coerce(%q): unexpected error: %v", input, err)
	}

	got, typeOk := result[argNameV].(bool)
	if !typeOk {
		t.Fatalf(
			"coerced value type: want bool, got %T",
			result[argNameV],
		)
	}

	if got {
		t.Error("coerced value: want false, got true")
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

			if testCase.want {
				assertBoolCoercedTrue(t, testCase.input)
			} else {
				assertBoolCoercedFalse(t, testCase.input)
			}
		})
	}
}

// Test_pattern_Coerce_boolTypeFailure verifies that an invalid bool
// token yields a *CoercionError.
func Test_pattern_Coerce_boolTypeFailure(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patVBool)
	if err != nil {
		t.Fatalf(fmtParseErr, err)
	}

	captures := map[string]string{argNameV: valueInvalidBool}

	_, err = pattern.Coerce(captures, tokens)
	if err == nil {
		t.Fatal(fmtCoerceExpNilErr)
	}

	var coercErr *pattern.CoercionError
	if !errors.As(err, &coercErr) {
		t.Fatalf(fmtCoerceErrType, err, err)
	}

	if coercErr.Expected != typeNameBool {
		t.Errorf(fmtCoerceErrExpected, typeNameBool, coercErr.Expected)
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

	tokens, err := pattern.Parse(patVDuration)
	if err != nil {
		t.Fatalf(fmtParseErr, err)
	}

	captures := map[string]string{argNameV: valueValidDuration}

	result, err := pattern.Coerce(captures, tokens)
	if err != nil {
		t.Fatalf(fmtCoerceUnexpectedErr, err)
	}

	got, typeOk := result[argNameV].(time.Duration)
	if !typeOk {
		t.Fatalf(
			"coerced value type: want time.Duration, got %T",
			result[argNameV],
		)
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

	tokens, err := pattern.Parse(patVDuration)
	if err != nil {
		t.Fatalf(fmtParseErr, err)
	}

	captures := map[string]string{argNameV: valueInvalidDuration}

	_, err = pattern.Coerce(captures, tokens)
	if err == nil {
		t.Fatal(fmtCoerceExpNilErr)
	}

	var coercErr *pattern.CoercionError
	if !errors.As(err, &coercErr) {
		t.Fatalf(fmtCoerceErrType, err, err)
	}

	if coercErr.Expected != typeNameDuration {
		t.Errorf(fmtCoerceErrExpected, typeNameDuration, coercErr.Expected)
	}
}

// Test_pattern_Coerce_missingCaptureKeyIgnored verifies that a
// KindPlaceholder token whose name is absent from the captures map is
// silently skipped rather than panicking or returning an error.
// Per AC5 and the implementation comment in coerce.go, this path
// guards an internal invariant; the registry must never trigger it.
func Test_pattern_Coerce_missingCaptureKeyIgnored(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patVInt)
	if err != nil {
		t.Fatalf(fmtParseErr, err)
	}

	// Pass an empty captures map — the key argNameV is absent.
	result, err := pattern.Coerce(map[string]string{}, tokens)
	if err != nil {
		t.Fatalf("Coerce with missing key: unexpected error: %v", err)
	}

	// The result map exists but the placeholder key must be absent.
	if _, present := result[argNameV]; present {
		t.Error("result[v]: key should be absent when capture was missing")
	}
}

// Test_pattern_Coerce_CoercionErrorMessage verifies that the
// CoercionError.Error() string contains the argument name, expected
// type, and got value in the documented format.
func Test_pattern_Coerce_CoercionErrorMessage(t *testing.T) {
	t.Parallel()

	tokens, err := pattern.Parse(patNInt)
	if err != nil {
		t.Fatalf(fmtParseErr, err)
	}

	_, err = pattern.Coerce(
		map[string]string{argNameN: valueInvalidInt},
		tokens,
	)
	if err == nil {
		t.Fatal("Coerce: expected error, got nil")
	}

	msg := err.Error()

	// The error message must reference the arg name.
	if !containsSubstring(msg, argNameN) {
		t.Errorf(
			"error message %q does not contain arg name %q",
			msg,
			argNameN,
		)
	}

	// The error message must reference the expected type.
	if !containsSubstring(msg, typeNameInt) {
		t.Errorf(
			"error message %q does not contain expected type %q",
			msg,
			typeNameInt,
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

// assertCoercedInt verifies that result[key] is an int equal to want.
func assertCoercedInt(
	t *testing.T,
	result map[string]any,
	key string,
	want int,
) {
	t.Helper()

	got, ok := result[key].(int)
	if !ok {
		t.Fatalf("%s type: want int, got %T", key, result[key])
	}

	if got != want {
		t.Errorf("%s: want %d, got %d", key, want, got)
	}
}

// assertCoercedString verifies that result[key] is a string equal to want.
func assertCoercedString(
	t *testing.T,
	result map[string]any,
	key, want string,
) {
	t.Helper()

	got, ok := result[key].(string)
	if !ok {
		t.Fatalf("%s type: want string, got %T", key, result[key])
	}

	if got != want {
		t.Errorf("%s: want %q, got %q", key, want, got)
	}
}

// assertCoercedDuration verifies result[key] is a time.Duration equal to want.
func assertCoercedDuration(
	t *testing.T,
	result map[string]any,
	key string,
	want time.Duration,
) {
	t.Helper()

	got, ok := result[key].(time.Duration)
	if !ok {
		t.Fatalf("%s type: want time.Duration, got %T", key, result[key])
	}

	if got != want {
		t.Errorf("%s: want %v, got %v", key, want, got)
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
		t.Fatalf(fmtParseErr, err)
	}

	captures, matched := pattern.Match(tokens, ac3Step)
	if !matched {
		t.Fatal("Match: expected true for AC3 step, got false")
	}

	result, err := pattern.Coerce(captures, tokens)
	if err != nil {
		t.Fatalf(fmtCoerceUnexpectedErr, err)
	}

	const (
		wantConnectorID = 1
		wantTimeout     = 30 * time.Second
	)

	assertCoercedInt(t, result, "connectorId", wantConnectorID)

	assertCoercedString(t, result, "idTag", "X")

	assertCoercedString(t, result, "station", valueValidStation)

	assertCoercedDuration(t, result, captureKeyTimeo, wantTimeout)
}

// ── helpers ────────────────────────────────────────────────────────────

// containsSubstring reports whether s contains sub. Kept as an
// inline helper to avoid importing "strings" solely for test assertions.
func containsSubstring(str, sub string) bool {
	if len(str) < len(sub) {
		return false
	}

	if len(sub) == wantZeroLen {
		return true
	}

	for idx := tokenIdx0; idx <= len(str)-len(sub); idx++ {
		if str[idx:idx+len(sub)] == sub {
			return true
		}
	}

	return false
}
