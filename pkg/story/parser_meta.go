// Package story — see parse.go for package documentation.
package story

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/octane-project/octane/pkg/story/ast"
	"github.com/octane-project/octane/pkg/story/diag"
	"github.com/octane-project/octane/pkg/story/lex"
)

// metaEntry is the internal representation of a single parsed meta line.
type metaEntry struct {
	key    string
	value  string
	line   int
	column int
}

// parseMeta implements T-001-20, T-001-21, T-001-23.
//
// It expects the token stream to begin with TokenMeta followed by one or
// more indented meta-entry lines. After all entries are collected it
// validates that required keys are present and that the Spec-Ref / helper
// tag contract is respected.
func (p *parser) parseMeta() (ast.Meta, error) {
	tok := p.lex.Next()
	if tok.Kind != lex.TokenMeta {
		return ast.Meta{
			Name:       "",
			ID:         "",
			SpecRef:    nil,
			Tags:       nil,
			Stations:   0,
			Timeout:    0,
			Parameters: nil,
			CacheTTL:   nil,
			Depends:    nil,
			Position:   ast.Position{Line: tok.Line, Column: tok.Column},
		}, fmt.Errorf(
			"%s:%d:%d: expected Meta section at top of file, got %s",
			p.file, tok.Line, tok.Column, tok.Kind,
		)
	}

	meta := ast.Meta{
		Name:       "",
		ID:         "",
		SpecRef:    nil,
		Tags:       nil,
		Stations:   0,
		Timeout:    0,
		Parameters: nil,
		CacheTTL:   nil,
		Depends:    nil,
		Position:   ast.Position{Line: tok.Line, Column: tok.Column},
	}

	tracker := &metaTracker{
		hasName:       false,
		hasID:         false,
		hasStations:   false,
		hasTags:       false,
		specRefLine:   0,
		specRefColumn: 0,
	}

	for p.lex.Peek().Kind == lex.TokenIndent {
		_ = p.lex.Next() // consume indent

		entry, err := p.parseMetaEntry()
		if err != nil {
			return ast.Meta{
				Name:       "",
				ID:         "",
				SpecRef:    nil,
				Tags:       nil,
				Stations:   0,
				Timeout:    0,
				Parameters: nil,
				CacheTTL:   nil,
				Depends:    nil,
				Position:   ast.Position{Line: 0, Column: 0},
			}, err
		}

		if err = p.applyMetaEntry(&meta, entry, tracker); err != nil {
			return ast.Meta{
				Name:       "",
				ID:         "",
				SpecRef:    nil,
				Tags:       nil,
				Stations:   0,
				Timeout:    0,
				Parameters: nil,
				CacheTTL:   nil,
				Depends:    nil,
				Position:   ast.Position{Line: 0, Column: 0},
			}, err
		}
	}

	if err := validateMetaRequired(p.file, meta, tracker); err != nil {
		return ast.Meta{
			Name:       "",
			ID:         "",
			SpecRef:    nil,
			Tags:       nil,
			Stations:   0,
			Timeout:    0,
			Parameters: nil,
			CacheTTL:   nil,
			Depends:    nil,
			Position:   ast.Position{Line: 0, Column: 0},
		}, err
	}

	return meta, nil
}

// metaTracker records which optional/required keys have been seen so that
// post-loop validation can detect missing keys without using a map.
type metaTracker struct {
	hasName       bool
	hasID         bool
	hasStations   bool
	hasTags       bool
	specRefLine   int
	specRefColumn int
}

// applyMetaEntry populates the correct ast.Meta field based on entry.key.
// It mutates meta in place and records presence of required keys in tracker.
func (p *parser) applyMetaEntry(
	meta *ast.Meta,
	entry metaEntry,
	tracker *metaTracker,
) error {
	switch entry.key {
	case "Name":
		meta.Name = entry.value
		tracker.hasName = true

	case "Id":
		meta.ID = entry.value
		tracker.hasID = true

	case "Spec-Ref":
		v := entry.value
		meta.SpecRef = &v
		tracker.specRefLine = entry.line
		tracker.specRefColumn = entry.column

	case "Tags":
		for _, tag := range splitTrimmed(entry.value, ",") {
			if tag != "" {
				meta.Tags = append(meta.Tags, tag)
				tracker.hasTags = true
			}
		}

	case "Stations":
		count, err := strconv.Atoi(entry.value)
		if err != nil {
			return fmt.Errorf(
				"%s:%d:%d: Stations value %q is not a valid integer",
				p.file, entry.line, entry.column, entry.value,
			)
		}

		meta.Stations = count
		tracker.hasStations = true

	case "Timeout":
		dur, err := time.ParseDuration(entry.value)
		if err != nil {
			return fmt.Errorf(
				"%s:%d:%d: Timeout value %q is not a valid duration: %w",
				p.file, entry.line, entry.column, entry.value, err,
			)
		}

		meta.Timeout = dur

	case "Parameters":
		// T-001-23: parse comma-separated parameter list.
		for _, param := range splitTrimmed(entry.value, ",") {
			if param != "" {
				meta.Parameters = append(meta.Parameters, param)
			}
		}

	case "Cache-TTL":
		dur, err := time.ParseDuration(entry.value)
		if err != nil {
			return fmt.Errorf(
				"%s:%d:%d: Cache-TTL value %q is not a valid duration: %w",
				p.file, entry.line, entry.column, entry.value, err,
			)
		}

		meta.CacheTTL = &dur

	case "Depends":
		// Delegate to parseDepends which reads subsequent indented lines.
		deps, err := p.parseDepends()
		if err != nil {
			return err
		}

		meta.Depends = deps
	default: // Unknown keys are silently tolerated for forward compatibility.
	}

	return nil
}

