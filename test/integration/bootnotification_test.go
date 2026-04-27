//go:build reference

package integration_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/evcoreco/octane/pkg/transport"
	"github.com/evcoreco/octane/pkg/wire"
)

// TestBootNotificationHandshake asserts that a BootNotification CALL sent to
// CitrineOS yields a CALLRESULT with status "Accepted" and the matching
// UniqueID, confirming end-to-end OCPP 1.6 framing over WebSocket.
func TestBootNotificationHandshake(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sta, err := transport.Dial(
		ctx,
		"ws://localhost:9210/CP001",
		transport.DialOptions{
			Subprotocols: []string{"ocpp1.6"},
		},
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer sta.Close()

	// Send BootNotification CALL
	// [2, "bn-001", "BootNotification", {"chargePointModel":"TestCP","chargePointVendor":"OCTANE"}]
	callFrame := []any{float64(2), "bn-001", "BootNotification", map[string]any{
		"chargePointModel":  "TestCP",
		"chargePointVendor": "OCTANE",
	}}
	if err = sta.Send(ctx, callFrame); err != nil {
		t.Fatalf("send BootNotification: %v", err)
	}

	// Receive CALLRESULT
	inbound, err := sta.Expect(ctx)
	if err != nil {
		t.Fatalf("expect response: %v", err)
	}

	result, err := wire.ParseResult(inbound)
	if err != nil {
		t.Fatalf("parse result: %v", err)
	}

	if result.UniqueID != "bn-001" {
		t.Errorf("UniqueID = %q, want %q", result.UniqueID, "bn-001")
	}

	var payload struct {
		Status string `json:"status"`
	}
	if err = json.Unmarshal(result.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if payload.Status != "Accepted" {
		t.Errorf("status = %q, want Accepted", payload.Status)
	}
}
