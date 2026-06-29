package ocpp16_test

import (
	"os"
	"testing"

	"github.com/evcoreco/octane/pkg/keywords/lifecycle"
	"github.com/evcoreco/octane/pkg/keywords/ocpp16"
	"github.com/evcoreco/octane/pkg/keywords/primitive"
)

// TestMain registers all keyword packages once before any test runs.
func TestMain(m *testing.M) {
	primitive.Register()
	lifecycle.Register()
	ocpp16.Register()
	os.Exit(m.Run())
}
