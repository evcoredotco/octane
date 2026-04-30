// Package story parses .story DSL files into an abstract syntax tree.
// It is the public entry point for the OCTANE story parser (T-001-26).
//
// The grammar is defined in ADR 0006:
//
//	story = meta_section background? setup? scenario+ teardown?
//
// Parse is the only exported function; all other symbols in this package
// are unexported helpers used by the section sub-parsers.
package story

import (
	"github.com/evcoreco/octane/pkg/story/ast"
	"github.com/evcoreco/octane/pkg/story/diag"
	"github.com/evcoreco/octane/pkg/story/lex"
)

// parser holds the per-file lexer and the source path used in diagnostics.
type parser struct {
	file string
	lex  lex.Lexer
}

// Parse parses a single .story file from its byte content.
// file is the filesystem path, used only in error messages.
// Returns (*ast.Story, nil) on success or (nil, error) on failure.
// The error will be one of the typed errors from pkg/story/diag.
func Parse(file string, src []byte) (*ast.Story, error) {
	p := &parser{
		file: file,
		lex:  lex.NewLexer(file, src),
	}

	return p.parseStory()
}

// skipLeadingComments discards any TokenComment tokens at the start of
// the file so that story files may begin with a file-level comment block.
// This is called once at the top of parseStory before parseMeta.
func (p *parser) skipLeadingComments() {
	for p.lex.Peek().Kind == lex.TokenComment {
		_ = p.lex.Next()
	}
}

// parseStory is the root production. It drives the top-level grammar:
//
//	story = meta_section background? setup? scenario+ teardown?
//
// After all sections are collected it calls validateParameters to check
// that every {placeholder} in step text is declared in Meta.Parameters.
func (p *parser) parseStory() (*ast.Story, error) {
	p.skipLeadingComments()

	startTok := p.lex.Peek()

	meta, err := p.parseMeta()
	if err != nil {
		return nil, err
	}

	sections, err := p.parseStorySections()
	if err != nil {
		return nil, err
	}

	err = validateParameters(
		p.file, meta,
		sections.scenarios, sections.background,
		sections.setup, sections.teardown,
	)
	if err != nil {
		return nil, err
	}

	return &ast.Story{
		Path:       p.file,
		Meta:       meta,
		Background: sections.background,
		Setup:      sections.setup,
		Scenarios:  sections.scenarios,
		Teardown:   sections.teardown,
		Position: ast.Position{
			Line:   startTok.Line,
			Column: startTok.Column,
		},
	}, nil
}

// storySections groups the four optional/required grammar sections.
type storySections struct {
	background []ast.Step
	setup      []ast.Step
	scenarios  []ast.Scenario
	teardown   []ast.Step
}

// parseStorySections parses Background?, Setup?, Scenario+, Teardown? and
// validates that at least one Scenario is present.
func (p *parser) parseStorySections() (storySections, error) {
	var (
		sections storySections
		err      error
	)

	sections.background, err = p.parseOptionalBackground()
	if err != nil {
		return storySections{}, err
	}

	sections.setup, err = p.parseOptionalSetup()
	if err != nil {
		return storySections{}, err
	}

	sections.scenarios, err = p.parseAllScenarios()
	if err != nil {
		return storySections{}, err
	}

	sections.teardown, err = p.parseOptionalTeardown()
	if err != nil {
		return storySections{}, err
	}

	return sections, p.expectEOF()
}

// parseOptionalBackground parses the Background section when present.
func (p *parser) parseOptionalBackground() ([]ast.Step, error) {
	if p.lex.Peek().Kind != lex.TokenBackground {
		return nil, nil
	}

	return p.parseBackground()
}

// parseOptionalSetup parses the Setup section when present.
func (p *parser) parseOptionalSetup() ([]ast.Step, error) {
	if p.lex.Peek().Kind != lex.TokenSetup {
		return nil, nil
	}

	return p.parseSetup()
}

// parseOptionalTeardown parses the Teardown section when present.
func (p *parser) parseOptionalTeardown() ([]ast.Step, error) {
	if p.lex.Peek().Kind != lex.TokenTeardown {
		return nil, nil
	}

	return p.parseTeardown()
}

// expectEOF consumes the next token and returns an error when it is not EOF.
func (p *parser) expectEOF() error {
	tok := p.lex.Next()
	if tok.Kind == lex.TokenEOF {
		return nil
	}

	return &diag.UnexpectedTokenError{
		File:     p.file,
		Line:     tok.Line,
		Column:   tok.Column,
		Got:      tok.Kind.String(),
		Expected: "EOF",
		Suggestion: "remove or relocate content after the final section " +
			"(Background, Scenario, Teardown)",
	}
}

// parseAllScenarios parses one or more Scenario blocks. It returns
// *diag.MissingSectionError when no Scenario is found.
func (p *parser) parseAllScenarios() ([]ast.Scenario, error) {
	var scenarios []ast.Scenario

	for p.lex.Peek().Kind == lex.TokenScenario {
		sc, err := p.parseScenario()
		if err != nil {
			return nil, err
		}

		scenarios = append(scenarios, sc)
	}

	if len(scenarios) == 0 {
		tok := p.lex.Peek()

		return nil, &diag.MissingSectionError{
			File:       p.file,
			Line:       tok.Line,
			Column:     tok.Column,
			Section:    "Scenario",
			Suggestion: "add at least one 'Scenario: <title>' block",
		}
	}

	return scenarios, nil
}
