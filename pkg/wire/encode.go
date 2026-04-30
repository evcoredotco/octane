package wire

import (
	"encoding/json"
	"fmt"
)

// Encode serializes an OCPP-J frame to canonical JSON.
//
// The frame is a []any whose first element is the numeric message type code
// (2, 3, or 4). Go's encoding/json marshals map keys in sorted order, so
// all nested objects in the output are key-sorted, producing identical wire
// bytes for identical logical frames.
//
// Returns an error only when the frame contains a value that encoding/json
// cannot represent (e.g. a channel or a function). For well-formed OCPP-J
// frames this will never occur.
func Encode(frame []any) ([]byte, error) {
	data, err := json.Marshal(frame)
	if err != nil {
		return nil, fmt.Errorf("wire: encode frame: %w", err)
	}

	return data, nil
}
