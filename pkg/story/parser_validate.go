package story

import (
	"github.com/evcoreco/octane/pkg/story/ast"
	"github.com/evcoreco/octane/pkg/story/diag"
)

// validateParameters implements T-001-25.
//
// It walks every step text across background, setup, all scenario steps, and
// teardown. For each step it extracts {placeholder} tokens using a linear
// scan (no regex). If any placeholder name is not present in meta.Parameters
// the function returns *diag.UnboundParameterError for the first offending
// step, including all unbound names found in that step.
//
// Placeholder syntax: {name} or {name:type}. The :type suffix is stripped
// before the lookup so that typed parameters still resolve correctly.
func validateParameters(
	file string,
	meta ast.Meta,
	scenarios []ast.Scenario,
	background []ast.Step,
	setup []ast.Step,
	teardown []ast.Step,
) error {
	paramSet := make(map[string]struct{}, len(meta.Parameters))

	for _, param := range meta.Parameters {
		paramSet[param] = struct{}{}
	}

	allGroups := buildStepGroups(scenarios, background, setup, teardown)

	for _, group := range allGroups {
		for _, step := range group {
			unbound := findUnbound(step.Text, paramSet)
			if len(unbound) == noUnbound {
				continue
			}

			return &diag.UnboundParameterError{
				File:       file,
				Line:       step.Position.Line,
				Column:     step.Position.Column,
				Parameters: unbound,
				StepText:   step.Text,
				Suggestion: "declare the parameter(s) under " +
					"'Parameters:' in the Meta section",
			}
		}
	}

	return nil
}

// fixedStepGroupCount is the number of fixed step groups:
// background, setup, teardown.
const fixedStepGroupCount = 3

// buildStepGroups assembles background, setup, scenario steps, and teardown
// into a single slice of step groups for uniform traversal. Grammar section
// order is preserved: background, setup, scenarios (in order), teardown.
func buildStepGroups(
	scenarios []ast.Scenario,
	background []ast.Step,
	setup []ast.Step,
	teardown []ast.Step,
) [][]ast.Step {
	groupCap := fixedStepGroupCount + len(scenarios)
	groups := make([][]ast.Step, emptyGroupCapacity, groupCap)
	groups = append(groups, background, setup)

	for _, sc := range scenarios {
		groups = append(groups, sc.Steps)
	}

	return append(groups, teardown)
}

// findUnbound returns the unbound placeholder names found in text. It uses a
// plain character-by-character scan to locate '{...}' spans, strips any
// ':type' suffix, and checks against the declared set. The returned slice is
// sorted and deduplicated.
func findUnbound(text string, declared map[string]struct{}) []string {
	seen := map[string]struct{}{}

	var unbound []string

	pos := textScanStart

	for pos < len(text) {
		name, advance := extractPlaceholderAt(text, pos)
		if name == emptyPlaceholder {
			pos += advance

			continue
		}

		recordUnbound(name, declared, seen, &unbound)

		pos += advance
	}

	insertionSortStrings(unbound)

	return unbound
}

// emptyPlaceholder is the empty string sentinel for placeholder names.
const emptyPlaceholder = ""

// noUnbound is the zero-length sentinel used to detect an empty unbound list.
// Required by the add-constant linter rule.
const noUnbound = 0

// noPlaceholderAdvance is the advance returned when no placeholder starts at
// the current position.
const noPlaceholderAdvance = 1

// placeholderOffset is the offset past the opening '{' when extracting the
// inner text of a placeholder. Also used for the closed-brace end-advance.
const placeholderOffset = 1

// emptyGroupCapacity is the zero capacity used when pre-allocating the
// initial step groups slice.
const emptyGroupCapacity = 0

// textScanStart is the initial byte position for the placeholder scanner
// in findUnbound. Required by add-constant.
const textScanStart = 0

// sortLowerBound is the inclusive lower bound for the inner loop index in
// insertionSortStrings. Required by add-constant.
const sortLowerBound = 0

// extractPlaceholderAt attempts to read a '{...}' placeholder starting at
// pos in text. It returns (name, advance) where advance is the number of
// bytes to move forward. When no placeholder is found advance is 1 and name
// is empty.
func extractPlaceholderAt(text string, pos int) (string, int) {
	if text[pos] != '{' {
		return emptyPlaceholder, noPlaceholderAdvance
	}

	end := pos + placeholderOffset

	for end < len(text) && text[end] != '}' && text[end] != '\n' {
		end++
	}

	if end >= len(text) || text[end] != '}' {
		return emptyPlaceholder, noPlaceholderAdvance
	}

	inner := text[pos+placeholderOffset : end]
	n := stripTypeSuffix(inner)

	return n, end - pos + placeholderOffset
}

// recordUnbound adds name to unbound when it is neither declared nor already
// seen.
func recordUnbound(
	name string,
	declared map[string]struct{},
	seen map[string]struct{},
	unbound *[]string,
) {
	if name == emptyPlaceholder {
		return
	}

	if _, ok := declared[name]; ok {
		return
	}

	if _, already := seen[name]; already {
		return
	}

	seen[name] = struct{}{}

	*unbound = append(*unbound, name)
}

// stripTypeSuffix removes the ':type' suffix from a placeholder inner
// string, returning only the name portion.
func stripTypeSuffix(inner string) string {
	for idx := range len(inner) {
		if inner[idx] == ':' {
			return inner[:idx]
		}
	}

	return inner
}

// sortStartIdx is the starting index for insertionSortStrings: the second
// element (index 1), since the first element is trivially sorted.
const sortStartIdx = 1

// sortPredecessorOffset is the offset from the current insertion-sort index
// to its immediate predecessor. Required by add-constant.
const sortPredecessorOffset = 1

// insertionSortStrings sorts slice in place using insertion sort. This is
// used instead of importing "sort" for the typically short unbound slice.
func insertionSortStrings(slice []string) {
	for idx := sortStartIdx; idx < len(slice); idx++ {
		key := slice[idx]
		jdx := idx - sortPredecessorOffset

		for jdx >= sortLowerBound && slice[jdx] > key {
			slice[jdx+sortPredecessorOffset] = slice[jdx]
			jdx--
		}

		slice[jdx+sortPredecessorOffset] = key
	}
}
