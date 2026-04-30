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
			if len(unbound) == 0 {
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
	groups := make([][]ast.Step, 0, fixedStepGroupCount+len(scenarios))
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

	pos := 0

	for pos < len(text) {
		if text[pos] != '{' {
			pos++

			continue
		}

		end := pos + 1

		for end < len(text) && text[end] != '}' && text[end] != '\n' {
			end++
		}

		if end >= len(text) || text[end] != '}' {
			pos++

			continue
		}

		inner := text[pos+1 : end]
		name := stripTypeSuffix(inner)

		if name != "" {
			if _, ok := declared[name]; !ok {
				if _, already := seen[name]; !already {
					seen[name] = struct{}{}

					unbound = append(unbound, name)
				}
			}
		}

		pos = end + 1
	}

	insertionSortStrings(unbound)

	return unbound
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

// insertionSortStrings sorts slice in place using insertion sort. This is
// used instead of importing "sort" for the typically short unbound slice.
func insertionSortStrings(slice []string) {
	for idx := 1; idx < len(slice); idx++ {
		key := slice[idx]
		jdx := idx - 1

		for jdx >= 0 && slice[jdx] > key {
			slice[jdx+1] = slice[jdx]
			jdx--
		}

		slice[jdx+1] = key
	}
}
