package ocpp16

import "fmt"

// pendingInfo holds the outbound request context stashed by "send" keywords
// so that the subsequent "response" keyword can correlate the CALLRESULT
// without needing the station handle or message ID in the step text.
//
// The connector fields are only set by StatusNotification send keywords
// so that the acknowledgment keyword can record the connector's new state.
type pendingInfo struct {
	station         string
	msgID           string
	action          string
	connectorID     int
	connectorStatus string
}

// reserveWaiting holds the station handle and ReserveNow CALL unique ID
// stashed by the ReserveNow.conf send keyword so that
// "the CSMS accepts the response without error" knows which station to
// call Expect on and which unique ID to watch for CALLERROR.
type reserveWaiting struct {
	station  string
	uniqueID string
}

const (
	// pendingKey is the stash key holding *pendingInfo from the most
	// recent "send" keyword. Popped by the matching "response" keyword.
	pendingKey = "ocpp16:pending"

	// lastPayloadKey is the stash key holding map[string]any from the
	// most recent CALLRESULT payload. Set by "response" keywords; read
	// (peek) by assertion keywords that inspect specific fields.
	lastPayloadKey = "ocpp16:last_payload"

	// reserveWaitingKey is the stash key holding *reserveWaiting set
	// after the station sends its ReserveNow.conf CALLRESULT. Popped
	// by "the CSMS accepts the response without error".
	reserveWaitingKey = "ocpp16:reserve_waiting"
)

// registeredKey returns the per-station stash key that records whether
// a successful BootNotification.conf with status "Accepted" has been
// received during this scenario execution.
func registeredKey(station string) string {
	return "ocpp16:registered:" + station
}

// connectorKey returns the per-station, per-connector stash key that
// holds the last status string reported by StatusNotification.
func connectorKey(station string, connectorID int) string {
	return fmt.Sprintf("ocpp16:connector:%s:%d", station, connectorID)
}

// msgCounterKey returns the per-station stash key for the incrementing
// message ID counter used by nextMsgID.
func msgCounterKey(station string) string {
	return "ocpp16:msg_counter:" + station
}

// reserveCallIDKey returns the per-station stash key holding the unique
// ID of the ReserveNow CALL received from the CSMS. Set by the
// "the CSMS sends ReserveNow" keyword; consumed by
// "station responds with ReserveNow.conf".
func reserveCallIDKey(station string) string {
	return "ocpp16:reserve_call_id:" + station
}
