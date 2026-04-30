package story

import (
	"fmt"
	"strings"

	"github.com/evcoreco/octane/pkg/story/ast"
	"github.com/evcoreco/octane/pkg/story/lex"
)

// indentedColumn is the minimum column value that indicates a token
// was preceded by at least one space of indentation.
const indentedColumn = 1

// indentSize is the minimum number of leading spaces for an indented step.
const indentSize = 4

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
// form "Scenario: <title>\n<steps>". Parallel blocks embedded within the
// scenario body are flattened: their steps are collected into the scenario's
// step list and the Parallel/End-Parallel markers are discarded.
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

	steps, err := p.parseScenarioBody()
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

// parseScenarioBody collects all steps in a Scenario, including steps inside
// Parallel blocks (which are flattened into the step list). The loop exits
// when neither an indented step nor a Parallel keyword is the next token.
func (p *parser) parseScenarioBody() ([]ast.Step, error) {
	var steps []ast.Step

	for {
		peek := p.lex.Peek()

		switch {
		case isIndentToken(peek):
			batch, err := p.parseSteps()
			if err != nil {
				return nil, err
			}

			steps = append(steps, batch...)

		case peek.Kind == lex.TokenParallel:
			_ = p.lex.Next() // consume Parallel

			inner, err := p.parseParallelBlock()
			if err != nil {
				return nil, err
			}

			steps = append(steps, inner...)

		default:
			return steps, nil
		}
	}
}

// parseParallelBlock collects steps between a Parallel and End-Parallel
// keyword pair. Parallel blocks are currently flattened into the containing
// step list (reserved for future concurrent execution semantics). The
// End-Parallel token is consumed before returning.
func (p *parser) parseParallelBlock() ([]ast.Step, error) {
	var steps []ast.Step

	for {
		peek := p.lex.Peek()

		if peek.Kind == lex.TokenEndParallel || peek.Kind == lex.TokenEOF {
			if peek.Kind == lex.TokenEndParallel {
				_ = p.lex.Next() // consume End-Parallel
			}

			return steps, nil
		}

		if isIndentToken(peek) {
			batch, err := p.parseSteps()
			if err != nil {
				return nil, err
			}

			steps = append(steps, batch...)

			continue
		}

		// Skip unrecognised tokens inside a Parallel block rather than
		// erroring, so that future Parallel sub-syntax does not break
		// existing parsers.
		_ = p.lex.Next()
	}
}

// isIndentToken reports whether tok is a TokenIndent with at least one level
// of indentation (four or more leading spaces).
func isIndentToken(tok lex.Token) bool {
	return tok.Kind == lex.TokenIndent && len(tok.Literal) >= indentSize
}

// isBareStepToken reports whether tok is a TokenIllegal produced by the
// lexer's no-colon path in scanMetaEntry. In that path the lexer consumes
// the leading spaces into the key literal's position, so the token arrives
// at column >= 2 without a preceding TokenIndent. Any TokenIllegal whose
// column is > 1 (i.e. it was indented) is treated as a bare action step.
func isBareStepToken(tok lex.Token) bool {
	return tok.Kind == lex.TokenIllegal && tok.Column > indentedColumn
}

// parseSteps collects step lines while the next token is an indent of at
// least four spaces, or a bare TokenIllegal produced from an indented line
// with no step keyword and no colon. Each step line produces an ast.Step.
// Bare text lines are stored with kind ast.StepAction, supporting teardown
// actions like "Disconnect station X".
func (p *parser) parseSteps() ([]ast.Step, error) {
	var steps []ast.Step

	for {
		peek := p.lex.Peek()

		switch {
		case isIndentToken(peek):
			_ = p.lex.Next() // consume the indent

			kwTok := p.lex.Next()

			kind, bare, err := stepKindFromToken(kwTok)
			if err != nil {
				return nil, fmt.Errorf(
					"%s:%d:%d: %w",
					p.file, kwTok.Line, kwTok.Column, err,
				)
			}

			if bare {
				// Bare text line: the illegal token's literal IS the text.
				steps = append(steps, ast.Step{
					Kind: ast.StepAction,
					Text: strings.TrimSpace(kwTok.Literal),
					Position: ast.Position{
						Line:   kwTok.Line,
						Column: kwTok.Column,
					},
				})

				continue
			}

			textTok := p.lex.Next()
			if textTok.Kind != lex.TokenText {
				return nil, fmt.Errorf(
					"%s:%d:%d: expected step text after %s keyword, got %s",
					p.file, textTok.Line, textTok.Column,
					kwTok.Literal, textTok.Kind,
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

		case isBareStepToken(peek):
			// The lexer emitted TokenIllegal directly (no preceding
			// TokenIndent) because scanMetaEntry found no colon. The
			// literal is the full bare text of the indented line.
			tok := p.lex.Next()
			steps = append(steps, ast.Step{
				Kind: ast.StepAction,
				Text: strings.TrimSpace(tok.Literal),
				Position: ast.Position{
					Line:   tok.Line,
					Column: tok.Column,
				},
			})

		default:
			return steps, nil
		}
	}
}

// stepKindFromToken maps a lexer token to the AST StepKind constant. The
// second return value is true when the token represents a bare text line
// (TokenIllegal used as StepAction) rather than a Gherkin keyword.
// An error is returned only for token kinds that cannot be interpreted as
// any step at all.
func stepKindFromToken(tok lex.Token) (ast.StepKind, bool, error) {
	switch tok.Kind {
	case lex.TokenGiven:
		return ast.StepGiven, false, nil
	case lex.TokenWhen:
		return ast.StepWhen, false, nil
	case lex.TokenThen:
		return ast.StepThen, false, nil
	case lex.TokenAnd:
		return ast.StepAnd, false, nil
	case lex.TokenBut:
		return ast.StepBut, false, nil
	case lex.TokenIllegal:
		// A bare indented line with no step keyword and no colon
		// (so the lexer emitted TokenIllegal). Treat as StepAction.
		return ast.StepAction, true, nil

	case lex.TokenEOF,
		lex.TokenNewline,
		lex.TokenComment,
		lex.TokenIndent,
		lex.TokenMeta,
		lex.TokenBackground,
		lex.TokenSetup,
		lex.TokenScenario,
		lex.TokenTeardown,
		lex.TokenParallel,
		lex.TokenEndParallel,
		lex.TokenMetaKey,
		lex.TokenColon,
		lex.TokenValue,
		lex.TokenText:
		return 0, false, fmt.Errorf(
			"expected step keyword (Given/When/Then/And/But), got %s", tok.Kind,
		)
	}

	// Unreachable: all TokenKind values handled above.
	return 0, false, fmt.Errorf(
		"expected step keyword (Given/When/Then/And/But), got %s", tok.Kind,
	)
}
