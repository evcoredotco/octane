// Package wire_test contains black-box unit tests for wire frame parsing
// (T-002-12).
package wire_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/octane-project/octane/pkg/wire"
)

// testUniqueID is the correlation identifier used in happy-path test frames.
const testUniqueID = "abc-123"

// TestParseCallHappyPath verifies that a well-formed CALL frame is decoded
// into the expected Call value.
func TestParseCallHappyPath(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeCall),
		testUniqueID,
		"BootNotification",
		map[string]any{"chargePointModel": "ACME"},
	}

	got, err := wire.ParseCall(frame)
	if err != nil {
		t.Fatalf("ParseCall returned unexpected error: %v", err)
	}

	if got.UniqueID != testUniqueID {
		t.Errorf("UniqueID: got %q, want %q", got.UniqueID, testUniqueID)
	}

	if got.Action != "BootNotification" {
		t.Errorf("Action: got %q, want %q", got.Action, "BootNotification")
	}

	if len(got.Payload) == 0 {
		t.Error("Payload must not be empty for a non-empty map")
	}
}

// TestParseResultHappyPath verifies that a well-formed CALLRESULT frame is
// decoded correctly.
func TestParseResultHappyPath(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeResult),
		testUniqueID,
		map[string]any{"status": "Accepted"},
	}

	got, err := wire.ParseResult(frame)
	if err != nil {
		t.Fatalf("ParseResult returned unexpected error: %v", err)
	}

	if got.UniqueID != testUniqueID {
		t.Errorf("UniqueID: got %q, want %q", got.UniqueID, testUniqueID)
	}

	if len(got.Payload) == 0 {
		t.Error("Payload must not be empty for a non-empty map")
	}
}

// TestParseErrorHappyPath verifies that a well-formed CALLERROR frame is
// decoded correctly.
func TestParseErrorHappyPath(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeError),
		testUniqueID,
		"NotImplemented",
		"The requested action is not implemented.",
		map[string]any{},
	}

	got, err := wire.ParseError(frame)
	if err != nil {
		t.Fatalf("ParseError returned unexpected error: %v", err)
	}

	if got.UniqueID != testUniqueID {
		t.Errorf("UniqueID: got %q, want %q", got.UniqueID, testUniqueID)
	}

	if got.ErrorCode != "NotImplemented" {
		t.Errorf("ErrorCode: got %q, want %q", got.ErrorCode, "NotImplemented")
	}

	if got.ErrorDescription != "The requested action is not implemented." {
		t.Errorf(
			"ErrorDescription: got %q, want %q",
			got.ErrorDescription,
			"The requested action is not implemented.",
		)
	}
}

// TestParseErrorNilDetails verifies that a CALLERROR with a null details
// element is accepted without error.
func TestParseErrorNilDetails(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeError),
		testUniqueID,
		"InternalError",
		"Something went wrong.",
		nil,
	}

	_, err := wire.ParseError(frame)
	if err != nil {
		t.Fatalf(
			"ParseError with nil details returned unexpected error: %v",
			err,
		)
	}
}

// callErrShape is a helper that asserts ParseCall returns an *ErrFrameShape
// whose Reason contains the expected substring.
func callErrShape(
	t *testing.T,
	frame []any,
	wantSubstr string,
) {
	t.Helper()

	_, err := wire.ParseCall(frame)
	if err == nil {
		t.Fatalf("ParseCall expected error containing %q, got nil", wantSubstr)
	}

	var fsErr *wire.ErrFrameShape

	if !errors.As(err, &fsErr) {
		t.Fatalf("ParseCall expected *ErrFrameShape, got %T: %v", err, err)
	}

	if fsErr.Reason == "" {
		t.Error("ErrFrameShape.Reason must not be empty")
	}

	if wantSubstr != "" && !strings.Contains(fsErr.Reason, wantSubstr) {
		t.Errorf("ErrFrameShape.Reason = %q, want substring %q", fsErr.Reason, wantSubstr)
	}
}

