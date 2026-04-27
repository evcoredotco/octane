// Package redact provides deny-by-default credential scrubbers for
// connection profile auth blocks and HTTP headers. Every field in an
// auth block is considered a credential; every sensitive HTTP header
// (Authorization, Cookie, Set-Cookie, X-Api-Key, Proxy-Authorization)
// is replaced by [Placeholder].
//
// Task: T-007-10.
package redact

// Placeholder is the string that replaces any redacted value.
const Placeholder = "<redacted>"

// AuthBlock redacts all credential fields from a connection profile auth
// block. The input is a map[string]any representing the auth block; the
// function returns a new map with every value replaced by [Placeholder].
//
// This is a deny-by-default redactor: all keys in the auth block are
// credentials, so all are replaced. The input map is never mutated.
func AuthBlock(auth map[string]any) map[string]any {
	out := make(map[string]any, len(auth))

	for key := range auth {
		out[key] = Placeholder
	}

	return out
}
