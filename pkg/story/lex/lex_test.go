// Package lex_test provides black-box unit tests for the story lexer.
// Every TokenKind and every documented error path is covered (T-001-12, AC1).
package lex_test

import (
	"testing"

	"github.com/evcoreco/octane/pkg/story/lex"
)

// Package-level sentinel constants shared across all test functions.
// Using consts (not vars) avoids the gochecknoglobals linter rule.
const (
	// emptyLiteral is the sentinel value meaning "do not check Literal".
	emptyLiteral = ""
	// unknownKind is the string that TokenKind.String() must never return.
	unknownKind = "Unknown"
	// fourSpaces is the four-space indent string used by the story DSL.
	fourSpaces = "    "
	// twoTokens is the numeric constant 2 used to avoid magic numbers.
	twoTokens = 2
	// zeroLineCol is the zero value for both Line and Column in Token literals.
	zeroLineCol = 0
	// firstLineCol is the value 1 used for the first Line/Column position.
	firstLineCol = 1
)

// runLexerTable is the shared table-runner used by all TestLexer_* functions.
// It iterates want and asserts each position via assertTokenAt.
func runLexerTable(
	t *testing.T,
	tests []struct {
		name  string
		input string
		want  []lex.Token
	},
) {
	t.Helper()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := lex.NewLexer("test.story", []byte(tc.input))
			for idx, want := range tc.want {
				assertTokenAt(t, idx, l.Next(), want)
			}
		})
	}
}

// TestLexer_SectionKeywordsMetaToScenario covers the Meta, Background,
// Setup, and Scenario section-level keywords at column 1.
func TestLexer_SectionKeywordsMetaToScenario(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []lex.Token
	}{
		{
			// Invariant: "Meta" at column 1 produces TokenMeta.
			name:  "section_keyword_Meta",
			input: "Meta\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenMeta,
					Literal: "Meta",
					Line:    firstLineCol,
					Column:  firstLineCol,
				},
			},
		},
		{
			// Invariant: "Background" at column 1 produces TokenBackground.
			name:  "section_keyword_Background",
			input: "Background\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenBackground,
					Literal: "Background",
					Line:    firstLineCol,
					Column:  firstLineCol,
				},
			},
		},
		{
			// Invariant: "Setup" at column 1 produces TokenSetup.
			name:  "section_keyword_Setup",
			input: "Setup\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenSetup,
					Literal: "Setup",
					Line:    firstLineCol,
					Column:  firstLineCol,
				},
			},
		},
		{
			// Invariant: "Scenario" at column 1 produces TokenScenario.
			name:  "section_keyword_Scenario",
			input: "Scenario\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenScenario,
					Literal: "Scenario",
					Line:    firstLineCol,
					Column:  firstLineCol,
				},
			},
		},
	}

	runLexerTable(t, tests)
}

// TestLexer_SectionKeywordsTeardownToEndParallel covers Teardown, Parallel,
// and End-Parallel section-level keywords at column 1.
func TestLexer_SectionKeywordsTeardownToEndParallel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []lex.Token
	}{
		{
			// Invariant: "Teardown" at column 1 produces TokenTeardown.
			name:  "section_keyword_Teardown",
			input: "Teardown\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenTeardown,
					Literal: "Teardown",
					Line:    firstLineCol,
					Column:  firstLineCol,
				},
			},
		},
		{
			// Invariant: "Parallel" at column 1 produces TokenParallel.
			name:  "section_keyword_Parallel",
			input: "Parallel\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenParallel,
					Literal: "Parallel",
					Line:    firstLineCol,
					Column:  firstLineCol,
				},
			},
		},
		{
			// Invariant: "End-Parallel" at column 1 produces
			// TokenEndParallel.
			name:  "section_keyword_EndParallel",
			input: "End-Parallel\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenEndParallel,
					Literal: "End-Parallel",
					Line:    firstLineCol,
					Column:  firstLineCol,
				},
			},
		},
	}

	runLexerTable(t, tests)
}

