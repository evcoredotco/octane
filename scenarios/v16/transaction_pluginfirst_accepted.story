# Transaction start, plugin-first variant, accepted.
#
# Validates the conformance of a CSMS to the plugin-first start
# sequence defined in OCPP-J 1.6. In this flow, the EV driver plugs
# the cable in *before* presenting an identification token. The
# station observes the plug-in event, advances the connector to
# "Preparing", and only then triggers identification and the
# StartTransaction.req.
#
# The specification-defined sequence under test is:
#
#   1. Connector transitions Available -> Preparing
#      (StatusNotification.req with status="Preparing").
#   2. CSMS acknowledges the status change.
#   3. Driver presents idTag; station sends Authorize.req.
#   4. CSMS responds with Authorize.conf carrying
#      idTagInfo.status="Accepted".
#   5. Station sends StartTransaction.req with the connectorId,
#      idTag, meterStart, and timestamp.
#   6. CSMS responds with StartTransaction.conf carrying
#      idTagInfo.status="Accepted" and a unique transactionId.
#   7. Connector transitions Preparing -> Charging
#      (StatusNotification.req with status="Charging").
#   8. CSMS acknowledges the final status change.
#
# This is the happy-path scenario: every CSMS response is "Accepted"
# and the transaction starts cleanly. Failure-path variants (rejected
# authorization, blocked tag, concurrent transaction) are covered by
# separate stories.
#
# Background note: this test assumes the operator has provisioned an
# idTag with status "Accepted" in the CSMS before the run. OCTANE
# cannot provision idTags over the wire (that would be a CSMS-
# specific adaptation, forbidden by constitution principle XII), so
# the precondition is declared in Background and the operator must
# satisfy it externally.

Meta
    Name:        Transaction start plugin-first accepted
    Id:          transaction_pluginfirst_accepted
    Spec-Ref:    OCPP-J 1.6 -6.13 StatusNotification, -6.2 Authorize, -6.16 StartTransaction
    Tags:        transaction, charging, wire-only
    Stations:    1
    Timeout:     60s
    Parameters:  connectorId, idTag, meterStart
    Depends:
      - id:    connector_status_available
        scope: per-station

Background
    Given the CSMS is reachable
    And   the operator has provisioned id token "{idTag}" with status "Accepted"

Scenario: Plug-in precedes authorization; transaction starts cleanly
    When  station "CP01" sends StatusNotification for connector {connectorId} with status "Preparing"
    Then  the CSMS acknowledges the status within 10 seconds

    When  station "CP01" sends Authorize with idTag "{idTag}"
    Then  the CSMS responds to Authorize with idTagInfo.status "Accepted" within 30 seconds

    When  station "CP01" starts a transaction on connector {connectorId} with idTag "{idTag}" and meterStart {meterStart}
    Then  the CSMS responds to StartTransaction with idTagInfo.status "Accepted" within 30 seconds
    And   the StartTransaction response assigns a positive transactionId

    When  station "CP01" sends StatusNotification for connector {connectorId} with status "Charging"
    Then  the CSMS acknowledges the status within 10 seconds

Teardown
    Disconnect station "CP01"
