# Authorization cache clear accepted.
#
# Validates that a CSMS implementing OCPP-J 1.6 -5.4 ClearCache sends
# a well-formed ClearCache.req CALL to the station and that the station
# can respond with status "Accepted".
#
# Per OCPP-J 1.6 -5.4 the ClearCache request has an empty payload;
# the station must clear its local authorization cache and respond with
# the outcome. OCTANE validates the wire exchange only.

Meta
    Name:        Authorization cache clear accepted
    Id:          cache_clear_accepted
    Spec-Ref:    OCPP-J 1.6 -5.4 ClearCache
    Tags:        cache, csms-initiated, wire-only
    Stations:    1
    Timeout:     30s
    Depends:
      - id:    station_boot_accepted
        scope: per-station

Background
    Given the CSMS is reachable
    And   station "CP01" is registered to the CSMS

Scenario: CSMS clears the authorization cache; station accepts
    When  the CSMS sends ClearCache to station "CP01" within 30 seconds
    Then  station "CP01" responds to ClearCache with status "Accepted"

Teardown
    Disconnect station "CP01"