// TestLexer_StepKeywordGiven covers the Given step keyword after a 4-space
// indent.
func TestLexer_StepKeywordGiven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []lex.Token
	}{
		{
			// Invariant: "Given" after 4-space indent produces TokenIndent
			// then TokenGiven then TokenText.
			name:  "step_keyword_Given",
			input: "    Given the station is ready\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenIndent,
					Literal: fourSpaces,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenGiven,
					Literal: "Given",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenText,
					Literal: "the station is ready",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
			},
		},
	}

	runLexerTable(t, tests)
}

// TestLexer_StepKeywordWhen covers the When step keyword after a 4-space
// indent.
func TestLexer_StepKeywordWhen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []lex.Token
	}{
		{
			// Invariant: "When" after 4-space indent produces TokenWhen.
			name:  "step_keyword_When",
			input: "    When the cable is plugged in\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenIndent,
					Literal: fourSpaces,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenWhen,
					Literal: "When",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenText,
					Literal: "the cable is plugged in",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
			},
		},
	}

	runLexerTable(t, tests)
}

// TestLexer_StepKeywordThen covers the Then step keyword after a 4-space
// indent.
func TestLexer_StepKeywordThen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []lex.Token
	}{
		{
			// Invariant: "Then" after 4-space indent produces TokenThen.
			name:  "step_keyword_Then",
			input: "    Then charging begins\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenIndent,
					Literal: fourSpaces,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenThen,
					Literal: "Then",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenText,
					Literal: "charging begins",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
			},
		},
	}

	runLexerTable(t, tests)
}

// TestLexer_StepKeywordsAndBut covers And and But step keywords after a
// 4-space indent.
func TestLexer_StepKeywordsAndBut(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []lex.Token
	}{
		{
			// Invariant: "And" after 4-space indent produces TokenAnd.
			name:  "step_keyword_And",
			input: "    And the LED is green\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenIndent,
					Literal: fourSpaces,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenAnd,
					Literal: "And",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenText,
					Literal: "the LED is green",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
			},
		},
		{
			// Invariant: "But" after 4-space indent produces TokenBut.
			name:  "step_keyword_But",
			input: "    But the session does not end\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenIndent,
					Literal: fourSpaces,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenBut,
					Literal: "But",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenText,
					Literal: "the session does not end",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
			},
		},
	}

	runLexerTable(t, tests)
}

// TestLexer_MetaEntry covers the meta-entry token sequence produced by a
// "Key: Value" indented line.
func TestLexer_MetaEntry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []lex.Token
	}{
		{
			// Invariant: a "Key: Value" indented line produces four tokens.
			name:  "meta_entry_correct",
			input: "    Name: Boot test\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenIndent,
					Literal: fourSpaces,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenMetaKey,
					Literal: "Name",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenColon,
					Literal: ":",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenValue,
					Literal: "Boot test",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
			},
		},
	}

	runLexerTable(t, tests)
}

// TestLexer_ScenarioHeader covers the scenario-header token sequence:
// TokenScenario, TokenColon, TokenText.
func TestLexer_ScenarioHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []lex.Token
	}{
		{
			// Invariant: "Scenario: My Title" emits Scenario, Colon, Text.
			name:  "scenario_header",
			input: "Scenario: My Title\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenScenario,
					Literal: "Scenario",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenColon,
					Literal: ":",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenText,
					Literal: "My Title",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
			},
		},
	}

	runLexerTable(t, tests)
}

