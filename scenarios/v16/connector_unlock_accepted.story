# Connector unlock accepted.
#
# Validates that a CSMS implementing OCPP-J 1.6 -5.15 UnlockConnector
# sends a well-formed UnlockConnector.req CALL to the station and that
# the station can respond with status "Unlocked".
#
# This is a CSMS-initiated sequence. OCTANE validates the wire exchange
# only; the physical unlock mechanism is out of scope.

Meta
    Name:        Connector unlock accepted
    Id:          connector_unlock_accepted
    Spec-Ref:    OCPP-J 1.6 -5.15 UnlockConnector
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

Scenario: CSMS unlocks a connector; station reports Unlocked
    When  the CSMS sends UnlockConnector with connectorId {connectorId}
          to station "CP01" within 30 seconds
    Then  station "CP01" responds to UnlockConnector with status "Unlocked"

Teardown
    Disconnect station "CP01"
