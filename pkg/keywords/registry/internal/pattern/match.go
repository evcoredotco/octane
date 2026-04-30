package pattern

import (
	"strings"
)

const (
	// startPos is the initial position index used when scanning tokens.
	startPos = 0

	// noMatch is the sentinel returned by consumeLiteral on failure.
	noMatch = -1

	// oneWord is the number of step words consumed by a single
	// KindPlaceholder token.
	oneWord = 1
)

// Match attempts to match a parsed keyword pattern (expressed as a
// slice of [Token] values produced by [Parse]) against a step text
// string. It returns the raw captured strings keyed by placeholder
// name and a boolean indicating whether the match succeeded.
//
// Matching rules:
//   - The step text is tokenised into whitespace-delimited words
//     before comparison begins. Leading, trailing, and runs of
//     internal whitespace in the step are all normalised.
//   - A [KindLiteral] token is split into words. Each word is
//     compared case-insensitively with the corresponding step word.
//     If any word does not match, the entire match fails.
//   - A [KindPlaceholder] token consumes exactly one step word and
//     stores its raw value in the returned map under the placeholder
//     name. No type coercion is performed; that responsibility
//     belongs to the coercer (see coerce.go, task T-003-12).
//   - The match succeeds only when every token has been consumed and
//     no step words remain unconsumed.
//
// An empty step string or a mismatch in word count returns
// (nil, false). A successful match with no placeholders returns
// (map[string]string{}, true) — a non-nil empty map.
func Match(
	tokens []Token,
	step string,
) (map[string]string, bool) {
	stepWords := splitWords(step)
	if len(stepWords) == startPos {
		return nil, false
	}

	captures := make(map[string]string)
	pos := startPos

	for idx := range tokens {
		newPos, ok := consumeToken(tokens[idx], stepWords, pos, captures)
		if !ok {
			return nil, false
		}

		pos = newPos
	}

	if pos != len(stepWords) {
		return nil, false
	}

	return captures, true
}

// consumeToken processes a single token against the step words at position
// pos, updating captures for placeholders. It returns the new position and
// whether the token was consumed successfully.
func consumeToken(
	tok Token,
	stepWords []string,
	pos int,
	captures map[string]string,
) (int, bool) {
	switch tok.Kind {
	case KindLiteral:
		newPos := consumeLiteral(tok.Text, stepWords, pos)
		if newPos == noMatch {
			return startPos, false
		}

		return newPos, true

	case KindPlaceholder:
		if pos >= len(stepWords) {
			return startPos, false
		}

		captures[tok.Name] = stepWords[pos]

		return pos + oneWord, true

	default:
		// Unknown token kinds are silently skipped; new token
		// kinds added in future must be handled explicitly above.
		return pos, true
	}
}

// splitWords splits s on any run of Unicode whitespace and returns
// the non-empty fields. It is equivalent to strings.Fields but kept
// as a named helper for clarity.
func splitWords(s string) []string {
	return strings.Fields(s)
}

// consumeLiteral splits literalText into words and verifies that the
// step words starting at pos match each literal word
// case-insensitively. It returns the new position (pos + number of
// literal words) on success, or -1 if any word does not match or
// there are not enough step words remaining.
func consumeLiteral(
	literalText string,
	stepWords []string,
	pos int,
) int {
	litWords := strings.Fields(literalText)

	if len(litWords) == startPos {
		return pos
	}

	if pos+len(litWords) > len(stepWords) {
		return noMatch
	}

	for offset, word := range litWords {
		if !strings.EqualFold(word, stepWords[pos+offset]) {
			return noMatch
		}
	}

	return pos + len(litWords)
}
