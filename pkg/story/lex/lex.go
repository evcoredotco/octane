// Package lex defines the token types and lexer interface for the .story
// DSL. This package is a leaf in the dependency graph: it imports nothing
// from the octane module and relies only on the standard library.
//
// The lexer contract is defined here (T-001-04); the byte-stream
// implementation is provided by T-001-10.
package lex

import "strings"

// emptyLiteral is the named empty-string sentinel for token Literal fields,
// required by the add-constant linter rule.
const emptyLiteral = ""

// initialPos is the starting byte offset in the source.
const initialPos = 0

// initialLine is the starting line number (1-based).
const initialLine = 1

// initialCol is the starting column number (1-based).
const initialCol = 1

// nextByte is the lookahead offset for checking the following byte.
const nextByte = 1

// noColon flags that no ':' separator has been found yet.
const noColon = -1

// TokenKind identifies the type of a lexical token produced by the lexer.
type TokenKind int

const (
	// TokenIllegal represents an unrecognised byte sequence.
	TokenIllegal TokenKind = iota

	// TokenEOF signals the end of the input.
	TokenEOF

	// TokenNewline represents a single newline character (after CRLF
	// normalisation to LF).
	TokenNewline

	// TokenComment represents a comment line starting with '#', up to
	// but not including the terminating newline.
	TokenComment

	// TokenIndent represents leading whitespace (exactly four spaces)
	// at the start of an indented line.
	TokenIndent

	// --- Section-level keywords (appear at column 1, unindented) ---

	// TokenMeta introduces the Meta section.
	TokenMeta

	// TokenBackground introduces the Background section.
	TokenBackground

	// TokenSetup introduces the Setup section.
	TokenSetup

	// TokenScenario introduces a Scenario section.
	TokenScenario

	// TokenTeardown introduces the Teardown section.
	TokenTeardown

	// TokenParallel introduces a parallel multi-station block (reserved
	// for future use).
	TokenParallel

	// TokenEndParallel closes a parallel multi-station block (reserved
	// for future use).
	TokenEndParallel

	// --- Step keywords (appear indented inside a section body) ---

	// TokenGiven introduces a precondition step.
	TokenGiven

	// TokenWhen introduces an action step.
	TokenWhen

	// TokenThen introduces an expected-outcome step.
	TokenThen

	// TokenAnd continues the preceding step kind.
	TokenAnd

	// TokenBut introduces a negative continuation of the preceding step kind.
	TokenBut

	// --- Meta-section tokens ---

	// TokenMetaKey represents a meta-header identifier such as "Name",
	// "Id", "Spec-Ref", or "Tags".
	TokenMetaKey

	// TokenColon represents the ':' separator between a meta key and
	// its value.
	TokenColon

	// TokenValue represents the trimmed text after the colon on a meta
	// line.
	TokenValue

	// --- Step text ---

	// TokenText represents the verbatim step text that follows a step
	// keyword.
	TokenText
)

// tokenKindName returns the human-readable name for kind, or "Unknown". Using
// a function-local map avoids a long switch and brings the cyclomatic
// complexity below the configured limit without introducing a package-level
// variable.
func tokenKindName(kind TokenKind) string {
	names := map[TokenKind]string{
		TokenIllegal:     "Illegal",
		TokenEOF:         "EOF",
		TokenNewline:     "Newline",
		TokenComment:     "Comment",
		TokenIndent:      "Indent",
		TokenMeta:        "Meta",
		TokenBackground:  "Background",
		TokenSetup:       "Setup",
		TokenScenario:    "Scenario",
		TokenTeardown:    "Teardown",
		TokenParallel:    "Parallel",
		TokenEndParallel: "EndParallel",
		TokenGiven:       "Given",
		TokenWhen:        "When",
		TokenThen:        "Then",
		TokenAnd:         "And",
		TokenBut:         "But",
		TokenMetaKey:     "MetaKey",
		TokenColon:       "Colon",
		TokenValue:       "Value",
		TokenText:        "Text",
	}

	if name, ok := names[kind]; ok {
		return name
	}

	return "Unknown"
}

// String returns a human-readable name for the token kind. The returned
// string is useful in error messages and debug output.
func (k TokenKind) String() string {
	return tokenKindName(k)
}

