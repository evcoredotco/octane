package transport

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coder/websocket"
)

const (
	// defaultHandshakeTimeout is used when DialOptions.HandshakeTimeout is
	// zero.
	defaultHandshakeTimeout = 30 * time.Second

	// defaultMaxFrameBytes is used when DialOptions.MaxFrameBytes is zero.
	// 1 MiB matches common OCPP-J server defaults.
	defaultMaxFrameBytes = int64(1 << 20)

	// inboundBufSize is the capacity of the inbound frame channel.
	// A slow keyword consumer is a programming error; this buffer prevents
	// the reader goroutine from blocking on a single slow Expect call while
	// still giving callers a window to drain.
	inboundBufSize = 64
)

// allowedSchemes lists the URL schemes accepted by Dial.
var allowedSchemes = map[string]struct{}{
	"ws":  {},
	"wss": {},
}

// Dial opens a WebSocket connection to rawURL and returns a [Station] handle.
//
// rawURL must use the "ws" or "wss" scheme. Dial performs the WebSocket
// handshake with the subprotocol list in opts.Subprotocols and validates the
// server's selection. If the server selects a subprotocol not in the list (or
// omits the header entirely), Dial closes the connection and returns
// [*ErrSubprotocolMismatch].
//
// TLS verification is on by default. Setting [DialOptions.InsecureSkipVerify]
// to true disables certificate validation and emits a WARNING via [log/slog].
//
// The context passed to Dial governs the WebSocket upgrade only. The returned
// [Station] uses its own internal context for the reader goroutine; cancelling
// the Dial context after the handshake has no effect on the station.
func Dial(
	ctx context.Context,
	rawURL string,
	opts DialOptions,
) (Station, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("transport: invalid URL %q: %w", rawURL, err)
	}

	if _, ok := allowedSchemes[parsed.Scheme]; !ok {
		return nil, fmt.Errorf(
			"transport: unsupported scheme %q in %q (want ws or wss)",
			parsed.Scheme,
			rawURL,
		)
	}

	tlsCfg := buildTLSConfig(rawURL, opts)
	maxBytes := opts.MaxFrameBytes

	if maxBytes == 0 {
		maxBytes = defaultMaxFrameBytes
	}

	timeout := opts.HandshakeTimeout
	if timeout == 0 {
		timeout = defaultHandshakeTimeout
	}

	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	wsOpts := &websocket.DialOptions{
		Subprotocols:    opts.Subprotocols,
		CompressionMode: websocket.CompressionDisabled,
		HTTPClient:      buildHTTPClient(tlsCfg),
	}

	conn, _, err := websocket.Dial(dialCtx, rawURL, wsOpts)
	if err != nil {
		return nil, wrapDialError(rawURL, err)
	}

	conn.SetReadLimit(maxBytes)

	if err = validateSubprotocol(conn, opts.Subprotocols); err != nil {
		_ = conn.Close(websocket.StatusNormalClosure, "subprotocol mismatch")

		return nil, err
	}

	return newStationHandle(conn), nil
}

// validateSubprotocol checks that the server chose a subprotocol from the
// requested list. It returns nil when the list is empty (caller has no
// preference). On mismatch it returns *ErrSubprotocolMismatch.
func validateSubprotocol(
	conn *websocket.Conn,
	subprotocols []string,
) error {
	if len(subprotocols) == 0 {
		return nil
	}

	negotiated := conn.Subprotocol()

	if contains(subprotocols, negotiated) {
		return nil
	}

	return &ErrSubprotocolMismatch{
		Requested: subprotocols,
		Got:       negotiated,
	}
}

// buildTLSConfig derives the *tls.Config to use for the dial.
//
// If opts.InsecureSkipVerify is true a warning is logged and the config's
// InsecureSkipVerify field is set. The original config is cloned to avoid
// mutating a value shared by the caller.
func buildTLSConfig(rawURL string, opts DialOptions) *tls.Config {
	if !opts.InsecureSkipVerify {
		return opts.TLSConfig
	}

	slog.Warn(
		"TLS certificate verification is disabled",
		"url", rawURL,
		"note", "insecure-skip-verify is set; every report will carry a "+
			"banner-level finding",
	)

	base := opts.TLSConfig
	if base == nil {
		base = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	cfg := base.Clone()
	cfg.InsecureSkipVerify = true //nolint:gosec // G402: intentional operator opt-in

	return cfg
}

// buildHTTPClient returns an *http.Client whose transport uses tlsCfg.
// If tlsCfg is nil the client's transport uses system defaults.
func buildHTTPClient(tlsCfg *tls.Config) *http.Client {
	if tlsCfg == nil {
		return nil // nil instructs the websocket library to use the default client
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
	}
}

// wrapDialError converts a raw dial error into a typed transport error where
// the cause is recognisable as a TLS failure.
func wrapDialError(rawURL string, cause error) error {
	if isTLSError(cause) {
		return &ErrTLSValidation{
			URL:   rawURL,
			Cause: cause,
		}
	}

	return fmt.Errorf("transport: dial %q: %w", rawURL, cause)
}

// isTLSError reports whether err originates from a TLS or x509 failure.
// It performs a substring heuristic on the error text because the
// crypto/tls package does not expose a stable exported error type for every
// handshake failure path.
func isTLSError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()

	for _, marker := range tlsErrorMarkers {
		if strings.Contains(msg, marker) {
			return true
		}
	}

	return false
}

// tlsErrorMarkers is the set of substrings that identify a TLS or x509
// failure in an error message returned by crypto/tls or net/http.
var tlsErrorMarkers = []string{
	"tls:",
	"x509:",
	"certificate",
	"handshake failure",
}

// contains reports whether needle is an element of haystack.
func contains(haystack []string, needle string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}

	return false
}
