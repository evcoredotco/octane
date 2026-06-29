package ocpp16_test

import (
	"testing"
	"time"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
	"github.com/evcoreco/octane/pkg/keywords/registry"
)

// ── named constants ──────────────────────────────────────────────────────────

const (
	sentFrameCount       = 1
	firstFrameIndex      = 0
	frameTypeIndex       = 0
	uniqueIDIndex        = 1
	actionIndex          = 2
	payloadIndex         = 3
	emptyCollectionCount = 0
	testNowYear          = 2024
	testNowMonth         = 1
	testNowDay           = 1
	testNowHour          = 12
	testNowMinute        = 0
	testNowSecond        = 0
	testNowNanosecond    = 0

	lastPayloadKeyTest = "ocpp16:last_payload"

	// stationHandle is the station handle used across ocpp16 tests.
	stationHandle = "CP01"

	// csmsUniqueID is a CSMS-assigned unique ID for CSMS-initiated CALL frames.
	csmsUniqueID = "csms-001"

	// defaultTimeout is the timeout passed to response keywords in tests.
	defaultTimeout = 5 * time.Second

	// statusAccepted is the OCPP status "Accepted".
	statusAccepted = "Accepted"

	// statusRejected is the OCPP status "Rejected".
	statusRejected = "Rejected"

	// statusBlocked is the OCPP idTagInfo status "Blocked".
	statusBlocked = "Blocked"

	// statusAvailable is the OCPP connector status "Available".
	statusAvailable = "Available"

	// statusCharging is the OCPP connector status "Charging".
	statusCharging = "Charging"

	// statusOperative is the OCPP ChangeAvailability type "Operative".
	statusOperative = "Operative"

	// statusUnlocked is the OCPP UnlockConnector status "Unlocked".
	statusUnlocked = "Unlocked"

	// resetTypeSoft is the OCPP Reset type "Soft".
	resetTypeSoft = "Soft"

	// resetTypeHard is the OCPP Reset type "Hard".
	resetTypeHard = "Hard"

	// idTagValue is the idTag used in authorize and transaction tests.
	idTagValue = "AABBCC"

	// currentTimeValid is a valid RFC 3339 timestamp for payload tests.
	currentTimeValid = "2024-01-01T12:00:00Z"

	// currentTimeInvalid is a string that is not a valid RFC 3339 timestamp.
	currentTimeInvalid = "not-a-date"

	// meterStartValue is the meterStart value used in transaction tests.
	meterStartValue = 100

	// meterStopValue is the meterStop value used in stop-transaction tests.
	meterStopValue = 500

	// transactionIDPositive is a positive transaction ID.
	transactionIDPositive = 42

	// transactionIDZero is the zero transaction ID (invalid).
	transactionIDZero = 0

	// connectorIDOne is connector identifier 1.
	connectorIDOne = 1

	// connectorIDTwo is connector identifier 2 (used for mismatch tests).
	connectorIDTwo = 2

	// heartbeatIntervalInRange is a heartbeat interval within a typical range.
	heartbeatIntervalInRange = 300

	// heartbeatIntervalTooLow is a heartbeat interval below the typical minimum.
	heartbeatIntervalTooLow = 5

	// heartbeatIntervalMin is the minimum acceptable heartbeat interval.
	heartbeatIntervalMin = 60

	// heartbeatIntervalMax is the maximum acceptable heartbeat interval.
	heartbeatIntervalMax = 86400

	// sampledValue is the meter value string used in MeterValues tests.
	sampledValue = "123.4"

	// reasonNormal is the reason label for boot/stop transaction tests.
	reasonNormal = "PowerUp"

	// stopReasonNormal is the reason for stop transaction.
	stopReasonNormal = "Local"

	// msgTypeCall is the OCPP-J message type for CALL frames.
	msgTypeCall = float64(2)

	// msgTypeCallResult is the OCPP-J message type for CALLRESULT frames.
	msgTypeCallResult = float64(3)

	// actionBootNotification is the OCPP action name for BootNotification.
	actionBootNotification = "BootNotification"

	// actionStatusNotification is the OCPP action name for StatusNotification.
	actionStatusNotification = "StatusNotification"

	// actionHeartbeat is the OCPP action name for Heartbeat.
	actionHeartbeat = "Heartbeat"

	// actionAuthorize is the OCPP action name for Authorize.
	actionAuthorize = "Authorize"

	// actionStartTransaction is the OCPP action name for StartTransaction.
	actionStartTransaction = "StartTransaction"

	// actionStopTransaction is the OCPP action name for StopTransaction.
	actionStopTransaction = "StopTransaction"

	// actionMeterValues is the OCPP action name for MeterValues.
	actionMeterValues = "MeterValues"

	// actionReset is the OCPP action name for Reset.
	actionReset = "Reset"

	// actionChangeAvailability is the OCPP action name for ChangeAvailability.
	actionChangeAvailability = "ChangeAvailability"

	// actionClearCache is the OCPP action name for ClearCache.
	actionClearCache = "ClearCache"

	// actionUnlockConnector is the OCPP action name for UnlockConnector.
	actionUnlockConnector = "UnlockConnector"

	// actionRemoteStartTransaction is the OCPP action name for RemoteStartTransaction.
	actionRemoteStartTransaction = "RemoteStartTransaction"

	// actionRemoteStopTransaction is the OCPP action name for RemoteStopTransaction.
	actionRemoteStopTransaction = "RemoteStopTransaction"
)

