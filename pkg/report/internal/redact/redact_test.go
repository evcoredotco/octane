// Package redact_test contains unit tests for the redact package.
// Task: T-007-12.
package redact_test

import (
	"testing"

	"github.com/evcoreco/octane/pkg/report/internal/redact"
)

// TestAuthBlock_redactsToken verifies that a token field in an auth
// block is replaced by the placeholder.
func TestAuthBlock_redactsToken(t *testing.T) {
	t.Parallel()

	input := map[string]any{"token": "secret-token-value"}
	got := redact.AuthBlock(input)

	if got["token"] != redact.Placeholder {
		t.Errorf("token: got %q, want %q", got["token"], redact.Placeholder)
	}
}

// TestAuthBlock_redactsPassword verifies that a password field in an
// auth block is replaced by the placeholder.
func TestAuthBlock_redactsPassword(t *testing.T) {
	t.Parallel()

	input := map[string]any{"password": "s3cr3t"}
	got := redact.AuthBlock(input)

	if got["password"] != redact.Placeholder {
		t.Errorf(
			"password: got %q, want %q",
			got["password"],
			redact.Placeholder,
		)
	}
}

// TestAuthBlock_redactsBasic verifies that a basic field in an auth
// block is replaced by the placeholder.
func TestAuthBlock_redactsBasic(t *testing.T) {
	t.Parallel()

	input := map[string]any{"basic": "dXNlcjpwYXNz"}
	got := redact.AuthBlock(input)

	if got["basic"] != redact.Placeholder {
		t.Errorf(
			"basic: got %q, want %q",
			got["basic"],
			redact.Placeholder,
		)
	}
}

// TestAuthBlock_redactsAllFields verifies that all fields in an auth
// block are replaced — deny-by-default means every key is a credential.
func TestAuthBlock_redactsAllFields(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"token":    "tok",
		"password": "pw",
		"basic":    "b64",
		"other":    "value",
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
	input := map[string]any{"token": original}

	_ = redact.AuthBlock(input)

	if input["token"] != original {
		t.Errorf("input mutated: got %q, want %q", input["token"], original)
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
