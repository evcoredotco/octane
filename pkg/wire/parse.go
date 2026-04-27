package wire

import (
	"encoding/json"
	"fmt"
)

// rawBytes converts a []any frame back to JSON bytes for diagnostic use.
// It returns an empty slice on marshal failure rather than propagating
// a secondary error.
func rawBytes(frame []any) []byte {
	raw, err := json.Marshal(frame)
	if err != nil {
		return []byte{}
	}

	return raw
}

// frameShape builds an *ErrFrameShape with the given reason and the first
// 256 bytes of the marshalled frame for diagnostics.
func frameShape(frame []any, reason string) *ErrFrameShape {
	const diagCap = 256

	raw := rawBytes(frame)

	if len(raw) > diagCap {
		raw = raw[:diagCap]
	}

	return &ErrFrameShape{
		Reason: reason,
		Raw:    raw,
	}
}

// stringAt asserts that frame[idx] is a non-empty string and returns it.
// If the assertion fails it returns a non-nil *ErrFrameShape.
func stringAt(
	frame []any,
	idx int,
	name string,
) (string, *ErrFrameShape) {
	val, ok := frame[idx].(string)
	if !ok {
		return "", frameShape(
			frame,
			fmt.Sprintf("element %d (%s) must be a string", idx, name),
		)
	}

	return val, nil
}

// mapAt asserts that frame[idx] is a map[string]any (or nil/absent) and
// returns the marshalled json.RawMessage. A nil map marshals to "null".
func mapAt(
	frame []any,
	idx int,
	name string,
) (json.RawMessage, *ErrFrameShape) {
	raw, err := json.Marshal(frame[idx])
	if err != nil {
		return nil, frameShape(
			frame,
			fmt.Sprintf("element %d (%s) could not be marshalled: %s",
				idx, name, err.Error()),
		)
	}

	switch frame[idx].(type) {
	case map[string]any, nil:
		return raw, nil
	default:
		return nil, frameShape(
			frame,
			fmt.Sprintf("element %d (%s) must be a JSON object or null", idx, name),
		)
	}
}

// ParseCall decodes a pre-decoded OCPP-J CALL frame into a Call value.
//
// A valid CALL frame has the shape [2, "<uniqueId>", "<Action>", {payload}].
// ParseCall validates:
//   - exactly 4 elements
//   - element[0] is float64(2)
//   - element[1] and element[2] are non-empty strings
//   - element[3] is a JSON object
//
// On any shape violation it returns a *ErrFrameShape with a precise Reason
// and the first 256 bytes of the re-marshalled raw frame.
func ParseCall(frame []any) (Call, error) {
	const wantLen = 4

	if len(frame) != wantLen {
		return Call{}, frameShape(
			frame,
			fmt.Sprintf(
				"expected array of length %d, got %d",
				wantLen, len(frame),
			),
		)
	}

	typeCode, ok := frame[0].(float64)
	if !ok || typeCode != MessageTypeCall {
		return Call{}, frameShape(
			frame,
			fmt.Sprintf(
				"element 0 (messageTypeId) must be %d", MessageTypeCall,
			),
		)
	}

	uniqueID, fsErr := stringAt(frame, 1, "uniqueId")
	if fsErr != nil {
		return Call{}, fsErr
	}

	action, fsErr := stringAt(frame, 2, "action")
	if fsErr != nil {
		return Call{}, fsErr
	}

	payload, fsErr := mapAt(frame, 3, "payload")
	if fsErr != nil {
		return Call{}, fsErr
	}

	return Call{
		UniqueID: uniqueID,
		Action:   action,
		Payload:  payload,
	}, nil
}

// ParseResult decodes a pre-decoded OCPP-J CALLRESULT frame into a Result.
//
// A valid CALLRESULT frame has the shape [3, "<uniqueId>", {payload}].
// ParseResult validates:
//   - exactly 3 elements
//   - element[0] is float64(3)
//   - element[1] is a non-empty string
//   - element[2] is a JSON object
//
// On any shape violation it returns a *ErrFrameShape with a precise Reason
// and the first 256 bytes of the re-marshalled raw frame.
func ParseResult(frame []any) (Result, error) {
	const wantLen = 3

	if len(frame) != wantLen {
		return Result{}, frameShape(
			frame,
			fmt.Sprintf(
				"expected array of length %d, got %d",
				wantLen, len(frame),
			),
		)
	}

	typeCode, ok := frame[0].(float64)
	if !ok || typeCode != MessageTypeResult {
		return Result{}, frameShape(
			frame,
			fmt.Sprintf(
				"element 0 (messageTypeId) must be %d", MessageTypeResult,
			),
		)
	}

	uniqueID, fsErr := stringAt(frame, 1, "uniqueId")
	if fsErr != nil {
		return Result{}, fsErr
	}

	payload, fsErr := mapAt(frame, 2, "payload")
	if fsErr != nil {
		return Result{}, fsErr
	}

	return Result{
		UniqueID: uniqueID,
		Payload:  payload,
	}, nil
}

// ParseError decodes a pre-decoded OCPP-J CALLERROR frame into a WireError.
//
// A valid CALLERROR frame has the shape:
//
//	[4, "<uniqueId>", "<errorCode>", "<errorDescription>", {details}]
//
// ParseError validates:
//   - exactly 5 elements
//   - element[0] is float64(4)
//   - elements[1], [2], and [3] are strings
//   - element[4] is a JSON object or null
//
// On any shape violation it returns a *ErrFrameShape with a precise Reason
// and the first 256 bytes of the re-marshalled raw frame.
func ParseError(frame []any) (WireError, error) {
	const wantLen = 5

	if len(frame) != wantLen {
		return WireError{}, frameShape(
			frame,
			fmt.Sprintf(
				"expected array of length %d, got %d",
				wantLen, len(frame),
			),
		)
	}

	typeCode, ok := frame[0].(float64)
	if !ok || typeCode != MessageTypeError {
		return WireError{}, frameShape(
			frame,
			fmt.Sprintf(
				"element 0 (messageTypeId) must be %d", MessageTypeError,
			),
		)
	}

	uniqueID, fsErr := stringAt(frame, 1, "uniqueId")
	if fsErr != nil {
		return WireError{}, fsErr
	}

	errorCode, fsErr := stringAt(frame, 2, "errorCode")
	if fsErr != nil {
		return WireError{}, fsErr
	}

	errorDesc, fsErr := stringAt(frame, 3, "errorDescription")
	if fsErr != nil {
		return WireError{}, fsErr
	}

	details, fsErr := mapAt(frame, 4, "details")
	if fsErr != nil {
		return WireError{}, fsErr
	}

	return WireError{
		UniqueID:         uniqueID,
		ErrorCode:        errorCode,
		ErrorDescription: errorDesc,
		Details:          details,
	}, nil
}
