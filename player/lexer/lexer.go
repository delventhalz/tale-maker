package lexer

import (
	"fmt"
	"tale/tokens"
	"unicode/utf8"
)

type Lexer struct {
	input string
	next rune
	peek rune
	pos int
	peekPos int
	readPos int
	line int
	col int
	captureStack []tokens.TokenType
	atBlockStart bool
	atLineStart bool
}

func (l *Lexer) read() (rune, int) {
	if l.readPos >= len(l.input) {
		return 0, 0
	}
	return utf8.DecodeRuneInString(l.input[l.readPos:])
}

func (l *Lexer) advance() {
	r, w := l.read()

	prev := l.next
	l.next = l.peek
	l.peek = r

	l.pos = l.peekPos
	l.peekPos = l.readPos
	l.readPos += w

	switch {
	case isEof(prev): // no increment at end of file
	case isWindowsBreakStart(prev) && isWindowsBreakEnd(l.next): // no increment for first char
	case isLineBreak(prev):
		l.line++
		l.col = 1
		l.atLineStart = true
	default:
		l.col++
		if l.atLineStart && !isWhitespace(prev) {
			l.atLineStart = false
		}
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
	l := &Lexer{
		input: input,
		line: 1,
		col: 1,
		atBlockStart: true,
		atLineStart: true,
	}

	l.next, l.peekPos = l.read()
	l.readPos = l.peekPos

	peek, peekWidth := l.read()
	l.peek = peek
	l.readPos = l.peekPos + peekWidth

	return l
}

func (l *Lexer) Next() tokens.Token {
	// Either returns a token or loops if position is a no-op.
	// Stops looping if it repeats a position (likely dev error)
	prevPos := -1

	for prevPos != l.pos {
		prevPos = l.pos

		if l.isCapturingAny(tokens.INPUT_HEADER, tokens.STATE_HEADER) {
			l.scanWhile(isNonBreakingSpace)
		}

		if l.isCapturing(tokens.ACTION) {
			l.scanWhile(isWhitespace)
		}

		if isEof(l.next) {
			return tokens.Token{tokens.EOF, "", l.line, l.col}
		}

		if isComment(l.next) && isCommentMarker(l.peek) {
			l.skipComment()
			continue // restart loop without returning token
		}

		// Check for end of header at end of line
		if l.isCapturingAny(tokens.INPUT_HEADER, tokens.STATE_HEADER) {
			var headerEnd string
			var endLine, endCol int

			if l.isCapturing(tokens.INPUT_HEADER) && isInputHeader(l.next) {
				headerEnd, endLine, endCol = l.scanWhile(isInputHeader)
			}
			if l.isCapturing(tokens.STATE_HEADER) && isStateHeader(l.next) {
				headerEnd, endLine, endCol = l.scanWhile(isStateHeader)
			}

			if headerEnd != "" {
				l.scanWhile(isNonBreakingSpace)
				if !isEndOfLine(l.next) {
					return tokens.Token{tokens.INVALID, headerEnd, endLine, endCol}
				}
			}

			if isEndOfLine(l.next) {
				l.endCurrentCapture()
				lineBreak, breakLine, breakCol := l.scanLineBreak()
				l.atBlockStart = true

				if headerEnd != "" {
					return tokens.Token{tokens.HEADER_END, headerEnd, endLine, endCol}
				}

				if lineBreak != "" {
					return tokens.Token{tokens.HEADER_END, lineBreak, breakLine, breakCol}
				}

				continue // restart loop to capture EOF
			}
		}

		if l.isCapturing(tokens.ACTION) {
			if isActionEnd(l.next) {
				end, line, col := l.scanNext()
				l.endCurrentCapture()
				return tokens.Token{tokens.ACTION_END, end, line, col}
			}
		}

		// Capturing an expression in a header or action
		if l.isCapturingAny(tokens.INPUT_HEADER, tokens.STATE_HEADER, tokens.ACTION) {
			if isOperator(l.next) {
				operator, opLine, opCol := l.scanOperator()
				return tokens.Token{getOperatorToken(operator), operator, opLine, opCol}
			}

			if isNumberStart(l.next) {
				number, numberLine, numberCol := l.scanWhileNumberLiteral()
				return tokens.Token{tokens.NUMBER, number, numberLine, numberCol}
			}

			if isAnyQuote(l.next) {
				quote, quoteLine, quoteCol := l.scanWhileQuotedText()
				return tokens.Token{tokens.TEXT, quote, quoteLine, quoteCol}
			}

			if isWordStart(l.next) {
				word, wordLine, wordCol := l.scanWhileWord()
				return tokens.Token{getWordToken(word), word, wordLine, wordCol}
			}

			invalid, invalidLine, invalidCol := l.scanNext()
			return tokens.Token{tokens.INVALID, invalid, invalidLine, invalidCol}
		}

		// Starting a Block Header
		if l.atLineStart {
			if isInputHeader(l.next) {
				header, line, col := l.scanWhile(isInputHeader)
				l.startCaptureOf(tokens.INPUT_HEADER)
				return tokens.Token{tokens.INPUT_HEADER, header, line, col}
			}

			if isStateHeader(l.next) {
				header, line, col := l.scanWhile(isStateHeader)
				l.startCaptureOf(tokens.STATE_HEADER)
				return tokens.Token{tokens.STATE_HEADER, header, line, col}
			}
		}

		// Starting an Action or Enclosing Action
		if isAction(l.next) {
			action, line, col := l.scanStartAction()
			l.startCaptureOf(tokens.ACTION)
			return tokens.Token{getActionToken(action), action, line, col}
		}

		// Block Text
		if l.atBlockStart {
			start, startLine, startCol := l.scanTextFromBlockStart()

			if (start == "") {
				continue
			}

			l.atBlockStart = false
			return tokens.Token{tokens.TEXT, start, startLine, startCol}
		}

		text, textLine, textCol := l.scanWhileBlockText()
		if text == "" {
			continue
		}

		return tokens.Token{tokens.TEXT, text, textLine, textCol}
	}

	panic(fmt.Sprintf("Repeat position [%d]! %q (%d:%d)", l.pos, l.next, l.line, l.col))
}
