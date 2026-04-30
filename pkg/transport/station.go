package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/coder/websocket"
)

// stationHandle is the concrete Station returned by Dial. It wraps a single
// *websocket.Conn and provides Send, Expect, and Close.
//
// The reader goroutine is started by newStationHandle and runs until the
// connection is closed or an unrecoverable read error occurs. All decoded
// frames are queued on inbound for consumption by Expect.
//
// readCancel cancels the reader goroutine's context so that Close() gives
// it a fast-teardown path under test (where the peer may stop sending
// without closing the TCP connection).
type stationHandle struct {
	conn       *websocket.Conn
	inbound    chan []any
	closed     chan struct{}
	once       sync.Once
	readCancel context.CancelFunc
	readErr    atomic.Pointer[error]
	maxBytes   int64
}

// newStationHandle allocates a stationHandle, starts the reader goroutine,
// and returns the value as a Station interface.
//
// The caller is responsible for calling conn.SetReadLimit before invoking
// newStationHandle. Closing the connection (via Close) causes the next
// conn.Read to return an error, which terminates the goroutine.
func newStationHandle(
	ctx context.Context,
	conn *websocket.Conn,
	maxBytes int64,
) Station {
	readerCtx, cancel := context.WithCancel(ctx)

	handle := &stationHandle{
		conn:       conn,
		inbound:    make(chan []any, inboundBufSize),
		closed:     make(chan struct{}),
		once:       sync.Once{},
		readCancel: cancel,
		readErr:    atomic.Pointer[error]{},
		maxBytes:   maxBytes,
	}

	go handle.readLoop(readerCtx)

	return handle
}

// Send encodes frame as a JSON array and writes it to the WebSocket.
//
// On a write failure after Close, it returns [*StationClosedError].
// Send is safe for concurrent use.
func (sta *stationHandle) Send(
	ctx context.Context,
	frame []any,
) error {
	data, err := json.Marshal(frame)
	if err != nil {
		return fmt.Errorf("transport: marshal frame: %w", err)
	}

	err = sta.conn.Write(ctx, websocket.MessageText, data)
	if err != nil {
		select {
		case <-sta.closed:
			return &StationClosedError{}
		default:
			return fmt.Errorf("transport: write frame: %w", err)
		}
	}

	return nil
}

// Expect blocks until an inbound OCPP-J frame is available, the context is
// cancelled, or the station is closed.
//
// Frames already buffered in the inbound queue are delivered even after
// Close is called — the channel is drained before StationClosedError is
// returned. If the connection was closed due to an oversized frame,
// [*FrameTooLargeError] is returned after the buffer is exhausted.
func (sta *stationHandle) Expect(ctx context.Context) ([]any, error) {
	select {
	case frame, ok := <-sta.inbound:
		if !ok {
			return nil, sta.closeError()
		}

		return frame, nil

	case <-ctx.Done():
		return nil, fmt.Errorf("transport: expect context: %w", ctx.Err())

	case <-sta.closed:
		// Drain one buffered frame before signalling closure so that
		// keyword authors can collect the final server response.
		select {
		case frame, ok := <-sta.inbound:
			if ok {
				return frame, nil
			}

			return nil, sta.closeError()
		default:
		}

		return nil, &StationClosedError{}
	}
}

// Close gracefully closes the WebSocket connection with status 1000 (normal
// closure). Subsequent calls to Close are no-ops and return nil. Close is
// safe for concurrent use.
func (sta *stationHandle) Close() error {
	sta.once.Do(func() {
		close(sta.closed)
		sta.readCancel()
		_ = sta.conn.Close(websocket.StatusNormalClosure, "")
	})

	return nil
}

// IsOpen reports whether the connection is currently open. It returns
// false once [Close] has been called. IsOpen is safe for concurrent use.
func (sta *stationHandle) IsOpen() bool {
	select {
	case <-sta.closed:
		return false
	default:
		return true
	}
}

// closeError returns the appropriate error for when the inbound channel closes.
// If the reader goroutine recorded a frame-size error it is returned as
// *FrameTooLargeError; otherwise *StationClosedError is returned.
func (sta *stationHandle) closeError() error {
	if errPtr := sta.readErr.Load(); errPtr != nil {
		return *errPtr
	}

	return &StationClosedError{}
}

// readLoop is the reader goroutine started by newStationHandle. It reads
// frames from the WebSocket, decodes each as a JSON array, and sends the
// result to the inbound channel.
//
// When conn.Read returns StatusMessageTooBig (1009), an *FrameTooLargeError is
// stored in readErr so that Expect can surface it to callers after the buffer
// is drained. All other read errors terminate the loop silently (the
// connection is already closed at that point).
func (sta *stationHandle) readLoop(ctx context.Context) {
	defer close(sta.inbound)

	for {
		msgType, data, err := sta.conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusMessageTooBig {
				frameErr := error(&FrameTooLargeError{
					Limit:  sta.maxBytes,
					Actual: -1,
				})
				sta.readErr.Store(&frameErr)
			}

			return
		}

		if msgType != websocket.MessageText {
			continue
		}

		var frame []any

		err = json.Unmarshal(data, &frame)
		if err != nil {
			// Non-JSON frames are silently dropped; they are not valid
			// OCPP-J and will surface as a missing Expect delivery to
			// the test scenario.
			continue
		}

		select {
		case sta.inbound <- frame:
		case <-sta.closed:
			return
		}
	}
}
