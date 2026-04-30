// Package pattern implements the {name:type} placeholder parser
// for OCTANE keyword patterns.
//
// A keyword pattern is a string composed of two kinds of segments:
// literal text (matched case-insensitively against step text) and
// typed placeholders of the form {name:type}. [Parse] splits a
// pattern string into an ordered slice of [Token] values so that
// the pattern matcher (see match.go) and the type coercer (see
// coerce.go) can operate on structured data rather than raw
// strings.
//
// Supported placeholder types are: string, int, float, bool,
// duration, station, and any. Any other type token is rejected
// with an error at parse time so that registration-time failures
// surface before any step is executed.
package pattern

import (
	"errors"
	"fmt"
	"strings"
)

// Sentinel errors returned by Parse and parsePlaceholder.
var (
	errBareCloseBrace   = errors.New("unexpected '}'")
	errUnclosedBrace    = errors.New("unclosed '{'")
	errEmptyPattern     = errors.New("pattern must not be empty")
	errEmptyPlaceholder = errors.New("empty placeholder")
	errMissingColon     = errors.New(
		"placeholder is missing the colon separator; use {name:type} syntax",
	)
	errEmptyName = errors.New("placeholder has an empty name")
	errEmptyType = errors.New(
		"placeholder has an empty type; " +
			"supported types: string, int, float, bool, duration, station, any",
	)
	errUnknownType = errors.New(
		"placeholder declares unknown type; " +
			"supported types: string, int, float, bool, duration, station, any",
	)
)

// Kind indicates whether a [Token] represents a literal segment or
// a typed placeholder.
type Kind int

const (
	// KindLiteral marks a token whose text must match the step
	// verbatim (case-insensitively, with flexible whitespace).
	KindLiteral Kind = iota + 1

	// KindPlaceholder marks a token that captures one word from
	// the step text and coerces it to the declared [PlaceholderType].
	KindPlaceholder
)

// emptyString is the named constant for an empty string, required
// by the add-constant linter rule when "" appears three or more times.
const emptyString = ""

// minInitialCap is the minimum initial capacity for the token slice.
const minInitialCap = 1

// capacityDivisor is the divisor for the capacity heuristic.
const capacityDivisor = 4

// zeroIdx is the zero index / empty-length sentinel.
const zeroIdx = 0

// fmtPlaceholderError is the repeated format string used when a
// placeholder is malformed; it is extracted to avoid the add-constant
// lint violation for string literals that appear three or more times.
const fmtPlaceholderError = "placeholder %q: %w in pattern %q"

// notFound is the sentinel value returned by strings.IndexByte when
// a byte is not found.
const notFound = -1

// bracketWidth is the character width of a single brace character
// '{' or '}', used for inclusive/exclusive slice bounds.
const bracketWidth = 1

// String returns a human-readable label for the Kind value.
func (k Kind) String() string {
	switch k {
	case KindLiteral:
		return "literal"
	case KindPlaceholder:
		return "placeholder"
	default:
		return "unknown"
	}
}

// PlaceholderType is the type declared inside a {name:type}
// placeholder. OCTANE supports exactly seven placeholder types;
// any other value is rejected by [Parse].
type PlaceholderType string

const (
	// TypeString accepts any whitespace-delimited token.
	TypeString PlaceholderType = "string"

	// TypeInt accepts a token that can be parsed as a base-10
	// integer (no fraction, no exponent).
	TypeInt PlaceholderType = "int"

	// TypeFloat accepts a token that can be parsed as a
	// float64 (decimal notation, no complex or scientific
	// notation for v1).
	TypeFloat PlaceholderType = "float"

	// TypeBool accepts "true" or "false" (case-insensitive).
	TypeBool PlaceholderType = "bool"

	// TypeDuration accepts tokens parseable by
	// [time.ParseDuration] (e.g., "30s", "1m30s", "500ms").
	TypeDuration PlaceholderType = "duration"

	// TypeStation accepts any whitespace-delimited token and
	// signals to the resolver that the captured value is a
	// station handle to be looked up in the runtime state. The
	// wire representation is a Go string; the semantic distinction
	// allows the resolver to validate handle existence.
	TypeStation PlaceholderType = "station"

	// TypeAny accepts any whitespace-delimited token and stores
	// it as a raw string without coercion. Use sparingly; prefer
	// a concrete type whenever the expected form is known.
	TypeAny PlaceholderType = "any"
)