// TestLexer_CommentAndEOF covers comment lines, empty-input EOF, and CRLF
// normalisation.
func TestLexer_CommentAndEOF(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []lex.Token
	}{
		{
			// Invariant: a line beginning with '#' produces TokenComment.
			name:  "comment_line",
			input: "# this is a comment\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenComment,
					Literal: "# this is a comment",
					Line:    firstLineCol,
					Column:  firstLineCol,
				},
			},
		},
		{
			// Invariant: empty input returns TokenEOF immediately.
			name:  "eof_empty_input",
			input: emptyLiteral,
			want: []lex.Token{
				{
					Kind:    lex.TokenEOF,
					Literal: emptyLiteral,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
			},
		},
		{
			// Invariant: "Meta\r\n" lexes identically to "Meta\n".
			name:  "crlf_normalised_to_lf",
			input: "Meta\r\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenMeta,
					Literal: "Meta",
					Line:    firstLineCol,
					Column:  firstLineCol,
				},
			},
		},
	}

	runLexerTable(t, tests)
}

// TestLexer_ErrorPathsIndent covers illegal-token cases caused by bad
// indentation: tabs, wrong indent width, and blank indented lines.
func TestLexer_ErrorPathsIndent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []lex.Token
	}{
		{
			// Invariant: a tab character produces TokenIllegal.
			name:  "tab_character_illegal",
			input: "\tGiven something\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenIllegal,
					Literal: "\t",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
			},
		},
		{
			// Invariant: 2-space indent (wrong width) produces TokenIllegal.
			name:  "wrong_indent_width_two_spaces",
			input: "  Given something\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenIllegal,
					Literal: emptyLiteral,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
			},
		},
		{
			// Invariant: 4 spaces then newline (blank indented line)
			// produces TokenIllegal.
			name:  "blank_indented_line",
			input: "    \n",
			want: []lex.Token{
				{
					Kind:    lex.TokenIllegal,
					Literal: fourSpaces,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
			},
		},
	}

	runLexerTable(t, tests)
}

// TestLexer_ErrorPathsContent covers illegal-token cases caused by content
// violations: keyword boundary, missing colon, and unknown unindented text.
func TestLexer_ErrorPathsContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []lex.Token
	}{
		{
			// Invariant: "Andromeda" after 4-space indent must NOT produce
			// TokenAnd; "And" is only a keyword when followed by space or
			// newline.
			name:  "step_keyword_boundary_Andromeda",
			input: "    Andromeda galaxy\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenIllegal,
					Literal: "Andromeda galaxy",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
			},
		},
		{
			// Invariant: an indented line with no colon produces
			// TokenIllegal (no TokenIndent emitted).
			name:  "meta_entry_without_colon",
			input: "    NoColonHere\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenIllegal,
					Literal: "NoColonHere",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
			},
		},
		{
			// Invariant: unknown unindented content produces TokenIllegal.
			name:  "unknown_unindented_content",
			input: "NotAKeyword\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenIllegal,
					Literal: "NotAKeyword",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
			},
		},
	}

	runLexerTable(t, tests)
}

// TestLexer_QueueOrdering verifies multi-token queue ordering: a meta entry
// line emits tokens in strict sequence.
func TestLexer_QueueOrdering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []lex.Token
	}{
		{
			// Invariant: a meta entry line emits tokens in strict order:
			// TokenIndent, TokenMetaKey, TokenColon, TokenValue.
			name:  "meta_entry_queue_order",
			input: "    Spec-Ref: TC-001\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenIndent,
					Literal: emptyLiteral,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenMetaKey,
					Literal: "Spec-Ref",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenColon,
					Literal: emptyLiteral,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenValue,
					Literal: "TC-001",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenEOF,
					Literal: emptyLiteral,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
			},
		},
	}

	runLexerTable(t, tests)
}

// TestLexer_LineTracking verifies that tokens on the second line carry the
// correct Line field.
func TestLexer_LineTracking(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []lex.Token
	}{
		{
			// Invariant: a token on line 2 carries the correct Line field.
			name:  "line_tracking_second_line",
			input: "Meta\nBackground\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenMeta,
					Literal: emptyLiteral,
					Line:    firstLineCol,
					Column:  firstLineCol,
				},
				{
					Kind:    lex.TokenBackground,
					Literal: emptyLiteral,
					Line:    twoTokens,
					Column:  firstLineCol,
				},
			},
		},
	}

	runLexerTable(t, tests)
}

