package ocpp16

import "fmt"

const (
	actionBootNotification        = "BootNotification"
	actionCancelReservation       = "CancelReservation"
	actionChangeAvailability      = "ChangeAvailability"
	actionChangeConfiguration     = "ChangeConfiguration"
	actionClearCache              = "ClearCache"
	actionGetConfiguration        = "GetConfiguration"
	actionRemoteStartTransaction  = "RemoteStartTransaction"
	actionRemoteStopTransaction   = "RemoteStopTransaction"
	actionReserveNow              = "ReserveNow"
	actionReset                   = "Reset"
	actionStartTransaction        = "StartTransaction"
	actionStatusNotification      = "StatusNotification"
	actionUnlockConnector         = "UnlockConnector"
	configKeyStartIndex           = 0
	emptyUniqueID                 = ""
	fieldConnectorID              = "connectorId"
	fieldIDTag                    = "idTag"
	fieldStatus                   = "status"
	fieldTimestamp                = "timestamp"
	iso8601SecondFormat           = "2006-01-02T15:04:05Z"
	noMessageCounter              = 0
	noNumericPayloadValue         = 0
	positiveTransactionIDBoundary = 0
	statusAccepted                = "Accepted"
	statusPass                    = "PASS"
)

const stationNotConnectedFormat = "ocpp16: station %q: not connected: %w"

func payloadString(payload map[string]any, field, context string) (string, error) {
	raw, exists := payload[field]
	if !exists {
		return "", fmt.Errorf("ocpp16: %s payload missing %s field", context, field)
	}

	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf(
			"ocpp16: %s payload %s has unexpected type %T (want string)",
			context, field, raw,
		)
	}

	return value, nil
}

func payloadNumber(payload map[string]any, field, context string) (float64, error) {
	raw, exists := payload[field]
	if !exists {
		return noNumericPayloadValue, fmt.Errorf("ocpp16: %s payload missing %s field", context, field)
	}

	value, ok := raw.(float64)
	if !ok {
		return noNumericPayloadValue, fmt.Errorf(
			"ocpp16: %s payload %s has unexpected type %T (want number)",
			context, field, raw,
		)
	}

	return value, nil
}
