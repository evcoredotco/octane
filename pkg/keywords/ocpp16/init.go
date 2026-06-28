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
		Func:        operatorProvisionedIdTag,
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
