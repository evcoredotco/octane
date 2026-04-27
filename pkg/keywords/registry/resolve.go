package registry

import (
	"errors"
	"sort"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/registry/internal/levenshtein"
	"github.com/evcoreco/octane/pkg/keywords/registry/internal/pattern"
)

// maxLevenshteinSuggestion is the inclusive upper bound on the
// Levenshtein edit distance for a pattern to be included as a
// "did you mean?" suggestion in [ErrNoMatch.Closest]. Patterns
// farther than this distance are not surfaced.
const maxLevenshteinSuggestion = 5

// Match is the successful result of a [Resolve] call. It pairs the
// matched [api.Keyword] with the bound [api.Args] whose values have
// been coerced to the Go types declared by the keyword pattern's
// {name:type} placeholders.
//
// The caller passes Match.Keyword.Func and Match.Args directly to the
// keyword execution layer (spec 005).
type Match struct {
	// Keyword is the registered keyword whose pattern matched the
	// step text.
	Keyword api.Keyword

	// Args holds the named parameter values extracted from the step
	// text and coerced to their declared Go types. Accessor calls
	// such as Args.Int("n") panic if the key is absent; see
	// [api.Args] for the full panic contract.
	Args api.Args
}

// Resolve matches step against every registered keyword that is
// eligible for the given ocppVersion and returns the first match in
// resolution order.
//
// Resolution order (per ADR 0007 and plan 003 §4):
//  1. Domain-layer keywords whose OCPPVersion equals ocppVersion, or
//     whose OCPPVersion is the zero value (version-agnostic domain
//     keyword), are preferred over primitive-layer keywords.
//  2. Within each layer, longer patterns (by character count) are
//     tried before shorter ones to resolve ambiguity in favour of
//     the more specific pattern.
//  3. The first pattern that matches step is returned immediately;
//     remaining patterns are not consulted.
//
// Eligibility rules:
//   - Primitive-layer keywords (LayerPrimitive) are always eligible
//     regardless of their OCPPVersion value.
//   - Domain-layer keywords (LayerDomain) are eligible when their
//     OCPPVersion equals ocppVersion or when their OCPPVersion is
//     the zero value (version-agnostic).
//
// Resolve returns [*ErrNoMatch] when no keyword matches step, and
// [*ErrTypeMismatch] when a pattern matches but a placeholder value
// cannot be coerced to its declared type.
func Resolve(step string, ocppVersion api.OCPPVersion) (Match, error) {
	all := All()
	candidates := eligibleCandidates(all, ocppVersion)

	for _, keyword := range candidates {
		matched, err := tryMatch(step, keyword)
		if err != nil {
			return Match{}, err
		}

		if matched != nil {
			return *matched, nil
		}
	}

	return Match{}, &ErrNoMatch{
		StepText: step,
		Closest:  closestPattern(step, all),
	}
}

// eligibleCandidates filters and sorts keywords from all for the
// given version. Domain keywords whose OCPPVersion is not ocppVersion
// (and is not zero) are excluded. The result is ordered by
// (Layer descending, len(Pattern) descending) so that domain keywords
// and more-specific patterns are tried first.
func eligibleCandidates(
	all []api.Keyword,
	ocppVersion api.OCPPVersion,
) []api.Keyword {
	out := make([]api.Keyword, 0, len(all))

	for _, keyword := range all {
		if !isEligible(keyword, ocppVersion) {
			continue
		}

		out = append(out, keyword)
	}

	sort.SliceStable(out, func(left, right int) bool {
		leftKeyword := out[left]
		rightKeyword := out[right]

		if leftKeyword.Layer != rightKeyword.Layer {
			// Higher Layer value wins (domain=2 before primitive=1).
			return leftKeyword.Layer > rightKeyword.Layer
		}

		// Longer patterns are more specific; try them first.
		return len(leftKeyword.Pattern) > len(rightKeyword.Pattern)
	})

	return out
}

// isEligible reports whether keyword should be considered during
// resolution for the given ocppVersion.
//
// Primitive-layer keywords are always eligible. Domain-layer keywords
// are eligible when their OCPPVersion matches ocppVersion or when
// their OCPPVersion is the zero value (treated as version-agnostic).
func isEligible(keyword api.Keyword, ocppVersion api.OCPPVersion) bool {
	if keyword.Layer == api.LayerPrimitive {
		return true
	}

	// Domain layer: zero OCPPVersion means version-agnostic.
	const zeroVersion api.OCPPVersion = 0

	return keyword.OCPPVersion == zeroVersion ||
		keyword.OCPPVersion == ocppVersion
}

// tryMatch attempts to match step against keyword's pattern. It
// returns a non-nil *Match on success, nil on a pattern mismatch, or
// a non-nil error when coercion fails.
//
// A *[ErrTypeMismatch] is returned when the pattern matches
// structurally but a placeholder value cannot be coerced to its
// declared type.
func tryMatch(step string, keyword api.Keyword) (*Match, error) {
	tokens, err := pattern.Parse(keyword.Pattern)
	if err != nil {
		// A malformed pattern should have been caught at Register
		// time; skip it defensively rather than surfacing a parse
		// error to the caller.
		return nil, nil //nolint:nilerr // defensive skip; not caller-visible
	}

	captures, matched := pattern.Match(tokens, step)
	if !matched {
		return nil, nil
	}

	coerced, coerceErr := pattern.Coerce(captures, tokens)
	if coerceErr != nil {
		var coercErr *pattern.CoercionError
		if errors.As(coerceErr, &coercErr) {
			return nil, &ErrTypeMismatch{
				ArgName:  coercErr.ArgName,
				Expected: coercErr.Expected,
				Got:      coercErr.Got,
			}
		}

		return nil, coerceErr
	}

	result := &Match{
		Keyword: keyword,
		Args:    api.NewArgs(coerced),
	}

	return result, nil
}

// closestPattern returns the registered pattern string whose
// Levenshtein distance to step is smallest and within
// [maxLevenshteinSuggestion]. An empty string is returned when no
// candidate is close enough.
func closestPattern(step string, all []api.Keyword) string {
	patterns := make([]string, len(all))
	for idx, keyword := range all {
		patterns[idx] = keyword.Pattern
	}

	closest := levenshtein.Closest(step, patterns)
	if closest == "" {
		return ""
	}

	if levenshtein.Distance(step, closest) > maxLevenshteinSuggestion {
		return ""
	}

	return closest
}
