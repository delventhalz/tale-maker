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
	capturedBlockStart bool
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

func isAction(r rune) bool {
	return r == '<'
}

func isActionEnd(r rune) bool {
	return r == '>'
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

func (l *Lexer) scanNext() (string, int, int) {
	line, col := l.line, l.col

	if l.atEndOfFile() {
		return "", line, col
	}

	next := l.current
	l.advance()
	return string(next), line, col
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
func (l *Lexer) scanWhileText() (string, int, int) {
	line, col := l.line, l.col
	padding := ""
	text := ""

	for !l.atEndOfFile() {
		linePadding, _, _ := l.scanWhile(isNonBreakingSpace)

		switch {
		// Line is a block header, capture up to end of last line
		case isHeader(l.current):
			return text, line, col

		// First non-whitespace is an action, capture padding if preceded by non-empty line
		case isAction(l.current):
			if text != "" {
				text += padding + linePadding
			}
			return text, line, col

		// Line is empty, only capture if not at start or end of block
		case l.atEndOfLine():
			if text == "" && !l.capturedBlockStart {
				l.scanLineBreak() // skip line break
				padding = ""
				line = l.line
				col = l.col
			} else {
				lineBreak, _, _ := l.scanLineBreak()
				padding += linePadding + lineBreak
			}

		default:
			lineEnd, _, _ := l.scanUntil(isAnyOf(isLineBreak, isAction))
			text += padding + linePadding + lineEnd
			padding = ""
			if isLineBreak(l.current) {
				lineBreak, _, _ := l.scanLineBreak()
				padding += lineBreak
			}

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

// Ends any matching captures IN ORDER from top of stack to bottom
func (l *Lexer) endAnyCapturesOf(ts ...tokens.TokenType) {
	for _, t := range ts {
		count := len(l.captureStack)
		if count > 0 && l.captureStack[count - 1] == t {
			l.endCurrentCapture()
		}
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
	// Either returns a token or loops if position is a no-op
	for {
		// In a Block Header
		if l.isCapturingAny(tokens.INPUT_HEADER, tokens.STATE_HEADER) {
			l.scanWhile(isNonBreakingSpace)

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

		// In an Action
		if l.isCapturingAny(tokens.ACTION, tokens.NAME) {
			l.scanWhile(isWhitespace)

			if l.atEndOfFile() {
				return tokens.Token{tokens.EOF, "", l.line, l.col}
			}

			if isActionEnd(l.current) {
				end, line, col := l.scanNext()
				l.endAnyCapturesOf(tokens.NAME, tokens.ACTION)
				return tokens.Token{tokens.ACTION_END, end, line, col}
			}
		}

		if l.isCapturing(tokens.NAME) {
			name, line, col := l.scanUntil(isAnyOf(isWhitespace, isActionEnd))
			l.endCurrentCapture()
			return tokens.Token{tokens.NAME, name, line, col}
		}

		if l.isCapturing(tokens.ACTION) {
			arg, line, col := l.scanUntil(isAnyOf(isWhitespace, isActionEnd))
			return tokens.Token{tokens.ARG, arg, line, col}
		}

		// Test for EOF after handling special captures states
		if l.atEndOfFile() {
			return tokens.Token{tokens.EOF, "", l.line, l.col}
		}

		// Starting a Block Header
		if isInputHeader(l.current) {
			header, line, col := l.scanWhile(isInputHeader)
			l.startCaptureOf(tokens.INPUT_HEADER)
			l.capturedBlockStart = false
			return tokens.Token{tokens.INPUT_HEADER, header, line, col}
		}

		if isStateHeader(l.current) {
			header, line, col := l.scanWhile(isStateHeader)
			l.startCaptureOf(tokens.STATE_HEADER)
			l.capturedBlockStart = false
			return tokens.Token{tokens.STATE_HEADER, header, line, col}
		}

		// Starting an Action
		if isAction(l.current) {
			action, line, col := l.scanNext()
			l.startCaptureOf(tokens.ACTION)
			l.startCaptureOf(tokens.NAME)
			return tokens.Token{tokens.ACTION, action, line, col}
		}

		// Text
		text, textLine, textCol := l.scanWhileText()

		if text == "" {
			continue
		}

		l.capturedBlockStart = true
		return tokens.Token{tokens.TEXT, text, textLine, textCol}
	}
}
