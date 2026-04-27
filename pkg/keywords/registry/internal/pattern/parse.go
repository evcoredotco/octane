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

// validTypes is the closed set of accepted placeholder type tokens.
// Keeping it as a map[PlaceholderType]struct{} gives O(1) lookup
// without regexp.
var validTypes = map[PlaceholderType]struct{}{
	TypeString:   {},
	TypeInt:      {},
	TypeFloat:    {},
	TypeBool:     {},
	TypeDuration: {},
	TypeStation:  {},
	TypeAny:      {},
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
	initialCap := len(pattern) / 4
	if initialCap < 1 {
		initialCap = 1
	}

	tokens := make([]Token, 0, initialCap)

	remaining := pattern

	for len(remaining) > 0 {
		openIdx := strings.IndexByte(remaining, '{')
		closeIdx := strings.IndexByte(remaining, '}')

		// Bare '}' before any '{' is malformed.
		if closeIdx != -1 && (openIdx == -1 || closeIdx < openIdx) {
			return nil, fmt.Errorf(
				"unexpected '}' at position %d in pattern %q",
				len(pattern)-len(remaining)+closeIdx,
				pattern,
			)
		}

		if openIdx == -1 {
			// No more placeholders; the rest is a literal.
			if lit := makeLiteral(remaining); lit.Text != "" {
				tokens = append(tokens, lit)
			}

			break
		}

		// Capture the literal segment before the '{'.
		if openIdx > 0 {
			if lit := makeLiteral(remaining[:openIdx]); lit.Text != "" {
				tokens = append(tokens, lit)
			}
		}

		remaining = remaining[openIdx:]

		// Find the matching '}'.
		closeIdx = strings.IndexByte(remaining, '}')
		if closeIdx == -1 {
			return nil, fmt.Errorf(
				"unclosed '{' in pattern %q",
				pattern,
			)
		}

		raw := remaining[:closeIdx+1]

		tok, err := parsePlaceholder(raw, pattern)
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, tok)

		remaining = remaining[closeIdx+1:]
	}

	if len(tokens) == 0 {
		return nil, errors.New("pattern must not be empty")
	}

	return tokens, nil
}

// makeLiteral builds a [KindLiteral] token from a raw string
// segment. The text is trimmed of leading and trailing whitespace
// before storage; internal whitespace is preserved so the matcher
// can treat it as a word boundary.
func makeLiteral(raw string) Token {
	return Token{
		Kind: KindLiteral,
		Text: strings.TrimSpace(raw),
		Name: "",
		Type: "",
	}
}

// parsePlaceholder parses a single "{name:type}" string and
// returns the corresponding [Token]. The fullPattern parameter is
// used only to produce human-readable error messages.
func parsePlaceholder(raw, fullPattern string) (Token, error) {
	// raw is guaranteed to start with '{' and end with '}'.
	inner := raw[1 : len(raw)-1]

	if inner == "" {
		return Token{}, fmt.Errorf(
			"empty placeholder %q in pattern %q",
			raw,
			fullPattern,
		)
	}

	colonIdx := strings.IndexByte(inner, ':')
	if colonIdx == -1 {
		return Token{}, fmt.Errorf(
			"placeholder %q is missing the colon separator "+
				"in pattern %q; use {name:type} syntax",
			raw,
			fullPattern,
		)
	}

	name := inner[:colonIdx]
	typePart := inner[colonIdx+1:]

	if name == "" {
		return Token{}, fmt.Errorf(
			"placeholder %q has an empty name in pattern %q",
			raw,
			fullPattern,
		)
	}

	if typePart == "" {
		return Token{}, fmt.Errorf(
			"placeholder %q has an empty type in pattern %q; "+
				"supported types: string, int, float, bool, "+
				"duration, station, any",
			raw,
			fullPattern,
		)
	}

	pType := PlaceholderType(typePart)

	if _, supported := validTypes[pType]; !supported {
		return Token{}, fmt.Errorf(
			"placeholder %q declares unknown type %q in pattern %q; "+
				"supported types: string, int, float, bool, "+
				"duration, station, any",
			raw,
			typePart,
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
