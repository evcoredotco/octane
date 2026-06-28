# Transaction remote start accepted.
#
# Validates that a CSMS implementing OCPP-J 1.6 -5.11
# RemoteStartTransaction sends a well-formed RemoteStartTransaction.req
# CALL to the station and that the station can respond with status
# "Accepted".
#
# This is a CSMS-initiated sequence. OCTANE waits for the CSMS to send
# the CALL within the given timeout and then responds on behalf of the
# station. The scenario depends on connector_status_available to ensure
# the connector is in the correct state before the remote start is sent.

Meta
    Name:        Transaction remote start accepted
    Id:          transaction_remotestart_accepted
    Spec-Ref:    OCPP-J 1.6 -5.11 RemoteStartTransaction
    Tags:        transaction, csms-initiated, wire-only
    Stations:    1
    Timeout:     60s
    Parameters:  connectorId, idTag
    Depends:
      - id:    connector_status_available
        scope: per-station

Background
    Given the CSMS is reachable
    And   station "CP01" is registered to the CSMS

Scenario: CSMS initiates a remote start; station accepts
    When  the CSMS sends RemoteStartTransaction with connectorId {connectorId}
          and idTag "{idTag}" to station "CP01" within 30 seconds
    Then  station "CP01" responds to RemoteStartTransaction with status "Accepted"

Teardown
    Disconnect station "CP01"
