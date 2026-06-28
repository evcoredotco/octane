// Package lifecycle provides OCPP 1.6 domain-layer keywords that manage
// the WebSocket connection lifecycle between a simulated charging station
// and the CSMS under test.
//
// These keywords implement the step patterns used in helper stories that
// establish a known connection state before conformance test scenarios run:
//
//   - station {station:string} connects to the CSMS
//   - the OCPP-J handshake completes within {timeout:duration}
//   - station {station:string} is in the connected state
//
// The CSMS WebSocket URL is sourced from [api.State.CSMSBaseURL]; run
// with --csms-endpoint (or set csmsEndpoint in octane.yml) to configure it.
// Per-station URLs are constructed as baseURL + "/" + stationHandle.
//
// Register this package's keywords by calling [Register] before
// running any stories that use lifecycle steps.
package lifecycle
