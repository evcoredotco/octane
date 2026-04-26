# BootNotification — central response with accepted registration.
#
# Validates that a CSMS implementing OCPP 2.0.1 §B01 BootNotification
# replies to a well-formed BootNotification.req with a
# BootNotificationResponse carrying status "Accepted" and a
# heartbeatInterval within the spec-permitted range.
#
# This is a single-station, wire-only conformance test. It depends on
# a successful WebSocket connection but does not assume any prior
# state on the CSMS.

Meta
    Name:        Boot notification with accepted registration
    Id:          boot_notification_accepted
    Spec-Ref:    OCPP 2.0.1 §B01 BootNotification
    Tags:        core, boot, wire-only
    Stations:    1
    Timeout:     30s
    Depends:
      - id:    station_connection_established
        scope: per-station

Background
    Given the CSMS is reachable

Scenario: Accepted registration on first boot
    When  station "CP01" sends BootNotification with reason "PowerUp"
    Then  the CSMS responds with status "Accepted" within 30 seconds
    And   the response includes a heartbeatInterval between 30 and 86400
    And   the response interval is a positive integer
    And   the response includes a currentTime in ISO-8601 format

Teardown
    Disconnect station "CP01"
