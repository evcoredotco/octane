package story

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/evcoreco/octane/pkg/story/ast"
	"github.com/evcoreco/octane/pkg/story/diag"
	"github.com/evcoreco/octane/pkg/story/lex"
)

// fmtPosExpected is the shared positional error format string used across
// parser_meta and parser_steps.
const fmtPosExpected = "%s:%d:%d: %w, got %s"

// emptySliceLen is the zero-length sentinel for make([]T, 0, n) calls.
const emptySliceLen = 0

// Sentinel errors for parser_meta parse failures.
var (
	errExpectedMetaSection = errors.New(
		"expected Meta section at top of file",
	)
	errStationsNotInt = errors.New(
		"stations value is not a valid integer",
	)
	errStationsOutOfRange = errors.New(
		"stations value is out of range; must be between 1 and 10000",
	)
	errExpectedMetaKey         = errors.New("expected meta key")
	errExpectedColonAfterKey   = errors.New("expected ':' after meta key")
	errExpectedValueAfterColon = errors.New("expected value after ':'")
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
// zeroMeta is the zero-value ast.Meta returned on parse failure.
func zeroMeta() ast.Meta {
	return ast.Meta{
		Name:       emptyStr,
		ID:         emptyStr,
		SpecRef:    nil,
		Tags:       nil,
		Stations:   tokenZeroPos,
		Timeout:    tokenZeroPos,
		Parameters: nil,
		CacheTTL:   nil,
		Depends:    nil,
		Position:   ast.Position{Line: tokenZeroPos, Column: tokenZeroPos},
	}
}

func (p *parser) parseMeta() (ast.Meta, error) {
	tok := p.lex.Next()
	if tok.Kind != lex.TokenMeta {
		return zeroMeta(), fmt.Errorf(
			fmtPosExpected,
			p.file, tok.Line, tok.Column, errExpectedMetaSection, tok.Kind,
		)
	}

	meta := ast.Meta{
		Name:       emptyStr,
		ID:         emptyStr,
		SpecRef:    nil,
		Tags:       nil,
		Stations:   tokenZeroPos,
		Timeout:    tokenZeroPos,
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
		specRefLine:   tokenZeroPos,
		specRefColumn: tokenZeroPos,
	}

	err := p.collectMetaEntries(&meta, tracker)
	if err != nil {
		return zeroMeta(), err
	}

	err = validateMetaRequired(p.file, meta, tracker)
	if err != nil {
		return zeroMeta(), err
	}

	return meta, nil
}

// isTopLevelIndent reports whether the next token is a standard four-space
// top-level meta indent.
func isTopLevelIndent(tok lex.Token) bool {
	return tok.Kind == lex.TokenIndent && tok.Literal == "    "
}

// collectMetaEntries reads all indented meta-entry lines and applies them
// to meta using the provided tracker. It stops when the next token is no
// longer a top-level indent.
func (p *parser) collectMetaEntries(
	meta *ast.Meta,
	tracker *metaTracker,
) error {
	for isTopLevelIndent(p.lex.Peek()) {
		_ = p.lex.Next() // consume indent

		entry, err := p.parseMetaEntry()
		if err != nil {
			return err
		}

		err = p.applyMetaEntry(meta, entry, tracker)
		if err != nil {
			return err
		}
	}

	return nil
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
	case "Stations":
		return p.applyMetaStations(meta, tracker, entry)

	case "Timeout":
		return p.applyMetaTimeout(meta, entry)

	case "Cache-TTL":
		return p.applyMetaCacheTTL(meta, entry)

	case "Depends":
		return p.applyMetaDepends(meta)

	default:
		applyMetaSimpleEntry(meta, entry, tracker)
	}

	return nil
}

// applyMetaSimpleEntry handles meta keys that never return an error:
// Name, Id, Spec-Ref, Tags, Parameters, and unknown keys (tolerated for
// forward compatibility).
func applyMetaSimpleEntry(
	meta *ast.Meta,
	entry metaEntry,
	tracker *metaTracker,
) {
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
		applyMetaTags(meta, tracker, entry.value)

	case "Parameters":
		applyMetaParameters(meta, entry.value)

	default: // Unknown keys are silently tolerated for forward compatibility.
	}
}

// applyMetaTags appends non-empty comma-separated tags to meta.Tags.
func applyMetaTags(
	meta *ast.Meta,
	tracker *metaTracker,
	value string,
) {
	for _, tag := range splitTrimmed(value, ",") {
		if tag != emptyStr {
			meta.Tags = append(meta.Tags, tag)
			tracker.hasTags = true
		}
	}
}

