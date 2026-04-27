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
// # Connection primitives (spec 004 §10, items 1–3, 9–10)
//
//   - "open a WebSocket to {url:string} as station {station:string}"
//   - "open a WebSocket to {url:string} as station {station:string} with subprotocol {subprotocol:string}"
//   - "close station {station:string}"
//   - "the connection on station {station:string} is open"
//   - "the connection on station {station:string} is closed"
//
// # Determinism
//
// Connection primitives do not consume the deterministic clock or PRNG;
// they have no timing behaviour of their own. The wait primitive
// (spec 004 §10, item 8) consumes [api.State.Now] and is implemented in
// the sibling file wait.go (T-004-20).
package primitive
