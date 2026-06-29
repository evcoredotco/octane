# Identification-first timeout returns the connector to Available.
#
# Validates the identification-first path where the station receives an
# accepted authorization, the connector enters Preparing, and no
# transaction starts before the configured connection timeout expires.
# The station then reports the connector as Available again and the CSMS
# acknowledges the status update.
#
# Background note: station connection, boot acceptance, and connector
# availability setup are included here so this story can run on an empty
# CSMS state. The accepted idTag remains an explicit operator
# precondition, because the station-side OCPP 1.6 flow has no portable way
# to provision authorization data in the CSMS under test.

Meta
    Name:        Transaction identification-first timeout returns connector available
    Id:          transaction_identificationfirst_connection_timeout_available
    Spec-Ref:    OCPP-J 1.6 -6.13 StatusNotification, -6.2 Authorize
    Tags:        transaction, charging, timeout, wire-only
    Stations:    1
    Timeout:     180s
    Parameters:  connectorId, valid_idTag, connectionTimeOut

Background
    Given the CSMS is reachable
    And   the operator has provisioned id token "{valid_idTag}" with status "Accepted"

Scenario: Authorization is accepted but plug-in times out before a transaction starts
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

    When  station "CP01" sends Authorize with idTag "{valid_idTag}"
    Then  the CSMS responds to Authorize with idTagInfo.status "Accepted" within 30 seconds

    When  station "CP01" sends StatusNotification for connector {connectorId} with status "Preparing"
    Then  the CSMS acknowledges the status within 10 seconds

    When  wait {connectionTimeOut}
    And   station "CP01" sends StatusNotification for connector {connectorId} with status "Available"
    Then  the CSMS acknowledges the status within 10 seconds
    And   connector {connectorId} of station "CP01" is in state "Available"

Teardown
    Disconnect station "CP01"
