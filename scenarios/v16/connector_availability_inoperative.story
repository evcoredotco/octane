# Connector availability change to Inoperative.
#
# Validates that a CSMS implementing OCPP-J 1.6 -5.2 ChangeAvailability
# sends a well-formed ChangeAvailability.req CALL with type "Inoperative"
# to the station and that the station can respond with status "Accepted".
#
# This is a CSMS-initiated sequence. Changing availability to Inoperative
# disables a connector so that it cannot start new transactions until the
# CSMS sets it back to Operative.

Meta
    Name:        Connector availability inoperative
    Id:          connector_availability_inoperative
    Spec-Ref:    OCPP-J 1.6 -5.2 ChangeAvailability
    Tags:        connector, csms-initiated, wire-only
    Stations:    1
    Timeout:     30s
    Parameters:  connectorId
    Depends:
      - id:    connector_status_available
        scope: per-station

Background
    Given the CSMS is reachable
    And   station "CP01" is registered to the CSMS

Scenario: CSMS changes connector availability to Inoperative; station accepts
    When  the CSMS sends ChangeAvailability with connectorId {connectorId}
          and type "Inoperative" to station "CP01" within 30 seconds
    Then  station "CP01" responds to ChangeAvailability with status "Accepted"

Teardown
    Disconnect station "CP01"
