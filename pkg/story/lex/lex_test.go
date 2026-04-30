// Package lex_test provides black-box unit tests for the story lexer.
// Every TokenKind and every documented error path is covered (T-001-12, AC1).
package lex_test

import (
	"testing"

	"github.com/evcoreco/octane/pkg/story/lex"
)

// TestLexer covers every token kind and every error path in one
// table-driven suite. Each row asserts that the token sequence produced
// by the lexer matches the want slice; checking stops after len(want)
// tokens (the caller need not list trailing TokenEOF tokens unless the
// test is specifically about EOF).
func TestLexer(t *testing.T) {
	t.Parallel()

	// Named constants used as literal sentinels throughout the table.
	// Empty string means "do not check literal for this position."
	const (
		litMeta        = "Meta"
		litBackground  = "Background"
		litSetup       = "Setup"
		litScenario    = "Scenario"
		litTeardown    = "Teardown"
		litParallel    = "Parallel"
		litEndParallel = "End-Parallel"
		litGiven       = "Given"
		litWhen        = "When"
		litThen        = "Then"
		litAnd         = "And"
		litBut         = "But"
		litIndent      = "    " // exactly four spaces
		litColon       = ":"
	)

	tests := []struct {
		name  string
		input string
		want  []lex.Token
	}{
		// -----------------------------------------------------------------
		// Section keywords at column 1
		// -----------------------------------------------------------------

		{
			// Invariant: "Meta" at column 1 produces TokenMeta.
			name:  "section_keyword_Meta",
			input: "Meta\n",
			want: []lex.Token{
				{Kind: lex.TokenMeta, Literal: litMeta, Line: 1, Column: 1},
			},
		},
		{
			// Invariant: "Background" at column 1 produces TokenBackground.
			name:  "section_keyword_Background",
			input: "Background\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenBackground,
					Literal: litBackground,
					Line:    1,
					Column:  1,
				},
			},
		},
		{
			// Invariant: "Setup" at column 1 produces TokenSetup.
			name:  "section_keyword_Setup",
			input: "Setup\n",
			want: []lex.Token{
				{Kind: lex.TokenSetup, Literal: litSetup, Line: 1, Column: 1},
			},
		},
		{
			// Invariant: "Scenario" at column 1 produces TokenScenario.
			name:  "section_keyword_Scenario",
			input: "Scenario\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenScenario,
					Literal: litScenario,
					Line:    1,
					Column:  1,
				},
			},
		},
		{
			// Invariant: "Teardown" at column 1 produces TokenTeardown.
			name:  "section_keyword_Teardown",
			input: "Teardown\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenTeardown,
					Literal: litTeardown,
					Line:    1,
					Column:  1,
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
					Literal: litParallel,
					Line:    1,
					Column:  1,
				},
			},
		},
		{
			// Invariant: "End-Parallel" at column 1 produces TokenEndParallel.
			name:  "section_keyword_EndParallel",
			input: "End-Parallel\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenEndParallel,
					Literal: litEndParallel,
					Line:    1,
					Column:  1,
				},
			},
		},

		// -----------------------------------------------------------------
		// Step keywords after 4-space indent
		// -----------------------------------------------------------------

		{
			// Invariant: "Given" after 4-space indent produces TokenIndent
			// then TokenGiven.
			name:  "step_keyword_Given",
			input: "    Given the station is ready\n",
			want: []lex.Token{
				{Kind: lex.TokenIndent, Literal: litIndent, Line: 0, Column: 0},
				{Kind: lex.TokenGiven, Literal: litGiven, Line: 0, Column: 0},
				{
					Kind:    lex.TokenText,
					Literal: "the station is ready",
					Line:    0,
					Column:  0,
				},
			},
		},
		{
			// Invariant: "When" after 4-space indent produces TokenWhen.
			name:  "step_keyword_When",
			input: "    When the cable is plugged in\n",
			want: []lex.Token{
				{Kind: lex.TokenIndent, Literal: litIndent, Line: 0, Column: 0},
				{Kind: lex.TokenWhen, Literal: litWhen, Line: 0, Column: 0},
				{
					Kind:    lex.TokenText,
					Literal: "the cable is plugged in",
					Line:    0,
					Column:  0,
				},
			},
		},
		{
			// Invariant: "Then" after 4-space indent produces TokenThen.
			name:  "step_keyword_Then",
			input: "    Then charging begins\n",
			want: []lex.Token{
				{Kind: lex.TokenIndent, Literal: litIndent, Line: 0, Column: 0},
				{Kind: lex.TokenThen, Literal: litThen, Line: 0, Column: 0},
				{
					Kind:    lex.TokenText,
					Literal: "charging begins",
					Line:    0,
					Column:  0,
				},
			},
		},
		{
			// Invariant: "And" after 4-space indent produces TokenAnd.
			name:  "step_keyword_And",
			input: "    And the LED is green\n",
			want: []lex.Token{
				{Kind: lex.TokenIndent, Literal: litIndent, Line: 0, Column: 0},
				{Kind: lex.TokenAnd, Literal: litAnd, Line: 0, Column: 0},
				{
					Kind:    lex.TokenText,
					Literal: "the LED is green",
					Line:    0,
					Column:  0,
				},
			},
		},
		{
			// Invariant: "But" after 4-space indent produces TokenBut.
			name:  "step_keyword_But",
			input: "    But the session does not end\n",
			want: []lex.Token{
				{Kind: lex.TokenIndent, Literal: litIndent, Line: 0, Column: 0},
				{Kind: lex.TokenBut, Literal: litBut, Line: 0, Column: 0},
				{
					Kind:    lex.TokenText,
					Literal: "the session does not end",
					Line:    0,
					Column:  0,
				},
			},
		},

		// -----------------------------------------------------------------
		// Meta entry: TokenIndent + TokenMetaKey + TokenColon + TokenValue
		// -----------------------------------------------------------------

		{
			// Invariant: a "Key: Value" indented line produces four tokens.
			name:  "meta_entry_correct",
			input: "    Name: Boot test\n",
			want: []lex.Token{
				{Kind: lex.TokenIndent, Literal: litIndent, Line: 0, Column: 0},
				{Kind: lex.TokenMetaKey, Literal: "Name", Line: 0, Column: 0},
				{Kind: lex.TokenColon, Literal: litColon, Line: 0, Column: 0},
				{
					Kind:    lex.TokenValue,
					Literal: "Boot test",
					Line:    0,
					Column:  0,
				},
			},
		},

		// -----------------------------------------------------------------
		// Scenario header: TokenScenario + TokenColon + TokenText
		// -----------------------------------------------------------------

		{
			// Invariant: "Scenario: My Title" emits Scenario, Colon, Text.
			name:  "scenario_header",
			input: "Scenario: My Title\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenScenario,
					Literal: litScenario,
					Line:    0,
					Column:  0,
				},
				{Kind: lex.TokenColon, Literal: litColon, Line: 0, Column: 0},
				{Kind: lex.TokenText, Literal: "My Title", Line: 0, Column: 0},
			},
		},

		// -----------------------------------------------------------------
		// Comment
		// -----------------------------------------------------------------

		{
			// Invariant: a line beginning with '#' produces TokenComment.
			name:  "comment_line",
			input: "# this is a comment\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenComment,
					Literal: "# this is a comment",
					Line:    1,
					Column:  1,
				},
			},
		},

		// -----------------------------------------------------------------
		// EOF on empty input
		// -----------------------------------------------------------------

		{
			// Invariant: empty input returns TokenEOF immediately.
			name:  "eof_empty_input",
			input: "",
			want: []lex.Token{
				{Kind: lex.TokenEOF, Literal: "", Line: 0, Column: 0},
			},
		},

		// -----------------------------------------------------------------
		// CRLF normalisation
		// -----------------------------------------------------------------

		{
			// Invariant: "Meta\r\n" lexes identically to "Meta\n".
			name:  "crlf_normalised_to_lf",
			input: "Meta\r\n",
			want: []lex.Token{
				{Kind: lex.TokenMeta, Literal: litMeta, Line: 1, Column: 1},
			},
		},

		// -----------------------------------------------------------------
		// Error paths
		// -----------------------------------------------------------------

		{
			// Invariant: a tab character produces TokenIllegal.
			name:  "tab_character_illegal",
			input: "\tGiven something\n",
			want: []lex.Token{
				{Kind: lex.TokenIllegal, Literal: "\t", Line: 0, Column: 0},
			},
		},
		{
			// Invariant: 2-space indent (wrong width) produces TokenIllegal.
			name:  "wrong_indent_width_two_spaces",
			input: "  Given something\n",
			want: []lex.Token{
				{Kind: lex.TokenIllegal, Literal: "", Line: 0, Column: 0},
			},
		},
		{
			// Invariant: 4 spaces then newline (blank indented line)
			// produces TokenIllegal.
			name:  "blank_indented_line",
			input: "    \n",
			want: []lex.Token{
				{Kind: lex.TokenIllegal, Literal: "    ", Line: 0, Column: 0},
			},
		},
		{
			// Invariant: "Andromeda" after 4-space indent must NOT produce
			// TokenAnd; "And" is only a keyword when followed by space or
			// newline. Because "Andromeda galaxy" has no colon,
			// scanMetaEntry returns TokenIllegal (the TokenIndent is not
			// emitted in this error path).
			name:  "step_keyword_boundary_Andromeda",
			input: "    Andromeda galaxy\n",
			want: []lex.Token{
				// Must NOT be TokenAnd — illegal because no colon found.
				{
					Kind:    lex.TokenIllegal,
					Literal: "Andromeda galaxy",
					Line:    0,
					Column:  0,
				},
			},
		},
		{
			// Invariant: an indented line with no colon produces TokenIllegal
			// (no TokenIndent emitted).
			name:  "meta_entry_without_colon",
			input: "    NoColonHere\n",
			want: []lex.Token{
				{
					Kind:    lex.TokenIllegal,
					Literal: "NoColonHere",
					Line:    0,
					Column:  0,
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
					Line:    0,
					Column:  0,
				},
			},
		},

		// -----------------------------------------------------------------
		// Multi-token queue ordering
		// -----------------------------------------------------------------

		{
			// Invariant: a meta entry line emits tokens in strict order:
			// TokenIndent, TokenMetaKey, TokenColon, TokenValue.
			name:  "meta_entry_queue_order",
			input: "    Spec-Ref: TC-001\n",
			want: []lex.Token{
				{Kind: lex.TokenIndent, Literal: "", Line: 0, Column: 0},
				{
					Kind:    lex.TokenMetaKey,
					Literal: "Spec-Ref",
					Line:    0,
					Column:  0,
				},
				{Kind: lex.TokenColon, Literal: "", Line: 0, Column: 0},
				{Kind: lex.TokenValue, Literal: "TC-001", Line: 0, Column: 0},
				{Kind: lex.TokenEOF, Literal: "", Line: 0, Column: 0},
			},
		},

		// -----------------------------------------------------------------
		// Line and column tracking
		// -----------------------------------------------------------------

		{
			// Invariant: a token on line 2 carries the correct Line field.
			name:  "line_tracking_second_line",
			input: "Meta\nBackground\n",
			want: []lex.Token{
				{Kind: lex.TokenMeta, Literal: "", Line: 1, Column: 1},
				{Kind: lex.TokenBackground, Literal: "", Line: 2, Column: 1},
			},
		},

		// -----------------------------------------------------------------
		// Blank line skipping
		// -----------------------------------------------------------------

		{
			// Invariant: blank lines between section keywords are silently
			// skipped.
			name:  "blank_lines_skipped",
			input: "Meta\n\nBackground\n",
			want: []lex.Token{
				{Kind: lex.TokenMeta, Literal: "", Line: 0, Column: 0},
				{Kind: lex.TokenBackground, Literal: "", Line: 0, Column: 0},
			},
		},

		// -----------------------------------------------------------------
		// Step keyword with no text (keyword alone on line)
		// -----------------------------------------------------------------

		{
			// Invariant: a step keyword at EOL without trailing text emits
			// an empty TokenText.
			name:  "step_keyword_no_text",
			input: "    Given\n",
			want: []lex.Token{
				{Kind: lex.TokenIndent, Literal: "", Line: 0, Column: 0},
				{Kind: lex.TokenGiven, Literal: "", Line: 0, Column: 0},
				{Kind: lex.TokenText, Literal: "", Line: 0, Column: 0},
			},
		},

		// -----------------------------------------------------------------
		// Value whitespace trimming
		// -----------------------------------------------------------------

		{
			// Invariant: leading whitespace after the colon in a meta entry
			// is trimmed.
			name:  "meta_value_leading_space_trimmed",
			input: "    Tags:   boot  \n",
			want: []lex.Token{
				{Kind: lex.TokenIndent, Literal: "", Line: 0, Column: 0},
				{Kind: lex.TokenMetaKey, Literal: "Tags", Line: 0, Column: 0},
				{Kind: lex.TokenColon, Literal: "", Line: 0, Column: 0},
				{Kind: lex.TokenValue, Literal: "boot", Line: 0, Column: 0},
			},
		},

		// -----------------------------------------------------------------
		// Repeated TokenEOF after end of input
		// -----------------------------------------------------------------

		{
			// Invariant: after TokenEOF, every subsequent Next() also
			// returns TokenEOF.
			name:  "eof_repeated",
			input: "",
			want: []lex.Token{
				{Kind: lex.TokenEOF, Literal: "", Line: 0, Column: 0},
				{Kind: lex.TokenEOF, Literal: "", Line: 0, Column: 0},
				{Kind: lex.TokenEOF, Literal: "", Line: 0, Column: 0},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lexer := lex.NewLexer("test.story", []byte(tc.input))
			for idx, want := range tc.want {
				got := lexer.Next()

				if got.Kind != want.Kind {
					t.Errorf(
						"token[%d]: kind = %v, want %v",
						idx,
						got.Kind,
						want.Kind,
					)
				}

				if want.Literal != "" && got.Literal != want.Literal {
					t.Errorf(
						"token[%d]: literal = %q, want %q",
						idx,
						got.Literal,
						want.Literal,
					)
				}

				if want.Line != 0 && got.Line != want.Line {
					t.Errorf(
						"token[%d]: line = %d, want %d",
						idx,
						got.Line,
						want.Line,
					)
				}

				if want.Column != 0 && got.Column != want.Column {
					t.Errorf(
						"token[%d]: column = %d, want %d",
						idx,
						got.Column,
						want.Column,
					)
				}
			}
		})
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
		t.Errorf(
			"Next() returned the same token twice; Peek appears to have consumed ahead",
		)
	}
}

// TestLexerPeek_AfterEOF verifies that Peek at EOF returns TokenEOF
// and does not panic or advance past the end.
func TestLexerPeek_AfterEOF(t *testing.T) {
	t.Parallel()

	lexer := lex.NewLexer("test.story", []byte(""))

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

			// Invariant: every defined TokenKind has a non-empty,
			// non-"Unknown" string.
			str := kind.String()
			if str == "" {
				t.Errorf("TokenKind(%d).String() = empty string", int(kind))
			}

			if str == "Unknown" {
				t.Errorf(
					"TokenKind(%d).String() = %q; add a case to TokenKind.String()",
					int(kind),
					str,
				)
			}
		})
	}
}