// ── keyword pattern constants ─────────────────────────────────────────────────

const (
	patternSendBoot          = "station {station:string} sends BootNotification with reason {reason:string}"
	patternCSMSResponds      = "the CSMS responds with status {status:string} within {timeout:duration}"
	patternStationRegistered = "station {station:string} is in the registered state"
	patternHBInterval        = "the response includes a heartbeatInterval between {min:int} and {max:int}"
	patternCurrentTime       = "the response includes a currentTime in ISO-8601 format"

	patternSendStatus     = "station {station:string} sends StatusNotification for connector {connectorId:int} with status {status:string}"
	patternCSMSAcksStatus = "the CSMS acknowledges the status within {timeout:duration}"
	patternConnectorState = "connector {connectorId:int} of station {station:string} is in state {state:string}"

	patternSendHeartbeat  = "station {station:string} sends Heartbeat"
	patternCSMSRespondsHB = "the CSMS responds to the Heartbeat within {timeout:duration}"
	patternHBCurrentTime  = "the Heartbeat response includes a currentTime in ISO-8601 format"

	patternSendAuthorize    = "station {station:string} sends Authorize with idTag {idTag:string}"
	patternCSMSRespondsAuth = "the CSMS responds to Authorize with idTagInfo.status {status:string} within {timeout:duration}"

	patternSendStartTx         = "station {station:string} starts a transaction on connector {connectorId:int} with idTag {idTag:string} and meterStart {meterStart:int}"
	patternCSMSRespondsStartTx = "the CSMS responds to StartTransaction with idTagInfo.status {status:string} within {timeout:duration}"
	patternPositiveTxID        = "the StartTransaction response assigns a positive transactionId"

	patternSendStopTx      = "station {station:string} stops transaction {transactionId:int} with meterStop {meterStop:int} and reason {reason:string}"
	patternCSMSAcceptsStop = "the CSMS accepts StopTransaction within {timeout:duration}"

	patternSendMeterValues = "station {station:string} sends MeterValues for connector {connectorId:int} with sampled value {value:string}"
	patternCSMSAcksMeter   = "the CSMS acknowledges MeterValues within {timeout:duration}"

	patternCSMSSendReset        = "the CSMS sends Reset with type {resetType:string} to station {station:string} within {timeout:duration}"
	patternStationRespondsReset = "station {station:string} responds to Reset with status {status:string}"

	patternCSMSSendAvail        = "the CSMS sends ChangeAvailability with connectorId {connectorId:int} and type {availType:string} to station {station:string} within {timeout:duration}"
	patternStationRespondsAvail = "station {station:string} responds to ChangeAvailability with status {status:string}"

	patternCSMSSendClearCache        = "the CSMS sends ClearCache to station {station:string} within {timeout:duration}"
	patternStationRespondsClearCache = "station {station:string} responds to ClearCache with status {status:string}"

	patternCSMSSendUnlock        = "the CSMS sends UnlockConnector with connectorId {connectorId:int} to station {station:string} within {timeout:duration}"
	patternStationRespondsUnlock = "station {station:string} responds to UnlockConnector with status {status:string}"

	patternCSMSSendRemoteStart        = "the CSMS sends RemoteStartTransaction with connectorId {connectorId:int} and idTag {idTag:string} to station {station:string} within {timeout:duration}"
	patternStationRespondsRemoteStart = "station {station:string} responds to RemoteStartTransaction with status {status:string}"

	patternCSMSSendRemoteStop        = "the CSMS sends RemoteStopTransaction with transactionId {transactionId:int} to station {station:string} within {timeout:duration}"
	patternStationRespondsRemoteStop = "station {station:string} responds to RemoteStopTransaction with status {status:string}"

	patternCSMSReachable       = "the CSMS is reachable"
	patternOperatorProvisioned = "the operator has provisioned id token {idTag:string} with status {status:string}"
	patternStationIsRegistered = "station {station:string} is registered to the CSMS"
)

