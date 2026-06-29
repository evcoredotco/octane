# Meter values periodic accepted.
#
# Validates that a CSMS implementing OCPP-J 1.6 -4.7 MeterValues
# correctly processes a periodic MeterValues.req sent by the station
# during an active transaction and responds with an empty MeterValues.conf.
#
# The station sends a single sampled energy reading (unit Wh) for the
# active connector. OCTANE validates that the CSMS acknowledges the
# message within the allowed timeout.
#
# This scenario depends on transaction_pluginfirst_accepted to ensure
# a transaction is already running before the meter values are sent.

Meta
    Name:        Meter values periodic accepted
    Id:          meter_values_periodic_accepted
    Spec-Ref:    OCPP-J 1.6 -4.7 MeterValues
    Tags:        metering, wire-only
    Stations:    1
    Timeout:     60s
    Parameters:  connectorId, valid_idTag, meterStart, meterValue
    Depends:
      - id:    transaction_pluginfirst_accepted
        scope: per-station

Background
    Given the CSMS is reachable
    And   the operator has provisioned id token "{valid_idTag}" with status "Accepted"

Scenario: Station sends a periodic MeterValues message; CSMS acknowledges
    When  station "CP01" sends MeterValues for connector {connectorId}
          with sampled value "{meterValue}"
    Then  the CSMS acknowledges MeterValues within 10 seconds

Teardown
    Disconnect station "CP01"
