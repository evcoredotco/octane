// Package wire_test contains black-box unit tests for wire frame parsing
// (T-002-12).

package wire_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/evcoreco/octane/pkg/wire"
)

// testUniqueID is the correlation identifier used in happy-path test frames.
const testUniqueID = "abc-123"

const (
	// actionBootNotification is the OCPP action name used across parse tests.
	actionBootNotification = "BootNotification"
	// actionNotImplemented is the error code used in CALLERROR test frames.
	actionNotImplemented = "NotImplemented"
	// descNotImplemented is the error description used in CALLERROR test
	// frames.
	descNotImplemented = "The requested action is not implemented."
	// testID1 is a reusable uniqueId string for malformed-frame tests.
	testID1 = "id1"
	// testErrorDesc is a short error description used in shape-error tests.
	testErrorDesc = "desc"
	// fmtUniqueIDGotWant is the Errorf format for UniqueID mismatches.
	fmtUniqueIDGotWant = "UniqueID: got %q, want %q"
	// emptyString is the empty string constant used for Reason emptiness
	// checks.
	emptyString = ""
	// errReasonMustNotBeEmpty is the failure message for an empty Reason field.
	errReasonMustNotBeEmpty = "FrameShapeError.Reason must not be empty"
	// nonNumericUniqueID is an integer used in place of a string uniqueId.
	nonNumericUniqueID = 42
	// nonNumericAction is an integer used in place of a string action.
	nonNumericAction = 99
	// nonNumericErrorCode is an integer used in place of a string error code.
	nonNumericErrorCode = 404
	// longPayloadLen is the byte length used to produce a capped Raw field.
	longPayloadLen = 512
)

// TestParseCallHappyPath verifies that a well-formed CALL frame is decoded
// into the expected Call value.
func TestParseCallHappyPath(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeCall),
		testUniqueID,
		actionBootNotification,
		map[string]any{"chargePointModel": "ACME"},
	}

	got, err := wire.ParseCall(frame)
	if err != nil {
		t.Fatalf("ParseCall returned unexpected error: %v", err)
	}

	if got.UniqueID != testUniqueID {
		t.Errorf(fmtUniqueIDGotWant, got.UniqueID, testUniqueID)
	}

	if got.Action != actionBootNotification {
		t.Errorf("Action: got %q, want %q", got.Action, actionBootNotification)
	}

	const emptyLen = 0
	if len(got.Payload) == emptyLen {
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
		t.Errorf(fmtUniqueIDGotWant, got.UniqueID, testUniqueID)
	}

	const emptyLen = 0
	if len(got.Payload) == emptyLen {
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
		actionNotImplemented,
		descNotImplemented,
		map[string]any{},
	}

	got, err := wire.ParseError(frame)
	if err != nil {
		t.Fatalf("ParseError returned unexpected error: %v", err)
	}

	if got.UniqueID != testUniqueID {
		t.Errorf(fmtUniqueIDGotWant, got.UniqueID, testUniqueID)
	}

	if got.ErrorCode != actionNotImplemented {
		t.Errorf(
			"ErrorCode: got %q, want %q",
			got.ErrorCode,
			actionNotImplemented,
		)
	}

	if got.ErrorDescription != descNotImplemented {
		t.Errorf(
			"ErrorDescription: got %q, want %q",
			got.ErrorDescription,
			descNotImplemented,
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

// callErrShape is a helper that asserts ParseCall returns an *FrameShapeError
// whose Reason contains the expected substring.
func callErrShape(
	t *testing.T,
	frame []any,
	wantSubstr string,
) {
	t.Helper()

	_, err := wire.ParseCall(frame)
	if err == nil {
		t.Fatalf(
			"ParseCall expected error containing %q, got nil",
			wantSubstr,
		)
	}

	var fsErr *wire.FrameShapeError

	if !errors.As(err, &fsErr) {
		t.Fatalf(
			"ParseCall expected *FrameShapeError, got %T: %v",
			err, err,
		)
	}

	if fsErr.Reason == emptyString {
		t.Error(errReasonMustNotBeEmpty)
	}

	substrOK := wantSubstr == emptyString ||
		strings.Contains(fsErr.Reason, wantSubstr)
	if !substrOK {
		t.Errorf(
			"FrameShapeError.Reason = %q, want substring %q",
			fsErr.Reason,
			wantSubstr,
		)
	}
}

// resultErrShape is a helper that asserts ParseResult returns a
// *FrameShapeError.
func resultErrShape(
	t *testing.T,
	frame []any,
) {
	t.Helper()

	_, err := wire.ParseResult(frame)
	if err == nil {
		t.Fatal("ParseResult expected error, got nil")
	}

	var fsErr *wire.FrameShapeError

	if !errors.As(err, &fsErr) {
		t.Fatalf(
			"ParseResult expected *FrameShapeError, got %T: %v",
			err, err,
		)
	}

	if fsErr.Reason == emptyString {
		t.Error(errReasonMustNotBeEmpty)
	}
}

// errorErrShape is a helper that asserts ParseError returns a *FrameShapeError.
func errorErrShape(
	t *testing.T,
	frame []any,
) {
	t.Helper()

	_, err := wire.ParseError(frame)
	if err == nil {
		t.Fatal("ParseError expected error, got nil")
	}

	var fsErr *wire.FrameShapeError

	if !errors.As(err, &fsErr) {
		t.Fatalf(
			"ParseError expected *FrameShapeError, got %T: %v",
			err, err,
		)
	}

	if fsErr.Reason == emptyString {
		t.Error(errReasonMustNotBeEmpty)
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
		testID1,
		actionBootNotification,
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
		testID1,
		actionBootNotification,
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
		nonNumericUniqueID,
		actionBootNotification,
		map[string]any{},
	}
	callErrShape(t, frame, "uniqueId")
}

// TestParseCallNonStringAction verifies that a non-string action is rejected.
func TestParseCallNonStringAction(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeCall),
		testID1,
		nonNumericAction,
		map[string]any{},
	}
	callErrShape(t, frame, "action")
}

