package integration_test

import (
	"os"
	"testing"

	"github.com/evcoreco/octane/pkg/keywords/lifecycle"
	"github.com/evcoreco/octane/pkg/keywords/primitive"
)

// TestMain registers all keyword packages before running any tests.
// Without this, keywords like "wait 0s" (primitive) would not be found
// by the runner's keyword resolver.
func TestMain(m *testing.M) {
	primitive.Register()
	lifecycle.Register()
	os.Exit(m.Run())
}