// validTypes returns the closed set of accepted placeholder type tokens.
// Keeping it as a map[PlaceholderType]struct{} gives O(1) lookup
// without regexp.
func validTypes() map[PlaceholderType]struct{} {
	return map[PlaceholderType]struct{}{
		TypeString:   {},
		TypeInt:      {},
		TypeFloat:    {},
		TypeBool:     {},
		TypeDuration: {},
		TypeStation:  {},
		TypeAny:      {},
	}
}

// Token is one segment of a parsed keyword pattern. Every pattern
// is a slice of tokens in left-to-right order; the matcher
// consumes step-text words against the token sequence.
//
// For a [KindLiteral] token only [Token.Text] is meaningful.
// For a [KindPlaceholder] token [Token.Name] and [Token.Type]
// carry the placeholder's name and declared type; [Token.Text] is
// set to the raw placeholder string (e.g., "{count:int}") for
// diagnostic use.
type Token struct {
	// Kind distinguishes a literal segment from a typed
	// placeholder.
	Kind Kind

	// Text is the raw string for this token as it appeared in
	// the pattern. For literals it is the literal word sequence;
	// for placeholders it is the full "{name:type}" string.
	Text string

	// Name is the placeholder name (e.g., "count" in
	// {count:int}). Empty for literal tokens.
	Name string

	// Type is the declared placeholder type. Zero for literal
	// tokens.
	Type PlaceholderType
}

// parseResult carries the output of one parseStep iteration.
type parseResult struct {
	// tokens is the updated token slice after processing.
	tokens []Token

	// remaining is the unprocessed portion of the pattern string.
	remaining string

	// done is true when the end of the pattern has been reached.
	done bool
}

// Parse splits a keyword pattern string into an ordered slice of
// [Token] values. It returns an error if the pattern contains any
// malformed placeholder.
//
// A placeholder is malformed when:
//   - it is empty (i.e., "{}").
//   - the colon separator is missing (i.e., "{name}").
//   - the name part is empty (i.e., "{:type}").
//   - the type part is empty (i.e., "{name:}").
//   - the type is not one of the seven supported values.
//   - a '{' is opened but never closed.
//
// Literal segments may contain any rune except '{' and '}'. A
// bare '}' with no preceding '{' is also an error.
//
// Parse pre-allocates a reasonable number of token slots to avoid
// unnecessary growth for typical patterns. Callers that cache
// parsed patterns (e.g., the registry) should store the returned
// slice directly; it is safe for concurrent read access after
// Parse returns.
func Parse(pattern string) ([]Token, error) {
	// Pre-allocate at least one slot; typical patterns are short.
	initialCap := max(len(pattern)/capacityDivisor, minInitialCap)

	tokens := make([]Token, zeroIdx, initialCap)

	remaining := pattern

	for len(remaining) > zeroIdx {
		res, err := parseStep(tokens, remaining, pattern)
		if err != nil {
			return nil, err
		}

		tokens = res.tokens
		remaining = res.remaining

		if res.done {
			break
		}
	}

	if len(tokens) == zeroIdx {
		return nil, errEmptyPattern
	}

	return tokens, nil
}

// parseStep processes one iteration of the Parse loop. It returns a
// parseResult carrying the updated tokens, remaining string, and done
// flag, plus any parse error.
func parseStep(
	tokens []Token,
	remaining string,
	pattern string,
) (parseResult, error) {
	openIdx := strings.IndexByte(remaining, '{')
	closeIdx := strings.IndexByte(remaining, '}')

	emptyResult := parseResult{
		tokens:    tokens,
		remaining: emptyString,
		done:      false,
	}

	err := checkBareClose(closeIdx, openIdx, remaining, pattern)
	if err != nil {
		return emptyResult, err
	}

	if openIdx == notFound {
		if lit := makeLiteral(remaining); lit.Text != emptyString {
			tokens = append(tokens, lit)
		}

		return parseResult{
			tokens:    tokens,
			remaining: emptyString,
			done:      true,
		}, nil
	}

	tokens = appendLeadingLiteral(tokens, remaining, openIdx)

	remaining = remaining[openIdx:]

	tok, newRemaining, err := parsePlaceholderFromRemaining(remaining, pattern)
	if err != nil {
		return emptyResult, err
	}

	return parseResult{
		tokens:    append(tokens, tok),
		remaining: newRemaining,
		done:      false,
	}, nil
}