// Token is the smallest unit produced by the lexer. Each token carries its
// kind, the exact source bytes that comprise it, and the 1-based line and
// column where it begins.
type Token struct {
	// Kind identifies the token type.
	Kind TokenKind

	// Literal holds the exact source bytes for this token.
	Literal string

	// Line is the 1-based line number where the token starts.
	Line int

	// Column is the 1-based byte offset from the start of the line
	// where the token starts.
	Column int
}

// Lexer tokenises a .story source byte slice. Implementations must
// normalise CRLF to LF before tokenising. After TokenEOF is returned,
// every subsequent call to Next must also return TokenEOF.
type Lexer interface {
	// Next returns the next token. After TokenEOF is returned,
	// subsequent calls continue to return TokenEOF.
	Next() Token

	// Peek returns the next token without consuming it. Consecutive
	// calls to Peek without an intervening Next return the same token.
	Peek() Token
}

// NewLexer returns a Lexer that tokenises src. The file parameter is used
// only for error messages. CRLF sequences in src are normalised to LF
// before any tokenisation occurs (T-001-11).
func NewLexer(_ string, src []byte) Lexer {
	return &lexer{
		src:   normaliseCRLF(src),
		pos:   initialPos,
		line:  initialLine,
		col:   initialCol,
		queue: nil,
	}
}

// normaliseCRLF replaces every \r\n pair with a single \n (T-001-11). Lone
// \r bytes that are not followed by \n are left as-is so that illegal-byte
// detection downstream can handle them.
func normaliseCRLF(src []byte) []byte {
	out := make([]byte, initialPos, len(src))

	for idx := initialPos; idx < len(src); idx++ {
		isCRLF := src[idx] == '\r' &&
			idx+nextByte < len(src) && src[idx+nextByte] == '\n'
		if isCRLF {
			out = append(out, '\n')
			idx++ // skip the paired \n; loop increment advances past it

			continue
		}

		out = append(out, src[idx])
	}

	return out
}

// lexer is the concrete byte-stream implementation of the Lexer interface.
//
// A single source line may produce multiple tokens (e.g. an indented meta
// entry emits TokenIndent, TokenMetaKey, TokenColon, TokenValue). The
// queue field holds pre-computed tokens that will be returned by
// subsequent calls to Next before scan() is invoked again.
type lexer struct {
	src   []byte
	pos   int
	line  int
	col   int
	queue []Token // FIFO of pre-scanned tokens not yet consumed
}

// Peek returns the next token without consuming it. Repeated calls without
// an intervening Next return the same token.
func (l *lexer) Peek() Token {
	tok := l.Next()
	l.queue = append([]Token{tok}, l.queue...)

	return tok
}

// Next returns the next token and advances the lexer position. After
// TokenEOF is first returned, every subsequent call also returns
// TokenEOF.
func (l *lexer) Next() Token {
	if len(l.queue) > initialPos {
		tok := l.queue[initialPos]
		l.queue = l.queue[nextByte:]

		return tok
	}

	return l.scan()
}

// scan reads from the current byte position and returns the next logical
// token. It may enqueue additional tokens into l.queue for multi-token
// lines before returning.
func (l *lexer) scan() Token {
	for {
		if l.pos >= len(l.src) {
			return l.eofToken()
		}

		nextByte := l.src[l.pos]

		switch nextByte {
		case '\n':
			// Blank line — skip silently and continue.
			l.advance()

		case '#':
			return l.scanComment()

		case ' ':
			return l.scanIndentedLine()

		case '\t':
			return l.scanIllegalByte()

		default:
			return l.scanSectionLine()
		}
	}
}

// scanComment emits TokenComment for a line beginning with '#'. The
// literal includes the '#' and all bytes through the end of the line.
// The terminating newline is consumed but not emitted.
func (l *lexer) scanComment() Token {
	startLine, startCol := l.line, l.col
	start := l.pos

	for l.pos < len(l.src) && l.src[l.pos] != '\n' {
		l.advance()
	}

	tok := Token{
		Kind:    TokenComment,
		Literal: string(l.src[start:l.pos]),
		Line:    startLine,
		Column:  startCol,
	}

	l.consumeNewline()

	return tok
}

// indentWidth is the exact number of spaces that constitute one level of
// indentation in the .story grammar (ADR 0006).
const indentWidth = 4

