# Configuration change accepted.
#
# Validates that a CSMS implementing OCPP-J 1.6 -5.3 ChangeConfiguration
# sends a well-formed ChangeConfiguration.req CALL with the expected key
# and value to the station and that the station can respond with status
# "Accepted".
#
# This is a CSMS-initiated sequence. OCTANE validates the wire exchange
# only; whether the station actually applies the configuration change is
# out of scope.

Meta
    Name:        Configuration change accepted
    Id:          configuration_change_accepted
    Spec-Ref:    OCPP-J 1.6 -5.3 ChangeConfiguration
    Tags:        configuration, csms-initiated, wire-only
    Stations:    1
    Timeout:     30s
    Parameters:  key, value
    Depends:
      - id:    station_boot_accepted
        scope: per-station

Background
    Given the CSMS is reachable
    And   station "CP01" is registered to the CSMS

Scenario: CSMS changes a configuration parameter; station accepts
    When  the CSMS sends ChangeConfiguration with key "{key}"
          and value "{value}" to station "CP01" within 30 seconds
    Then  station "CP01" responds to ChangeConfiguration with status "Accepted"

Teardown
    Disconnect station "CP01"
