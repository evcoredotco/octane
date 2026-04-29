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

	var background []ast.Step

	if p.lex.Peek().Kind == lex.TokenBackground {
		background, err = p.parseBackground()
		if err != nil {
			return nil, err
		}
	}

	var setup []ast.Step

	if p.lex.Peek().Kind == lex.TokenSetup {
		setup, err = p.parseSetup()
		if err != nil {
			return nil, err
		}
	}

	var scenarios []ast.Scenario

	for p.lex.Peek().Kind == lex.TokenScenario {
		sc, scErr := p.parseScenario()
		if scErr != nil {
			return nil, scErr
		}

		scenarios = append(scenarios, sc)
	}

	if len(scenarios) == 0 {
		tok := p.lex.Peek()

		return nil, &diag.ErrMissingSection{
			File:       p.file,
			Line:       tok.Line,
			Column:     tok.Column,
			Section:    "Scenario",
			Suggestion: "add at least one 'Scenario: <title>' block",
		}
	}

	var teardown []ast.Step

	if p.lex.Peek().Kind == lex.TokenTeardown {
		teardown, err = p.parseTeardown()
		if err != nil {
			return nil, err
		}
	}

	tok := p.lex.Next()
	if tok.Kind != lex.TokenEOF {
		return nil, &diag.ErrUnexpectedToken{
			File:     p.file,
			Line:     tok.Line,
			Column:   tok.Column,
			Got:      tok.Kind.String(),
			Expected: "EOF",
			Suggestion: "remove or relocate content after the final section " +
				"(Background, Scenario, Teardown)",
		}
	}

	err = validateParameters(p.file, meta, scenarios, background, setup, teardown)
	if err != nil {
		return nil, err
	}

	return &ast.Story{
		Path:       p.file,
		Meta:       meta,
		Background: background,
		Setup:      setup,
		Scenarios:  scenarios,
		Teardown:   teardown,
		Position: ast.Position{
			Line:   startTok.Line,
			Column: startTok.Column,
		},
	}, nil
}
