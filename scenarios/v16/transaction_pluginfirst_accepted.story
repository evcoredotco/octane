# Transaction start, plugin-first variant, accepted.
#
# Validates the conformance of a CSMS to the plugin-first start
# sequence defined in OCPP-J 1.6. The story starts from a fresh
# station connection, performs registration, reports the connector as
# Available, and then exercises the plug-in-before-identification
# transaction path.
#
# The transaction portion under test is:
#
#   1. The station reports the selected connector as Preparing.
#   2. The CSMS acknowledges the Preparing status update.
#   3. The station asks the CSMS to authorize the presented idTag.
#   4. The CSMS accepts the idTag.
#   5. The station starts a transaction for the connector and idTag.
#   6. The CSMS accepts the transaction and assigns a transactionId.
#   7. The station reports the selected connector as Charging.
#   8. The CSMS acknowledges the Charging status update.
#
# This is the happy-path scenario: every CSMS response is "Accepted"
# and the transaction starts cleanly. Failure-path variants (rejected
# authorization, blocked tag, concurrent transaction) are covered by
# separate stories.
#
# Background note: the station boot and connector availability setup
# are part of this story so the scenario does not rely on helper-story
# state. The idTag acceptance policy still belongs to the CSMS under
# test; OCTANE records the required idTag status as an explicit
# precondition because OCPP 1.6 does not provide a CSMS-neutral way for
# the station side to provision authorization data.

Meta
    Name:        Transaction start plugin-first accepted
    Id:          transaction_pluginfirst_accepted
    Spec-Ref:    OCPP-J 1.6 -6.13 StatusNotification, -6.2 Authorize, -6.16 StartTransaction
    Tags:        transaction, charging, wire-only
    Stations:    1
    Timeout:     120s
    Parameters:  connectorId, valid_idTag, meterStart

Background
    Given the CSMS is reachable
    And   the operator has provisioned id token "{valid_idTag}" with status "Accepted"

Scenario: Plug-in precedes authorization; transaction starts cleanly
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

Teardown
    Disconnect station "CP01"
