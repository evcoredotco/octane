// Package primitive registers the transport-level primitive keywords for
// OCTANE's story DSL (spec 004).
//
// Primitive keywords are OCPP-version-agnostic. They speak only OCPP-J
// framing and know nothing about OCPP message semantics. They are the
// escape hatch for story authors whose CSMS uses extension messages
// without a matching domain keyword, and the building block that domain
// keyword authors delegate to.
//
// All keywords in this package are registered at [api.LayerPrimitive]
// with a zero [api.OCPPVersion] and self-register in init() via
// [registry.Register]. Importing this package is sufficient to
// activate the primitives; no further setup is required.
//
// # Connection primitives (spec 004 -10, items 1–3, 9–10)
//
//   - "open a WebSocket to {url:string} as station {station:string}"
//   - "open a WebSocket to {url:string} as station {station:string}
//     with subprotocol {subprotocol:string}"
//   - "close station {station:string}"
//   - "the connection on station {station:string} is open"
//   - "the connection on station {station:string} is closed"
//
// # Send primitives (spec 004 -10, items 4–5)
//
//   - "send raw frame {frame:any} on station {station:string}"
//   - "send raw bytes {bytes:string} on station {station:string}"
//
// The frame primitive accepts a []any value (the decoded Go form of an
// OCPP-J JSON array). The bytes primitive accepts a hex-encoded string
// and is intended for negative-path conformance testing of malformed or
// extension frames (spec 004 OQ1).
//
// # Expect primitives (spec 004 -10, items 6–7)
//
//   - "expect any frame on station {station:string} within {timeout:duration}"
//   - "expect a frame of type {messageType:int} on station
//     {station:string} within {timeout:duration}"
//
// Both expect keywords derive their deadline from [api.State.Now] so that
// deterministic-clock scenarios never advance real wall time. When the
// deadline elapses before a matching frame arrives, they return
// [*TimeoutError].
//
// # Wait primitive (spec 004 -10, item 8)
//
//   - "wait {duration:duration}"
//
// # Determinism
//
// All timing in this package is routed through [api.State.Now] and
// [api.State.Sleep]; neither [time.Now] nor [time.Sleep] is called
// directly. This satisfies constitution principle IV: deterministic clock
// injection produces byte-identical reports across runs.
package primitive
