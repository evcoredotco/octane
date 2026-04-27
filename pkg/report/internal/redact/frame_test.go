// Task: T-007-10 (security review fix).
package redact_test

import (
	"encoding/json"
	"testing"

	"github.com/octane-project/octane/pkg/report/internal/redact"
)

func Test_redact_Frame_credentialField(t *testing.T) {
	t.Parallel()

	// Simulated OCPP-J SetNetworkProfile payload containing an auth block.
	raw := []byte(`[2,"abc","SetNetworkProfile",{"configurationSlot":1,` +
		`"connectionData":{"ocppVersion":"OCPP20","ocppTransport":"JSON",` +
		`"ocppCsmsUrl":"wss://csms.example.com","messageTimeout":30,` +
		`"securityProfile":2,"identityDocument":{"type":"Basic",` +
		`"password":"s3cr3t","idToken":"station-01"}}}]`)

	scrubbed := redact.Frame(raw)

	var parsed []any
	if err := json.Unmarshal(scrubbed, &parsed); err != nil {
		t.Fatalf("scrubbed frame is not valid JSON: %v", err)
	}

	// Verify password and idToken are redacted.
	payload, _ := parsed[3].(map[string]any)
	connData, _ := payload["connectionData"].(map[string]any)
	identity, _ := connData["identityDocument"].(map[string]any)

	if identity["password"] != redact.Placeholder {
		t.Errorf("password not redacted: got %v", identity["password"])
	}

	if identity["idToken"] != redact.Placeholder {
		t.Errorf("idToken not redacted: got %v", identity["idToken"])
	}

	// Non-credential fields must be preserved.
	if connData["ocppVersion"] != "OCPP20" {
		t.Errorf(
			"ocppVersion unexpectedly modified: got %v",
			connData["ocppVersion"],
		)
	}
}

func Test_redact_Frame_jwtInString(t *testing.T) {
	t.Parallel()

	raw := []byte(
		`[3,"abc",{"error":"invalid token eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ4In0.abc123"}]`,
	)

	scrubbed := redact.Frame(raw)

	var parsed []any
	if err := json.Unmarshal(scrubbed, &parsed); err != nil {
		t.Fatalf("scrubbed frame is not valid JSON: %v", err)
	}

	result, _ := parsed[2].(map[string]any)

	errMsg, _ := result["error"].(string)
	if errMsg == "" {
		t.Fatal("error field missing after scrub")
	}

	if errMsg == "invalid token eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ4In0.abc123" {
		t.Errorf("JWT not redacted in error string: %q", errMsg)
	}
}

func Test_redact_Frame_invalidJSON(t *testing.T) {
	t.Parallel()

	raw := []byte(`not-json`)
	scrubbed := redact.Frame(raw)

	// Invalid JSON is returned unchanged.
	if string(scrubbed) != string(raw) {
		t.Errorf("invalid JSON was modified: got %q, want %q", scrubbed, raw)
	}
}

func Test_redact_Frame_nilFrame(t *testing.T) {
	t.Parallel()

	scrubbed := redact.Frame(nil)

	// nil input should return nil or empty, not panic.
	_ = scrubbed
}

func Test_redact_FindingMessage_jwt(t *testing.T) {
	t.Parallel()

	msg := "TLS handshake failed: token eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ4In0.abc123 rejected"
	scrubbed := redact.FindingMessage(msg)

	if scrubbed == msg {
		t.Errorf("JWT not redacted in finding message: %q", scrubbed)
	}
}

func Test_redact_FindingMessage_noSensitiveData(t *testing.T) {
	t.Parallel()

	msg := "assertion failed: expected BootNotification response within 30s"
	scrubbed := redact.FindingMessage(msg)

	if scrubbed != msg {
		t.Errorf("clean message was modified: got %q, want %q", scrubbed, msg)
	}
}
