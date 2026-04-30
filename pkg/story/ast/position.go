package ast

// Position records a source location within a .story file.
type Position struct {
	// Line is the 1-based line number.
	Line int `json:"line"`
	// Column is the 1-based column number (byte offset from line start).
	Column int `json:"column"`
}
