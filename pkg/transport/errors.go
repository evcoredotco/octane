package transport

import "fmt"

// SubprotocolMismatchError is returned by Dial when the CSMS's WebSocket
// upgrade response selects a subprotocol not in the requested list, or
// omits the header entirely.
//
// The error carries both sides of the negotiation so that diagnostics can
// produce an actionable message without additional context.
type SubprotocolMismatchError struct {
	// Requested contains the subprotocols sent in the Sec-WebSocket-Protocol
	// request header, in preference order.
	Requested []string
	// Got is the subprotocol returned by the CSMS in its upgrade response.
	// An empty string indicates the header was absent in the response.
	Got string
}

// Error implements the error interface.
func (e *SubprotocolMismatchError) Error() string {
	if e.Got == "" {
		return fmt.Sprintf(
			"transport: subprotocol mismatch: requested %v, CSMS returned none",
			e.Requested,
		)
	}

	return fmt.Sprintf(
		"transport: subprotocol mismatch: requested %v, CSMS returned %q",
		e.Requested,
		e.Got,
	)
}

// TLSValidationError is returned by Dial when the TLS handshake fails due
// to certificate validation (expired cert, untrusted CA, hostname mismatch).
//
// The [Cause] field wraps the underlying x509 or tls error so that callers
// can use errors.As to inspect the certificate details.
type TLSValidationError struct {
	// URL is the endpoint that was dialled.
	URL string
	// Cause is the underlying x509 or tls error that triggered the failure.
	Cause error
}

// Error implements the error interface.
func (e *TLSValidationError) Error() string {
	return fmt.Sprintf(
		"transport: TLS validation failed for %q: %v",
		e.URL,
		e.Cause,
	)
}

// Unwrap returns the underlying cause, enabling errors.As and errors.Is
// inspection of the x509 certificate error.
func (e *TLSValidationError) Unwrap() error {
	return e.Cause
}

// FrameTooLargeError is returned by Expect when an inbound frame exceeds
// the configured [DialOptions.MaxFrameBytes] limit.
//
// The frame is discarded and the connection remains open; subsequent calls
// to Expect continue to deliver frames.
type FrameTooLargeError struct {
	// Limit is the configured maximum frame size in bytes.
	Limit int64
	// Actual is the actual frame size in bytes. -1 means the size was
	// unavailable (e.g., the WebSocket library signalled the limit was
	// exceeded without exposing the exact byte count).
	Actual int64
}

// Error implements the error interface.
func (e *FrameTooLargeError) Error() string {
	if e.Actual < 0 {
		return fmt.Sprintf(
			"transport: inbound frame exceeds limit of %d bytes"+
				" (actual size unknown)",
			e.Limit,
		)
	}

	return fmt.Sprintf(
		"transport: inbound frame too large: limit %d bytes, got %d bytes",
		e.Limit,
		e.Actual,
	)
}

// StationClosedError is returned by Send or Expect when the Station has
// already been closed.
//
// Use errors.As or errors.Is to detect this condition:
//
//	var closed *transport.StationClosedError
//	if errors.As(err, &closed) { ... }
//	// or
//	if errors.Is(err, &transport.StationClosedError{}) { ... }
type StationClosedError struct{}

// Error implements the error interface.
func (*StationClosedError) Error() string {
	return "transport: station is closed"
}

// Is reports whether target is an *StationClosedError, enabling
// errors.Is matching.
func (*StationClosedError) Is(target error) bool {
	_, ok := target.(*StationClosedError)

	return ok
}
