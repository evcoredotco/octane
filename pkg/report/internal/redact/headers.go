// Package redact provides deny-by-default credential scrubbers.
// This file implements HTTP header redaction.
//
// Task: T-007-11.
package redact

import "regexp"

// sensitiveHeaderRE matches HTTP header names that carry credentials.
// The pattern is case-insensitive.
var sensitiveHeaderRE = regexp.MustCompile(
	`(?i)^(authorization|cookie|set-cookie|x-api-key|proxy-authorization)$`,
)

// Header returns [Placeholder] when the header name matches the
// sensitive pattern, or the original value otherwise.
func Header(name, value string) string {
	if sensitiveHeaderRE.MatchString(name) {
		return Placeholder
	}

	return value
}
