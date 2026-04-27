// Package wire_test contains the JSON float64-as-int coercion tests for the
// wire package (T-002-13).
//
// When encoding/json decodes a JSON number into an any (interface{}) value it
// always produces a float64, never an int. This test confirms that ParseCall,
// ParseResult, and ParseError accept the float64 type codes that the standard
// library produces when unmarshalling real OCPP-J frames received over the
// wire.
package wire_test

import (
	"encoding/json"
	"testing"

	"github.com/octane-project/octane/pkg/wire"
)

// TestJSONFloat64CoercionCall verifies that decoding a CALL frame JSON string
// into []any yields float64(2) at element 0, and that ParseCall accepts it.
//
// This guards against the common mistake of comparing the type code to int(2)
// instead of float64(2).
func TestJSONFloat64CoercionCall(t *testing.T) {
	t.Parallel()

	raw := `[2, "id1", "BootNotification", {}]`

	var frame []any

	if err := json.Unmarshal([]byte(raw), &frame); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	typeCode, ok := frame[0].(float64)
	if !ok {
		t.Fatalf(
			"element 0: expected float64 from JSON decode, got %T",
			frame[0],
		)
	}

	if typeCode != float64(wire.MessageTypeCall) {
		t.Errorf(
			"element 0: got %v, want float64(%d)",
			typeCode, wire.MessageTypeCall,
		)
	}

	if _, err := wire.ParseCall(frame); err != nil {
		t.Fatalf(
			"ParseCall rejected a frame with float64 type code: %v",
			err,
		)
	}
}

// TestJSONFloat64CoercionResult verifies that decoding a CALLRESULT frame
// produces float64(3) at element 0, and that ParseResult accepts it.
func TestJSONFloat64CoercionResult(t *testing.T) {
	t.Parallel()

	raw := `[3, "id2", {"status": "Accepted"}]`

	var frame []any

	if err := json.Unmarshal([]byte(raw), &frame); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	typeCode, ok := frame[0].(float64)
	if !ok {
		t.Fatalf(
			"element 0: expected float64 from JSON decode, got %T",
			frame[0],
		)
	}

	if typeCode != float64(wire.MessageTypeResult) {
		t.Errorf(
			"element 0: got %v, want float64(%d)",
			typeCode, wire.MessageTypeResult,
		)
	}

	if _, err := wire.ParseResult(frame); err != nil {
		t.Fatalf(
			"ParseResult rejected a frame with float64 type code: %v",
			err,
		)
	}
}

// TestJSONFloat64CoercionError verifies that decoding a CALLERROR frame
// produces float64(4) at element 0, and that ParseError accepts it.
func TestJSONFloat64CoercionError(t *testing.T) {
	t.Parallel()

	raw := `[4, "id3", "NotImplemented", "not impl", {}]`

	var frame []any

	if err := json.Unmarshal([]byte(raw), &frame); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	typeCode, ok := frame[0].(float64)
	if !ok {
		t.Fatalf(
			"element 0: expected float64 from JSON decode, got %T",
			frame[0],
		)
	}

	if typeCode != float64(wire.MessageTypeError) {
		t.Errorf(
			"element 0: got %v, want float64(%d)",
			typeCode, wire.MessageTypeError,
		)
	}

	if _, err := wire.ParseError(frame); err != nil {
		t.Fatalf(
			"ParseError rejected a frame with float64 type code: %v",
			err,
		)
	}
}

// TestJSONFloat64CoercionIntCodeRejected verifies that a frame constructed
// with a native Go int type code (not float64) is rejected by ParseCall.
//
// This is the inverse of the acceptance test above: code that constructs
// frames manually must use float64 as the type code, or use the Encode
// function which always produces proper JSON.
func TestJSONFloat64CoercionIntCodeRejected(t *testing.T) {
	t.Parallel()

	// Construct a frame with int(2), not float64(2).
	frame := []any{
		int(wire.MessageTypeCall),
		"id1",
		"BootNotification",
		map[string]any{},
	}

	_, err := wire.ParseCall(frame)
	if err == nil {
		t.Fatal(
			"ParseCall accepted int(2) as a type code; " +
				"only float64(2) is valid from JSON-decoded frames",
		)
	}
}
