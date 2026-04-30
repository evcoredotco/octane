// Package story — see parse.go for package documentation.

package story

import (
	"fmt"
	"strings"

	"github.com/evcoreco/octane/pkg/story/ast"
	"github.com/evcoreco/octane/pkg/story/diag"
	"github.com/evcoreco/octane/pkg/story/lex"
)

// subIndentMinLen is the length threshold distinguishing a Depends
// sub-indent from a top-level indent (4 spaces or one tab).
const subIndentMinLen = 4

// emptyStr is a named empty string required by the add-constant rule.
const emptyStr = ""

// noEntryIdx is the sentinel value meaning no Depends entry has been started.
const noEntryIdx = -1

// isSubIndent returns true when tok is an indent token whose literal
// is longer than the standard four-space top-level indent.
func isSubIndent(tok lex.Token) bool {
	return tok.Kind == lex.TokenIndent && len(tok.Literal) > subIndentMinLen
}

// dependsEntry accumulates the fields for a single Depends bullet while it
// is being parsed. It is converted to ast.Dependency once the bullet is
// complete.
type dependsEntry struct {
	id       string
	scope    ast.Scope
	pos      ast.Position
	idSet    bool
	scopeSet bool
}

// toDependency converts a completed dependsEntry to ast.Dependency, applying
// the default scope when none was specified.
func (de *dependsEntry) toDependency() ast.Dependency {
	scope := de.scope
	if !de.scopeSet {
		scope = ast.ScopePerStation
	}

	return ast.Dependency{
		ID:       de.id,
		Scope:    scope,
		Position: de.pos,
	}
}

// parseDepends parses the indented YAML-style Depends block that follows a
// "Depends:" meta key. The lexer has already consumed the "Depends:" key,
// colon, and (empty) value tokens. This function reads subsequent indented
// lines of the form:
//
//   - id:    <story-id>
//     scope: <per-station|per-run|global>
//
// Each bullet (lines whose MetaKey literal starts with "-") begins a new
// Dependency entry. The scope field defaults to ScopePerStation when absent.
// On the first malformed entry this function returns
// *diag.MalformedDependsError.
func (p *parser) parseDepends() ([]ast.Dependency, error) {
	var (
		deps       []ast.Dependency
		cur        *dependsEntry
		entryIndex int
	)

	entryIndex = noEntryIdx

	for isSubIndent(p.lex.Peek()) {
		_ = p.lex.Next() // consume the sub-indent token

		keyTok := p.lex.Peek()
		if keyTok.Kind != lex.TokenMetaKey {
			break // end of Depends block
		}

		_ = p.lex.Next() // consume key

		colonTok, valTok, err := p.consumeColonValue(keyTok, entryIndex)
		if err != nil {
			return nil, err
		}

		_ = colonTok

		keyLit := strings.TrimSpace(keyTok.Literal)
		valLit := strings.TrimSpace(valTok.Literal)

		if strings.HasPrefix(keyLit, "-") {
			flushed, ok, flushErr := flushEntry(p.file, cur, entryIndex)
			if flushErr != nil {
				return nil, flushErr
			}

			if ok {
				deps = append(deps, flushed)
			}

			entryIndex++

			cur = &dependsEntry{
				id:    emptyStr,
				scope: ast.ScopePerStation,
				pos: ast.Position{
					Line:   keyTok.Line,
					Column: keyTok.Column,
				},
				idSet:    false,
				scopeSet: false,
			}

			subKey := strings.TrimSpace(strings.TrimPrefix(keyLit, "-"))

			applyErr := applySubKey(
				cur,
				subKey,
				valLit,
				valTok,
				p.file,
				entryIndex,
			)
			if applyErr != nil {
				return nil, applyErr
			}

			continue
		}

		if cur == nil {
			continue // ignore lines before the first bullet
		}

		applyErr := applySubKey(cur, keyLit, valLit, valTok, p.file, entryIndex)
		if applyErr != nil {
			return nil, applyErr
		}
	}

	flushed, ok, err := flushEntry(p.file, cur, entryIndex)
	if err != nil {
		return nil, err
	}

	if ok {
		deps = append(deps, flushed)
	}

	return deps, nil
}

