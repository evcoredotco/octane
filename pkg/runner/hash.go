package runner

import "crypto/sha256"

// sha256Of returns the SHA-256 digest of b. It is the single
// place in the runner that imports crypto/sha256, so the
// dependency is centralised and easy to audit.
func sha256Of(b []byte) [32]byte {
	return sha256.Sum256(b)
}