// TestLexer_BlankLinesAndStepNoText covers blank-line skipping and a step
// keyword at EOL without trailing text.
func TestLexer_BlankLinesAndStepNoText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []lex.Token
	}{
		{
			// Invariant: blank lines between section keywords are silently
			// skipped.
			name:  "blank_lines_skipped",
			input: "Meta\n\nBackground\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenMeta,
					Literal: emptyLiteral,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenBackground,
					Literal: emptyLiteral,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
			},
		},
		{
			// Invariant: a step keyword at EOL without trailing text emits
			// an empty TokenText.
			name:  "step_keyword_no_text",
			input: "    Given\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenIndent,
					Literal: emptyLiteral,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenGiven,
					Literal: emptyLiteral,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenText,
					Literal: emptyLiteral,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
			},
		},
	}

	runLexerTable(t, tests)
}

// TestLexer_MetaValueTrimming covers leading-whitespace trimming of meta
// values after the colon.
func TestLexer_MetaValueTrimming(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []lex.Token
	}{
		{
			// Invariant: leading whitespace after the colon in a meta entry
			// is trimmed.
			name:  "meta_value_leading_space_trimmed",
			input: "    Tags:   boot  \n",
			want: []lex.Token{
				{
					Kind:    lex.TokenIndent,
					Literal: emptyLiteral,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenMetaKey,
					Literal: "Tags",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenColon,
					Literal: emptyLiteral,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenValue,
					Literal: "boot",
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
			},
		},
	}

	runLexerTable(t, tests)
}

// TestLexer_EOFRepeated verifies that every Next() call after the first
// TokenEOF also returns TokenEOF (idempotent EOF contract).
func TestLexer_EOFRepeated(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []lex.Token
	}{
		{
			// Invariant: after TokenEOF, every subsequent Next() also
			// returns TokenEOF.
			name:  "eof_repeated",
			input: emptyLiteral,
			want: []lex.Token{
				{
					Kind:    lex.TokenEOF,
					Literal: emptyLiteral,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenEOF,
					Literal: emptyLiteral,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
				{
					Kind:    lex.TokenEOF,
					Literal: emptyLiteral,
					Line:    zeroLineCol,
					Column:  zeroLineCol,
				},
			},
		},
	}

	runLexerTable(t, tests)
}

// assertTokenAt checks that got matches want at position idx.
// It is called from the inner loop of the TestLexer_* functions to keep
// each loop body flat and below the cognitive-complexity limit.
// When want.Literal == emptyLiteral the literal field is not checked.
// When want.Line / want.Column == 0 the position field is not checked.
func assertTokenAt(t *testing.T, idx int, got, want lex.Token) {
	t.Helper()

	if got.Kind != want.Kind {
		t.Errorf(
			"token[%d]: kind = %v, want %v",
			idx,
			got.Kind,
			want.Kind,
		)
	}

	if want.Literal != emptyLiteral && got.Literal != want.Literal {
		t.Errorf(
			"token[%d]: literal = %q, want %q",
			idx,
			got.Literal,
			want.Literal,
		)
	}

	if want.Line != zeroLineCol && got.Line != want.Line {
		t.Errorf(
			"token[%d]: line = %d, want %d",
			idx,
			got.Line,
			want.Line,
		)
	}

	if want.Column != zeroLineCol && got.Column != want.Column {
		t.Errorf(
			"token[%d]: column = %d, want %d",
			idx,
			got.Column,
			want.Column,
		)
	}
}

