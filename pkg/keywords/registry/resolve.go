package registry

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/registry/internal/levenshtein"
	"github.com/evcoreco/octane/pkg/keywords/registry/internal/pattern"
)

const (
	// maxLevenshteinSuggestion is the inclusive upper bound on the
	// Levenshtein edit distance for a pattern to be included as a
	// "did you mean?" suggestion in [NoMatchError.Closest]. Patterns
	// farther than this distance are not surfaced.
	maxLevenshteinSuggestion = 5

	// noSuggestion is returned by closestPattern when no close match is found.
	noSuggestion = ""

	// emptySliceLen is the zero-capacity sentinel for make([]T, 0, n) calls.
	emptySliceLen = 0

	// zeroLayer is the zero-value sentinel for api.Layer fields.
	zeroLayer api.Layer = 0

	// zeroOCPPVersion is the zero-value sentinel for api.OCPPVersion fields.
	zeroOCPPVersion api.OCPPVersion = 0
)

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
// Resolve returns [*NoMatchError] when no keyword matches step, and
// [*TypeMismatchError] when a pattern matches but a placeholder value
// cannot be coerced to its declared type.
func Resolve(step string, ocppVersion api.OCPPVersion) (Match, error) {
	all := All()
	candidates := eligibleCandidates(all, ocppVersion)

	for _, keyword := range candidates {
		matched, ok, err := tryMatch(step, keyword)
		if err != nil {
			return Match{}, err
		}

		if ok {
			return matched, nil
		}
	}

	return Match{}, &NoMatchError{
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
	out := make([]api.Keyword, emptySliceLen, len(all))

	for _, keyword := range all {
		if !isEligible(keyword, ocppVersion) {
			continue
		}

		out = append(out, keyword)
	}

	slices.SortStableFunc(out, func(left, right api.Keyword) int {
		if left.Layer != right.Layer {
			// Higher Layer value wins (domain=2 before primitive=1).
			return cmp.Compare(right.Layer, left.Layer)
		}

		// Longer patterns are more specific; try them first.
		return cmp.Compare(len(right.Pattern), len(left.Pattern))
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
	return keyword.OCPPVersion == zeroOCPPVersion ||
		keyword.OCPPVersion == ocppVersion
}

// tryMatch attempts to match step against keyword's pattern. It
// returns a non-nil *Match on success, nil on a pattern mismatch, or
// a non-nil error when coercion fails.
//
// A *[TypeMismatchError] is returned when the pattern matches
// structurally but a placeholder value cannot be coerced to its
// declared type.
func tryMatch(step string, keyword api.Keyword) (Match, bool, error) {
	tokens, err := pattern.Parse(keyword.Pattern)
	if err != nil {
		// A malformed pattern should have been caught at Register
		// time; skip it defensively rather than surfacing a parse
		// error to the caller.
		return Match{ //nolint:nilerr // defensive skip; not caller-visible
			Keyword: api.Keyword{
				Pattern:     "",
				Layer:       zeroLayer,
				OCPPVersion: zeroOCPPVersion,
				Func:        nil,
			},
			Args: api.Args{},
		}, false, nil
	}

	captures, matched := pattern.Match(tokens, step)
	if !matched {
		return Match{
			Keyword: api.Keyword{
				Pattern:     "",
				Layer:       zeroLayer,
				OCPPVersion: zeroOCPPVersion,
				Func:        nil,
			},
			Args: api.Args{},
		}, false, nil
	}

	coerced, coerceErr := pattern.Coerce(captures, tokens)
	if coerceErr != nil {
		var coercErr *pattern.CoercionError
		if errors.As(coerceErr, &coercErr) {
			return Match{}, false, &TypeMismatchError{
				ArgName:  coercErr.ArgName,
				Expected: coercErr.Expected,
				Got:      coercErr.Got,
			}
		}

		return Match{}, false, fmt.Errorf(
			"registry: coerce args: %w",
			coerceErr,
		)
	}

	result := Match{
		Keyword: keyword,
		Args:    api.NewArgs(coerced),
	}

	return result, true, nil
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
	if closest == noSuggestion {
		return noSuggestion
	}

	if levenshtein.Distance(step, closest) > maxLevenshteinSuggestion {
		return noSuggestion
	}

	return closest
}
