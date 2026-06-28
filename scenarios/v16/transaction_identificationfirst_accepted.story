# Transaction start, identification-first variant, accepted.
#
# Validates the conformance of a CSMS to the identification-first
# start sequence defined in OCPP-J 1.6. In this flow, the EV driver
# presents an identification token *before* plugging the cable in
# (typical for stations with an RFID reader at the unit but a separate
# cable management surface). Authorization happens against the
# Available connector; the cable is plugged in afterward and the
# transaction starts.
#
# The specification-defined sequence under test is:
#
#   1. Driver presents idTag while connector is Available; station
#      sends Authorize.req.
#   2. CSMS responds with Authorize.conf carrying
#      idTagInfo.status="Accepted".
#   3. Driver plugs in; connector transitions Available -> Preparing
#      (StatusNotification.req with status="Preparing").
#   4. CSMS acknowledges the status change.
#   5. Station sends StartTransaction.req with the connectorId,
#      previously authorized idTag, meterStart, and timestamp.
#   6. CSMS responds with StartTransaction.conf carrying
#      idTagInfo.status="Accepted" and a unique transactionId.
#   7. Connector transitions Preparing -> Charging
#      (StatusNotification.req with status="Charging").
#   8. CSMS acknowledges the final status change.
#
# The functional contract differs from plugin-first only in step
# ordering: Authorize precedes the Preparing status. The CSMS must
# accept either ordering as conformant, because the OCPP-J
# specification permits both modes. This test verifies the
# identification-first ordering specifically.
#
# Background note: as with the plugin-first variant, this test
# assumes the operator has provisioned an idTag with status
# "Accepted" in the CSMS before the run.

Meta
    Name:        Transaction start identification-first accepted
    Id:          transaction_identificationfirst_accepted
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

Scenario: Authorization precedes plug-in; transaction starts cleanly
    When  station "CP01" sends Authorize with idTag "{idTag}"
    Then  the CSMS responds to Authorize with idTagInfo.status "Accepted" within 30 seconds

    When  station "CP01" sends StatusNotification for connector {connectorId} with status "Preparing"
    Then  the CSMS acknowledges the status within 10 seconds

    When  station "CP01" starts a transaction on connector {connectorId} with idTag "{idTag}" and meterStart {meterStart}
    Then  the CSMS responds to StartTransaction with idTagInfo.status "Accepted" within 30 seconds
    And   the StartTransaction response assigns a positive transactionId

    When  station "CP01" sends StatusNotification for connector {connectorId} with status "Charging"
    Then  the CSMS acknowledges the status within 10 seconds

Teardown
    Disconnect station "CP01"
