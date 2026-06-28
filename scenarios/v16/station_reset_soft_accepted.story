# Station reset soft accepted.
#
# Validates that a CSMS implementing OCPP-J 1.6 -5.13 Reset sends a
# well-formed Reset.req CALL with type "Soft" to the station and that
# the station can respond with status "Accepted".
#
# A Soft reset requests the station to gracefully complete any active
# transactions before restarting. OCTANE validates the wire exchange
# only; actual station restart behaviour is out of scope.

Meta
    Name:        Station reset soft accepted
    Id:          station_reset_soft_accepted
    Spec-Ref:    OCPP-J 1.6 -5.13 Reset
    Tags:        station, csms-initiated, wire-only
    Stations:    1
    Timeout:     30s
    Depends:
      - id:    station_boot_accepted
        scope: per-station

Background
    Given the CSMS is reachable
    And   station "CP01" is registered to the CSMS

Scenario: CSMS sends Soft reset; station accepts
    When  the CSMS sends Reset with type "Soft" to station "CP01" within 30 seconds
    Then  station "CP01" responds to Reset with status "Accepted"

Teardown
    Disconnect station "CP01"
