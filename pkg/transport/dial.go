package transport

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/coder/websocket"
)

const (
	// defaultHandshakeTimeout is used when DialOptions.HandshakeTimeout is
	// zero.
	defaultHandshakeTimeout = 30 * time.Second

	// defaultMaxFrameBytes is used when DialOptions.MaxFrameBytes is zero
	// or negative. 1 MiB matches common OCPP-J server defaults.
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
		return nil, fmt.Errorf("transport: invalid URL: %w", err)
	}

	if _, ok := allowedSchemes[parsed.Scheme]; !ok {
		return nil, fmt.Errorf(
			"transport: unsupported scheme %q (want ws or wss)",
			parsed.Scheme,
		)
	}

	safeURL := sanitizeURL(parsed)

	tlsCfg := buildTLSConfig(safeURL, opts)
	maxBytes := opts.MaxFrameBytes

	if maxBytes <= 0 {
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
		return nil, wrapDialError(safeURL, err)
	}

	conn.SetReadLimit(maxBytes)

	err = validateSubprotocol(conn, opts.Subprotocols)
	if err != nil {
		_ = conn.Close(websocket.StatusNormalClosure, "subprotocol mismatch")

		return nil, err
	}

	return newStationHandle(conn, maxBytes), nil
}

// sanitizeURL strips userinfo (credentials) from the parsed URL so that the
// result is safe to embed in log messages and error strings. This prevents
// credential leakage (CWE-532 / constitution principle X).
func sanitizeURL(parsed *url.URL) string {
	safe := *parsed
	safe.User = nil

	return safe.String()
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
// A TLS 1.2 minimum version floor is always enforced — if the caller supplies
// a config with a lower MinVersion, it is raised to tls.VersionTLS12 per NIST
// SP 800-52 Rev. 2 and the OCA Security Profile for OCPP 1.6.
//
// If opts.InsecureSkipVerify is true a warning is logged and the config's
// InsecureSkipVerify field is set. The original config is cloned to avoid
// mutating a value shared by the caller.
func buildTLSConfig(safeURL string, opts DialOptions) *tls.Config {
	if !opts.InsecureSkipVerify && opts.TLSConfig == nil {
		return nil // nil → system defaults, always TLS 1.2+ in modern Go
	}

	base := opts.TLSConfig
	if base == nil {
		base = &tls.Config{}
	}

	cfg := base.Clone()

	// Enforce TLS 1.2 floor regardless of caller-supplied value.
	if cfg.MinVersion < tls.VersionTLS12 {
		cfg.MinVersion = tls.VersionTLS12
	}

	if opts.InsecureSkipVerify {
		slog.Warn(
			"TLS certificate verification is disabled",
			"url", safeURL,
			"note", "insecure-skip-verify is set; every report will carry a "+
				"banner-level finding",
		)

		cfg.InsecureSkipVerify = true
	}

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
// the cause is recognisable as a TLS failure. safeURL has userinfo stripped.
func wrapDialError(safeURL string, cause error) error {
	if isTLSError(cause) {
		return &ErrTLSValidation{
			URL:   safeURL,
			Cause: cause,
		}
	}

	return fmt.Errorf("transport: dial %q: %w", safeURL, cause)
}

// isTLSError reports whether err originates from a TLS or x509 failure.
// Prefixes "tls: " and "x509: " are checked first; "handshake failure" is
// kept as a fallback for older library versions. The broad "certificate"
// marker was removed to avoid false positives from CSMS error bodies.
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

// tlsErrorMarkers identifies TLS/x509 errors by error-string prefix.
// Deliberately narrow: "certificate" was removed (too broad — a CSMS could
// return "certificate" in an HTTP 400 body, causing a false ErrTLSValidation).
var tlsErrorMarkers = []string{
	"tls: ",
	"x509: ",
	"handshake failure",
}

// contains reports whether needle is an element of haystack.
func contains(haystack []string, needle string) bool {
	return slices.Contains(haystack, needle)
}