// ── helpers ──────────────────────────────────────────────────────────────────

// resolveFunc returns the keyword Func registered for the given pattern.
// It fails the test immediately if no matching keyword is found.
func resolveFunc(t *testing.T, pattern string) api.Func {
	t.Helper()

	for _, kw := range registry.All() {
		if kw.Pattern == pattern {
			return kw.Func
		}
	}

	t.Fatalf("resolveFunc: no keyword registered for pattern %q", pattern)

	return nil
}

// newState creates a mock.State with stationHandle registered to the given station.
func newState(t *testing.T, station *mock.Station) *mock.State {
	t.Helper()

	state := mock.NewMockState()
	state.RegisterStation(stationHandle, station)
	state.SetNow(time.Date(
		testNowYear,
		testNowMonth,
		testNowDay,
		testNowHour,
		testNowMinute,
		testNowSecond,
		testNowNanosecond,
		time.UTC,
	))

	return state
}

func requireSentCallPayload(
	t *testing.T,
	station *mock.Station,
	operation string,
	action string,
) map[string]any {
	t.Helper()

	frames := station.SentFrames()
	if len(frames) != sentFrameCount {
		t.Fatalf("%s: want 1 sent frame, got %d", operation, len(frames))
	}

	frame := frames[firstFrameIndex]
	if frame[frameTypeIndex] != msgTypeCall {
		t.Errorf(
			"frame[0]: want %v (CALL), got %v",
			msgTypeCall,
			frame[frameTypeIndex],
		)
	}

	if frame[actionIndex] != action {
		t.Errorf("frame[2]: want %q, got %v", action, frame[actionIndex])
	}

	payload, ok := frame[payloadIndex].(map[string]any)
	if !ok {
		t.Fatalf("frame[3]: want map[string]any, got %T", frame[payloadIndex])
	}

	return payload
}

func requireSentCallResultPayload(
	t *testing.T,
	station *mock.Station,
	operation string,
) map[string]any {
	t.Helper()

	frames := station.SentFrames()
	if len(frames) != sentFrameCount {
		t.Fatalf("%s: want 1 sent frame, got %d", operation, len(frames))
	}

	frame := frames[firstFrameIndex]
	if frame[frameTypeIndex] != msgTypeCallResult {
		t.Errorf(
			"frame[0]: want %v (CALLRESULT), got %v",
			msgTypeCallResult,
			frame[frameTypeIndex],
		)
	}

	if frame[uniqueIDIndex] != csmsUniqueID {
		t.Errorf(
			"frame[1]: want %q, got %v",
			csmsUniqueID,
			frame[uniqueIDIndex],
		)
	}

	payload, ok := frame[actionIndex].(map[string]any)
	if !ok {
		t.Fatalf("frame[2]: want map[string]any, got %T", frame[actionIndex])
	}

	return payload
}
