# EV-side disconnect stops the transaction and releases the connector.
#
# Validates the charging-session path where the EV-side cable is removed
# while a transaction is active. The station reports the connector as
# SuspendedEV, stops the active transaction with reason EVDisconnected,
# reports Finishing while the charge-point-side cable is still attached,
# and finally reports Available after the local plug is removed.
#
# Background note: the story establishes a fresh connection, registers the
# station, makes the connector available, authorizes the idTag, and starts
# a transaction before exercising the disconnect sequence. The accepted
# idTag remains an explicit operator precondition, because OCPP 1.6 has no
# station-side command that can provision authorization data in a generic
# CSMS.

Meta
    Name:        EV-side disconnect stops transaction and unlocks connector
    Id:          transaction_evside_disconnect_true_true
    Spec-Ref:    OCPP-J 1.6 -6.13 StatusNotification, -6.2 Authorize, -6.16 StartTransaction, -6.47 StopTransaction
    Tags:        transaction, charging, disconnect, wire-only
    Stations:    1
    Timeout:     180s
    Parameters:  connectorId, valid_idTag, meterStart, meterStop

Background
    Given the CSMS is reachable
    And   the operator has provisioned id token "{valid_idTag}" with status "Accepted"

Scenario: EV-side disconnect stops charging and returns the connector to Available
    When  station "CP01" connects to the CSMS
    Then  the OCPP-J handshake completes within 5 seconds
    And   station "CP01" is in the connected state

    When  station "CP01" sends BootNotification with reason "PowerUp"
    Then  the CSMS responds with status "Accepted" within 30 seconds
    And   station "CP01" is in the registered state

    When  station "CP01" sends StatusNotification for connector 0 with status "Available"
    Then  the CSMS acknowledges the status within 10 seconds

    When  station "CP01" sends StatusNotification for connector {connectorId} with status "Available"
    Then  the CSMS acknowledges the status within 10 seconds
    And   connector {connectorId} of station "CP01" is in state "Available"

    When  station "CP01" sends StatusNotification for connector {connectorId} with status "Preparing"
    Then  the CSMS acknowledges the status within 10 seconds

    When  station "CP01" sends Authorize with idTag "{valid_idTag}"
    Then  the CSMS responds to Authorize with idTagInfo.status "Accepted" within 30 seconds

    When  station "CP01" starts a transaction on connector {connectorId} with idTag "{valid_idTag}" and meterStart {meterStart}
    Then  the CSMS responds to StartTransaction with idTagInfo.status "Accepted" within 30 seconds
    And   the StartTransaction response assigns a positive transactionId

    When  station "CP01" sends StatusNotification for connector {connectorId} with status "Charging"
    Then  the CSMS acknowledges the status within 10 seconds
    And   connector {connectorId} of station "CP01" is in state "Charging"

    When  station "CP01" sends StatusNotification for connector {connectorId} with status "SuspendedEV"
    Then  the CSMS acknowledges the status within 10 seconds

    When  station "CP01" stops transaction 0 with meterStop {meterStop} and reason "EVDisconnected"
    Then  the CSMS accepts StopTransaction within 30 seconds

    When  station "CP01" sends StatusNotification for connector {connectorId} with status "Finishing"
    Then  the CSMS acknowledges the status within 10 seconds

    When  station "CP01" sends StatusNotification for connector {connectorId} with status "Available"
    Then  the CSMS acknowledges the status within 10 seconds
    And   connector {connectorId} of station "CP01" is in state "Available"

Teardown
    Disconnect station "CP01"
