# Configuration get accepted.
#
# Validates that a CSMS implementing OCPP-J 1.6 §5.7 GetConfiguration
# sends a well-formed GetConfiguration.req CALL to the station and that
# the station can respond with a configurationKey array containing at
# least one entry.
#
# Per OCPP-J 1.6 §5.7 the request may optionally include a list of keys
# to retrieve; when no keys are specified the station returns all
# supported configuration parameters. OCTANE responds with one generic
# key entry to satisfy the protocol exchange.

Meta
    Name:        Configuration get accepted
    Id:          configuration_get_accepted
    Spec-Ref:    OCPP-J 1.6 §5.7 GetConfiguration
    Tags:        configuration, csms-initiated, wire-only
    Stations:    1
    Timeout:     30s
    Depends:
      - id:    station_boot_accepted
        scope: per-station

Background
    Given the CSMS is reachable
    And   station "CP01" is registered to the CSMS

Scenario: CSMS retrieves station configuration; station responds with one key
    When  the CSMS sends GetConfiguration to station "CP01" within 30 seconds
    Then  station "CP01" responds to GetConfiguration with 1 configuration keys

Teardown
    Disconnect station "CP01"
