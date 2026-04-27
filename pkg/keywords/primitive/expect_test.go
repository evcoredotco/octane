// Package primitive_test exercises the expect primitive keywords
// (spec 004 §10, items 6–7) against mock.MockState and mock.MockStation.
//
// Task: T-004-14
// AC3: "expect any frame on station {station:string} within {timeout:duration}"
// returns nil when a frame arrives within the timeout and stashes the frame.
// AC4: When no frame arrives within the timeout the keyword returns *ErrTimeout
// carrying the configured timeout and the deterministic-clock deadline.
package primitive_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
	// Named import registers all primitive keywords at init() time and
	// provides access to primitive.ErrTimeout for typed error assertions.
	"github.com/evcoreco/octane/pkg/keywords/primitive"
)

// ── Named constants ───────────────────────────────────────────────────────────

const (
	// handleExpect is the station handle name used across expect tests.
	handleExpect = "CP05"

	// patternExpectAny is the step text for the expect-any-frame keyword.
	patternExpectAny = "expect any frame on station {station:string} within {timeout:duration}"

	// patternExpectOfType is the step text for the expect-frame-of-type keyword.
	patternExpectOfType = "expect a frame of type {messageType:int} on station" +
		" {station:string} within {timeout:duration}"

	// timeoutShort is a very short deadline used to trigger timeout behaviour.
	timeoutShort = time.Millisecond

	// timeoutGenerous is a long deadline used in happy-path tests where the
	// mock returns a frame immediately and no real time elapses.
	timeoutGenerous = 10 * time.Second

	// messageTypeCALL is the OCPP-J message-type code for a CALL frame.
	messageTypeCALL = 2

	// messageTypeCALLRESULT is the OCPP-J message-type code for a CALLRESULT.
	messageTypeCALLRESULT = 3
)

// frozenNow is a fixed deterministic clock value injected into MockState
// so that deadline calculations are reproducible across runs
// (constitution principle IV).
var frozenNow = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// ── expectAnyFrame tests ──────────────────────────────────────────────────────

// Test_primitive_expectAnyFrame_HappyPath verifies that when a frame is
// pre-queued on the mock station the keyword returns nil (AC3).
func Test_primitive_expectAnyFrame_HappyPath(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	state.SetNow(frozenNow)

	station := mock.NewMockStation()
	station.QueueFrame(
		[]any{
			float64(messageTypeCALL),
			"msg-100",
			"BootNotification",
			map[string]any{},
		},
	)

	state.RegisterStation(handleExpect, station)

	keywordFunc := resolveFunc(t, patternExpectAny)

	args := api.NewArgs(map[string]any{
		"station": handleExpect,
		"timeout": timeoutGenerous,
	})

	// Invariant: keyword must return nil when a frame is available.
	err := keywordFunc(context.Background(), state, args)
	if err != nil {
		t.Fatalf("expectAnyFrame (happy path): unexpected error: %v", err)
	}
}

// Test_primitive_expectAnyFrame_Timeout verifies that when no frame is
// available and the context deadline elapses the keyword returns *ErrTimeout
// (AC4).
func Test_primitive_expectAnyFrame_Timeout(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	state.SetNow(frozenNow)

	// Empty mock station — Expect will block until context expires.
	station := mock.NewMockStation()
	state.RegisterStation(handleExpect, station)

	keywordFunc := resolveFunc(t, patternExpectAny)

	args := api.NewArgs(map[string]any{
		"station": handleExpect,
		"timeout": timeoutShort,
	})

	err := keywordFunc(context.Background(), state, args)

	// Invariant: a missing frame must produce *ErrTimeout.
	if err == nil {
		t.Fatal("expectAnyFrame (no frames): want error, got nil")
	}

	var timeoutErr *primitive.ErrTimeout

	if !errors.As(err, &timeoutErr) {
		t.Fatalf(
			"expectAnyFrame (no frames): want *primitive.ErrTimeout via errors.As, got %T: %v",
			err,
			err,
		)
	}

	// Invariant: ErrTimeout must carry the configured station handle.
	if timeoutErr.Station != handleExpect {
		t.Errorf(
			"ErrTimeout.Station: want %q, got %q",
			handleExpect,
			timeoutErr.Station,
		)
	}

	// Invariant: ErrTimeout must carry the configured timeout duration.
	if timeoutErr.Timeout != timeoutShort {
		t.Errorf(
			"ErrTimeout.Timeout: want %v, got %v",
			timeoutShort,
			timeoutErr.Timeout,
		)
	}

	// Invariant: ErrTimeout.Deadline must equal frozenNow + timeout (deterministic clock).
	wantDeadline := frozenNow.Add(timeoutShort)

	if !timeoutErr.Deadline.Equal(wantDeadline) {
		t.Errorf(
			"ErrTimeout.Deadline: want %v, got %v",
			wantDeadline,
			timeoutErr.Deadline,
		)
	}
}

// ── expectFrameOfType tests ───────────────────────────────────────────────────

// Test_primitive_expectFrameOfType_HappyPath verifies that when a queued frame
// has the correct message-type code at index 0 the keyword returns nil (AC3).
func Test_primitive_expectFrameOfType_HappyPath(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	state.SetNow(frozenNow)

	station := mock.NewMockStation()
	// Queue a CALLRESULT frame (type 3).
	station.QueueFrame([]any{
		float64(
			messageTypeCALLRESULT,
		), "msg-200", map[string]any{"status": "Accepted"},
	})

	state.RegisterStation(handleExpect, station)

	keywordFunc := resolveFunc(t, patternExpectOfType)

	args := api.NewArgs(map[string]any{
		"messageType": messageTypeCALLRESULT,
		"station":     handleExpect,
		"timeout":     timeoutGenerous,
	})

	// Invariant: keyword must return nil when a matching frame is queued.
	err := keywordFunc(context.Background(), state, args)
	if err != nil {
		t.Fatalf("expectFrameOfType (happy path): unexpected error: %v", err)
	}
}

// Test_primitive_expectFrameOfType_WrongTypeThenTimeout verifies that frames
// with the wrong message-type code are silently skipped and an eventual timeout
// produces *ErrTimeout (AC4).
func Test_primitive_expectFrameOfType_WrongTypeThenTimeout(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	state.SetNow(frozenNow)

	station := mock.NewMockStation()
	// Queue only CALL frames (type 2); the keyword expects type 3.
	station.QueueFrame(
		[]any{
			float64(messageTypeCALL),
			"msg-300",
			"BootNotification",
			map[string]any{},
		},
	)

	state.RegisterStation(handleExpect, station)

	keywordFunc := resolveFunc(t, patternExpectOfType)

	args := api.NewArgs(map[string]any{
		"messageType": messageTypeCALLRESULT,
		"station":     handleExpect,
		"timeout":     timeoutShort,
	})

	err := keywordFunc(context.Background(), state, args)

	// Invariant: wrong-type frames must be skipped; eventual timeout returns *ErrTimeout.
	if err == nil {
		t.Fatal(
			"expectFrameOfType (wrong type): want error, got nil",
		)
	}

	var timeoutErr *primitive.ErrTimeout

	if !errors.As(err, &timeoutErr) {
		t.Fatalf(
			"expectFrameOfType (wrong type): want *primitive.ErrTimeout via errors.As, got %T: %v",
			err,
			err,
		)
	}

	// Invariant: ErrTimeout must identify the correct station handle.
	if timeoutErr.Station != handleExpect {
		t.Errorf(
			"ErrTimeout.Station: want %q, got %q",
			handleExpect,
			timeoutErr.Station,
		)
	}
}
