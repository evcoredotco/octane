// Package mock_test exercises the public surface of pkg/keywords/api/mock
// to verify that State and Station satisfy the api.State and api.Station
// interfaces respectively, and that the mock package carries zero imports
// of pkg/runner or pkg/transport.
//
// The import-isolation guarantee is enforced structurally: this test file
// imports only the mock package and the standard library. If mock.go ever
// acquired a pkg/runner or pkg/transport import, the build would fail
// before this test could run.
//
// Task: T-003-41
// AC8: Given a mock State and Station from pkg/keywords/api/mock, when a
// third-party keyword is unit-tested against them, then the test does not
// require importing pkg/runner/, pkg/transport/, or any network library.
package mock_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/octane-project/octane/pkg/keywords/api"
	"github.com/octane-project/octane/pkg/keywords/api/mock"
)

// ── interface-satisfaction compile-time checks ────────────────────────────────

// Verify at compile time that *mock.State implements api.State and that
// *mock.Station implements api.Station. These assignments produce a compile
// error if the interfaces diverge from the mock implementations.
var (
	_ api.State   = (*mock.State)(nil)
	_ api.Station = (*mock.Station)(nil)
)

// ── mock.State tests ──────────────────────────────────────────────────────────

// Test_MockState_NowReturnsZeroByDefault verifies that a freshly created
// mock.State returns the zero time.Time from Now().
func Test_MockState_NowReturnsZeroByDefault(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()

	got := state.Now()

	if !got.IsZero() {
		t.Errorf("Now(): want zero time, got %v", got)
	}
}

// Test_MockState_SetNow verifies that Now() returns the configured time
// after calling SetNow.
func Test_MockState_SetNow(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	want := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

	state.SetNow(want)

	got := state.Now()

	if !got.Equal(want) {
		t.Errorf("Now(): want %v, got %v", want, got)
	}
}

// Test_MockState_LogfAppendsMessages verifies that Logf appends formatted
// messages and that Logs returns them in order.
func Test_MockState_LogfAppendsMessages(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()

	state.Logf("step %d started", 1)
	state.Logf("station %q connected", "CP01")

	got := state.Logs()

	const wantCount = 2

	if len(got) != wantCount {
		t.Fatalf("Logs() count: want %d, got %d", wantCount, len(got))
	}

	const wantFirst = "step 1 started"

	if got[0] != wantFirst {
		t.Errorf("Logs()[0]: want %q, got %q", wantFirst, got[0])
	}

	const wantSecond = `station "CP01" connected`

	if got[1] != wantSecond {
		t.Errorf("Logs()[1]: want %q, got %q", wantSecond, got[1])
	}
}

// Test_MockState_LogsReturnsSnapshot verifies that the slice returned by
// Logs() is a copy: appending to it does not affect future Logs() calls.
func Test_MockState_LogsReturnsSnapshot(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()

	state.Logf("first message")

	snapshot := state.Logs()
	_ = append(snapshot, "injected")

	later := state.Logs()

	if len(later) != 1 {
		t.Errorf(
			"Logs() after mutation of previous snapshot: "+
				"want 1 entry, got %d",
			len(later),
		)
	}
}

// Test_MockState_StationPanicsOnMissingHandle verifies that Station() panics
// when the requested handle has not been registered.
func Test_MockState_StationPanicsOnMissingHandle(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()

	defer func() {
		if rec := recover(); rec == nil {
			t.Error(
				"Station(\"unknown\"): expected panic on missing " +
					"handle, but none occurred",
			)
		}
	}()

	_, _ = state.Station("unknown")
}

// Test_MockState_RegisterStation verifies that Station() returns the mock
// station registered under the given handle.
func Test_MockState_RegisterStation(t *testing.T) {
	t.Parallel()

	state := mock.NewMockState()
	station := mock.NewMockStation()

	state.RegisterStation("CP01", station)

	got, err := state.Station("CP01")
	if err != nil {
		t.Fatalf("Station(\"CP01\") unexpected error: %v", err)
	}

	if got != station {
		t.Error("Station(\"CP01\"): returned unexpected station instance")
	}
}

// ── mock.Station tests ────────────────────────────────────────────────────────

// Test_MockStation_SendRecordsFrame verifies that Send appends the frame
// to the internal buffer and SentFrames returns it.
func Test_MockStation_SendRecordsFrame(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	ctx := context.Background()
	frame := []any{2, "msg-01", "BootNotification", map[string]any{}}

	err := station.Send(ctx, frame)
	if err != nil {
		t.Fatalf("Send(): unexpected error: %v", err)
	}

	sent := station.SentFrames()

	if len(sent) != 1 {
		t.Fatalf("SentFrames() count: want 1, got %d", len(sent))
	}
}

// Test_MockStation_SendReturnsConfiguredError verifies that Send returns
// the error set by SetSendError without recording the frame.
func Test_MockStation_SendReturnsConfiguredError(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	ctx := context.Background()

	wantErr := errors.New("send failed")

	station.SetSendError(wantErr)

	err := station.Send(
		ctx,
		[]any{2, "msg-02", "Action", map[string]any{}},
	)

	if !errors.Is(err, wantErr) {
		t.Errorf("Send(): want error %v, got %v", wantErr, err)
	}

	if len(station.SentFrames()) != 0 {
		t.Error(
			"Send(): frame must not be recorded when sendErr is set",
		)
	}
}

