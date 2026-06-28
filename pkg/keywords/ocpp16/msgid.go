package ocpp16

import (
	"fmt"
	"strings"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// nextMsgID generates the next unique message identifier for a CALL
// originating from station for action. It uses a per-station counter
// stored in the stash so that IDs are deterministic and unique within
// a scenario — no math/rand or time.Now are used (constitution principle IV).
//
// The generated ID has the form "octane-{action}-{n}" where n begins
// at 1 and increments with each call for the same station.
func nextMsgID(state api.State, station, action string) string {
	key := msgCounterKey(station)
	anyVal, exists := state.Pop(key)
	n := noMessageCounter

	if exists {
		var ok bool

		n, ok = anyVal.(int)
		if !ok {
			n = noMessageCounter
		}
	}

	n++
	state.Stash(key, n)

	return fmt.Sprintf("octane-%s-%d", strings.ToLower(action), n)
}
