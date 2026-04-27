package pattern

import (
	"strings"
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
	if len(stepWords) == 0 {
		return nil, false
	}

	captures := make(map[string]string)
	pos := 0

	for idx := range tokens {
		tok := tokens[idx]

		switch tok.Kind {
		case KindLiteral:
			pos = consumeLiteral(tok.Text, stepWords, pos)
			if pos < 0 {
				return nil, false
			}

		case KindPlaceholder:
			if pos >= len(stepWords) {
				return nil, false
			}

			captures[tok.Name] = stepWords[pos]
			pos++
		}
	}

	if pos != len(stepWords) {
		return nil, false
	}

	return captures, true
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

	if len(litWords) == 0 {
		return pos
	}

	if pos+len(litWords) > len(stepWords) {
		return -1
	}

	for offset, word := range litWords {
		if !strings.EqualFold(word, stepWords[pos+offset]) {
			return -1
		}
	}

	return pos + len(litWords)
}
