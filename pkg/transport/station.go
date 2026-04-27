package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/coder/websocket"
)

// stationHandle is the concrete Station returned by Dial. It wraps a single
// *websocket.Conn and provides Send, Expect, and Close.
//
// The reader goroutine is started by newStationHandle and runs until the
// connection is closed or an unrecoverable read error occurs. All decoded
// frames are queued on inbound for consumption by Expect.
type stationHandle struct {
	conn     *websocket.Conn
	inbound  chan []any
	closed   chan struct{}
	once     sync.Once
	maxBytes int64
}

// newStationHandle allocates a stationHandle, starts the reader goroutine,
// and returns the value as a Station interface.
//
// The reader goroutine's lifetime is controlled by readerCtx: closing the
// connection (via Close) causes the next conn.Read to return an error, which
// terminates the goroutine.
func newStationHandle(
	conn *websocket.Conn,
	maxBytes int64,
) Station {
	handle := &stationHandle{
		conn:     conn,
		inbound:  make(chan []any, inboundBufSize),
		closed:   make(chan struct{}),
		once:     sync.Once{},
		maxBytes: maxBytes,
	}

	go handle.readLoop()

	return handle
}

// Send encodes frame as a JSON array and writes it to the WebSocket.
//
// It returns [*ErrStationClosed] immediately if the station has already been
// closed. On a write failure the underlying websocket error is wrapped and
// returned. Send is safe for concurrent use.
func (sta *stationHandle) Send(
	ctx context.Context,
	frame []any,
) error {
	select {
	case <-sta.closed:
		return &ErrStationClosed{}
	default:
	}

	data, err := json.Marshal(frame)
	if err != nil {
		return fmt.Errorf("transport: marshal frame: %w", err)
	}

	err = sta.conn.Write(ctx, websocket.MessageText, data)
	if err != nil {
		select {
		case <-sta.closed:
			return &ErrStationClosed{}
		default:
			return fmt.Errorf("transport: write frame: %w", err)
		}
	}

	return nil
}

// Expect blocks until an inbound OCPP-J frame is available, the context is
// cancelled, or the station is closed.
//
// Frames are delivered in FIFO order matching the order in which the reader
// goroutine queued them. Expect returns [*ErrStationClosed] when the station
// has been closed and the inbound queue is drained.
func (sta *stationHandle) Expect(ctx context.Context) ([]any, error) {
	select {
	case <-sta.closed:
		return nil, &ErrStationClosed{}
	default:
	}

	select {
	case frame, ok := <-sta.inbound:
		if !ok {
			return nil, &ErrStationClosed{}
		}

		return frame, nil

	case <-ctx.Done():
		return nil, ctx.Err()

	case <-sta.closed:
		return nil, &ErrStationClosed{}
	}
}

// Close gracefully closes the WebSocket connection with status 1000 (normal
// closure). Subsequent calls to Close are no-ops and return nil. Close is
// safe for concurrent use.
func (sta *stationHandle) Close() error {
	sta.once.Do(func() {
		close(sta.closed)
		_ = sta.conn.Close(websocket.StatusNormalClosure, "")
	})

	return nil
}

// readLoop is the reader goroutine started by newStationHandle. It reads
// frames from the WebSocket, decodes each as a JSON array, and sends the
// result to the inbound channel.
//
// When conn.Read returns an error (including the error produced when
// MaxFrameBytes is exceeded, which closes the connection at the library
// level), the loop closes the inbound channel and returns, signalling Expect
// callers that no further frames will arrive.
//
// readLoop uses a background context derived from the connection lifetime,
// not from any caller context, so that frames already in flight are drained
// regardless of the Dial context state.
func (sta *stationHandle) readLoop() {
	defer close(sta.inbound)

	ctx := context.Background()

	for {
		msgType, data, err := sta.conn.Read(ctx)
		if err != nil {
			return
		}

		if msgType != websocket.MessageText {
			continue
		}

		var frame []any

		if err = json.Unmarshal(data, &frame); err != nil {
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
