package lexer

import (
	"tale/tokens"
	"unicode/utf8"
)

type Lexer struct {
	input string
	pos int
	nextPos int
	line int
	col int
	current rune
	captureStack []tokens.TokenType
}

func isLineBreak(r rune) bool {
	return r == '\n' || r == '\r'
}

func isWindowsLineBreak(first rune, second rune) bool {
	return first == '\r' && second == '\n'
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

func isAnyOf[T any](tests ...func(T) bool) func(T) bool {
	return func(val T) bool {
		for _, test := range tests {
			if test(val) {
				return true
			}
		}
		return false
	}
}

func (l *Lexer) read() {
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.current = r
	l.nextPos = l.pos + w
}

func (l *Lexer) advance() {
	prev := l.current

	l.pos = l.nextPos
	l.read()

	switch {
	case isWindowsLineBreak(prev, l.current): // no increment for first char
	case isLineBreak(prev):
		l.line++
		l.col = 1
	default:
		l.col++
	}
}

func (l *Lexer) atEndOfFile() bool {
	return l.pos >= len(l.input)
}

func (l *Lexer) atEndOfLine() bool {
	return l.atEndOfFile() || isLineBreak(l.current)
}

func (l *Lexer) scanLineBreak() (string, int, int) {
	line, col := l.line, l.col

	if !isLineBreak(l.current) || l.atEndOfFile() {
		return "", line, col
	}

	breakStart := l.current
	l.advance()

	if isWindowsLineBreak(breakStart, l.current) {
		lineBreak := string(breakStart) + string(l.current)
		l.advance()
		return lineBreak, line, col
	}

	return string(breakStart), line, col
}

func (l *Lexer) scanWhile(test func (rune) bool) (string, int, int) {
	line, col := l.line, l.col
	scanned := ""

	// End when current rune fails test
	for !l.atEndOfFile() && test(l.current) {
		scanned += string(l.current)
		l.advance()
	}

	return scanned, line, col
}

func (l *Lexer) scanUntil(test func (rune) bool) (string, int, int) {
	// End when current rune passes test
	return l.scanWhile(func (r rune) bool {
		return !test(r)
	})
}

// Entirety of each contentful line is captured (including enclosed empty lines)
// but empty lines before and after text content is dropped
func (l *Lexer) scanWhileText(startPadding string, startLine, startCol int) (string, int, int) {
	padding := startPadding
	line := startLine
	col := startCol
	text := ""

	for l.pos < len(l.input) {
		lineStart, _, _ := l.scanWhile(isNonBreakingSpace)

		switch {
		// Line is a block header, capture up to end of last line
		case isHeader(l.current):
			return text, line, col

		// Line is empty, only capture if non-empty lines are before and after
		case l.atEndOfLine():
			if text == "" {
				l.scanLineBreak() // skip line break
				padding = ""
				line = l.line
				col = l.col
			} else {
				lineBreak, _, _ := l.scanLineBreak()
				padding += lineStart + lineBreak
			}

		default:
			lineEnd, _, _ := l.scanUntil(isLineBreak)
			lineBreak, _, _ := l.scanLineBreak()
			text += padding + lineStart + lineEnd
			padding += lineBreak
		}
	}

	return text, line, col
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
	l := &Lexer{input: input, line: 1, col: 1}
	l.read()
	return l
}

func (l *Lexer) Next() tokens.Token {
	// Either returns a token or loops if current line is ignored
	for {
		startPadding, startLine, startCol := l.scanWhile(isNonBreakingSpace)

		// In a Block Header
		if l.isCapturingAny(tokens.INPUT_HEADER, tokens.STATE_HEADER) {
			var headerEnd string
			var endLine, endCol int

			if l.isCapturing(tokens.INPUT_HEADER) && isInputHeader(l.current) {
				headerEnd, endLine, endCol = l.scanWhile(isInputHeader)
			}
			if l.isCapturing(tokens.STATE_HEADER) && isStateHeader(l.current) {
				headerEnd, endLine, endCol = l.scanWhile(isStateHeader)
			}

			// Empty to end of line (with or without header chars) and header ends
			l.scanWhile(isNonBreakingSpace)
			if l.atEndOfLine() {
				l.endCurrentCapture()
				lineBreak, breakLine, breakCol := l.scanLineBreak()

				if headerEnd == "" {
					return tokens.Token{tokens.HEADER_END, lineBreak, breakLine, breakCol}
				}

				return tokens.Token{tokens.HEADER_END, headerEnd, endLine, endCol}
			}

			var arg string
			var argLine, argCol int

			if l.isCapturing(tokens.STATE_HEADER) {
				arg, argLine, argCol = l.scanUntil(isAnyOf(isWhitespace, isStateHeader))
			} else {
				arg, argLine, argCol = l.scanUntil(isAnyOf(isWhitespace, isInputHeader))
			}

			return tokens.Token{tokens.ARG, arg, argLine, argCol}
		}

		// Test for EOF only after handling possible implicit header end
		if l.atEndOfFile() {
			return tokens.Token{tokens.EOF, "", l.line, l.col}
		}

		// Starting a Block Header
		if isInputHeader(l.current) {
			header, line, col := l.scanWhile(isInputHeader)
			l.startCaptureOf(tokens.INPUT_HEADER)
			return tokens.Token{tokens.INPUT_HEADER, header, line, col}
		}

		if isStateHeader(l.current) {
			header, line, col := l.scanWhile(isStateHeader)
			l.startCaptureOf(tokens.STATE_HEADER)
			return tokens.Token{tokens.STATE_HEADER, header, line, col}
		}

		// Text
		text, textLine, textCol := l.scanWhileText(startPadding, startLine, startCol)

		if text == "" {
			continue
		}

		return tokens.Token{tokens.TEXT, text, textLine, textCol}
	}


}
