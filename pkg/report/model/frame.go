package model

// Frame holds a single OCPP-J wire frame.
type Frame struct {
	// Raw is the raw OCPP-J JSON bytes for this frame.
	Raw []byte
}
