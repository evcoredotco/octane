package registry

import "testing"

// serialTestEnv is set by resetGlobalRegistry. Its only purpose is to make
// t.Setenv refuse to run in a parallel test.
const serialTestEnv = "OCTANE_REGISTRY_SERIAL_TEST"

// resetGlobalRegistry clears the global registry for a test.
//
// The registry is shared by every test in the package, so a test that
// resets it must not run in parallel with any other test. t.Setenv panics
// when the test, or one of its parents, calls t.Parallel, so a parallel
// caller fails at once instead of intermittently.
func resetGlobalRegistry(t *testing.T) {
	t.Helper()
	t.Setenv(serialTestEnv, "1")

	reset()
}
