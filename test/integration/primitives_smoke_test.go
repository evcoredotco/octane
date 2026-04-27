//go:build reference

// Package integration_test contains integration tests that run against the
// pinned CitrineOS instance (see test/reference/citrineos.version).
//
// Task: T-004-31
// AC6: The smoke story primitives_only.story executes end-to-end against the
// pinned CitrineOS using only primitive keywords and the received frame is a
// CALLRESULT (type 3) for the BootNotification.
package integration_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/octane-project/octane/pkg/keywords/api"
	"github.com/octane-project/octane/pkg/keywords/api/mock"
	// Blank import registers all primitive keywords at init() time.
	_ "github.com/octane-project/octane/pkg/keywords/primitive"
	"github.com/octane-project/octane/pkg/keywords/registry"
	"github.com/octane-project/octane/pkg/wire"
)

// ── Named constants ───────────────────────────────────────────────────────────

const (
	// citrineOSURL is the WebSocket endpoint for the pinned CitrineOS instance.
	// The path encodes the station identity (CP001) per the CitrineOS routing
	// convention used across the integration test suite.
	citrineOSURL = "ws://localhost:9210/CP001"

	// stationHandle is the runtime handle name used to refer to the station
	// throughout the smoke scenario.
	stationHandle = "CP001"

	// subprotocolOCPP16Smoke is the OCPP 1.6 subprotocol identifier offered
	// during the WebSocket upgrade handshake.
	subprotocolOCPP16Smoke = "ocpp1.6"

	// messageIDSmoke is the OCPP-J unique identifier for the BootNotification
	// CALL sent in the smoke scenario.
	messageIDSmoke = "msg-001"

	// messageTypeCALLRESULT is the OCPP-J message type code for a CALLRESULT
	// frame (type 3 per OCPP-J §3.4).  Used to validate the received frame.
	messageTypeCALLRESULT = float64(3)

	// smokeTimeout is the wall-clock budget for the entire smoke scenario
	// (open + send + expect + close).
	smokeTimeout = 30 * time.Second

	// expectTimeout is the duration passed to the "expect any frame" keyword
	// step; it must be less than smokeTimeout.
	expectTimeout = 20 * time.Second
)

// resolveStep finds the registered primitive keyword whose pattern matches
// stepText and returns its Func.  The test fails immediately if no match is
// found — a missing primitive registration is a production bug.
func resolveStep(t *testing.T, stepText string) api.Func {
	t.Helper()

	match, err := registry.Resolve(stepText, api.OCPP16)
	if err != nil {
		t.Fatalf("registry.Resolve(%q): %v", stepText, err)
	}

	return match.Keyword.Func
}

