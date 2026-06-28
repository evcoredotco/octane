// Package ocpp16 provides OCPP 1.6J domain-layer keywords for OCTANE
// conformance stories. It covers the message exchanges defined in
// OCPP-J 1.6: BootNotification, StatusNotification, Heartbeat,
// Authorize, StartTransaction, and ReserveNow.
//
// Keywords are registered via [Register], which must be called exactly
// once at program startup — typically from cmd/octane Execute() alongside
// primitive.Register() and lifecycle.Register().
//
// All frame serialization uses pkg/wire for OCPP-J envelope construction.
// Payload fields use map[string]any so that no local OCPP 1.6 type
// declarations are needed (ADR 0020).
//
// # Stash contract
//
// "Send" keywords stash [pendingInfo] under [pendingKey] so that the
// following "response" keyword can correlate the CALLRESULT without
// repeating the station name or message ID in the step text.
//
// "Response" keywords pop [pendingKey], receive the frame, unmarshal the
// payload, and stash it under [lastPayloadKey] for subsequent assertion
// keywords that inspect specific fields.
//
// See stash.go for the full set of stash keys and their value types.
package ocpp16
