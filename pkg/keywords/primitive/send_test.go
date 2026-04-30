// Package primitive_test exercises the send primitive keywords
// (spec 004 §10, items 4–5) against mock.MockState and mock.MockStation.
//
// Task: T-004-14
// AC2: "send raw frame {frame:any} on station {station:string}" encodes the
// frame and emits it on the station's wire.
// AC2: "send raw bytes {bytes:string} on station {station:string}" decodes a
// hex string and sends the resulting frame.

package primitive_test

import (
	"context"
	"errors"
	"testing"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
	_ "github.com/evcoreco/octane/pkg/keywords/primitive" // blank import
)

// ── Package-level sentinel errors ────────────────────────────────────────────

// errSendStub is the sentinel returned by the mock station during error tests.
var errSendStub = errors.New("stub: send failed")

// ── Named constants ───────────────────────────────────────────────────────────

const (
	// handleSend is the station handle name used across send tests.
	handleSend = "CP04"

	// patternSendRawFrame is the step text for the send-raw-frame keyword.
	patternSendRawFrame = "send raw frame {frame:any}" +
		" on station {station:string}"

	// patternSendRawBytes is the step text for the send-raw-bytes keyword.
	patternSendRawBytes = "send raw bytes {bytes:string}" +
		" on station {station:string}"

	// hexValidFrame is a valid hex-encoded OCPP-J CALL frame:
	// [2,"id","Action",{}] → as JSON bytes encoded to hex.
	// JSON: [2,"id","Action",{}]
	hexValidFrame = "5b322c226964222c22416374696f6e222c7b7d5d"

	// hexMalformed is a hex string that decodes to bytes that are not
	// valid JSON (and therefore not a JSON array).
	hexMalformed = "zzzz"

	// ocppCallType is the OCPP-J message type for Call frames.
	ocppCallType = float64(2)

	// wantOneSent is the expected SentFrames count after one successful send.
	wantOneSent = 1

	// noSentFrames is the expected SentFrames count when no frame was sent.
	noSentFrames = 0
)

// ── sendRawFrame tests ────────────────────────────────────────────────────────

// Test_primitive_sendRawFrame_HappyPath verifies that the keyword delivers the
// frame to SentFrames() exactly once and with the correct contents (AC2).
func Test_primitive_sendRawFrame_HappyPath(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	station := mock.NewMockStation()
	state.RegisterStation(handleSend, station)

	keywordFunc := resolveFunc(t, patternSendRawFrame)

	// Invariant: the frame passed as []any must appear in SentFrames().
	frame := []any{
		ocppCallType,
		"msg-001",
		"BootNotification",
		map[string]any{},
	}

	args := api.NewArgs(map[string]any{
		"frame":   frame,
		"station": handleSend,
	})

	err := keywordFunc(context.Background(), state, args)
	if err != nil {
		t.Fatalf("sendRawFrame: unexpected error: %v", err)
	}

	sent := station.SentFrames()

	if len(sent) != wantOneSent {
		t.Fatalf("SentFrames(): want 1 frame, got %d", len(sent))
	}

	if len(sent[0]) != len(frame) {
		t.Errorf(
			"SentFrames()[0] length: want %d, got %d",
			len(frame),
			len(sent[0]),
		)
	}
}

// Test_primitive_sendRawFrame_FrameNotSlice verifies that passing a non-[]any
// value for {frame:any} returns an error (AC2).
func Test_primitive_sendRawFrame_FrameNotSlice(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	station := mock.NewMockStation()
	state.RegisterStation(handleSend, station)

	keywordFunc := resolveFunc(t, patternSendRawFrame)

	// Invariant: a non-[]any frame value must produce a non-nil error.
	args := api.NewArgs(map[string]any{
		"frame":   "not-a-slice",
		"station": handleSend,
	})

	err := keywordFunc(context.Background(), state, args)
	if err == nil {
		t.Fatal("sendRawFrame with string frame: want error, got nil")
	}
}

// Test_primitive_sendRawFrame_SendError verifies that an error from
// Station.Send is wrapped and returned by the keyword (AC2).
func Test_primitive_sendRawFrame_SendError(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	station := mock.NewMockStation()

	// Configure the mock to fail on Send.
	station.SetSendError(errSendStub)

	state.RegisterStation(handleSend, station)

	keywordFunc := resolveFunc(t, patternSendRawFrame)

	frame := []any{
		ocppCallType,
		"msg-002",
		"BootNotification",
		map[string]any{},
	}

	args := api.NewArgs(map[string]any{
		"frame":   frame,
		"station": handleSend,
	})

	err := keywordFunc(context.Background(), state, args)

	// Invariant: a Send failure must be surfaced by the keyword.
	if err == nil {
		t.Fatal("sendRawFrame on Send error: want error, got nil")
	}

	if !errors.Is(err, errSendStub) {
		t.Errorf("sendRawFrame: want errors.Is(err, errSendStub), got %v", err)
	}
}

// ── sendRawBytes tests ────────────────────────────────────────────────────────

// Test_primitive_sendRawBytes_HappyPath verifies that a valid hex string is
// decoded, parsed as a JSON array, and delivered via Station.Send (AC2).
func Test_primitive_sendRawBytes_HappyPath(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	station := mock.NewMockStation()
	state.RegisterStation(handleSend, station)

	keywordFunc := resolveFunc(t, patternSendRawBytes)

	args := api.NewArgs(map[string]any{
		"bytes":   hexValidFrame,
		"station": handleSend,
	})

	err := keywordFunc(context.Background(), state, args)
	if err != nil {
		t.Fatalf("sendRawBytes with valid hex: unexpected error: %v", err)
	}

	// Invariant: exactly one frame must have been sent after decoding.
	sent := station.SentFrames()

	if len(sent) != wantOneSent {
		t.Fatalf("SentFrames(): want 1 frame, got %d", len(sent))
	}
}

// Test_primitive_sendRawBytes_MalformedHex verifies that an invalid hex string
// returns an error without sending any frame (AC2).
func Test_primitive_sendRawBytes_MalformedHex(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	station := mock.NewMockStation()
	state.RegisterStation(handleSend, station)

	keywordFunc := resolveFunc(t, patternSendRawBytes)

	// Invariant: a hex-decode failure must produce a non-nil error.
	args := api.NewArgs(map[string]any{
		"bytes":   hexMalformed,
		"station": handleSend,
	})

	err := keywordFunc(context.Background(), state, args)
	if err == nil {
		t.Fatal("sendRawBytes with malformed hex: want error, got nil")
	}

	// Invariant: no frame should have been sent on decode failure.
	sent := station.SentFrames()

	if len(sent) != noSentFrames {
		t.Errorf(
			"sendRawBytes with malformed hex: SentFrames() want %d, got %d",
			noSentFrames,
			len(sent),
		)
	}
}
