// Package levenshtein provides edit-distance helpers used by the
// keyword registry to produce "did you mean?" suggestions when a
// step text does not match any registered pattern.
//
// Both exported functions are case-insensitive: inputs are
// lower-cased before comparison so that "BootNotification" and
// "bootnotification" have a distance of zero.
package levenshtein

import "strings"

// zeroDistance is the edit distance between two identical strings.
const zeroDistance = 0

// zeroLen is the empty-string length sentinel used in base-case checks.
const zeroLen = 0

// matrixOffset is the +1 added to string lengths when sizing the DP matrix.
const matrixOffset = 1

// startIndex is the first non-base-case row/column in the DP matrix.
const startIndex = 1

// noCandidate flags an uninitialised best-distance (no candidates seen yet).
const noCandidate = -1

// emptyResult is the empty string returned when no candidates are provided.
const emptyResult = ""

// zeroCol is the base-case column index (converting to empty string).
const zeroCol = 0

// zeroRow is the base-case row index (converting from empty string).
const zeroRow = 0

// prevOffset is the index offset used to access the previous row/column.
const prevOffset = 1

// editCost is the unit cost of a single insertion, deletion, or substitution.
const editCost = 1

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
		return zeroDistance
	}

	if len(src) == zeroLen {
		return len(tgt)
	}

	if len(tgt) == zeroLen {
		return len(src)
	}

	// Build the full rows×cols matrix where rows = len(src)+1 and
	// cols = len(tgt)+1. matrix[row][col] holds the edit distance
	// between src[:row] and tgt[:col].
	rows := len(src) + matrixOffset
	cols := len(tgt) + matrixOffset

	matrix := make([][]int, rows)
	for row := range matrix {
		matrix[row] = make([]int, cols)
	}

	// Base cases: converting to/from the empty string costs one
	// operation per character.
	for row := range rows {
		matrix[row][zeroCol] = row
	}

	for col := range cols {
		matrix[zeroRow][col] = col
	}

	for row := startIndex; row < rows; row++ {
		for col := startIndex; col < cols; col++ {
			if src[row-prevOffset] == tgt[col-prevOffset] {
				matrix[row][col] = matrix[row-prevOffset][col-prevOffset]
			} else {
				matrix[row][col] = editCost + minOf3(
					matrix[row-prevOffset][col],            // deletion
					matrix[row][col-prevOffset],            // insertion
					matrix[row-prevOffset][col-prevOffset], // substitution
				)
			}
		}
	}

	return matrix[rows-prevOffset][cols-prevOffset]
}

// Closest returns the element of candidates whose Levenshtein
// distance to needle is smallest. When two or more candidates share
// the minimum distance the one that is lexicographically first
// (after lower-casing) is returned, giving a deterministic result
// independent of slice order.
//
// An empty string is returned when candidates is empty.
func Closest(needle string, candidates []string) string {
	if len(candidates) == zeroLen {
		return emptyResult
	}

	best := emptyResult
	bestDist := noCandidate

	for _, candidate := range candidates {
		dist := Distance(needle, candidate)

		switch {
		case bestDist < zeroDistance || dist < bestDist:
			best = candidate
			bestDist = dist
		case dist == bestDist &&
			strings.ToLower(candidate) < strings.ToLower(best):
			best = candidate
		default:
			// dist > bestDist: current candidate is further; skip.
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
