# Helper story — station boot accepted.
#
# Performs the OCPP BootNotification handshake and waits for the
# CSMS to accept the registration. This helper layers on
# station_connection_established and is the most commonly used
# prerequisite across the conformance suite, because almost every
# OCPP scenario assumes a station that has booted and been accepted.
#
# Helper stories carry no Spec-Ref. The boot behavior they exercise
# IS specified in OCPP-J 1.6 §6.5 BootNotification, but the
# conformance assertion for that section lives in the dedicated
# boot_notification_accepted story; this helper merely needs the
# state and trusts that the conformance test verifies it elsewhere.

Meta
    Name:      Station boot accepted
    Id:        station_boot_accepted
    Tags:      helper, lifecycle, wire-only
    Stations:  1
    Timeout:   30s
    Depends:
      - id:    station_connection_established
        scope: per-station

Scenario: Station registers with the CSMS and is accepted
    When  station "CP01" sends BootNotification with reason "PowerUp"
    Then  the CSMS responds with status "Accepted" within 30 seconds
    And   station "CP01" is in the registered state