// validateMetaRequired checks that all required Meta keys are present and
// that the Spec-Ref / helper-tag contract is satisfied. It returns a typed
// diag error on the first violation found.
func validateMetaRequired(
	file string,
	meta ast.Meta,
	tracker *metaTracker,
) error {
	if !tracker.hasName {
		return &diag.ErrMissingKey{
			File:       file,
			Line:       meta.Position.Line,
			Column:     meta.Position.Column,
			Key:        "Name",
			Suggestion: "add 'Name: <test name>' to the Meta section",
		}
	}

	if !tracker.hasID {
		return &diag.ErrMissingKey{
			File:       file,
			Line:       meta.Position.Line,
			Column:     meta.Position.Column,
			Key:        "Id",
			Suggestion: "add 'Id: <snake_case_id>' to the Meta section",
		}
	}

	if !tracker.hasStations {
		return &diag.ErrMissingKey{
			File:       file,
			Line:       meta.Position.Line,
			Column:     meta.Position.Column,
			Key:        "Stations",
			Suggestion: "add 'Stations: 1' to the Meta section",
		}
	}

	if !tracker.hasTags {
		return &diag.ErrMissingKey{
			File:       file,
			Line:       meta.Position.Line,
			Column:     meta.Position.Column,
			Key:        "Tags",
			Suggestion: "add 'Tags: <comma-separated tags>' to the Meta section",
		}
	}

	return validateSpecRef(file, meta, tracker)
}

// validateSpecRef enforces the helper-tag vs Spec-Ref contract:
//   - helper stories must NOT have Spec-Ref
//   - conformance stories MUST have Spec-Ref
func validateSpecRef(file string, meta ast.Meta, tracker *metaTracker) error {
	isHelper := containsTag(meta.Tags, "helper")

	if isHelper && meta.SpecRef != nil {
		return &diag.ErrSpecRefOnHelper{
			File:       file,
			Line:       tracker.specRefLine,
			Column:     tracker.specRefColumn,
			SpecRef:    *meta.SpecRef,
			Suggestion: "remove the Spec-Ref key from helper stories",
		}
	}

	if !isHelper && meta.SpecRef == nil {
		return &diag.ErrMissingSpecRef{
			File:       file,
			Line:       meta.Position.Line,
			Column:     meta.Position.Column,
			Suggestion: "add 'Spec-Ref: <section>' or add the 'helper' tag",
		}
	}

	return nil
}

// parseMetaEntry reads a TokenMetaKey, TokenColon, TokenValue triple from
// the token stream. The TokenIndent has already been consumed by parseMeta.
func (p *parser) parseMetaEntry() (metaEntry, error) {
	keyTok := p.lex.Next()
	if keyTok.Kind != lex.TokenMetaKey {
		return metaEntry{
			key:    "",
			value:  "",
			line:   keyTok.Line,
			column: keyTok.Column,
		}, fmt.Errorf(
			"%s:%d:%d: expected meta key, got %s",
			p.file, keyTok.Line, keyTok.Column, keyTok.Kind,
		)
	}

	colonTok := p.lex.Next()
	if colonTok.Kind != lex.TokenColon {
		return metaEntry{
			key:    "",
			value:  "",
			line:   colonTok.Line,
			column: colonTok.Column,
		}, fmt.Errorf(
			"%s:%d:%d: expected ':' after meta key %q, got %s",
			p.file, colonTok.Line, colonTok.Column, keyTok.Literal, colonTok.Kind,
		)
	}

	valTok := p.lex.Next()
	if valTok.Kind != lex.TokenValue {
		return metaEntry{
			key:    "",
			value:  "",
			line:   valTok.Line,
			column: valTok.Column,
		}, fmt.Errorf(
			"%s:%d:%d: expected value after ':', got %s",
			p.file, valTok.Line, valTok.Column, valTok.Kind,
		)
	}

	return metaEntry{
		key:    strings.TrimSpace(keyTok.Literal),
		value:  strings.TrimSpace(valTok.Literal),
		line:   valTok.Line,
		column: valTok.Column,
	}, nil
}

// splitTrimmed splits s by sep and trims whitespace from each element.
func splitTrimmed(s, sep string) []string {
	parts := strings.Split(s, sep)

	result := make([]string, 0, len(parts))

	for _, part := range parts {
		result = append(result, strings.TrimSpace(part))
	}

	return result
}

// containsTag reports whether tags contains the given tag (case-sensitive).
func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}

	return false
}
