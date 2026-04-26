# Authorize — same id token used at two stations concurrently.
#
# Validates that a CSMS implementing OCPP 2.0.1 §C01 Authorize rejects
# the second concurrent authorization attempt with status
# "ConcurrentTx" when the same id token is already authorized at
# another station with an active transaction.
#
# This scenario is wire-verifiable: both authorize attempts and both
# responses are observable on the OCPP wire; no privileged CSMS
# access is required.

Meta
    Name:        Authorize concurrent transaction rejection
    Id:          authorize_concurrent_rejected
    Spec-Ref:    OCPP 2.0.1 §C01 Authorize
    Tags:        core, authorization, multi-station
    Stations:    2
    Timeout:     30s
    Depends:
      - id:    station_boot_accepted
        scope: per-station

Background
    Given the CSMS is reachable
    And   the operator has provisioned id token "VID:0001" with status "Accepted"

Scenario: Same id token authorized at two stations concurrently
    When  station "CP01" starts a transaction with id token "VID:0001"
    And   the CSMS authorizes id token "VID:0001" for station "CP01"
    Parallel
        When  station "CP02" sends Authorize with id token "VID:0001"
    End-Parallel
    Then  the CSMS rejects "CP02" with status "ConcurrentTx"
    And   station "CP01"'s transaction remains active

Teardown
    Disconnect station "CP02"
    Stop transaction at station "CP01"
    Disconnect station "CP01"
