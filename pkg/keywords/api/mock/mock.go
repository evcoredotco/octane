// Package mock provides in-memory test doubles for the [api.State] and
// [api.Station] interfaces defined in pkg/keywords/api.
//
// Keyword authors use this package to unit-test their keyword functions
// without importing pkg/runner/, pkg/transport/, or any network library.
// The package depends solely on the Go standard library and on the api
// package it is doubling.
//
// Typical usage:
//
//	state := mock.NewMockState()
//	station := mock.NewMockStation()
//	state.RegisterStation("CP01", station)
//	station.QueueFrame([]any{3, "msg-01", "BootNotificationResponse",
//	    map[string]any{"currentTime": "2024-01-01T00:00:00Z",
//	        "interval": 300, "status": "Accepted"}})
//
//	err := myKeyword(context.Background(), state, args)
//	// assert err, station.SentFrames(), state.Logs(), …
//
// Task: T-003-40
// AC8: Given a mock State and Station from pkg/keywords/api/mock, when a
// third-party keyword is unit-tested against them, then the test does not
// require importing pkg/runner/, pkg/transport/, or any network library.
package mock

import (
	"context"
	"fmt"
	"time"

	"github.com/octane-project/octane/pkg/keywords/api"
)

// State is an in-memory implementation of [api.State] for use in keyword
// unit tests. It holds a map of named mock stations, a configurable
// frozen clock, and an append-only log buffer.
//
// The zero value is not usable; create instances via [NewMockState].
type State struct {
	// stations holds the registered mock stations keyed by handle name.
	stations map[string]api.Station

	// frozenTime is the fixed time returned by [State.Now].
	frozenTime time.Time

	// logs accumulates every Logf call for later assertion.
	logs []string
}

// NewMockState returns a ready-to-use *[State] with an empty station map,
// a zero-value frozen time, and an empty log buffer.
//
// The frozen time defaults to the zero value of [time.Time]. Call
// [State.SetNow] to configure a specific instant before the test runs
// the keyword under test.
func NewMockState() *State {
	return &State{
		stations:   make(map[string]api.Station),
		frozenTime: time.Time{},
		logs:       nil,
	}
}

// Station returns the [api.Station] registered under the given handle.
// It panics if no station has been registered for that name, because
// an unresolved station handle is a test-setup bug, not a runtime error.
//
// Call [State.RegisterStation] before running the keyword under test.
func (s *State) Station(handle string) (api.Station, error) {
	station, found := s.stations[handle]
	if !found {
		panic(fmt.Sprintf(
			"mock.State: station %q not registered; "+
				"call RegisterStation before running the keyword",
			handle,
		))
	}

	return station, nil
}

// Now returns the frozen time configured via [State.SetNow]. The default
// value is the zero value of [time.Time]. Keywords must call
// [api.State.Now] instead of [time.Now] to remain deterministic
// (constitution principle IV).
func (s *State) Now() time.Time {
	return s.frozenTime
}

// Logf appends a formatted message to the internal log buffer. The format
// string and arguments follow [fmt.Sprintf] conventions. Retrieve the
// accumulated messages with [State.Logs].
func (s *State) Logf(format string, args ...any) {
	s.logs = append(s.logs, fmt.Sprintf(format, args...))
}

// SetNow sets the frozen time returned by [State.Now]. Call this before
// running the keyword under test to control timestamp behaviour.
func (s *State) SetNow(frozenTime time.Time) {
	s.frozenTime = frozenTime
}

// RegisterStation adds station under handle so that [State.Station] can
// return it. Registering the same handle twice replaces the previous entry
// without panicking, which simplifies test-table setups that reuse the same
// handle name across rows.
func (s *State) RegisterStation(handle string, station api.Station) {
	s.stations[handle] = station
}

// Logs returns a copy of all messages passed to [State.Logf] in the order
// they were emitted. The returned slice is a snapshot; subsequent Logf
// calls do not affect it.
func (s *State) Logs() []string {
	result := make([]string, len(s.logs))

	copy(result, s.logs)

	return result
}