// applyMetaStations parses and validates the Stations integer value.
func (p *parser) applyMetaStations(
	meta *ast.Meta,
	tracker *metaTracker,
	entry metaEntry,
) error {
	count, err := strconv.Atoi(entry.value)
	if err != nil {
		return fmt.Errorf(
			"%s:%d:%d: %w: %q",
			p.file, entry.line, entry.column,
			errStationsNotInt, entry.value,
		)
	}

	const minStations, maxStations = 1, 10000

	if count < minStations || count > maxStations {
		return fmt.Errorf(
			"%s:%d:%d: %w: got %d",
			p.file, entry.line, entry.column,
			errStationsOutOfRange, count,
		)
	}

	meta.Stations = count
	tracker.hasStations = true

	return nil
}

// applyMetaTimeout parses the Timeout duration value.
func (p *parser) applyMetaTimeout(meta *ast.Meta, entry metaEntry) error {
	dur, err := time.ParseDuration(entry.value)
	if err != nil {
		return fmt.Errorf(
			"%s:%d:%d: Timeout value %q is not a valid duration: %w",
			p.file, entry.line, entry.column, entry.value, err,
		)
	}

	meta.Timeout = dur

	return nil
}

// applyMetaParameters appends non-empty comma-separated parameters to
// meta.Parameters.
func applyMetaParameters(meta *ast.Meta, value string) {
	for _, param := range splitTrimmed(value, ",") {
		if param != emptyStr {
			meta.Parameters = append(meta.Parameters, param)
		}
	}
}

// applyMetaCacheTTL parses the Cache-TTL duration value.
func (p *parser) applyMetaCacheTTL(meta *ast.Meta, entry metaEntry) error {
	dur, err := time.ParseDuration(entry.value)
	if err != nil {
		return fmt.Errorf(
			"%s:%d:%d: Cache-TTL value %q is not a valid duration: %w",
			p.file, entry.line, entry.column, entry.value, err,
		)
	}

	meta.CacheTTL = &dur

	return nil
}

// applyMetaDepends delegates to parseDepends and assigns the result.
func (p *parser) applyMetaDepends(meta *ast.Meta) error {
	deps, err := p.parseDepends()
	if err != nil {
		return err
	}

	meta.Depends = deps

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
		return &diag.MissingKeyError{
			File:       file,
			Line:       meta.Position.Line,
			Column:     meta.Position.Column,
			Key:        "Name",
			Suggestion: "add 'Name: <test name>' to the Meta section",
		}
	}

	if !tracker.hasID {
		return &diag.MissingKeyError{
			File:       file,
			Line:       meta.Position.Line,
			Column:     meta.Position.Column,
			Key:        "Id",
			Suggestion: "add 'Id: <snake_case_id>' to the Meta section",
		}
	}

	if !tracker.hasStations {
		return &diag.MissingKeyError{
			File:       file,
			Line:       meta.Position.Line,
			Column:     meta.Position.Column,
			Key:        "Stations",
			Suggestion: "add 'Stations: 1' to the Meta section",
		}
	}

	if !tracker.hasTags {
		return &diag.MissingKeyError{
			File:   file,
			Line:   meta.Position.Line,
			Column: meta.Position.Column,
			Key:    "Tags",
			Suggestion: "add 'Tags: <comma-separated tags>'" +
				" to the Meta section",
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
		return &diag.SpecRefOnHelperError{
			File:       file,
			Line:       tracker.specRefLine,
			Column:     tracker.specRefColumn,
			SpecRef:    *meta.SpecRef,
			Suggestion: "remove the Spec-Ref key from helper stories",
		}
	}

	if !isHelper && meta.SpecRef == nil {
		return &diag.MissingSpecRefError{
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
				key:    emptyStr,
				value:  emptyStr,
				line:   keyTok.Line,
				column: keyTok.Column,
			}, fmt.Errorf(
				"%s:%d:%d: %w, got %s",
				p.file, keyTok.Line, keyTok.Column,
				errExpectedMetaKey, keyTok.Kind,
			)
	}

	colonTok := p.lex.Next()
	if colonTok.Kind != lex.TokenColon {
		return metaEntry{
				key:    emptyStr,
				value:  emptyStr,
				line:   colonTok.Line,
				column: colonTok.Column,
			}, fmt.Errorf(
				"%s:%d:%d: %w %q, got %s",
				p.file,
				colonTok.Line,
				colonTok.Column,
				errExpectedColonAfterKey,
				keyTok.Literal,
				colonTok.Kind,
			)
	}

	valTok := p.lex.Next()
	if valTok.Kind != lex.TokenValue {
		return metaEntry{
				key:    emptyStr,
				value:  emptyStr,
				line:   valTok.Line,
				column: valTok.Column,
			}, fmt.Errorf(
				"%s:%d:%d: %w, got %s",
				p.file,
				valTok.Line,
				valTok.Column,
				errExpectedValueAfterColon,
				valTok.Kind,
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

	result := make([]string, emptySliceLen, len(parts))

	for _, part := range parts {
		result = append(result, strings.TrimSpace(part))
	}

	return result
}

// containsTag reports whether tags contains the given tag (case-sensitive).
func containsTag(tags []string, tag string) bool {
	return slices.Contains(tags, tag)
}