// TestPrimitivesSmoke_BootNotificationCALLRESULT asserts that the full
// primitives_only.story smoke scenario executes end-to-end against CitrineOS:
//  1. Opens a WebSocket to ws://localhost:9210/CP001 with subprotocol ocpp1.6.
//  2. Asserts the connection is open.
//  3. Sends a BootNotification CALL frame (type 2).
//  4. Expects any inbound frame within 20 s.
//  5. Validates that the inbound frame is a CALLRESULT (type 3) for msg-001.
//  6. Closes the station and asserts the connection is closed.
func TestPrimitivesSmoke_BootNotificationCALLRESULT(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), smokeTimeout)
	defer cancel()

	state := mock.NewMockState()
	state.SetNow(time.Now())

	// ── Step 1: open WebSocket with subprotocol ───────────────────────────────

	// Invariant: the open keyword registers a station handle in state.
	openFunc := resolveStep(
		t,
		"open a WebSocket to {url:string} as station {station:string}"+
			" with subprotocol {subprotocol:string}",
	)

	openArgs := api.NewArgs(map[string]any{
		"url":         citrineOSURL,
		"station":     stationHandle,
		"subprotocol": subprotocolOCPP16Smoke,
	})

	if err := openFunc(ctx, state, openArgs); err != nil {
		t.Fatalf("open WebSocket: %v", err)
	}

	// ── Step 2: assert connection is open ─────────────────────────────────────

	// Invariant: IsOpen returns true immediately after a successful open.
	isOpenFunc := resolveStep(
		t,
		"the connection on station {station:string} is open",
	)

	isOpenArgs := api.NewArgs(map[string]any{"station": stationHandle})

	if err := isOpenFunc(ctx, state, isOpenArgs); err != nil {
		t.Fatalf("connection-open assertion: %v", err)
	}

	// ── Step 3: send BootNotification CALL ────────────────────────────────────

	// Invariant: sendRawFrame forwards the OCPP-J array to the CSMS wire.
	sendFunc := resolveStep(
		t,
		"send raw frame {frame:any} on station {station:string}",
	)

	bootNotificationCall := []any{
		float64(2),
		messageIDSmoke,
		"BootNotification",
		map[string]any{
			"reason": "PowerUp",
			"chargingStation": map[string]any{
				"model":      "ACME",
				"vendorName": "Test",
			},
		},
	}

	sendArgs := api.NewArgs(map[string]any{
		"frame":   bootNotificationCall,
		"station": stationHandle,
	})

	if err := sendFunc(ctx, state, sendArgs); err != nil {
		t.Fatalf("send raw frame: %v", err)
	}

	// ── Step 4: expect any inbound frame within the timeout ───────────────────

	// The expect keyword stashes the received frame in the station's scratch
	// space via the mock station's Expect call.  We retrieve it directly
	// from the live station handle to validate its shape.
	sta, stErr := state.Station(stationHandle)
	if stErr != nil {
		t.Fatalf("state.Station(%q): %v", stationHandle, stErr)
	}

	expectCtx, expectCancel := context.WithTimeout(ctx, expectTimeout)
	defer expectCancel()

	// Invariant: a CALLRESULT frame arrives within the timeout when CitrineOS
	// is reachable and the BootNotification payload is well-formed.
	inbound, expectErr := sta.Expect(expectCtx)
	if expectErr != nil {
		t.Fatalf("expect frame: %v", expectErr)
	}

	// ── Step 5: validate the received frame is a CALLRESULT for msg-001 ───────

	// Invariant: the first element of the frame is float64(3) (CALLRESULT).
	if len(inbound) == 0 {
		t.Fatal("received empty frame")
	}

	msgType, ok := inbound[0].(float64)
	if !ok {
		t.Fatalf("frame[0] type: want float64, got %T", inbound[0])
	}

	if msgType != messageTypeCALLRESULT {
		t.Errorf(
			"frame[0] (messageType): want %.0f (CALLRESULT), got %.0f",
			messageTypeCALLRESULT,
			msgType,
		)
	}

	// Invariant: ParseResult succeeds and UniqueID matches the sent CALL.
	result, parseErr := wire.ParseResult(inbound)
	if parseErr != nil {
		t.Fatalf("wire.ParseResult: %v", parseErr)
	}

	if result.UniqueID != messageIDSmoke {
		t.Errorf(
			"CALLRESULT UniqueID: want %q, got %q",
			messageIDSmoke,
			result.UniqueID,
		)
	}

	// Invariant: the payload carries an "Accepted" status from CitrineOS.
	var payload struct {
		Status string `json:"status"`
	}

	if err := json.Unmarshal(result.Payload, &payload); err != nil {
		t.Fatalf("unmarshal BootNotification response payload: %v", err)
	}

	if payload.Status != "Accepted" {
		t.Errorf(
			"BootNotification response status: want %q, got %q",
			"Accepted",
			payload.Status,
		)
	}

	// ── Step 6: close the station and assert it is closed ────────────────────

	// Invariant: close keyword calls Station.Close; IsOpen returns false.
	closeFunc := resolveStep(t, "close station {station:string}")

	closeArgs := api.NewArgs(map[string]any{"station": stationHandle})

	if err := closeFunc(ctx, state, closeArgs); err != nil {
		t.Fatalf("close station: %v", err)
	}

	isClosedFunc := resolveStep(
		t,
		"the connection on station {station:string} is closed",
	)

	if err := isClosedFunc(ctx, state, isOpenArgs); err != nil {
		t.Fatalf("connection-closed assertion: %v", err)
	}

	_ = sta
}