// scanIndentedLine handles a line whose first character is a space.
// The minimum valid indentation is exactly 4 spaces (one indent level);
// lines with more spaces are sub-indented continuation lines (e.g. Depends
// sub-entries). Lines with fewer than 4 leading spaces are TokenIllegal.
//
// The TokenIndent literal carries all leading spaces so callers can
// determine the indentation depth (len(tok.Literal) / indentWidth).
func (l *lexer) scanIndentedLine() Token {
	startLine, startCol := l.line, l.col

	count := l.countLeadingSpaces()

	if count < indentWidth {
		return l.scanUnderIndented(startLine, startCol, count)
	}

	return l.scanProperlyIndented(startLine, startCol, count)
}

// scanUnderIndented handles a line with fewer than four leading spaces, which
// is not a valid indent level. It consumes the malformed line and returns
// TokenIllegal.
func (l *lexer) scanUnderIndented(
	startLine, startCol, count int,
) Token {
	raw := l.src[l.pos : l.pos+count]

	for range count {
		l.advance()
	}

	start := l.pos

	for l.pos < len(l.src) && l.src[l.pos] != '\n' {
		l.advance()
	}

	literal := string(raw) + string(l.src[start:l.pos])
	l.consumeNewline()

	return Token{
		Kind:    TokenIllegal,
		Literal: literal,
		Line:    startLine,
		Column:  startCol,
	}
}

// scanProperlyIndented handles a line with at least four leading spaces.
// It dispatches to blank-line, comment, step-keyword, or meta-entry
// handling.
func (l *lexer) scanProperlyIndented(
	startLine, startCol, count int,
) Token {
	indentLiteral := strings.Repeat(" ", count)

	for range count {
		l.advance()
	}

	indentTok := Token{
		Kind:    TokenIndent,
		Literal: indentLiteral,
		Line:    startLine,
		Column:  startCol,
	}

	// Blank indented line (only spaces before newline / EOF).
	if l.pos >= len(l.src) || l.src[l.pos] == '\n' {
		l.consumeNewline()

		return Token{
			Kind:    TokenIllegal,
			Literal: indentLiteral,
			Line:    startLine,
			Column:  startCol,
		}
	}

	// Comment after indent — emit TokenIndent and let next scan() pick
	// up the '#'.
	if l.src[l.pos] == '#' {
		return indentTok
	}

	// Try step keywords first.
	if res := l.tryStepKeyword(); res.matched {
		l.enqueue(res.textTok)

		return l.withIndentQueued(indentTok, res.keywordTok)
	}

	// Otherwise treat as a meta entry: Key: Value.
	return l.scanMetaEntry(indentTok)
}

// withIndentQueued prepends kwTok to the front of the queue so that the
// sequence emitted is: indentTok (returned now), kwTok (next), and
// whatever was already in the queue after that (e.g. textTok).
func (l *lexer) withIndentQueued(indentTok, kwTok Token) Token {
	l.queue = append([]Token{kwTok}, l.queue...)

	return indentTok
}

// stepMatchResult carries the result of a step-keyword match attempt.
type stepMatchResult struct {
	keywordTok Token
	textTok    Token
	matched    bool
}

// tryStepKeyword checks whether the current position starts with one of
// the five step keywords. If so, it consumes the keyword and the
// following step text up to EOL and returns the populated stepMatchResult.
func (l *lexer) tryStepKeyword() stepMatchResult {
	type stepEntry struct {
		text string
		kind TokenKind
	}

	steps := [...]stepEntry{
		{"Given", TokenGiven},
		{"When", TokenWhen},
		{"Then", TokenThen},
		{"And", TokenAnd},
		{"But", TokenBut},
	}

	for _, step := range steps {
		if res := l.matchStepKeyword(step.text, step.kind); res.matched {
			return res
		}
	}

	illegal := Token{
		Kind:    TokenIllegal,
		Literal: emptyLiteral,
		Line:    initialPos,
		Column:  initialPos,
	}

	return stepMatchResult{
		keywordTok: illegal,
		textTok:    illegal,
		matched:    false,
	}
}

