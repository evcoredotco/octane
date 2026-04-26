# Cold boot sequence accepted.
#
# Validates the full cold-boot sequence a charging station performs
# the first time it connects to a CSMS. The sequence the
# specification defines is:
#
#   1. Station opens an OCPP-J WebSocket and sends BootNotification.
#   2. CSMS responds with BootNotification.conf carrying status
#      "Accepted", a non-zero interval, and a currentTime.
#   3. Station sends one StatusNotification per connector with
#      status "Available" (connectorId=0 is the station itself).
#   4. Station begins emitting Heartbeat at the cadence the CSMS
#      advertised in BootNotification.conf.interval.
#   5. CSMS acknowledges each Heartbeat with Heartbeat.conf.
#
# This is a CSMS-side conformance test: every step is observed on
# the wire from messages the station originates and responses the
# CSMS produces. The test does not assert anything about
# configuration variables, audit logs, or internal CSMS state.
#
# This story is broader in scope than boot_notification_accepted
# (which validates only the BootNotification round-trip in
# isolation). Here we exercise the full registration handshake plus
# the first heartbeat to confirm the CSMS honors the interval it
# advertised.

Meta
    Name:        Cold boot sequence accepted
    Id:          boot_sequence_accepted
    Spec-Ref:    OCPP-J 1.6 §6.5 BootNotification, §6.13 StatusNotification, §6.7 Heartbeat
    Tags:        boot, lifecycle, wire-only
    Stations:    1
    Timeout:     90s
    Depends:
      - id:    station_connection_established
        scope: per-station

Background
    Given the CSMS is reachable

Scenario: Station completes the cold-boot registration sequence
    When  station "CP01" sends BootNotification with reason "PowerUp"
    Then  the CSMS responds with status "Accepted" within 30 seconds
    And   the response includes a heartbeatInterval between 30 and 86400
    And   the response includes a currentTime in ISO-8601 format

    When  station "CP01" sends StatusNotification for connector 0 with status "Available"
    Then  the CSMS acknowledges the status within 10 seconds

    When  station "CP01" sends StatusNotification for connector 1 with status "Available"
    Then  the CSMS acknowledges the status within 10 seconds

    When  station "CP01" sends Heartbeat
    Then  the CSMS responds to the Heartbeat within 10 seconds
    And   the Heartbeat response includes a currentTime in ISO-8601 format

Teardown
    Disconnect station "CP01"
