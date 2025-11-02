package lexer

import (
	"tale/tokens"
	"unicode/utf8"
)

type Lexer struct {
	input string
	pos int
	nextPos int
	current rune
	captureStack []tokens.TokenType
}

func isLineBreak(r rune) bool {
	return r == '\n' || r == '\r'
}

func isNonBreakingSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isWhitespace(r rune) bool {
	return isNonBreakingSpace(r) || isLineBreak(r)
}

func isInputHeader(r rune) bool {
	return r == '>'
}

func isStateHeader(r rune) bool {
	return r == '='
}

func isHeader(r rune) bool {
	return isInputHeader(r) || isStateHeader(r)
}

func (l *Lexer) read() {
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.current = r
	l.nextPos = l.pos + w
}

func (l *Lexer) advance() {
	l.pos = l.nextPos
	l.read()
}

func (l *Lexer) atEndOfFile() bool {
	return l.pos >= len(l.input)
}

func (l *Lexer) atEndOfLine() bool {
	return l.atEndOfFile() || isLineBreak(l.current)
}

func (l *Lexer) scanNext() string {
	if l.atEndOfFile() {
		return ""
	}

	prev := l.current
	l.advance()
	return string(prev)
}

func (l *Lexer) scanWhile(test func (rune) bool) string {
	scanned := ""

	// End when first rune fails test
	for !l.atEndOfFile() && test(l.current) {
		scanned += l.scanNext()
	}

	return scanned
}

func (l *Lexer) scanUntil(test func (rune) bool) string {
	// End when first rune passes test
	return l.scanWhile(func (r rune) bool {
		return !test(r)
	})
}

func (l *Lexer) scanWhileText() string {
	text := ""
	empties := ""

	for l.pos < len(l.input) {
		lineStart := l.scanWhile(isNonBreakingSpace)

		switch {
		// Line is a block header, capture up to end of last line
		case isHeader(l.current):
			return text

		// Line is empty, only capture if non-empty lines are before and after
		case l.atEndOfLine():
			if text == "" {
				l.scanNext() // skip line break
			} else {
				empties += lineStart
				empties += l.scanNext() // include line break
			}

		default:
			text += empties
			text += lineStart
			text += l.scanUntil(isLineBreak)
			empties += l.scanNext() // include line break
		}
	}

	return text
}

func (l *Lexer) startCaptureOf(t tokens.TokenType) {
	l.captureStack = append(l.captureStack, t)
}

func (l *Lexer) endCurrentCapture() {
	if len(l.captureStack) > 0 {
		l.captureStack = l.captureStack[:len(l.captureStack) - 1]
	}
}

func (l *Lexer) isCapturing(t tokens.TokenType) bool {
	if len(l.captureStack) == 0 {
		return false
	}
	return l.captureStack[len(l.captureStack) - 1] == t
}

func (l *Lexer) isCapturingAny(ts ...tokens.TokenType) bool {
	if len(l.captureStack) == 0 {
		return false
	}

	for _, t := range ts {
		if l.captureStack[len(l.captureStack) - 1] == t {
			return true
		}
	}

	return false
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.read()
	return l
}

func (l *Lexer) Next() tokens.Token {
	if isNonBreakingSpace(l.current) {
		l.scanWhile(isNonBreakingSpace)
	}

	if l.atEndOfFile() {
		return tokens.Token{tokens.EOF, ""}
	}

	// In a Block Header
	if l.isCapturingAny(tokens.INPUT_HEADER, tokens.STATE_HEADER) {
		if isNonBreakingSpace(l.current) {
			l.scanWhile(isNonBreakingSpace)
		}

		arg := ""

		if l.isCapturing(tokens.INPUT_HEADER) && isInputHeader(l.current) {
			arg += l.scanWhile(isInputHeader)
		}

		if l.isCapturing(tokens.STATE_HEADER) && isStateHeader(l.current) {
			arg += l.scanWhile(isStateHeader)
		}

		if isNonBreakingSpace(l.current) {
			l.scanWhile(isNonBreakingSpace)
		}

		// Hit end of line (with or without header characters) header is done
		if l.atEndOfLine() {
			l.endCurrentCapture()
			lineBreak := l.scanNext()

			if (arg == "") {
				return tokens.Token{tokens.HEADER_END, lineBreak}
			}

			return tokens.Token{tokens.HEADER_END, arg}
		}

		arg += l.scanUntil(isWhitespace)
		return tokens.Token{tokens.ARG, arg}
	}

	// N in a block header, safe to advance past line breaks
	if isWhitespace(l.current) {
		l.scanWhile(isWhitespace)
	}

	// Check EOF again
	if l.atEndOfFile() {
		return tokens.Token{tokens.EOF, ""}
	}

	// Starting a Block Header
	if isInputHeader(l.current) {
		header := l.scanWhile(isInputHeader)
		l.startCaptureOf(tokens.INPUT_HEADER)
		return tokens.Token{tokens.INPUT_HEADER, header}
	}

	if isStateHeader(l.current) {
		header := l.scanWhile(isStateHeader)
		l.startCaptureOf(tokens.STATE_HEADER)
		return tokens.Token{tokens.STATE_HEADER, header}
	}

	// Text
	text := l.scanWhileText()
	return tokens.Token{tokens.TEXT, text}
}
