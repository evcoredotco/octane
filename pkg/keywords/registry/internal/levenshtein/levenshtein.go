// Package levenshtein provides edit-distance helpers used by the
// keyword registry to produce "did you mean?" suggestions when a
// step text does not match any registered pattern.
//
// Both exported functions are case-insensitive: inputs are
// lower-cased before comparison so that "BootNotification" and
// "bootnotification" have a distance of zero.
package levenshtein

import "strings"

// Distance returns the Levenshtein edit distance between strings
// src and tgt. The comparison is case-insensitive: both inputs are
// lower-cased before the distance is computed, so "Hello" and
// "hello" have distance zero.
//
// The implementation uses a standard O(m*n) dynamic-programming
// matrix. For the short keyword patterns used in OCTANE (typically
// fewer than 120 characters) this is well within acceptable bounds.
func Distance(src, tgt string) int {
	src = strings.ToLower(src)
	tgt = strings.ToLower(tgt)

	// Handle the degenerate cases first to avoid allocating a
	// matrix when one or both strings are empty.
	if src == tgt {
		return 0
	}

	if len(src) == 0 {
		return len(tgt)
	}

	if len(tgt) == 0 {
		return len(src)
	}

	// Build the full rows×cols matrix where rows = len(src)+1 and
	// cols = len(tgt)+1. matrix[row][col] holds the edit distance
	// between src[:row] and tgt[:col].
	rows := len(src) + 1
	cols := len(tgt) + 1

	matrix := make([][]int, rows)
	for row := range matrix {
		matrix[row] = make([]int, cols)
	}

	// Base cases: converting to/from the empty string costs one
	// operation per character.
	for row := range rows {
		matrix[row][0] = row
	}

	for col := range cols {
		matrix[0][col] = col
	}

	for row := 1; row < rows; row++ {
		for col := 1; col < cols; col++ {
			if src[row-1] == tgt[col-1] {
				matrix[row][col] = matrix[row-1][col-1]
			} else {
				matrix[row][col] = 1 + minOf3(
					matrix[row-1][col],   // deletion
					matrix[row][col-1],   // insertion
					matrix[row-1][col-1], // substitution
				)
			}
		}
	}

	return matrix[rows-1][cols-1]
}

// Closest returns the element of candidates whose Levenshtein
// distance to needle is smallest. When two or more candidates share
// the minimum distance the one that is lexicographically first
// (after lower-casing) is returned, giving a deterministic result
// independent of slice order.
//
// An empty string is returned when candidates is empty.
func Closest(needle string, candidates []string) string {
	if len(candidates) == 0 {
		return ""
	}

	best := ""
	bestDist := -1

	for _, candidate := range candidates {
		dist := Distance(needle, candidate)

		switch {
		case bestDist < 0 || dist < bestDist:
			best = candidate
			bestDist = dist
		case dist == bestDist &&
			strings.ToLower(candidate) < strings.ToLower(best):
			best = candidate
		}
	}

	return best
}

// minOf3 returns the smallest of three integers.
func minOf3(first, second, third int) int {
	if first <= second && first <= third {
		return first
	}

	if second <= third {
		return second
	}

	return third
}
