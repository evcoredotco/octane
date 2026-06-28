# Transaction stop accepted.
#
# Validates that a CSMS implementing OCPP-J 1.6 -6.47 StopTransaction
# correctly processes a StopTransaction.req sent by the station and
# responds with a StopTransaction.conf acknowledging the request.
#
# The station reports reason "Local" (driver-initiated stop). Per
# OCPP-J 1.6 -6.47 the confirmation payload is an acknowledgement
# with an optional idTagInfo; OCTANE only verifies that a well-formed
# response arrives within the allowed timeout.
#
# This scenario depends on transaction_pluginfirst_accepted to ensure
# a transaction is already running before the stop is sent.

Meta
    Name:        Transaction stop accepted
    Id:          transaction_stop_accepted
    Spec-Ref:    OCPP-J 1.6 -6.47 StopTransaction
    Tags:        transaction, charging, wire-only
    Stations:    1
    Timeout:     60s
    Parameters:  connectorId, idTag, meterStart, transactionId, meterStop
    Depends:
      - id:    transaction_pluginfirst_accepted
        scope: per-station

Background
    Given the CSMS is reachable
    And   the operator has provisioned id token "{idTag}" with status "Accepted"

Scenario: Station stops an accepted transaction; CSMS acknowledges
    When  station "CP01" stops transaction {transactionId} with meterStop {meterStop} and reason "Local"
    Then  the CSMS accepts StopTransaction within 30 seconds

Teardown
    Disconnect station "CP01"
