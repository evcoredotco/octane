//go:build ignore

// Package main is a sample stand-alone program that demonstrates how an
// external keyword author writes and unit-tests a keyword function using
// only pkg/keywords/api and pkg/keywords/api/mock.
//
// This file carries the "ignore" build tag so it is never compiled as
// part of "go test ./..." — it is documentation-as-code. Run it directly
// with:
//
//	go run ./pkg/keywords/api/mock/testdata/external/keyword.go
//
// The program exercises mock.State and mock.Station without importing any
// of pkg/runner/, pkg/transport/, or network-layer packages, satisfying
// spec 003 AC8.
//
// Task: T-003-42
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/api/mock"
)

// bootNotificationFrame is a minimal OCPP 1.6 BootNotification CALL frame
// expressed as a decoded Go []any value per ADR 0006.
//
// NOTE (ADR 0020): When github.com/evcoreco/ocpp16types is available, the
// payload map[string]any below should be replaced with the typed struct:
//
//	import ocpp16 "github.com/evcoreco/ocpp16types"
//	payload := ocpp16.BootNotificationRequest{
//	    ChargePointModel:  "ACME-500",
//	    ChargePointVendor: "ACME Corp",
//	}
//
// Until then, this file deliberately constructs raw wire frames because
// it demonstrates the primitive/wire layer only. Real domain keywords
// use typed structs exclusively.
var bootNotificationFrame = []any{
	2,
	"msg-boot-001",
	"BootNotification",
	map[string]any{
		"chargePointModel":  "ACME-500",
		"chargePointVendor": "ACME Corp",
	},
}

// bootNotificationResponseFrame is the corresponding CALLRESULT frame that
// the CSMS echoes back.
var bootNotificationResponseFrame = []any{
	3,
	"msg-boot-001",
	map[string]any{
		"currentTime": "2024-01-01T00:00:00Z",
		"interval":    300,
		"status":      "Accepted",
	},
}

// sendBootNotification is a sample domain-layer keyword that:
//  1. Resolves the station handle from Args.
//  2. Logs the current time via state.Now().
//  3. Sends a BootNotification CALL frame to the station.
//  4. Waits for the CALLRESULT response.
//  5. Returns an error if the response status is not "Accepted".
//
// Notice that the function body never imports pkg/runner, pkg/transport,
// or any network library — it operates entirely through the api.State and
// api.Station interfaces.
func sendBootNotification(
	ctx context.Context,
	state api.State,
	args api.Args,
) error {
	handle := args.Station("station")

	station, err := state.Station(handle)
	if err != nil {
		return fmt.Errorf(
			"sendBootNotification: station %q not available: %w",
			handle,
			err,
		)
	}

	state.Logf(
		"sending BootNotification from %q at %v",
		handle,
		state.Now().Format(time.RFC3339),
	)

	if err = station.Send(ctx, bootNotificationFrame); err != nil {
		return fmt.Errorf(
			"sendBootNotification: send failed: %w",
			err,
		)
	}

	response, err := station.Expect(ctx)
	if err != nil {
		return fmt.Errorf(
			"sendBootNotification: expect failed: %w",
			err,
		)
	}

	return interpretResponse(response)
}

// interpretResponse reads the decoded CALLRESULT frame and returns an
// error when the CSMS rejected the BootNotification.
func interpretResponse(response []any) error {
	const minFrameFields = 3

	if len(response) < minFrameFields {
		return errors.New(
			"sendBootNotification: malformed response frame",
		)
	}

	payload, ok := response[2].(map[string]any)
	if !ok {
		return errors.New(
			"sendBootNotification: response payload is not an object",
		)
	}

	status, _ := payload["status"].(string)
	if status != "Accepted" {
		return fmt.Errorf(
			"sendBootNotification: CSMS rejected BootNotification "+
				"with status %q",
			status,
		)
	}

	return nil
}

// main runs a self-contained demonstration:
//   - builds a mock.State and mock.Station
//   - registers the station under "CP01"
//   - queues the CALLRESULT response
//   - invokes the keyword function
//   - prints the result and logged messages
func main() {
	state := mock.NewMockState()
	station := mock.NewMockStation()

	state.RegisterStation("CP01", station)
	state.SetNow(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))

	station.QueueFrame(bootNotificationResponseFrame)

	args := api.NewArgs(map[string]any{
		"station": "CP01",
	})

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	err := sendBootNotification(ctx, state, args)

	fmt.Println("=== Keyword result ===")

	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("PASS")

	fmt.Println("\n=== State logs ===")

	for _, line := range state.Logs() {
		fmt.Println(" ", line)
	}

	fmt.Println("\n=== Sent frames ===")

	for idx, frame := range station.SentFrames() {
		fmt.Printf("  [%d] %v\n", idx, frame)
	}
}
