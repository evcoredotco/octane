// Package serialize provides a test-only JSON serializer for ast.Story values.
// It is internal to pkg/story and must not be imported by production code.
//
// Byte-determinism is guaranteed by construction: encoding/json serializes
// struct fields in declaration order, and the AST was designed with ordered
// slices rather than maps (constitution principle IV), so no custom key
// sorting is required.
package serialize

import (
	"encoding/json"

	"github.com/evcoreco/octane/pkg/story/ast"
)

// Serialize returns a compact, byte-deterministic JSON encoding of story.
// It uses encoding/json which serializes struct fields in declaration order
// (deterministic by construction since structs have no map keys).
// This function is for test use only; it is not part of the public API.
func Serialize(story *ast.Story) ([]byte, error) {
	return json.Marshal(story)
}