// matchStepKeyword attempts to match a single step keyword at the current
// position. It returns a stepMatchResult with matched=true on success.
func (l *lexer) matchStepKeyword(
	text string,
	kind TokenKind,
) stepMatchResult {
	noMatch := Token{
		Kind:    TokenIllegal,
		Literal: emptyLiteral,
		Line:    initialPos,
		Column:  initialPos,
	}

	if !l.hasPrefix(text) {
		return stepMatchResult{
			keywordTok: noMatch,
			textTok:    noMatch,
			matched:    false,
		}
	}

	// Ensure the keyword is followed by a space or end-of-line.
	afterKeyword := l.pos + len(text)
	if afterKeyword < len(l.src) &&
		l.src[afterKeyword] != ' ' &&
		l.src[afterKeyword] != '\n' {
		return stepMatchResult{
			keywordTok: noMatch,
			textTok:    noMatch,
			matched:    false,
		}
	}

	kwLine, kwCol := l.line, l.col
	l.consumeBytes(len(text))

	// Consume the single separating space if present.
	if l.pos < len(l.src) && l.src[l.pos] == ' ' {
		l.advance()
	}

	kwTok := Token{
		Kind:    kind,
		Literal: text,
		Line:    kwLine,
		Column:  kwCol,
	}
	textTok := l.scanToEOLasText()

	return stepMatchResult{keywordTok: kwTok, textTok: textTok, matched: true}
}

// scanMetaEntry scans a meta-entry line of the form "Key: Value" after
// the 4-space indent has already been consumed. It enqueues
// TokenMetaKey, TokenColon, and TokenValue and returns TokenIndent.
//
// If no colon is found on the line, a TokenIllegal is returned instead.
func (l *lexer) scanMetaEntry(indentTok Token) Token {
	keyLine, keyCol := l.line, l.col
	start := l.pos

	colonPos := l.findColonOnLine()

	if colonPos < 0 {
		return l.scanIllegalLineFrom(start, keyLine, keyCol)
	}

	// Consume up to (but not including) the colon.
	for l.pos < colonPos {
		l.advance()
	}

	keyLiteral := string(l.src[start:l.pos])

	colonLine, colonCol := l.line, l.col
	l.advance() // consume ':'

	// Collect the value: trim leading whitespace, then read to EOL.
	valLine, valCol := l.line, l.col
	valStart := l.pos

	for l.pos < len(l.src) && l.src[l.pos] != '\n' {
		l.advance()
	}

	valLiteral := strings.TrimSpace(string(l.src[valStart:l.pos]))
	l.consumeNewline()

	l.enqueue(
		Token{
			Kind:    TokenMetaKey,
			Literal: keyLiteral,
			Line:    keyLine,
			Column:  keyCol,
		},
		Token{
			Kind:    TokenColon,
			Literal: ":",
			Line:    colonLine,
			Column:  colonCol,
		},
		Token{
			Kind:    TokenValue,
			Literal: valLiteral,
			Line:    valLine,
			Column:  valCol,
		},
	)

	return indentTok
}

// findColonOnLine scans forward from the current position to find the first
// ':' character before the end of the current line. It returns the byte
// index of the colon, or noColon (-1) when none is found.
func (l *lexer) findColonOnLine() int {
	srcLen := len(l.src)

	for scanIdx := l.pos; scanIdx < srcLen; scanIdx++ {
		if l.src[scanIdx] == '\n' {
			break
		}

		if l.src[scanIdx] == ':' {
			return scanIdx
		}
	}

	return noColon
}

// scanIllegalLineFrom consumes the rest of the current line starting from
// start and returns a TokenIllegal.
func (l *lexer) scanIllegalLineFrom(start, line, col int) Token {
	for l.pos < len(l.src) && l.src[l.pos] != '\n' {
		l.advance()
	}

	literal := string(l.src[start:l.pos])
	l.consumeNewline()

	return Token{
		Kind:    TokenIllegal,
		Literal: literal,
		Line:    line,
		Column:  col,
	}
}

// scanSectionLine handles an unindented line. It tries to match one of
// the section-level keywords. After the keyword it may find a colon and
// trailing text (Scenario lines). Unknown content produces TokenIllegal.
func (l *lexer) scanSectionLine() Token {
	type sectionEntry struct {
		text string
		kind TokenKind
	}

	// End-Parallel must precede Parallel to avoid prefix mis-match.
	sections := [...]sectionEntry{
		{"End-Parallel", TokenEndParallel},
		{"Background", TokenBackground},
		{"Teardown", TokenTeardown},
		{"Parallel", TokenParallel},
		{"Scenario", TokenScenario},
		{"Setup", TokenSetup},
		{"Meta", TokenMeta},
	}

	kwLine, kwCol := l.line, l.col

	for _, sect := range sections {
		if !l.hasPrefix(sect.text) {
			continue
		}

		l.consumeBytes(len(sect.text))

		kwTok := Token{
			Kind:    sect.kind,
			Literal: sect.text,
			Line:    kwLine,
			Column:  kwCol,
		}

		l.consumeSectionLineRemainder()

		return kwTok
	}

	// Unknown unindented content.
	return l.scanIllegalToEOL(kwLine, kwCol)
}