// resultErrShape is a helper that asserts ParseResult returns an *ErrFrameShape.
func resultErrShape(
	t *testing.T,
	frame []any,
) {
	t.Helper()

	_, err := wire.ParseResult(frame)
	if err == nil {
		t.Fatal("ParseResult expected error, got nil")
	}

	var fsErr *wire.ErrFrameShape

	if !errors.As(err, &fsErr) {
		t.Fatalf("ParseResult expected *ErrFrameShape, got %T: %v", err, err)
	}

	if fsErr.Reason == "" {
		t.Error("ErrFrameShape.Reason must not be empty")
	}
}

// errorErrShape is a helper that asserts ParseError returns an *ErrFrameShape.
func errorErrShape(
	t *testing.T,
	frame []any,
) {
	t.Helper()

	_, err := wire.ParseError(frame)
	if err == nil {
		t.Fatal("ParseError expected error, got nil")
	}

	var fsErr *wire.ErrFrameShape

	if !errors.As(err, &fsErr) {
		t.Fatalf("ParseError expected *ErrFrameShape, got %T: %v", err, err)
	}

	if fsErr.Reason == "" {
		t.Error("ErrFrameShape.Reason must not be empty")
	}
}

// TestParseCallWrongLength verifies that frames with incorrect element counts
// are rejected.
func TestParseCallWrongLength(t *testing.T) {
	t.Parallel()

	callErrShape(t, []any{}, "got 0")
	callErrShape(t, []any{float64(2), "id"}, "got 2")
	callErrShape(
		t,
		[]any{float64(2), "id", "Act", map[string]any{}, "extra"},
		"got 5",
	)
}

// TestParseCallWrongMessageType verifies that a frame with the wrong message
// type code at element 0 is rejected.
func TestParseCallWrongMessageType(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeResult),
		"id1",
		"BootNotification",
		map[string]any{},
	}
	callErrShape(t, frame, "must be 2")
}

// TestParseCallNonNumericTypeCode verifies that a non-float64 type code is
// rejected.
func TestParseCallNonNumericTypeCode(t *testing.T) {
	t.Parallel()

	frame := []any{
		"2",
		"id1",
		"BootNotification",
		map[string]any{},
	}
	callErrShape(t, frame, "must be 2")
}

// TestParseCallNonStringUniqueId verifies that a non-string uniqueId is
// rejected.
func TestParseCallNonStringUniqueId(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeCall),
		42,
		"BootNotification",
		map[string]any{},
	}
	callErrShape(t, frame, "uniqueId")
}

// TestParseCallNonStringAction verifies that a non-string action is rejected.
func TestParseCallNonStringAction(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeCall),
		"id1",
		99,
		map[string]any{},
	}
	callErrShape(t, frame, "action")
}

// TestParseCallNonMapPayload verifies that a non-map payload is rejected.
func TestParseCallNonMapPayload(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeCall),
		"id1",
		"BootNotification",
		"not-a-map",
	}
	callErrShape(t, frame, "payload")
}

// TestParseResultWrongLength verifies that frames with incorrect element
// counts are rejected by ParseResult.
func TestParseResultWrongLength(t *testing.T) {
	t.Parallel()

	resultErrShape(t, []any{})
	resultErrShape(t, []any{float64(3), "id"})
	resultErrShape(t, []any{float64(3), "id", map[string]any{}, "extra"})
}

// TestParseResultWrongMessageType verifies that the wrong type code at
// element 0 is rejected by ParseResult.
func TestParseResultWrongMessageType(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeCall),
		"id1",
		map[string]any{},
	}
	resultErrShape(t, frame)
}

// TestParseResultNonStringUniqueId verifies that a non-string uniqueId is
// rejected by ParseResult.
func TestParseResultNonStringUniqueId(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeResult),
		true,
		map[string]any{},
	}
	resultErrShape(t, frame)
}