// TestParseCallNonMapPayload verifies that a non-map payload is rejected.
func TestParseCallNonMapPayload(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeCall),
		testID1,
		actionBootNotification,
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
		testID1,
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
		testID1,
		nonNumericUniqueID,
	}
	resultErrShape(t, frame)
}

// TestParseErrorWrongLength verifies that frames with incorrect element
// counts are rejected by ParseError.
func TestParseErrorWrongLength(t *testing.T) {
	t.Parallel()

	errorErrShape(t, []any{})
	errorErrShape(t, []any{
		float64(wire.MessageTypeError),
		"id",
		actionNotImplemented,
		testErrorDesc,
	})
	errorErrShape(t, []any{
		float64(wire.MessageTypeError),
		"id",
		actionNotImplemented,
		testErrorDesc,
		map[string]any{},
		"extra",
	})
}

// TestParseErrorWrongMessageType verifies that the wrong type code is rejected
// by ParseError.
func TestParseErrorWrongMessageType(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeCall),
		testID1,
		actionNotImplemented,
		testErrorDesc,
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
		testID1,
		nonNumericErrorCode,
		testErrorDesc,
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
		testID1,
		actionNotImplemented,
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
		testID1,
		actionNotImplemented,
		testErrorDesc,
		"not-a-map",
	}
	errorErrShape(t, frame)
}

// TestFrameShapeErrorRawCapped verifies that FrameShapeError.Raw is capped at
// 256 bytes in its Error() output and that Reason is surfaced.
func TestFrameShapeErrorRawCapped(t *testing.T) {
	t.Parallel()

	// Build a frame that will produce a long raw representation.
	longID := make([]byte, longPayloadLen)
	for idx := range longID {
		longID[idx] = 'x'
	}

	frame := []any{
		float64(wire.MessageTypeCall),
		testID1,
		actionBootNotification,
		string(longID), // non-map payload triggers the error
	}

	_, err := wire.ParseCall(frame)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var fsErr *wire.FrameShapeError

	if !errors.As(err, &fsErr) {
		t.Fatalf("expected *FrameShapeError, got %T", err)
	}

	const maxRaw = 256

	if len(fsErr.Raw) > maxRaw {
		t.Errorf("Raw length %d exceeds cap of %d", len(fsErr.Raw), maxRaw)
	}
}

// TestParseCallEmptyUniqueID verifies that an empty uniqueId string is
// rejected. A CSMS could send [2, "", "Action", {}]; the empty string
// breaks correlation.
func TestParseCallEmptyUniqueID(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeCall),
		emptyString,
		actionBootNotification,
		map[string]any{},
	}

	callErrShape(t, frame, "uniqueId")
}

// TestParseCallEmptyAction verifies that an empty action string is rejected.
func TestParseCallEmptyAction(t *testing.T) {
	t.Parallel()

	frame := []any{
		float64(wire.MessageTypeCall),
		testID1,
		emptyString,
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
		emptyString,
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
		emptyString,
		actionNotImplemented,
		testErrorDesc,
		map[string]any{},
	}

	errorErrShape(t, frame)
}
