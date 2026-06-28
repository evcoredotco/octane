package lifecycle

import (
	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/registry"
)

// Register adds every lifecycle keyword to the global registry.
// Call this once from the CLI entry point (Execute) before running any stories.
// It uses the same explicit-registration pattern as pkg/keywords/primitive.
func Register() {
	registerConnectionLifecycleKeywords()
}

func registerConnectionLifecycleKeywords() {
	registry.Register(api.Keyword{
		Pattern:     "station {station:string} connects to the CSMS",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        connectToCSMS,
	})

	registry.Register(api.Keyword{
		Pattern: "the OCPP-J handshake completes within" +
			" {timeout:duration}",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        handshakeCompletes,
	})

	registry.Register(api.Keyword{
		Pattern:     "station {station:string} is in the connected state",
		Layer:       api.LayerDomain,
		OCPPVersion: api.OCPP16,
		Func:        assertConnectedState,
	})
}