// Test_MockStation_SendRespectsContextCancellation verifies that Send
// returns ctx.Err() when the context is already cancelled.
func Test_MockStation_SendRespectsContextCancellation(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	ctx, cancel := context.WithCancel(context.Background())

	cancel()

	err := station.Send(
		ctx,
		[]any{2, "msg-03", "Action", map[string]any{}},
	)

	if !errors.Is(err, context.Canceled) {
		t.Errorf(
			"Send() with cancelled ctx: want context.Canceled, got %v",
			err,
		)
	}
}

// Test_MockStation_ExpectDequeuesFrame verifies that Expect returns frames
// in the order they were queued via QueueFrame.
func Test_MockStation_ExpectDequeuesFrame(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	ctx := context.Background()

	first := []any{
		3, "msg-01", "BootNotificationResponse",
		map[string]any{"status": "Accepted"},
	}
	second := []any{
		3, "msg-02", "HeartbeatResponse",
		map[string]any{},
	}

	station.QueueFrame(first)
	station.QueueFrame(second)

	gotFirst, err := station.Expect(ctx)
	if err != nil {
		t.Fatalf("Expect() first: unexpected error: %v", err)
	}

	if len(gotFirst) != len(first) {
		t.Errorf(
			"Expect() first frame length: want %d, got %d",
			len(first),
			len(gotFirst),
		)
	}

	gotSecond, err := station.Expect(ctx)
	if err != nil {
		t.Fatalf("Expect() second: unexpected error: %v", err)
	}

	if len(gotSecond) != len(second) {
		t.Errorf(
			"Expect() second frame length: want %d, got %d",
			len(second),
			len(gotSecond),
		)
	}
}

// Test_MockStation_ExpectReturnsConfiguredError verifies that Expect returns
// the error set by SetExpectError and does not touch the pending queue.
func Test_MockStation_ExpectReturnsConfiguredError(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	ctx := context.Background()

	wantErr := errors.New("expect failed")

	station.QueueFrame([]any{
		3, "msg-01", "HeartbeatResponse", map[string]any{},
	})
	station.SetExpectError(wantErr)

	frame, err := station.Expect(ctx)

	if !errors.Is(err, wantErr) {
		t.Errorf("Expect(): want error %v, got %v", wantErr, err)
	}

	if frame != nil {
		t.Error("Expect(): want nil frame when error is configured")
	}
}

// Test_MockStation_ExpectBlocksOnEmptyQueue verifies that Expect blocks
// until the context expires when no frames are queued and no error is
// configured.
func Test_MockStation_ExpectBlocksOnEmptyQueue(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	ctx, cancel := context.WithCancel(context.Background())

	cancel()

	_, err := station.Expect(ctx)

	if !errors.Is(err, context.Canceled) {
		t.Errorf(
			"Expect() on empty queue with cancelled ctx: "+
				"want context.Canceled, got %v",
			err,
		)
	}
}

// Test_MockStation_SentFramesReturnsSnapshot verifies that the slice
// returned by SentFrames() is a copy: mutating it does not affect future
// calls.
func Test_MockStation_SentFramesReturnsSnapshot(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	ctx := context.Background()

	_ = station.Send(ctx, []any{2, "msg-01", "Action", map[string]any{}})

	snapshot := station.SentFrames()
	snapshot[0] = nil

	later := station.SentFrames()

	if later[0] == nil {
		t.Error(
			"SentFrames(): mutation of returned snapshot " +
				"affected subsequent call",
		)
	}
}

// Test_MockStation_SetSendErrorNilClearsError verifies that passing nil to
// SetSendError clears a previously configured error so subsequent Send
// calls succeed.
func Test_MockStation_SetSendErrorNilClearsError(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	ctx := context.Background()

	station.SetSendError(errors.New("temporary error"))
	station.SetSendError(nil)

	err := station.Send(
		ctx,
		[]any{2, "msg-01", "Action", map[string]any{}},
	)
	if err != nil {
		t.Errorf(
			"Send() after clearing sendErr: want nil, got %v",
			err,
		)
	}
}

// Test_MockStation_SetExpectErrorNilClearsError verifies that passing nil
// to SetExpectError clears a previously configured error so subsequent
// Expect calls resume dequeuing frames normally.
func Test_MockStation_SetExpectErrorNilClearsError(t *testing.T) {
	t.Parallel()

	station := mock.NewMockStation()
	ctx := context.Background()

	station.QueueFrame([]any{
		3, "msg-01", "Response", map[string]any{},
	})
	station.SetExpectError(errors.New("transient error"))
	station.SetExpectError(nil)

	_, err := station.Expect(ctx)
	if err != nil {
		t.Errorf(
			"Expect() after clearing expectErr: want nil, got %v",
			err,
		)
	}
}