// consumeColonValue consumes a TokenColon and then a TokenValue from the
// stream and returns both tokens. It returns a typed error on any mismatch.
// The first return is the colon token, the second is the value token.
func (p *parser) consumeColonValue(
	keyTok lex.Token,
	entryIndex int,
) (lex.Token, lex.Token, error) {
	colonTok := p.lex.Peek()
	if colonTok.Kind != lex.TokenColon {
		return lex.Token{
				Kind:    lex.TokenIllegal,
				Literal: emptyStr,
				Line:    colonTok.Line,
				Column:  colonTok.Column,
			}, lex.Token{
				Kind:    lex.TokenIllegal,
				Literal: emptyStr,
				Line:    0,
				Column:  0,
			}, &diag.MalformedDependsError{
				File:       p.file,
				Line:       keyTok.Line,
				Column:     keyTok.Column,
				EntryIndex: max(entryIndex, 0),
				Reason:     "expected colon after key",
				Suggestion: "use the form '  - id: <story-id>'",
			}
	}

	_ = p.lex.Next() // consume ':'

	valTok := p.lex.Next()
	if valTok.Kind != lex.TokenValue {
		return colonTok, lex.Token{
				Kind:    lex.TokenIllegal,
				Literal: emptyStr,
				Line:    valTok.Line,
				Column:  valTok.Column,
			}, &diag.MalformedDependsError{
				File:       p.file,
				Line:       colonTok.Line,
				Column:     colonTok.Column,
				EntryIndex: max(entryIndex, 0),
				Reason:     "expected value after colon",
				Suggestion: "use the form '  - id: <story-id>'",
			}
	}

	return colonTok, valTok, nil
}

// applySubKey sets the id or scope field on cur based on subKey and val.
// Unknown sub-keys are silently tolerated.
func applySubKey(
	cur *dependsEntry,
	subKey string,
	val string,
	valTok lex.Token,
	file string,
	entryIndex int,
) error {
	switch subKey {
	case "id":
		cur.id = val
		cur.idSet = true

	case "scope":
		scopeVal, err := parseScope(val)
		if err != nil {
			return &diag.MalformedDependsError{
				File:       file,
				Line:       valTok.Line,
				Column:     valTok.Column,
				EntryIndex: entryIndex,
				Reason:     err.Error(),
				Suggestion: "scope must be per-station, per-run, or global",
			}
		}

		cur.scope = scopeVal
		cur.scopeSet = true
	default: // Unknown sub-key; tolerate for forward compatibility.
	}

	return nil
}

// flushEntry validates a completed dependsEntry and converts it to an
// ast.Dependency. If cur is nil (no entry has started) it returns
// (zero, false, nil).
func flushEntry(
	file string,
	cur *dependsEntry,
	entryIndex int,
) (ast.Dependency, bool, error) {
	if cur == nil {
		return ast.Dependency{
			ID:       emptyStr,
			Scope:    0,
			Position: ast.Position{Line: 0, Column: 0},
		}, false, nil
	}

	if !cur.idSet {
		return ast.Dependency{}, false, &diag.MalformedDependsError{
			File:       file,
			Line:       cur.pos.Line,
			Column:     cur.pos.Column,
			EntryIndex: entryIndex,
			Reason:     "missing id field",
			Suggestion: "add '  - id: <story-id>' to the Depends block",
		}
	}

	dep := cur.toDependency()

	return dep, true, nil
}

// parseScope converts a raw string to an ast.Scope value. It returns a
// non-nil error when the value is not one of the three recognised literals.
func parseScope(raw string) (ast.Scope, error) {
	switch raw {
	case "per-station":
		return ast.ScopePerStation, nil
	case "per-run":
		return ast.ScopePerRun, nil
	case "global":
		return ast.ScopeGlobal, nil
	default:
		return 0, fmt.Errorf("unknown scope value %q", raw)
	}
}
