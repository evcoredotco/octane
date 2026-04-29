// Package registry provides the global keyword registry for OCTANE's
// story DSL.
//
// Keywords self-register at package init() time via [Register].
// The registry is protected by a [sync.RWMutex] so that multiple
// packages may call [Register] from their init() functions without
// data races, and [All] may be called concurrently from test
// goroutines.
//
// Registration panics immediately on a collision: if two keywords
// share the same (Layer, OCPPVersion, Pattern) triple, the second
// [Register] call panics with a message that names both the new and
// existing registrant by their call sites (file and line number).
//
// [All] returns a fresh, stably-sorted copy of every registered
// keyword in (Layer ascending, OCPPVersion ascending, Pattern
// lexicographic) order, satisfying constitution principle IV
// (determinism).
package registry

import (
	"fmt"
	"runtime"
	"sort"
	"sync"

	"github.com/evcoreco/octane/pkg/keywords/api"
)

// entry pairs a keyword with the formatted call site at which it
// was registered. The call site is used to produce collision panic
// messages that name both registrants.
type entry struct {
	// keyword is the registered keyword.
	keyword api.Keyword

	// caller is the formatted "file:line" string of the Register
	// call that added this keyword. It is captured via
	// runtime.Callers at registration time.
	caller string
}

// registryKey is the uniqueness key for a keyword registration.
// Two keywords with the same (Layer, OCPPVersion, Pattern) triple
// are a collision and the second registration panics.
type registryKey struct {
	// layer is the keyword's layer (primitive or domain).
	layer api.Layer

	// version is the keyword's OCPP version scope.
	version api.OCPPVersion

	// pattern is the keyword's step-matching pattern string.
	pattern string
}

// global is the package-level registry state. It is never replaced;
// only its fields are mutated under mu.
var global = struct { //nolint:exhaustruct // zero value is correct initial state
	mu      sync.RWMutex
	entries []entry
	index   map[registryKey]string // maps key → formatted caller
}{
	index: make(map[registryKey]string),
}

// Register adds keyword to the global keyword registry. It panics if a
// keyword with the same (Layer, OCPPVersion, Pattern) triple has
// already been registered. The panic message includes the formatted
// call sites of both the existing and the new registrant so that
// keyword authors can locate the conflict quickly.
//
// Register is goroutine-safe. It is intended to be called from
// package init() functions; calling it after program startup is
// permitted but unusual.
func Register(keyword api.Keyword) {
	caller := callerLocation(
		2,
	)

	key := registryKey{
		layer:   keyword.Layer,
		version: keyword.OCPPVersion,
		pattern: keyword.Pattern,
	}

	global.mu.Lock()
	defer global.mu.Unlock()

	if existing, dup := global.index[key]; dup {
		panic(fmt.Sprintf(
			"registry: keyword collision on pattern %q "+
				"(layer=%s, ocpp=%s): "+
				"existing registrant at %s, "+
				"new registrant at %s",
			keyword.Pattern,
			keyword.Layer,
			keyword.OCPPVersion,
			existing,
			caller,
		))
	}

	global.index[key] = caller
	global.entries = append(global.entries, entry{
		keyword: keyword,
		caller:  caller,
	})
}

// All returns a fresh slice containing every registered keyword,
// sorted stably by (Layer ascending, OCPPVersion ascending, Pattern
// lexicographic). Callers may modify the returned slice without
// affecting the registry.
//
// The sort order satisfies constitution principle IV (determinism):
// repeated calls to All with the same set of registered keywords
// always return the same ordering regardless of registration order.
func All() []api.Keyword {
	global.mu.RLock()
	defer global.mu.RUnlock()

	result := make([]api.Keyword, len(global.entries))
	for idx, ent := range global.entries {
		result[idx] = ent.keyword
	}

	sort.SliceStable(result, func(left, right int) bool {
		lkw := result[left]
		rkw := result[right]

		if lkw.Layer != rkw.Layer {
			return lkw.Layer < rkw.Layer
		}

		if lkw.OCPPVersion != rkw.OCPPVersion {
			return lkw.OCPPVersion < rkw.OCPPVersion
		}

		return lkw.Pattern < rkw.Pattern
	})

	return result
}

// reset clears the global registry. It is provided for test
// teardown only; production code must never call it. Tests that
// call reset must restore the registry to a clean state before
// other test functions run.
func reset() {
	global.mu.Lock()
	defer global.mu.Unlock()

	global.entries = global.entries[:0]
	global.index = make(map[registryKey]string)
}

// callerLocation returns a "file:line" string for the call frame
// at the given skip depth. skip=1 identifies the direct caller of
// callerLocation; skip=2 identifies the caller's caller, and so on.
// If the frame cannot be resolved the function returns "<unknown>".
func callerLocation(skip int) string {
	var pcs [1]uintptr

	count := runtime.Callers(skip+1, pcs[:])
	if count == 0 {
		return "<unknown>"
	}

	frames := runtime.CallersFrames(pcs[:count])

	frame, _ := frames.Next()
	if frame.File == "" {
		return "<unknown>"
	}

	return fmt.Sprintf("%s:%d", frame.File, frame.Line)
}
