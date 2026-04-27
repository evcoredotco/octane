// Package report provides the emitter layer that converts a
// runner.RunResult into operator-facing output formats (JSON, Robot XML).
// Callers obtain an emitter-specific options struct (JSONOptions,
// RobotXMLOptions) and pass it to the corresponding WriteXxx function.
//
// Task: T-007-02.
package report

// JSONOptions controls the JSON emitter behaviour.
type JSONOptions struct {
	// NoTraceOnPass suppresses wire trace data for stories that passed.
	// When true, the emitted trace_present JSON field is false and the
	// trace object is omitted for passing stories.
	NoTraceOnPass bool

	// OctaneVersion is embedded in the report header. Defaults to "dev"
	// when empty.
	OctaneVersion string
}

// RobotXMLOptions controls the Robot XML emitter behaviour.
type RobotXMLOptions struct {
	// SuiteName is the <suite name=""> attribute in the output XML.
	// Defaults to "OCTANE Conformance" when empty.
	SuiteName string
}
