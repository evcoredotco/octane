// Package lex defines the token types and lexer interface for the .story
// DSL. This package is a leaf in the dependency graph: it imports nothing
// from the octane module and relies only on the standard library.
//
// The lexer contract is defined here (T-001-04); the byte-stream
// implementation is provided by T-001-10.
package lex

// TokenKind identifies the type of a lexical token produced by the lexer.
type TokenKind int

const (
	// TokenIllegal represents an unrecognised byte sequence.
	TokenIllegal TokenKind = iota

	// TokenEOF signals the end of the input.
	TokenEOF

	// TokenNewline represents a single newline character (after CRLF
	// normalisation to LF).
	TokenNewline

	// TokenComment represents a comment line starting with '#', up to
	// but not including the terminating newline.
	TokenComment

	// TokenIndent represents leading whitespace (exactly four spaces)
	// at the start of an indented line.
	TokenIndent

	// --- Section-level keywords (appear at column 1, unindented) ---

	// TokenMeta introduces the Meta section.
	TokenMeta

	// TokenBackground introduces the Background section.
	TokenBackground

	// TokenSetup introduces the Setup section.
	TokenSetup

	// TokenScenario introduces a Scenario section.
	TokenScenario

	// TokenTeardown introduces the Teardown section.
	TokenTeardown

	// TokenParallel introduces a parallel multi-station block (reserved
	// for future use).
	TokenParallel

	// TokenEndParallel closes a parallel multi-station block (reserved
	// for future use).
	TokenEndParallel

	// --- Step keywords (appear indented inside a section body) ---

	// TokenGiven introduces a precondition step.
	TokenGiven

	// TokenWhen introduces an action step.
	TokenWhen

	// TokenThen introduces an expected-outcome step.
	TokenThen

	// TokenAnd continues the preceding step kind.
	TokenAnd

	// TokenBut introduces a negative continuation of the preceding step kind.
	TokenBut

	// --- Meta-section tokens ---

	// TokenMetaKey represents a meta-header identifier such as "Name",
	// "Id", "Spec-Ref", or "Tags".
	TokenMetaKey

	// TokenColon represents the ':' separator between a meta key and
	// its value.
	TokenColon

	// TokenValue represents the trimmed text after the colon on a meta
	// line.
	TokenValue

	// --- Step text ---

	// TokenText represents the verbatim step text that follows a step
	// keyword.
	TokenText
)

// String returns a human-readable name for the token kind. The returned
// string is useful in error messages and debug output.
func (k TokenKind) String() string {
	switch k {
	case TokenIllegal:
		return "Illegal"
	case TokenEOF:
		return "EOF"
	case TokenNewline:
		return "Newline"
	case TokenComment:
		return "Comment"
	case TokenIndent:
		return "Indent"
	case TokenMeta:
		return "Meta"
	case TokenBackground:
		return "Background"
	case TokenSetup:
		return "Setup"
	case TokenScenario:
		return "Scenario"
	case TokenTeardown:
		return "Teardown"
	case TokenParallel:
		return "Parallel"
	case TokenEndParallel:
		return "EndParallel"
	case TokenGiven:
		return "Given"
	case TokenWhen:
		return "When"
	case TokenThen:
		return "Then"
	case TokenAnd:
		return "And"
	case TokenBut:
		return "But"
	case TokenMetaKey:
		return "MetaKey"
	case TokenColon:
		return "Colon"
	case TokenValue:
		return "Value"
	case TokenText:
		return "Text"
	default:
		return "Unknown"
	}
}

// Token is the smallest unit produced by the lexer. Each token carries its
// kind, the exact source bytes that comprise it, and the 1-based line and
// column where it begins.
type Token struct {
	// Kind identifies the token type.
	Kind TokenKind

	// Literal holds the exact source bytes for this token.
	Literal string

	// Line is the 1-based line number where the token starts.
	Line int

	// Column is the 1-based byte offset from the start of the line
	// where the token starts.
	Column int
}

// Lexer tokenises a .story source byte slice. Implementations must
// normalise CRLF to LF before tokenising. After TokenEOF is returned,
// every subsequent call to Next must also return TokenEOF.
type Lexer interface {
	// Next returns the next token. After TokenEOF is returned,
	// subsequent calls continue to return TokenEOF.
	Next() Token

	// Peek returns the next token without consuming it. Consecutive
	// calls to Peek without an intervening Next return the same token.
	Peek() Token
}

// NewLexer returns a Lexer that tokenises src. The file parameter is used
// only for error messages. The implementation is provided by T-001-10.
func NewLexer(file string, src []byte) Lexer {
	panic("not yet implemented: see T-001-10")
}
