# Connector reservation faulted.
#
# Validates that a CSMS implementing OCPP-J 1.6 -6.40 ReserveNow
# correctly handles the case where a charging station rejects a
# reservation request by responding with status "Faulted". The CSMS
# must accept the response without raising an OCPP-level error;
# whether or not it surfaces the rejection to upstream operator
# tooling is out of OCTANE's wire-only scope.
#
# This is a CSMS-initiated, single-station scenario. The dependency
# chain ensures the connector under test is in the "Available" state
# (per OCPP-J 1.6 -4.7) before the reservation request is sent;
# without that prerequisite, the CSMS may legitimately respond
# differently and the test would fail for the wrong reason.

Meta
    Name:        Connector reservation faulted
    Id:          connector_reservation_faulted
    Spec-Ref:    OCPP-J 1.6 -6.40 ReserveNow
    Tags:        reservation, csms-initiated, wire-only
    Stations:    1
    Timeout:     30s
    Parameters:  connectorId, idTag
    Depends:
      - id:    connector_status_available
        scope: per-station

Background
    Given the CSMS is reachable
    And   station "CP01" is registered to the CSMS

Scenario: CSMS handles a Faulted reservation response
    When  the CSMS sends ReserveNow with connectorId {connectorId}
          and idTag "{idTag}" to station "CP01" within 30 seconds
    Then  station "CP01" responds with ReserveNow.conf status "Faulted"
    And   the CSMS accepts the response without error within 10 seconds

Teardown
    Disconnect station "CP01"
