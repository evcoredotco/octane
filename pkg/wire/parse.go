package wire

import (
	"encoding/json"
	"fmt"
)

// emptyStr is the named empty-string constant required by add-constant.
const emptyStr = ""

// msgTypeIDElem is the index of the messageTypeId element in OCPP-J frames.
const msgTypeIDElem = 0

// uniqueIDElem is the index of the uniqueId element in OCPP-J frames.
const uniqueIDElem = 1

// callActionElem is the index of the action element in CALL frames.
const callActionElem = 2

// callPayloadElem is the index of the payload element in CALL frames.
const callPayloadElem = 3

// resultPayloadElem is the index of the payload element in CALLRESULT frames.
const resultPayloadElem = 2

// errCodeElem is the index of the errorCode element in CALLERROR frames.
const errCodeElem = 2

// errDescElem is the index of the errorDescription element in CALLERROR frames.
const errDescElem = 3

// errDetailsElem is the index of the details element in CALLERROR frames.
const errDetailsElem = 4

// fmtWrongLen is the format string for wrong-length frame errors.
const fmtWrongLen = "expected array of length %d, got %d"

// fmtWrongMsgType is the format string for wrong messageTypeId errors.
const fmtWrongMsgType = "element 0 (messageTypeId) must be %d"

// fmtWrongElemType is the format string for wrong element type errors.
const fmtWrongElemType = "element %d (%s) must be a JSON object or null"

// fieldUniqueID is the OCPP-J field name for the correlation identifier.
const fieldUniqueID = "uniqueId"

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

// frameShape builds an *FrameShapeError with the given reason and the first
// 256 bytes of the marshalled frame for diagnostics.
func frameShape(frame []any, reason string) *FrameShapeError {
	const diagCap = 256

	raw := rawBytes(frame)

	if len(raw) > diagCap {
		raw = raw[:diagCap]
	}

	return &FrameShapeError{
		Reason: reason,
		Raw:    raw,
	}
}

// stringAt asserts that frame[idx] is a non-empty string and returns it.
// If the assertion fails or the value is empty it returns a non-nil
// *FrameShapeError.
func stringAt(
	frame []any,
	idx int,
	name string,
) (string, *FrameShapeError) {
	val, ok := frame[idx].(string)
	if !ok {
		return emptyStr, frameShape(
			frame,
			fmt.Sprintf(
				"element %d (%s) must be a non-empty string",
				idx,
				name,
			),
		)
	}

	if val == emptyStr {
		return emptyStr, frameShape(
			frame,
			fmt.Sprintf(
				"element %d (%s) must be a non-empty string",
				idx,
				name,
			),
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
) (json.RawMessage, *FrameShapeError) {
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
			fmt.Sprintf(fmtWrongElemType, idx, name),
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
// On any shape violation it returns a *FrameShapeError with a precise Reason
// and the first 256 bytes of the re-marshalled raw frame.
func ParseCall(frame []any) (Call, error) {
	const wantLen = 4

	if len(frame) != wantLen {
		return Call{}, frameShape(
			frame,
			fmt.Sprintf(fmtWrongLen, wantLen, len(frame)),
		)
	}

	typeCode, ok := frame[msgTypeIDElem].(float64)
	if !ok || typeCode != MessageTypeCall {
		return Call{}, frameShape(
			frame,
			fmt.Sprintf(fmtWrongMsgType, MessageTypeCall),
		)
	}

	uniqueID, fsErr := stringAt(frame, uniqueIDElem, fieldUniqueID)
	if fsErr != nil {
		return Call{}, fsErr
	}

	action, fsErr := stringAt(frame, callActionElem, "action")
	if fsErr != nil {
		return Call{}, fsErr
	}

	payload, fsErr := mapAt(frame, callPayloadElem, "payload")
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
// On any shape violation it returns a *FrameShapeError with a precise Reason
// and the first 256 bytes of the re-marshalled raw frame.
func ParseResult(frame []any) (Result, error) {
	const wantLen = 3

	if len(frame) != wantLen {
		return Result{}, frameShape(
			frame,
			fmt.Sprintf(fmtWrongLen, wantLen, len(frame)),
		)
	}

	typeCode, ok := frame[msgTypeIDElem].(float64)
	if !ok || typeCode != MessageTypeResult {
		return Result{}, frameShape(
			frame,
			fmt.Sprintf(fmtWrongMsgType, MessageTypeResult),
		)
	}

	uniqueID, fsErr := stringAt(frame, uniqueIDElem, fieldUniqueID)
	if fsErr != nil {
		return Result{}, fsErr
	}

	payload, fsErr := mapAt(frame, resultPayloadElem, "payload")
	if fsErr != nil {
		return Result{}, fsErr
	}

	return Result{
		UniqueID: uniqueID,
		Payload:  payload,
	}, nil
}

// ParseError decodes a pre-decoded OCPP-J CALLERROR frame into an Error.
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
// On any shape violation it returns a *FrameShapeError with a precise Reason
// and the first 256 bytes of the re-marshalled raw frame.
func ParseError(frame []any) (Error, error) {
	const wantLen = 5

	if len(frame) != wantLen {
		return Error{}, frameShape(
			frame,
			fmt.Sprintf(fmtWrongLen, wantLen, len(frame)),
		)
	}

	typeCode, ok := frame[msgTypeIDElem].(float64)
	if !ok || typeCode != MessageTypeError {
		return Error{}, frameShape(
			frame,
			fmt.Sprintf(fmtWrongMsgType, MessageTypeError),
		)
	}

	uniqueID, fsErr := stringAt(frame, uniqueIDElem, fieldUniqueID)
	if fsErr != nil {
		return Error{}, fsErr
	}

	errorCode, fsErr := stringAt(frame, errCodeElem, "errorCode")
	if fsErr != nil {
		return Error{}, fsErr
	}

	errorDesc, fsErr := stringAt(frame, errDescElem, "errorDescription")
	if fsErr != nil {
		return Error{}, fsErr
	}

	details, fsErr := mapAt(frame, errDetailsElem, "details")
	if fsErr != nil {
		return Error{}, fsErr
	}

	return Error{
		UniqueID:         uniqueID,
		ErrorCode:        errorCode,
		ErrorDescription: errorDesc,
		Details:          details,
	}, nil
}
