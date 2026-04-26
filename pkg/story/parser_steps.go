// Package story — see parse.go for package documentation.
package story

import (
	"fmt"

	"github.com/octane-project/octane/pkg/story/ast"
	"github.com/octane-project/octane/pkg/story/lex"
)

// parseBackground implements T-001-24: expects TokenBackground followed by
// one or more indented step lines. Returns the collected steps.
func (p *parser) parseBackground() ([]ast.Step, error) {
	tok := p.lex.Next()
	if tok.Kind != lex.TokenBackground {
		return nil, fmt.Errorf(
			"%s:%d:%d: expected Background keyword, got %s",
			p.file, tok.Line, tok.Column, tok.Kind,
		)
	}

	return p.parseSteps()
}

// parseSetup implements T-001-24: expects TokenSetup followed by one or
// more indented step lines. Returns the collected steps.
func (p *parser) parseSetup() ([]ast.Step, error) {
	tok := p.lex.Next()
	if tok.Kind != lex.TokenSetup {
		return nil, fmt.Errorf(
			"%s:%d:%d: expected Setup keyword, got %s",
			p.file, tok.Line, tok.Column, tok.Kind,
		)
	}

	return p.parseSteps()
}

// parseTeardown implements T-001-24: expects TokenTeardown followed by one
// or more indented step lines. Returns the collected steps.
func (p *parser) parseTeardown() ([]ast.Step, error) {
	tok := p.lex.Next()
	if tok.Kind != lex.TokenTeardown {
		return nil, fmt.Errorf(
			"%s:%d:%d: expected Teardown keyword, got %s",
			p.file, tok.Line, tok.Column, tok.Kind,
		)
	}

	return p.parseSteps()
}

// parseScenario implements T-001-24: parses a single Scenario section of the
// form "Scenario: <title>\n<steps>".
func (p *parser) parseScenario() (ast.Scenario, error) {
	scTok := p.lex.Next()
	if scTok.Kind != lex.TokenScenario {
		return ast.Scenario{}, fmt.Errorf(
			"%s:%d:%d: expected Scenario keyword, got %s",
			p.file, scTok.Line, scTok.Column, scTok.Kind,
		)
	}

	colonTok := p.lex.Next()
	if colonTok.Kind != lex.TokenColon {
		return ast.Scenario{}, fmt.Errorf(
			"%s:%d:%d: expected ':' after Scenario, got %s",
			p.file, colonTok.Line, colonTok.Column, colonTok.Kind,
		)
	}

	titleTok := p.lex.Next()
	if titleTok.Kind != lex.TokenText {
		return ast.Scenario{}, fmt.Errorf(
			"%s:%d:%d: expected scenario title text after 'Scenario:', got %s",
			p.file, titleTok.Line, titleTok.Column, titleTok.Kind,
		)
	}

	steps, err := p.parseSteps()
	if err != nil {
		return ast.Scenario{}, err
	}

	return ast.Scenario{
		Name:  titleTok.Literal,
		Steps: steps,
		Position: ast.Position{
			Line:   scTok.Line,
			Column: scTok.Column,
		},
	}, nil
}

// parseSteps collects step lines while the next token is TokenIndent. Each
// step line produces an ast.Step with Kind, Text, and Position populated.
func (p *parser) parseSteps() ([]ast.Step, error) {
	var steps []ast.Step

	for p.lex.Peek().Kind == lex.TokenIndent {
		_ = p.lex.Next() // consume the indent

		kwTok := p.lex.Next()

		kind, err := stepKindFromToken(kwTok.Kind)
		if err != nil {
			return nil, fmt.Errorf(
				"%s:%d:%d: %w",
				p.file, kwTok.Line, kwTok.Column, err,
			)
		}

		textTok := p.lex.Next()
		if textTok.Kind != lex.TokenText {
			return nil, fmt.Errorf(
				"%s:%d:%d: expected step text after %s keyword, got %s",
				p.file, textTok.Line, textTok.Column, kwTok.Literal, textTok.Kind,
			)
		}

		steps = append(steps, ast.Step{
			Kind: kind,
			Text: textTok.Literal,
			Position: ast.Position{
				Line:   kwTok.Line,
				Column: kwTok.Column,
			},
		})
	}

	return steps, nil
}

// stepKindFromToken maps a lexer step-keyword token to the AST StepKind
// constant. It returns an error for any token that is not a step keyword.
func stepKindFromToken(kind lex.TokenKind) (ast.StepKind, error) {
	switch kind {
	case lex.TokenGiven:
		return ast.StepGiven, nil
	case lex.TokenWhen:
		return ast.StepWhen, nil
	case lex.TokenThen:
		return ast.StepThen, nil
	case lex.TokenAnd:
		return ast.StepAnd, nil
	case lex.TokenBut:
		return ast.StepBut, nil
	default:
		return 0, fmt.Errorf(
			"expected step keyword (Given/When/Then/And/But), got %s", kind,
		)
	}
}
