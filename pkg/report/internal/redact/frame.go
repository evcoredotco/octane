package redact

import (
	"encoding/json"
	"regexp"
)

// credentialFieldRE matches JSON object keys that are likely to carry
// credentials in OCPP-J payloads and similar structures.
// The pattern is case-insensitive and requires an exact key match.
var credentialFieldRE = regexp.MustCompile(
	`(?i)^(password|passphrase|privatekey|clientcertificate|bearertoken|` +
		`idtoken|accesstoken|clientsecret|sharedsecret|apikey|secretkey|` +
		`authorization|credential)$`,
)

// jwtRE matches strings that look like Base64url-encoded JWT tokens
// (three dot-separated segments starting with "eyJ"). These may appear
// in error messages or description fields.
var jwtRE = regexp.MustCompile(
	`eyJ[A-Za-z0-9\-_]+\.[A-Za-z0-9\-_]+\.[A-Za-z0-9\-_]*`,
)

// Frame scrubs a raw OCPP-J wire frame (a JSON byte slice) by replacing
// the values of credential-bearing fields with [Placeholder] and masking
// JWT patterns in string values. The frame is parsed as JSON, scrubbed
// recursively, and re-serialized. If parsing fails the original bytes are
// returned unchanged (the frame is treated as opaque data).
func Frame(raw []byte) []byte {
	var value any

	err := json.Unmarshal(raw, &value)
	if err != nil {
		return raw
	}

	scrubbed := scrubValue(value)

	out, err := json.Marshal(scrubbed)
	if err != nil {
		return raw
	}

	return out
}

// scrubValue recursively scrubs JSON values. Objects have credential keys
// replaced; arrays have each element scrubbed; strings have JWT patterns
// masked; other primitives are returned as-is.
func scrubValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return scrubObject(typed)
	case []any:
		return scrubArray(typed)
	case string:
		return scrubString(typed)
	default:
		return value
	}
}

// scrubObject replaces values for credential-bearing keys and recurses into
// non-credential values.
func scrubObject(obj map[string]any) map[string]any {
	out := make(map[string]any, len(obj))

	for key, val := range obj {
		if credentialFieldRE.MatchString(key) {
			out[key] = Placeholder
		} else {
			out[key] = scrubValue(val)
		}
	}

	return out
}

// scrubArray recurses into each element of a JSON array.
func scrubArray(arr []any) []any {
	out := make([]any, len(arr))

	for idx, el := range arr {
		out[idx] = scrubValue(el)
	}

	return out
}

// scrubString masks JWT-like patterns inside arbitrary string values.
// Non-JWT strings are returned unchanged.
func scrubString(s string) string {
	return jwtRE.ReplaceAllString(s, Placeholder)
}

// FindingMessage scrubs a finding message string by masking JWT patterns
// and any HTTP header values that match the sensitive header regex.
// It is used to prevent credential leakage via runner error strings.
func FindingMessage(msg string) string {
	return jwtRE.ReplaceAllString(msg, Placeholder)
}
