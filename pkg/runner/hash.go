package runner

import "crypto/sha256"

// sha256DigestSize is the byte length of a SHA-256 digest.
// It equals sha256.Size but expressed as a typed array dimension
// so that the add-constant linter does not flag the literal 32.
const sha256DigestSize = sha256.Size

// sha256Of returns the SHA-256 digest of b. It is the single
// place in the runner that imports crypto/sha256, so the
// dependency is centralised and easy to audit.
func sha256Of(b []byte) [sha256DigestSize]byte {
	return sha256.Sum256(b)
}

// safeUint64 converts a non-negative int to uint64 safely.
// If n is negative (which should not occur given validated inputs),
// it is clamped to 0 to prevent gosec G115 integer overflow issues.
func safeUint64(n int) uint64 {
	if n < 0 {
		return 0
	}

	return uint64(n)
}
