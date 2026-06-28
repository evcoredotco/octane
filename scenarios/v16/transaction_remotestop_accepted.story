# Transaction remote stop accepted.
#
# Validates that a CSMS implementing OCPP-J 1.6 §5.12
# RemoteStopTransaction sends a well-formed RemoteStopTransaction.req
# CALL to the station and that the station can respond with status
# "Accepted".
#
# This is a CSMS-initiated sequence. OCTANE waits for the CSMS to send
# the CALL within the given timeout and then responds on behalf of the
# station. The scenario depends on transaction_pluginfirst_accepted to
# ensure a transaction is already running before the remote stop is sent.

Meta
    Name:        Transaction remote stop accepted
    Id:          transaction_remotestop_accepted
    Spec-Ref:    OCPP-J 1.6 §5.12 RemoteStopTransaction
    Tags:        transaction, csms-initiated, wire-only
    Stations:    1
    Timeout:     60s
    Parameters:  connectorId, idTag, meterStart, transactionId
    Depends:
      - id:    transaction_pluginfirst_accepted
        scope: per-station

Background
    Given the CSMS is reachable
    And   station "CP01" is registered to the CSMS

Scenario: CSMS remotely stops a running transaction; station accepts
    When  the CSMS sends RemoteStopTransaction with transactionId {transactionId}
          to station "CP01" within 30 seconds
    Then  station "CP01" responds to RemoteStopTransaction with status "Accepted"

Teardown
    Disconnect station "CP01"
