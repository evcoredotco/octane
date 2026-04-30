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

// dependsLineResult carries the updated state after processing one
// sub-indented line in a Depends block.
type dependsLineResult struct {
	cur        *dependsEntry
	deps       []ast.Dependency
	entryIndex int
	done       bool
}

// startBulletResult carries the updated state after starting a new Depends
// bullet.
type startBulletResult struct {
	cur        *dependsEntry
	deps       []ast.Dependency
	entryIndex int
}

// colonValueResult carries the two tokens produced by consumeColonValue.
type colonValueResult struct {
	colonTok lex.Token
	valTok   lex.Token
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
		res, err := p.parseDependsLine(cur, entryIndex, deps)
		if err != nil {
			return nil, err
		}

		if res.done {
			break
		}

		cur = res.cur
		entryIndex = res.entryIndex
		deps = res.deps
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

// parseDependsLine processes one sub-indented line inside a Depends block.
// The done field in the result is true when the line is not a MetaKey and
// the block ends. On a new bullet it flushes the current entry, increments
// the index, and creates a fresh dependsEntry.
func (p *parser) parseDependsLine(
	cur *dependsEntry,
	entryIndex int,
	deps []ast.Dependency,
) (dependsLineResult, error) {
	_ = p.lex.Next() // consume the sub-indent token

	keyTok := p.lex.Peek()
	if keyTok.Kind != lex.TokenMetaKey {
		return dependsLineResult{
			cur: cur, entryIndex: entryIndex, deps: deps, done: true,
		}, nil
	}

	_ = p.lex.Next() // consume key

	cv, colonErr := p.consumeColonValue(keyTok, entryIndex)
	if colonErr != nil {
		return dependsLineResult{
			cur: cur, entryIndex: entryIndex, deps: deps, done: false,
		}, colonErr
	}

	keyLit := strings.TrimSpace(keyTok.Literal)
	valLit := strings.TrimSpace(cv.valTok.Literal)

	if !strings.HasPrefix(keyLit, "-") {
		applyErr := applyNonBulletKey(
			cur, keyLit, valLit, cv.valTok, p.file, entryIndex,
		)

		return dependsLineResult{
			cur: cur, entryIndex: entryIndex, deps: deps, done: false,
		}, applyErr
	}

	br, err := p.startNewBullet(cur, entryIndex, deps, keyTok, valLit, cv.valTok)

	return dependsLineResult{
		cur: br.cur, entryIndex: br.entryIndex, deps: br.deps, done: false,
	}, err
}

// applyNonBulletKey applies a sub-key line that does not start a new bullet
// (i.e. keyLit does not begin with "-"). It is a no-op when cur is nil.
func applyNonBulletKey(
	cur *dependsEntry,
	keyLit string,
	valLit string,
	valTok lex.Token,
	file string,
	entryIndex int,
) error {
	if cur == nil {
		return nil
	}

	return applySubKey(cur, keyLit, valLit, valTok, file, entryIndex)
}

// startNewBullet flushes the current entry, increments the index, creates a
// fresh dependsEntry, and applies the sub-key from the new bullet line.
func (p *parser) startNewBullet(
	cur *dependsEntry,
	entryIndex int,
	deps []ast.Dependency,
	keyTok lex.Token,
	valLit string,
	valTok lex.Token,
) (startBulletResult, error) {
	flushed, ok, flushErr := flushEntry(p.file, cur, entryIndex)
	if flushErr != nil {
		return startBulletResult{
			cur: cur, entryIndex: entryIndex, deps: deps,
		}, flushErr
	}

	if ok {
		deps = append(deps, flushed)
	}

	entryIndex++

	newCur := &dependsEntry{
		id:    emptyStr,
		scope: ast.ScopePerStation,
		pos: ast.Position{
			Line:   keyTok.Line,
			Column: keyTok.Column,
		},
		idSet:    false,
		scopeSet: false,
	}

	keyLit := strings.TrimSpace(keyTok.Literal)
	subKey := strings.TrimSpace(strings.TrimPrefix(keyLit, "-"))

	applyErr := applySubKey(newCur, subKey, valLit, valTok, p.file, entryIndex)
	if applyErr != nil {
		return startBulletResult{
			cur: newCur, entryIndex: entryIndex, deps: deps,
		}, applyErr
	}

	return startBulletResult{cur: newCur, entryIndex: entryIndex, deps: deps}, nil
}

// consumeColonValue consumes a TokenColon and then a TokenValue from the
// stream and returns a colonValueResult. It returns a typed error on any
// mismatch.
func (p *parser) consumeColonValue(
	keyTok lex.Token,
	entryIndex int,
) (colonValueResult, error) {
	illegalZero := lex.Token{
		Kind:    lex.TokenIllegal,
		Literal: emptyStr,
		Line:    0,
		Column:  0,
	}

	colonTok := p.lex.Peek()
	if colonTok.Kind != lex.TokenColon {
		illegalColon := lex.Token{
			Kind:    lex.TokenIllegal,
			Literal: emptyStr,
			Line:    colonTok.Line,
			Column:  colonTok.Column,
		}

		return colonValueResult{colonTok: illegalColon, valTok: illegalZero},
			&diag.MalformedDependsError{
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
		illegalVal := lex.Token{
			Kind:    lex.TokenIllegal,
			Literal: emptyStr,
			Line:    valTok.Line,
			Column:  valTok.Column,
		}

		return colonValueResult{colonTok: colonTok, valTok: illegalVal},
			&diag.MalformedDependsError{
				File:       p.file,
				Line:       colonTok.Line,
				Column:     colonTok.Column,
				EntryIndex: max(entryIndex, 0),
				Reason:     "expected value after colon",
				Suggestion: "use the form '  - id: <story-id>'",
			}
	}

	return colonValueResult{colonTok: colonTok, valTok: valTok}, nil
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