// TestLexerPeek verifies the Peek contract: consecutive calls to Peek
// without an intervening Next return the same token, and Next after
// Peek returns that same token again (consuming it).
func TestLexerPeek(t *testing.T) {
	t.Parallel()

	const input = "Meta\nBackground\n"

	lexer := lex.NewLexer("test.story", []byte(input))

	// Invariant: two consecutive Peek calls return the same token.
	first := lexer.Peek()
	second := lexer.Peek()

	if first.Kind != second.Kind {
		t.Errorf(
			"Peek[0].Kind = %v, Peek[1].Kind = %v; want identical",
			first.Kind,
			second.Kind,
		)
	}

	if first.Literal != second.Literal {
		t.Errorf(
			"Peek[0].Literal = %q, Peek[1].Literal = %q; want identical",
			first.Literal,
			second.Literal,
		)
	}

	// Invariant: Next after Peek returns the same token (consumes the peek).
	consumed := lexer.Next()

	if consumed.Kind != first.Kind {
		t.Errorf("Next() after Peek() = %v, want %v", consumed.Kind, first.Kind)
	}

	if consumed.Literal != first.Literal {
		t.Errorf(
			"Next() after Peek() literal = %q, want %q",
			consumed.Literal,
			first.Literal,
		)
	}

	// Invariant: Peek does not advance the stream; Next after consuming
	// the peeked token returns the following token, not a repeat.
	next := lexer.Next()

	if next.Kind == first.Kind && next.Literal == first.Literal {
		t.Error(
			"Next() returned the same token twice;" +
				" Peek appears to have consumed ahead",
		)
	}
}

// TestLexerPeek_AfterEOF verifies that Peek at EOF returns TokenEOF
// and does not panic or advance past the end.
func TestLexerPeek_AfterEOF(t *testing.T) {
	t.Parallel()

	lexer := lex.NewLexer("test.story", []byte(emptyLiteral))

	// Invariant: Peek on empty input returns TokenEOF.
	p := lexer.Peek()
	if p.Kind != lex.TokenEOF {
		t.Errorf("Peek() on empty input = %v, want %v", p.Kind, lex.TokenEOF)
	}

	// Invariant: Next after Peek at EOF also returns TokenEOF.
	n := lexer.Next()
	if n.Kind != lex.TokenEOF {
		t.Errorf(
			"Next() after Peek() at EOF = %v, want %v",
			n.Kind,
			lex.TokenEOF,
		)
	}
}

// assertKindStringValid checks that a TokenKind's String() result is neither
// empty nor the sentinel "Unknown" value.
func assertKindStringValid(t *testing.T, kind lex.TokenKind) {
	t.Helper()

	str := kind.String()

	if str == emptyLiteral {
		t.Errorf("TokenKind(%d).String() = empty string", int(kind))
	}

	if str == unknownKind {
		t.Errorf(
			"TokenKind(%d).String() = %q; add a case to tokenKindName()",
			int(kind),
			str,
		)
	}
}

// TestLexer_TokenKindString verifies that TokenKind.String() returns a
// non-empty, non-"Unknown" value for every defined token kind.
func TestLexer_TokenKindString(t *testing.T) {
	t.Parallel()

	kinds := []lex.TokenKind{
		lex.TokenIllegal,
		lex.TokenEOF,
		lex.TokenNewline,
		lex.TokenComment,
		lex.TokenIndent,
		lex.TokenMeta,
		lex.TokenBackground,
		lex.TokenSetup,
		lex.TokenScenario,
		lex.TokenTeardown,
		lex.TokenParallel,
		lex.TokenEndParallel,
		lex.TokenGiven,
		lex.TokenWhen,
		lex.TokenThen,
		lex.TokenAnd,
		lex.TokenBut,
		lex.TokenMetaKey,
		lex.TokenColon,
		lex.TokenValue,
		lex.TokenText,
	}

	for _, kind := range kinds {
		t.Run(kind.String(), func(t *testing.T) {
			t.Parallel()

			assertKindStringValid(t, kind)
		})
	}
}