// Station is an in-memory implementation of [api.Station] for use in
// keyword unit tests. It records every frame passed to [Station.Send] and
// serves pre-queued frames from [Station.Expect].
//
// Note: this type is named Station and lives in the mock package; it
// implements the [api.Station] interface but is distinct from it. When
// both appear in the same file, qualify the interface as [api.Station]
// and this concrete type as [mock.Station] to avoid confusion.
//
// The zero value is not usable; create instances via [NewMockStation].
type Station struct {
	// sentFrames accumulates every frame passed to Send.
	sentFrames [][]any

	// pendingFrames is the FIFO queue consumed by Expect.
	pendingFrames [][]any

	// sendErr is returned by Send when non-nil.
	sendErr error

	// expectErr is returned by Expect when non-nil (takes priority over
	// pendingFrames).
	expectErr error

	// open tracks the connection state. True means the station is open.
	open bool
}

// NewMockStation returns a ready-to-use *[Station] with empty frame
// buffers, nil errors, and the connection open (IsOpen returns true
// until [Station.Close] is called).
func NewMockStation() *Station {
	return &Station{
		sentFrames:    nil,
		pendingFrames: nil,
		sendErr:       nil,
		expectErr:     nil,
		open:          true,
	}
}

// Send records frame in the internal sent-frames buffer and returns the
// configured send error (nil by default). The context is checked for
// cancellation before appending; if ctx.Err() is non-nil, Send returns
// that error immediately without recording the frame.
//
// Use [Station.SentFrames] after the keyword returns to assert the
// sequence and content of outbound frames.
func (st *Station) Send(ctx context.Context, frame []any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if st.sendErr != nil {
		return st.sendErr
	}

	clone := make([]any, len(frame))

	copy(clone, frame)

	st.sentFrames = append(st.sentFrames, clone)

	return nil
}

// Expect dequeues and returns the next frame from the pre-queued buffer.
// If a non-nil expect error has been configured via [Station.SetExpectError],
// it is returned immediately without consuming the queue. If the queue is
// empty and no error is configured, Expect blocks until the context expires
// and returns ctx.Err().
func (st *Station) Expect(ctx context.Context) ([]any, error) {
	if st.expectErr != nil {
		return nil, st.expectErr
	}

	if len(st.pendingFrames) > 0 {
		frame := st.pendingFrames[0]
		st.pendingFrames = st.pendingFrames[1:]

		return frame, nil
	}

	<-ctx.Done()

	return nil, ctx.Err()
}

// SentFrames returns a copy of every frame recorded by [Station.Send] in
// the order they were sent. The returned slice is a snapshot; subsequent
// Send calls do not affect it.
func (st *Station) SentFrames() [][]any {
	result := make([][]any, len(st.sentFrames))

	for idx, frame := range st.sentFrames {
		clone := make([]any, len(frame))

		copy(clone, frame)

		result[idx] = clone
	}

	return result
}

// QueueFrame appends frame to the back of the pending-frames FIFO queue.
// Each [Station.Expect] call dequeues from the front, so frames are
// delivered in the order they were queued.
func (st *Station) QueueFrame(frame []any) {
	clone := make([]any, len(frame))

	copy(clone, frame)

	st.pendingFrames = append(st.pendingFrames, clone)
}

// SetSendError configures the error that [Station.Send] returns on every
// subsequent call. Pass nil to clear a previously set error.
func (st *Station) SetSendError(err error) {
	st.sendErr = err
}

// SetExpectError configures the error that [Station.Expect] returns on
// every subsequent call, regardless of the pending-frames queue.
// Pass nil to clear a previously set error.
func (st *Station) SetExpectError(err error) {
	st.expectErr = err
}

// Close marks the station as closed. Subsequent calls to [Station.IsOpen]
// return false. Close is idempotent and always returns nil.
func (st *Station) Close() error {
	st.open = false

	return nil
}

// IsOpen reports whether the station has not yet been closed.
// It returns true from construction until [Station.Close] is called.
func (st *Station) IsOpen() bool {
	return st.open
}
