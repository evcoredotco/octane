# Connector availability change to Operative.
#
# Validates that a CSMS implementing OCPP-J 1.6 §5.2 ChangeAvailability
# sends a well-formed ChangeAvailability.req CALL with type "Operative"
# to the station and that the station can respond with status "Accepted".
#
# This is a CSMS-initiated sequence. Changing availability to Operative
# re-enables a connector that was previously set Inoperative.

Meta
    Name:        Connector availability operative
    Id:          connector_availability_operative
    Spec-Ref:    OCPP-J 1.6 §5.2 ChangeAvailability
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

Scenario: CSMS changes connector availability to Operative; station accepts
    When  the CSMS sends ChangeAvailability with connectorId {connectorId}
          and type "Operative" to station "CP01" within 30 seconds
    Then  station "CP01" responds to ChangeAvailability with status "Accepted"

Teardown
    Disconnect station "CP01"
