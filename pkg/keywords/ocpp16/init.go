package ocpp16

import (
	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/registry"
)

// Register adds every OCPP 1.6 domain keyword to the global registry.
// Call this once from the CLI entry point (Execute) before running any stories.
func Register() {
	registerPreconditionKeywords()
	registerBootKeywords()
	registerStatusKeywords()
	registerHeartbeatKeywords()
	registerTeardownKeywords()
	registerAuthorizeKeywords()
	registerTransactionKeywords()
	registerReserveKeywords()
	registerStopTransactionKeywords()
	registerMeterValuesKeywords()
	registerRemoteStartKeywords()
	registerRemoteStopKeywords()
	registerResetKeywords()
	registerUnlockKeywords()
	registerAvailabilityKeywords()
	registerGetConfigurationKeywords()
	registerChangeConfigurationKeywords()
	registerClearCacheKeywords()
	registerCancelReservationKeywords()
}

func registerPreconditionKeywords() {
	registry.Register(api.Keyword{
		Pattern:     "the CSMS is reachable",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        csmsIsReachable,
	})

	registry.Register(api.Keyword{
		Pattern:     "the operator has provisioned id token {idTag:string} with status {status:string}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        operatorProvisionedIDTag,
	})

	registry.Register(api.Keyword{
		Pattern:     "station {station:string} is registered to the CSMS",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        stationIsRegistered,
	})
}

func registerBootKeywords() {
	registry.Register(api.Keyword{
		Pattern:     "station {station:string} sends BootNotification with reason {reason:string}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        sendBootNotification,
	})

	registry.Register(api.Keyword{
		Pattern:     "the CSMS responds with status {status:string} within {timeout:duration}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        csmsRespondsWithStatus,
	})

	registry.Register(api.Keyword{
		Pattern:     "station {station:string} is in the registered state",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        stationIsInRegisteredState,
	})

	registry.Register(api.Keyword{
		Pattern:     "the response includes a heartbeatInterval between {min:int} and {max:int}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        responseIncludesHeartbeatInterval,
	})

	registry.Register(api.Keyword{
		Pattern:     "the response includes a currentTime in ISO-8601 format",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        responseIncludesCurrentTime,
	})
}

func registerStatusKeywords() {
	registry.Register(api.Keyword{
		Pattern: "station {station:string} sends StatusNotification for connector" +
			" {connectorId:int} with status {status:string}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        sendStatusNotification,
	})

	registry.Register(api.Keyword{
		Pattern:     "the CSMS acknowledges the status within {timeout:duration}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        csmsAcknowledgesStatus,
	})

	registry.Register(api.Keyword{
		Pattern:     "connector {connectorId:int} of station {station:string} is in state {state:string}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        connectorIsInState,
	})
}

func registerHeartbeatKeywords() {
	registry.Register(api.Keyword{
		Pattern:     "station {station:string} sends Heartbeat",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        sendHeartbeat,
	})

	registry.Register(api.Keyword{
		Pattern:     "the CSMS responds to the Heartbeat within {timeout:duration}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        csmsRespondsToHeartbeat,
	})

	registry.Register(api.Keyword{
		Pattern:     "the Heartbeat response includes a currentTime in ISO-8601 format",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        heartbeatResponseIncludesCurrentTime,
	})
}

func registerTeardownKeywords() {
	registry.Register(api.Keyword{
		Pattern:     "Disconnect station {station:string}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        disconnectStation,
	})
}

func registerAuthorizeKeywords() {
	registry.Register(api.Keyword{
		Pattern:     "station {station:string} sends Authorize with idTag {idTag:string}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        sendAuthorize,
	})

	registry.Register(api.Keyword{
		Pattern: "the CSMS responds to Authorize with idTagInfo.status" +
			" {status:string} within {timeout:duration}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        csmsRespondsToAuthorize,
	})
}

func registerTransactionKeywords() {
	registry.Register(api.Keyword{
		Pattern: "station {station:string} starts a transaction on connector" +
			" {connectorId:int} with idTag {idTag:string} and meterStart {meterStart:int}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        sendStartTransaction,
	})

	registry.Register(api.Keyword{
		Pattern: "the CSMS responds to StartTransaction with idTagInfo.status" +
			" {status:string} within {timeout:duration}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        csmsRespondsToStartTransaction,
	})

	registry.Register(api.Keyword{
		Pattern:     "the StartTransaction response assigns a positive transactionId",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        startTransactionAssignsPositiveTransactionID,
	})
}

func registerReserveKeywords() {
	registry.Register(api.Keyword{
		Pattern: "the CSMS sends ReserveNow with connectorId {connectorId:int}" +
			" and idTag {idTag:string} to station {station:string} within {timeout:duration}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        csmsEnqueuesReserveNow,
	})

	registry.Register(api.Keyword{
		Pattern:     "station {station:string} responds with ReserveNow.conf status {status:string}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        stationRespondsWithReserveNow,
	})

	registry.Register(api.Keyword{
		Pattern:     "the CSMS accepts the response without error within {timeout:duration}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        csmsAcceptsReserveResponse,
	})
}