// consumeSectionLineRemainder handles the rest of an unindented section-
// keyword line. If the next byte is ':' it emits TokenColon and TokenText;
// otherwise it silently consumes up to the newline.
func (l *lexer) consumeSectionLineRemainder() {
	if l.pos >= len(l.src) || l.src[l.pos] != ':' {
		// No colon — consume the rest of the line silently.
		for l.pos < len(l.src) && l.src[l.pos] != '\n' {
			l.advance()
		}

		l.consumeNewline()

		return
	}

	colonLine, colonCol := l.line, l.col
	l.advance() // consume ':'

	colonTok := Token{
		Kind:    TokenColon,
		Literal: ":",
		Line:    colonLine,
		Column:  colonCol,
	}

	// Skip optional space after colon.
	if l.pos < len(l.src) && l.src[l.pos] == ' ' {
		l.advance()
	}

	textTok := l.scanToEOLasText()
	l.enqueue(colonTok, textTok)
}

// scanToEOLasText reads from the current position to the end of the line,
// trims surrounding whitespace, and returns a TokenText. The newline is
// consumed.
func (l *lexer) scanToEOLasText() Token {
	startLine, startCol := l.line, l.col
	start := l.pos

	for l.pos < len(l.src) && l.src[l.pos] != '\n' {
		l.advance()
	}

	literal := strings.TrimSpace(string(l.src[start:l.pos]))
	l.consumeNewline()

	return Token{
		Kind:    TokenText,
		Literal: literal,
		Line:    startLine,
		Column:  startCol,
	}
}

// scanIllegalToEOL reads from the current position to end of line and
// returns TokenIllegal positioned at startLine/startCol. The newline is
// consumed.
func (l *lexer) scanIllegalToEOL(startLine, startCol int) Token {
	start := l.pos

	for l.pos < len(l.src) && l.src[l.pos] != '\n' {
		l.advance()
	}

	literal := string(l.src[start:l.pos])
	l.consumeNewline()

	return Token{
		Kind:    TokenIllegal,
		Literal: literal,
		Line:    startLine,
		Column:  startCol,
	}
}

// scanIllegalByte emits TokenIllegal for a single forbidden byte (e.g. tab).
func (l *lexer) scanIllegalByte() Token {
	startLine, startCol := l.line, l.col
	b := l.src[l.pos]
	l.advance()

	return Token{
		Kind:    TokenIllegal,
		Literal: string([]byte{b}),
		Line:    startLine,
		Column:  startCol,
	}
}

// eofToken returns a TokenEOF positioned at the current (past-the-end)
// location.
func (l *lexer) eofToken() Token {
	return Token{
		Kind:    TokenEOF,
		Literal: "",
		Line:    l.line,
		Column:  l.col,
	}
}

// advance moves the cursor forward by one byte and updates the line and
// column counters. A '\n' byte increments the line counter and resets the
// column to 1 for the next byte.
func (l *lexer) advance() {
	if l.pos >= len(l.src) {
		return
	}

	if l.src[l.pos] == '\n' {
		l.line++
		l.col = initialCol
	} else {
		l.col++
	}

	l.pos++
}

// consumeNewline advances past a '\n' if the current byte is a newline.
func (l *lexer) consumeNewline() {
	if l.pos < len(l.src) && l.src[l.pos] == '\n' {
		l.advance()
	}
}

// hasPrefix reports whether the source starting at the current position
// begins with the given prefix string.
func (l *lexer) hasPrefix(prefix string) bool {
	end := l.pos + len(prefix)
	if end > len(l.src) {
		return false
	}

	return string(l.src[l.pos:end]) == prefix
}

// consumeBytes advances the cursor by exactly n bytes.
func (l *lexer) consumeBytes(n int) {
	for range n {
		l.advance()
	}
}

// countLeadingSpaces returns the number of consecutive space bytes at the
// current position without advancing the cursor.
func (l *lexer) countLeadingSpaces() int {
	count := 0

	for l.pos+count < len(l.src) && l.src[l.pos+count] == ' ' {
		count++
	}

	return count
}

// enqueue appends tokens to the back of the lookahead queue.
func (l *lexer) enqueue(toks ...Token) {
	l.queue = append(l.queue, toks...)
}