// TestParseResultNonMapPayload verifies that a non-map payload is rejected by
// ParseResult.
func TestParseResultNonMapPayload(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeResult),
		"id1",
		42,
	}
	resultErrShape(t, frame)
}

// TestParseErrorWrongLength verifies that frames with incorrect element
// counts are rejected by ParseError.
func TestParseErrorWrongLength(t *testing.T) {
	t.Parallel()

	errorErrShape(t, []any{})
	errorErrShape(t, []any{float64(4), "id", "NotImplemented", "desc"})
	errorErrShape(t, []any{
		float64(4), "id", "NotImplemented", "desc", map[string]any{}, "extra",
	})
}

// TestParseErrorWrongMessageType verifies that the wrong type code is rejected
// by ParseError.
func TestParseErrorWrongMessageType(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeCall),
		"id1",
		"NotImplemented",
		"desc",
		map[string]any{},
	}
	errorErrShape(t, frame)
}

// TestParseErrorNonStringErrorCode verifies that a non-string errorCode is
// rejected.
func TestParseErrorNonStringErrorCode(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeError),
		"id1",
		404,
		"desc",
		map[string]any{},
	}
	errorErrShape(t, frame)
}

// TestParseErrorNonStringDescription verifies that a non-string
// errorDescription is rejected.
func TestParseErrorNonStringDescription(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeError),
		"id1",
		"NotImplemented",
		false,
		map[string]any{},
	}
	errorErrShape(t, frame)
}

// TestParseErrorNonMapDetails verifies that a non-map, non-nil details element
// is rejected.
func TestParseErrorNonMapDetails(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeError),
		"id1",
		"NotImplemented",
		"desc",
		"not-a-map",
	}
	errorErrShape(t, frame)
}

// TestErrFrameShapeRawCapped verifies that ErrFrameShape.Raw is capped at
// 256 bytes in its Error() output and that Reason is surfaced.
func TestErrFrameShapeRawCapped(t *testing.T) {
	t.Parallel()

	// Build a frame that will produce a long raw representation.
	longID := make([]byte, 512)
	for idx := range longID {
		longID[idx] = 'x'
	}

	frame := []any{
		float64(wire.MessageTypeCall),
		"id1",
		"BootNotification",
		string(longID), // non-map payload triggers the error
	}

	_, err := wire.ParseCall(frame)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var fsErr *wire.ErrFrameShape

	if !errors.As(err, &fsErr) {
		t.Fatalf("expected *ErrFrameShape, got %T", err)
	}

	const maxRaw = 256

	if len(fsErr.Raw) > maxRaw {
		t.Errorf("Raw length %d exceeds cap of %d", len(fsErr.Raw), maxRaw)
	}
}

// TestParseCallEmptyUniqueID verifies that an empty uniqueId string is rejected.
// A CSMS could send [2, "", "Action", {}]; the empty string breaks correlation.
func TestParseCallEmptyUniqueID(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeCall),
		"",
		"BootNotification",
		map[string]any{},
	}

	callErrShape(t, frame, "uniqueId")
}

// TestParseCallEmptyAction verifies that an empty action string is rejected.
func TestParseCallEmptyAction(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeCall),
		"id1",
		"",
		map[string]any{},
	}

	callErrShape(t, frame, "action")
}

// TestParseResultEmptyUniqueID verifies that an empty uniqueId in a CALLRESULT
// is rejected.
func TestParseResultEmptyUniqueID(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeResult),
		"",
		map[string]any{},
	}

	resultErrShape(t, frame)
}

// TestParseErrorEmptyUniqueID verifies that an empty uniqueId in a CALLERROR
// is rejected.
func TestParseErrorEmptyUniqueID(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeError),
		"",
		"NotImplemented",
		"desc",
		map[string]any{},
	}

	errorErrShape(t, frame)
}
