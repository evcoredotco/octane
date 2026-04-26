# Helper story — connector status reported as Available.
#
# Sends an OCPP StatusNotification declaring the station's primary
# connector as "Available". This helper layers on top of
# station_boot_accepted and is the standard prerequisite for any
# scenario that operates on a connector (reservation, transactions,
# meter values, etc.).
#
# Helper stories carry no Spec-Ref.

Meta
    Name:      Connector reported as Available
    Id:        connector_status_available
    Tags:      helper, lifecycle, wire-only
    Stations:  1
    Timeout:   10s
    Depends:
      - id:    station_boot_accepted
        scope: per-station

Scenario: Station declares its connector available to the CSMS
    When  station "CP01" sends StatusNotification for connector 1 with status "Available"
    Then  the CSMS acknowledges the status within 10 seconds
    And   connector 1 of station "CP01" is in state "Available"
