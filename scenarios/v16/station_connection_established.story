# Helper story — station connection established.
#
# Establishes a WebSocket connection to the CSMS using the OCPP-J
# subprotocol declared by the connection profile. This helper is the
# foundation of the lifecycle prerequisite chain used by most
# conformance stories.
#
# Helper stories carry no Spec-Ref because they do not assert
# conformance to a specification section in their own right; they
# bring the system to a known state so that downstream conformance
# tests can run from a defined starting point.

Meta
    Name:      Station connection established
    Id:        station_connection_established
    Tags:      helper, lifecycle, wire-only
    Stations:  1
    Timeout:   10s

Scenario: Station opens an OCPP-J WebSocket to the CSMS
    When  station "CP01" connects to the CSMS
    Then  the OCPP-J handshake completes within 5 seconds
    And   station "CP01" is in the connected state
