package primitive

import (
	"github.com/evcoreco/octane/pkg/keywords/api"
	"github.com/evcoreco/octane/pkg/keywords/registry"
)

func init() {
	registerConnectionKeywords()
	registerExpectKeywords()
	registerSendKeywords()
	registerStatusKeywords()
	registerWaitKeywords()
}

func registerConnectionKeywords() {
	registry.Register(api.Keyword{
		Pattern: "open a WebSocket to {url:string}" +
			" as station {station:string}",
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        openWebSocket,
	})

	registry.Register(api.Keyword{
		Pattern: "open a WebSocket to {url:string} as station" +
			" {station:string} with subprotocol {subprotocol:string}",
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        openWebSocketWithSubprotocol,
	})

	registry.Register(api.Keyword{
		Pattern:     "close station {station:string}",
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        closeStation,
	})
}

func registerExpectKeywords() {
	registry.Register(api.Keyword{
		Pattern: "expect any frame on station" +
			" {station:string} within {timeout:duration}",
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        expectAnyFrame,
	})

	registry.Register(api.Keyword{
		Pattern: "expect a frame of type {messageType:int} on station" +
			" {station:string} within {timeout:duration}",
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        expectFrameOfType,
	})
}

func registerSendKeywords() {
	registry.Register(api.Keyword{
		Pattern:     "send raw frame {frame:any} on station {station:string}",
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        sendRawFrame,
	})

	registry.Register(api.Keyword{
		Pattern: "send raw bytes {bytes:string}" +
			" on station {station:string}",
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        sendRawBytes,
	})
}

func registerStatusKeywords() {
	registry.Register(api.Keyword{
		Pattern:     "the connection on station {station:string} is open",
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        assertConnectionOpen,
	})

	registry.Register(api.Keyword{
		Pattern:     "the connection on station {station:string} is closed",
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        assertConnectionClosed,
	})
}

func registerWaitKeywords() {
	registry.Register(api.Keyword{
		Pattern:     "wait {duration:duration}",
		Layer:       api.LayerPrimitive,
		OCPPVersion: 0,
		Func:        waitDuration,
	})
}
