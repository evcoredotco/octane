# Pure-protocol scenario: malformed BootNotification.req is rejected.
#
# Verifies that a CSMS implementing OCPP-J 2.0.1 §3 emits a CALLERROR
# with code "FormatViolation" when receiving an OCPP-J CALL frame
# whose payload does not match the BootNotification.req JSON schema.
#
# This scenario verifies wire-level robustness only. It does not
# require any CSMS-side state and runs identically against every
# OCPP 2.0.1 CSMS. The "wire-only" + "pure-protocol" tag combination
# tells the runner this scenario can be invoked without any prior
# fixture state.

Meta
    Name:        Malformed boot notification rejected
    Id:          boot_notification_malformed
    Spec-Ref:    OCPP-J 2.0.1 §3 Wire Protocol
    Tags:        core, boot, wire-only, pure-protocol
    Stations:    1
    Timeout:     10s
    Depends:
      - id:    station_connection_established
        scope: per-station

Background
    Given the CSMS is reachable

Scenario: CSMS rejects an OCPP-J frame with malformed payload
    When  station "CP01" sends a raw frame "[2,\"abc\",\"BootNotification\",{}]"
    Then  the CSMS responds with a CALLERROR within 10 seconds
    And   the CALLERROR error code equals "FormatViolation"

Teardown
    Disconnect station "CP01"
