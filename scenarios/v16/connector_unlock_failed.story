# Connector unlock failed.
#
# Validates that a CSMS implementing OCPP-J 1.6 -5.15 UnlockConnector
# correctly handles a station response of status "UnlockFailed". This
# occurs when the station cannot physically release the connector lock
# (e.g., the lock mechanism is stuck or the cable is under tension).
#
# The CSMS must accept the UnlockFailed response without raising a
# protocol error; how it surfaces the failure to the operator is out
# of OCTANE's wire-only scope.

Meta
    Name:        Connector unlock failed
    Id:          connector_unlock_failed
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

Scenario: CSMS tries to unlock a connector; station reports UnlockFailed
    When  the CSMS sends UnlockConnector with connectorId {connectorId}
          to station "CP01" within 30 seconds
    Then  station "CP01" responds to UnlockConnector with status "UnlockFailed"

Teardown
    Disconnect station "CP01"
