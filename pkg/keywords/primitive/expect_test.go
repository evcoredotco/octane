// Package primitive_test exercises the expect primitive keywords
// (spec 004 §10, items 6–7) against mock.MockState and mock.MockStation.
//
// Task: T-004-14
// AC3: "expect any frame on station {station:string} within {timeout:duration}"
// returns nil when a frame arrives within the timeout and stashes the frame.
// AC4: When no frame arrives within the timeout the keyword returns
// *TimeoutError carrying the configured timeout and the deterministic-clock
// deadline.

package primitive_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
	"github.com/evcoreco/octane/pkg/keywords/primitive" // exposes TimeoutError
)

// ── Named constants ──────────────────────────────────────────────────────────

const (
	// frozenYear is the year component of the deterministic clock value.
	frozenYear = 2026

	// handleExpect is the station handle name used across expect tests.
	handleExpect = "CP05"

	// patternExpectAny is the step text for the expect-any-frame keyword.
	patternExpectAny = "expect any frame on station {station:string}" +
		" within {timeout:duration}"

	// patternExpectOfType is the step text for the expect-frame-of-type
	// keyword.
	patternExpectOfType = "expect a frame of type {messageType:int}" +
		" on station {station:string} within {timeout:duration}"

	// timeoutShort is a very short deadline used to trigger timeout behaviour.
	timeoutShort = time.Millisecond

	// timeoutGenerous is a long deadline used in happy-path tests where the
	// mock returns a frame immediately and no real time elapses.
	timeoutGenerous = 10 * time.Second

	// messageTypeCALL is the OCPP-J message-type code for a CALL frame.
	messageTypeCALL = 2

	// messageTypeCALLRESULT is the OCPP-J message-type code for a CALLRESULT.
	messageTypeCALLRESULT = 3

	// frozenDayFirst is the first day of the month used in frozenNow.
	frozenDayFirst = 1

	// zeroTimeField is the zero value for hour/min/sec/nsec in time.Date.
	zeroTimeField = 0
)

// frozenNow returns a fixed deterministic clock value to inject into MockState
// so that deadline calculations are reproducible across runs
// (constitution principle IV).
func frozenNow() time.Time {
	return time.Date(
		frozenYear, time.January, frozenDayFirst,
		zeroTimeField, zeroTimeField, zeroTimeField, zeroTimeField,
		time.UTC,
	)
}

// ── expectAnyFrame tests ─────────────────────────────────────────────────────

// Test_primitive_expectAnyFrame_HappyPath verifies that when a frame is
// pre-queued on the mock station the keyword returns nil (AC3).
func Test_primitive_expectAnyFrame_HappyPath(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	state.SetNow(frozenNow())

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
// available and the context deadline elapses the keyword returns *TimeoutError
// (AC4).
func Test_primitive_expectAnyFrame_Timeout(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	state.SetNow(frozenNow())

	// Empty mock station — Expect will block until context expires.
	station := mock.NewMockStation()
	state.RegisterStation(handleExpect, station)

	keywordFunc := resolveFunc(t, patternExpectAny)

	args := api.NewArgs(map[string]any{
		"station": handleExpect,
		"timeout": timeoutShort,
	})

	err := keywordFunc(context.Background(), state, args)

	// Invariant: a missing frame must produce *TimeoutError.
	if err == nil {
		t.Fatal("expectAnyFrame (no frames): want error, got nil")
	}

	var timeoutErr *primitive.TimeoutError

	if !errors.As(err, &timeoutErr) {
		t.Fatalf(
			"expectAnyFrame: want *primitive.TimeoutError, got %T: %v",
			err,
			err,
		)
	}

	// Invariant: TimeoutError must carry the configured station handle.
	if timeoutErr.Station != handleExpect {
		t.Errorf(
			"TimeoutError.Station: want %q, got %q",
			handleExpect,
			timeoutErr.Station,
		)
	}

	// Invariant: TimeoutError must carry the configured timeout duration.
	if timeoutErr.Timeout != timeoutShort {
		t.Errorf(
			"TimeoutError.Timeout: want %v, got %v",
			timeoutShort,
			timeoutErr.Timeout,
		)
	}

	// Invariant: TimeoutError.Deadline must equal frozenNow + timeout.
	wantDeadline := frozenNow().Add(timeoutShort)

	if !timeoutErr.Deadline.Equal(wantDeadline) {
		t.Errorf(
			"TimeoutError.Deadline: want %v, got %v",
			wantDeadline,
			timeoutErr.Deadline,
		)
	}
}

// ── expectFrameOfType tests ──────────────────────────────────────────────────

// Test_primitive_expectFrameOfType_HappyPath verifies that when a queued frame
// has the correct message-type code at index 0 the keyword returns nil (AC3).
func Test_primitive_expectFrameOfType_HappyPath(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	state.SetNow(frozenNow())

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
// produces *TimeoutError (AC4).
func Test_primitive_expectFrameOfType_WrongTypeThenTimeout(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	state.SetNow(frozenNow())

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

	// Invariant: wrong-type frames are skipped; timeout returns *TimeoutError.
	if err == nil {
		t.Fatal(
			"expectFrameOfType (wrong type): want error, got nil",
		)
	}

	var timeoutErr *primitive.TimeoutError

	if !errors.As(err, &timeoutErr) {
		t.Fatalf(
			"expectFrameOfType: want *primitive.TimeoutError, got %T: %v",
			err,
			err,
		)
	}

	// Invariant: TimeoutError must identify the correct station handle.
	if timeoutErr.Station != handleExpect {
		t.Errorf(
			"TimeoutError.Station: want %q, got %q",
			handleExpect,
			timeoutErr.Station,
		)
	}
}