// checkBareClose returns an error when a '}' appears before any '{'.
func checkBareClose(
	closeIdx, openIdx int,
	remaining, pattern string,
) error {
	isBareClose := closeIdx != notFound &&
		(openIdx == notFound || closeIdx < openIdx)

	if !isBareClose {
		return nil
	}

	return fmt.Errorf(
		"%w at position %d in pattern %q",
		errBareCloseBrace,
		len(pattern)-len(remaining)+closeIdx,
		pattern,
	)
}

// appendLeadingLiteral appends the literal segment before the first '{'
// to tokens when openIdx > 0.
func appendLeadingLiteral(
	tokens []Token,
	remaining string,
	openIdx int,
) []Token {
	if openIdx <= zeroIdx {
		return tokens
	}

	lit := makeLiteral(remaining[:openIdx])
	if lit.Text != emptyString {
		tokens = append(tokens, lit)
	}

	return tokens
}

// parsePlaceholderFromRemaining extracts the placeholder token from remaining
// (which starts with '{'), advances remaining past the closing '}', and
// returns the token and the new remaining string.
func parsePlaceholderFromRemaining(
	remaining, pattern string,
) (Token, string, error) {
	closeIdx := strings.IndexByte(remaining, '}')
	if closeIdx == notFound {
		return Token{}, emptyString,
			fmt.Errorf("%w in pattern %q", errUnclosedBrace, pattern)
	}

	raw := remaining[:closeIdx+bracketWidth]

	tok, err := parsePlaceholder(raw, pattern)
	if err != nil {
		return Token{}, emptyString, err
	}

	return tok, remaining[closeIdx+bracketWidth:], nil
}

// makeLiteral builds a [KindLiteral] token from a raw string
// segment. The text is trimmed of leading and trailing whitespace
// before storage; internal whitespace is preserved so the matcher
// can treat it as a word boundary.
func makeLiteral(raw string) Token {
	return Token{
		Kind: KindLiteral,
		Text: strings.TrimSpace(raw),
		Name: emptyString,
		Type: emptyString,
	}
}

// parsePlaceholder parses a single "{name:type}" string and
// returns the corresponding [Token]. The fullPattern parameter is
// used only to produce human-readable error messages.
func parsePlaceholder(raw, fullPattern string) (Token, error) {
	// raw is guaranteed to start with '{' and end with '}'.
	inner := raw[bracketWidth : len(raw)-bracketWidth]

	if inner == emptyString {
		return Token{}, fmt.Errorf(
			"%w %q in pattern %q",
			errEmptyPlaceholder,
			raw,
			fullPattern,
		)
	}

	before, after, ok := strings.Cut(inner, ":")
	if !ok {
		return Token{}, fmt.Errorf(
			fmtPlaceholderError,
			raw,
			errMissingColon,
			fullPattern,
		)
	}

	name := before
	typePart := after

	if name == emptyString {
		return Token{}, fmt.Errorf(
			fmtPlaceholderError,
			raw,
			errEmptyName,
			fullPattern,
		)
	}

	if typePart == emptyString {
		return Token{}, fmt.Errorf(
			fmtPlaceholderError,
			raw,
			errEmptyType,
			fullPattern,
		)
	}

	pType := PlaceholderType(typePart)

	if _, supported := validTypes()[pType]; !supported {
		return Token{}, fmt.Errorf(
			"placeholder %q type %q: %w in pattern %q",
			raw,
			typePart,
			errUnknownType,
			fullPattern,
		)
	}

	return Token{
		Kind: KindPlaceholder,
		Text: raw,
		Name: name,
		Type: pType,
	}, nil
}
