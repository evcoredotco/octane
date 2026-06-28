# Connector cancel reservation accepted.
#
# Validates that a CSMS implementing OCPP-J 1.6 -5.17 CancelReservation
# sends a well-formed CancelReservation.req CALL with the expected
# reservationId to the station and that the station can respond with
# status "Accepted".
#
# Per OCPP-J 1.6 -5.17 the station must respond with "Accepted" if the
# reservationId matches an active reservation, or "Rejected" if it does
# not. This story validates the accepted path. The connector_reservation_faulted
# dependency is used for connection setup; no active reservation is required
# for the CancelReservation exchange to succeed per the spec.

Meta
    Name:        Connector cancel reservation accepted
    Id:          connector_cancelreservation_accepted
    Spec-Ref:    OCPP-J 1.6 -5.17 CancelReservation
    Tags:        reservation, csms-initiated, wire-only
    Stations:    1
    Timeout:     30s
    Parameters:  connectorId, idTag, reservationId
    Depends:
      - id:    connector_reservation_faulted
        scope: per-station

Background
    Given the CSMS is reachable
    And   station "CP01" is registered to the CSMS

Scenario: CSMS cancels a reservation; station accepts
    When  the CSMS sends CancelReservation with reservationId {reservationId}
          to station "CP01" within 30 seconds
    Then  station "CP01" responds to CancelReservation with status "Accepted"

Teardown
    Disconnect station "CP01"