func registerStopTransactionKeywords() {
	registry.Register(api.Keyword{
		Pattern: "station {station:string} stops transaction {transactionId:int}" +
			" with meterStop {meterStop:int} and reason {reason:string}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        sendStopTransaction,
	})

	registry.Register(api.Keyword{
		Pattern:     "the CSMS accepts StopTransaction within {timeout:duration}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        csmsAcceptsStopTransaction,
	})
}

func registerMeterValuesKeywords() {
	registry.Register(api.Keyword{
		Pattern: "station {station:string} sends MeterValues for connector" +
			" {connectorId:int} with sampled value {value:string}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        sendMeterValues,
	})

	registry.Register(api.Keyword{
		Pattern:     "the CSMS acknowledges MeterValues within {timeout:duration}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        csmsAcknowledgesMeterValues,
	})
}

func registerRemoteStartKeywords() {
	registry.Register(api.Keyword{
		Pattern: "the CSMS sends RemoteStartTransaction with connectorId {connectorId:int}" +
			" and idTag {idTag:string} to station {station:string} within {timeout:duration}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        csmsEnqueuesRemoteStart,
	})

	registry.Register(api.Keyword{
		Pattern:     "station {station:string} responds to RemoteStartTransaction with status {status:string}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        stationRespondsToRemoteStart,
	})
}

func registerRemoteStopKeywords() {
	registry.Register(api.Keyword{
		Pattern: "the CSMS sends RemoteStopTransaction with transactionId {transactionId:int}" +
			" to station {station:string} within {timeout:duration}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        csmsEnqueuesRemoteStop,
	})

	registry.Register(api.Keyword{
		Pattern:     "station {station:string} responds to RemoteStopTransaction with status {status:string}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        stationRespondsToRemoteStop,
	})
}

func registerResetKeywords() {
	registry.Register(api.Keyword{
		Pattern: "the CSMS sends Reset with type {resetType:string}" +
			" to station {station:string} within {timeout:duration}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        csmsEnqueuesReset,
	})

	registry.Register(api.Keyword{
		Pattern:     "station {station:string} responds to Reset with status {status:string}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        stationRespondsToReset,
	})
}

func registerUnlockKeywords() {
	registry.Register(api.Keyword{
		Pattern: "the CSMS sends UnlockConnector with connectorId {connectorId:int}" +
			" to station {station:string} within {timeout:duration}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        csmsEnqueuesUnlockConnector,
	})

	registry.Register(api.Keyword{
		Pattern:     "station {station:string} responds to UnlockConnector with status {status:string}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        stationRespondsToUnlockConnector,
	})
}

func registerAvailabilityKeywords() {
	registry.Register(api.Keyword{
		Pattern: "the CSMS sends ChangeAvailability with connectorId {connectorId:int}" +
			" and type {availType:string} to station {station:string} within {timeout:duration}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        csmsEnqueuesChangeAvailability,
	})

	registry.Register(api.Keyword{
		Pattern:     "station {station:string} responds to ChangeAvailability with status {status:string}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        stationRespondsToChangeAvailability,
	})
}

func registerGetConfigurationKeywords() {
	registry.Register(api.Keyword{
		Pattern:     "the CSMS sends GetConfiguration to station {station:string} within {timeout:duration}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        csmsEnqueuesGetConfiguration,
	})

	registry.Register(api.Keyword{
		Pattern: "station {station:string} responds to GetConfiguration" +
			" with {count:int} configuration keys",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        stationRespondsWithGetConfiguration,
	})
}

func registerChangeConfigurationKeywords() {
	registry.Register(api.Keyword{
		Pattern: "the CSMS sends ChangeConfiguration with key {key:string}" +
			" and value {value:string} to station {station:string} within {timeout:duration}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        csmsEnqueuesChangeConfiguration,
	})

	registry.Register(api.Keyword{
		Pattern:     "station {station:string} responds to ChangeConfiguration with status {status:string}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        stationRespondsToChangeConfiguration,
	})
}

func registerClearCacheKeywords() {
	registry.Register(api.Keyword{
		Pattern:     "the CSMS sends ClearCache to station {station:string} within {timeout:duration}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        csmsEnqueuesClearCache,
	})

	registry.Register(api.Keyword{
		Pattern:     "station {station:string} responds to ClearCache with status {status:string}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        stationRespondsToClearCache,
	})
}

func registerCancelReservationKeywords() {
	registry.Register(api.Keyword{
		Pattern: "the CSMS sends CancelReservation with reservationId {reservationId:int}" +
			" to station {station:string} within {timeout:duration}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        csmsEnqueuesCancelReservation,
	})

	registry.Register(api.Keyword{
		Pattern:     "station {station:string} responds to CancelReservation with status {status:string}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        stationRespondsToCancelReservation,
	})
}
