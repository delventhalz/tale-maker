package lexer

import (
	"fmt"
	"tale/tokens"
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
	// Either returns a token or loops if position is a no-op.
	// Stops looping if it repeats a position (likely dev error)
	prevPos := -1

	for prevPos != l.pos {
		prevPos = l.pos

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
		}

		// In an Action or Insert
		if l.isCapturingAny(tokens.ACTION, tokens.INSERT) {
			l.scanWhile(isWhitespace)

			if l.atEndOfFile() {
				return tokens.Token{tokens.EOF, "", l.line, l.col}
			}
		}

		if l.isCapturing(tokens.ACTION) {
			if isActionEnd(l.current) {
				end, line, col := l.scanNext()
				l.endCurrentCapture()
				return tokens.Token{tokens.ACTION_END, end, line, col}
			}
		}

		if l.isCapturing(tokens.INSERT) {
			if isInsertEnd(l.current) {
				end, line, col := l.scanNext()
				l.endCurrentCapture()
				return tokens.Token{tokens.INSERT_END, end, line, col}
			}
		}

		// Capturing an expression in a header, action, or insert
		if l.isCapturingAny(tokens.INPUT_HEADER, tokens.STATE_HEADER, tokens.ACTION, tokens.INSERT) {
			if isNumberStart(l.current) {
				number, numberline, numberCol := l.scanWhileNumberLiteral()
				return tokens.Token{tokens.NUMBER, number, numberline, numberCol}
			}
			if isAnyQuote(l.current) {
				text, textLint, textCol := l.scanWhileTextLiteral()
				return tokens.Token{tokens.TEXT, text, textLint, textCol}
			}
			word, wordLine, wordCol := l.scanWhileWord()
			return tokens.Token{getWordToken(word), word, wordLine, wordCol}
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
			return tokens.Token{tokens.ACTION, action, line, col}
		}

		// Starting an Insert
		if isInsert(l.current) {
			insert, line, col := l.scanNext()
			l.startCaptureOf(tokens.INSERT)
			return tokens.Token{tokens.INSERT, insert, line, col}
		}

		// Text
		text, textLine, textCol := l.scanWhileTextBlock()

		if text == "" {
			continue
		}

		l.capturedBlockStart = true
		return tokens.Token{tokens.TEXT, text, textLine, textCol}
	}

	panic(fmt.Sprintf("Repeat position [%d]! %q (%d:%d)", l.pos, l.current, l.line, l.col))
}
