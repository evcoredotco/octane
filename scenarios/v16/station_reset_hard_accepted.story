# Station reset hard accepted.
#
# Validates that a CSMS implementing OCPP-J 1.6 §5.13 Reset sends a
# well-formed Reset.req CALL with type "Hard" to the station and that
# the station can respond with status "Accepted".
#
# A Hard reset requests an immediate restart without waiting for active
# transactions to complete. OCTANE validates the wire exchange only;
# actual station restart behaviour is out of scope.

Meta
    Name:        Station reset hard accepted
    Id:          station_reset_hard_accepted
    Spec-Ref:    OCPP-J 1.6 §5.13 Reset
    Tags:        station, csms-initiated, wire-only
    Stations:    1
    Timeout:     30s
    Depends:
      - id:    station_boot_accepted
        scope: per-station

Background
    Given the CSMS is reachable
    And   station "CP01" is registered to the CSMS

Scenario: CSMS sends Hard reset; station accepts
    When  the CSMS sends Reset with type "Hard" to station "CP01" within 30 seconds
    Then  station "CP01" responds to Reset with status "Accepted"

Teardown
    Disconnect station "CP01"
