package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Hash returns the SHA-256 hex digest of the cache key tuple.
//
// The fields are joined in lexicographic field-name order with
// colon separators, matching the key derivation algorithm defined
// in ADR 0016 §"Cache key derivation":
//
//	csms_endpoint_sha : octane_version : ocpp_version :
//	parameter_sha : scope_key : story_content_sha : test_id
//
// The returned lower-case hex string is 64 characters long and is
// used as the filesystem path component under the cache directory
// (with a two-character fanout prefix per ADR 0016 §"Layout").
func (k Key) Hash() string {
	// Fields are ordered lexicographically by JSON/ADR field name:
	//   csms_endpoint_sha, octane_version, ocpp_version,
	//   parameter_sha, scope_key, story_content_sha, test_id.
	parts := []string{
		k.CSMSEndpointSHA,
		k.OctaneVersion,
		k.OCPPVersion,
		k.ParameterSHA,
		k.ScopeKey,
		k.StoryContentSHA,
		k.TestID,
	}

	sum := sha256.Sum256([]byte(strings.Join(parts, ":")))

	return hex.EncodeToString(sum[:])
}
