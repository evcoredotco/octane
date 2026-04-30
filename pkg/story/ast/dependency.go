package ast

// Scope controls how many times a prerequisite story runs relative to its
// dependent story.
type Scope int

const (
	// ScopePerStation runs the prerequisite once per station handle.
	// This is the default when no scope is specified in the Depends block.
	ScopePerStation Scope = iota + 1
	// ScopePerRun runs the prerequisite once per octane run invocation,
	// regardless of station count.
	ScopePerRun
	// ScopeGlobal runs the prerequisite once across the cache validity
	// window (see CacheTTL).
	ScopeGlobal
)

// String returns the canonical string representation of a Scope value.
func (s Scope) String() string {
	switch s {
	case ScopePerStation:
		return "per-station"
	case ScopePerRun:
		return "per-run"
	case ScopeGlobal:
		return "global"
	default:
		return "unknown"
	}
}

// Dependency declares a prerequisite story that must run (and pass) before
// the current story executes. Declared in the Depends: Meta block.
type Dependency struct {
	// ID is the stable snake_case identifier of the prerequisite story.
	// It must match the ID Meta key of the target story exactly.
	ID string `json:"id"`
	// Scope controls how many times the prerequisite is executed relative
	// to the dependent story. Defaults to ScopePerStation when absent.
	Scope Scope `json:"scope"`
	// Position is the source location of this dependency entry in the
	// .story file, used for diagnostic messages.
	Position Position `json:"position"`
}
