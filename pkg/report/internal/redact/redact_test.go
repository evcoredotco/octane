// Package redact_test contains unit tests for the redact package.
// Task: T-007-12.
package redact_test

import (
	"testing"

	"github.com/evcoreco/octane/pkg/report/internal/redact"
)

const (
	// authKeyToken is the credential key name for bearer tokens.
	authKeyToken = "token"

	// authKeyPassword is the credential key name for passwords.
	authKeyPassword = "password"

	// authKeyBasic is the credential key name for Basic auth credentials.
	authKeyBasic = "basic"
)

// TestAuthBlock_redactsToken verifies that a token field in an auth
// block is replaced by the placeholder.
func TestAuthBlock_redactsToken(t *testing.T) {
	t.Parallel()

	input := map[string]any{authKeyToken: "secret-token-value"}
	got := redact.AuthBlock(input)

	if got[authKeyToken] != redact.Placeholder {
		t.Errorf(
			"token: got %q, want %q",
			got[authKeyToken],
			redact.Placeholder,
		)
	}
}

// TestAuthBlock_redactsPassword verifies that a password field in an
// auth block is replaced by the placeholder.
func TestAuthBlock_redactsPassword(t *testing.T) {
	t.Parallel()

	input := map[string]any{authKeyPassword: "s3cr3t"}
	got := redact.AuthBlock(input)

	if got[authKeyPassword] != redact.Placeholder {
		t.Errorf(
			"password: got %q, want %q",
			got[authKeyPassword],
			redact.Placeholder,
		)
	}
}

// TestAuthBlock_redactsBasic verifies that a basic field in an auth
// block is replaced by the placeholder.
func TestAuthBlock_redactsBasic(t *testing.T) {
	t.Parallel()

	input := map[string]any{authKeyBasic: "dXNlcjpwYXNz"}
	got := redact.AuthBlock(input)

	if got[authKeyBasic] != redact.Placeholder {
		t.Errorf(
			"basic: got %q, want %q",
			got[authKeyBasic],
			redact.Placeholder,
		)
	}
}

// TestAuthBlock_redactsAllFields verifies that all fields in an auth
// block are replaced — deny-by-default means every key is a credential.
func TestAuthBlock_redactsAllFields(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		authKeyToken:    "tok",
		authKeyPassword: "pw",
		authKeyBasic:    "b64",
		"other":         "value",
	}
	got := redact.AuthBlock(input)

	for key, val := range got {
		if val != redact.Placeholder {
			t.Errorf("key %q: got %q, want %q", key, val, redact.Placeholder)
		}
	}
}

// TestAuthBlock_doesNotMutateInput verifies that AuthBlock returns a new
// map and does not modify the original.
func TestAuthBlock_doesNotMutateInput(t *testing.T) {
	t.Parallel()

	original := "original-value"
	input := map[string]any{authKeyToken: original}

	_ = redact.AuthBlock(input)

	if input[authKeyToken] != original {
		t.Errorf(
			"input mutated: got %q, want %q",
			input[authKeyToken],
			original,
		)
	}
}

// TestHeader_redactsAuthorization verifies that the Authorization header
// is replaced by the placeholder.
func TestHeader_redactsAuthorization(t *testing.T) {
	t.Parallel()

	got := redact.Header("Authorization", "Bearer token123")

	if got != redact.Placeholder {
		t.Errorf("Authorization: got %q, want %q", got, redact.Placeholder)
	}
}

// TestHeader_redactsCookie verifies that the Cookie header is replaced
// by the placeholder.
func TestHeader_redactsCookie(t *testing.T) {
	t.Parallel()

	got := redact.Header("Cookie", "session=abc")

	if got != redact.Placeholder {
		t.Errorf("Cookie: got %q, want %q", got, redact.Placeholder)
	}
}

// TestHeader_redactsXAPIKey verifies that the X-Api-Key header is
// replaced by the placeholder.
func TestHeader_redactsXAPIKey(t *testing.T) {
	t.Parallel()

	got := redact.Header("X-Api-Key", "key-abc123")

	if got != redact.Placeholder {
		t.Errorf("X-Api-Key: got %q, want %q", got, redact.Placeholder)
	}
}

// TestHeader_preservesSafeHeader verifies that a non-sensitive header
// such as Content-Type is returned unchanged.
func TestHeader_preservesSafeHeader(t *testing.T) {
	t.Parallel()

	value := "application/json"
	got := redact.Header("Content-Type", value)

	if got != value {
		t.Errorf("Content-Type: got %q, want %q", got, value)
	}
}
