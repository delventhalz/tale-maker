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
	capturedBlockStart bool
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
	case isWindowsBreakStart(prev) && isWindowsBreakEnd(l.next): // no increment for first char
	case isLineBreak(prev):
		l.line++
		l.col = 1
	default:
		l.col++
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
	l := &Lexer{ input: input, line: 1, col: 1 }

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

		if l.isCapturingAny(tokens.ACTION, tokens.INSERT) {
			l.scanWhile(isWhitespace)
		}

		if isEof(l.next) {
			return tokens.Token{tokens.EOF, "", l.line, l.col}
		}

		if isComment(l.next) && isCommentMarker(l.peek) {
			l.scanUntil(isCommentEnd)
			l.scanNext()
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

				if headerEnd != "" {
					return tokens.Token{tokens.HEADER_END, headerEnd, endLine, endCol}
				}

				if lineBreak != "" {
					return tokens.Token{tokens.HEADER_END, lineBreak, breakLine, breakCol}
				}

				continue // Restart loop to capture EOF
			}
		}

		if l.isCapturing(tokens.ACTION) {
			if isActionEnd(l.next) {
				end, line, col := l.scanNext()
				l.endCurrentCapture()
				return tokens.Token{tokens.ACTION_END, end, line, col}
			}
		}

		if l.isCapturing(tokens.INSERT) {
			if isInsertEnd(l.next) {
				end, line, col := l.scanNext()
				l.endCurrentCapture()
				return tokens.Token{tokens.INSERT_END, end, line, col}
			}
		}

		// Capturing an expression in a header, action, or insert
		if l.isCapturingAny(tokens.INPUT_HEADER, tokens.STATE_HEADER, tokens.ACTION, tokens.INSERT) {
			if isNumberStart(l.next) {
				number, numberLine, numberCol := l.scanWhileNumberLiteral()
				if isScannedMinus(number) {
					return tokens.Token{tokens.INVALID, number, numberLine, numberCol}
				}
				return tokens.Token{tokens.NUMBER, number, numberLine, numberCol}
			}

			if isAnyQuote(l.next) {
				quote, quoteLine, quoteCol := l.scanWhileQuotedText()
				return tokens.Token{tokens.QUOTED_TEXT, quote, quoteLine, quoteCol}
			}

			if isWordStart(l.next) {
				word, wordLine, wordCol := l.scanWhileWord()
				return tokens.Token{getWordToken(word), word, wordLine, wordCol}
			}

			invalid, invalidLine, invalidCol := l.scanNext()
			return tokens.Token{tokens.INVALID, invalid, invalidLine, invalidCol}
		}

		// Starting a Block Header
		if isInputHeader(l.next) {
			header, line, col := l.scanWhile(isInputHeader)
			l.startCaptureOf(tokens.INPUT_HEADER)
			l.capturedBlockStart = false
			return tokens.Token{tokens.INPUT_HEADER, header, line, col}
		}

		if isStateHeader(l.next) {
			header, line, col := l.scanWhile(isStateHeader)
			l.startCaptureOf(tokens.STATE_HEADER)
			l.capturedBlockStart = false
			return tokens.Token{tokens.STATE_HEADER, header, line, col}
		}

		// Starting an Action or Enclosing Action
		if isAction(l.next) {
			action, line, col := l.scanStartAction()
			l.startCaptureOf(tokens.ACTION)
			return tokens.Token{getActionToken(action), action, line, col}
		}

		// Starting an Insert
		if isInsert(l.next) {
			insert, line, col := l.scanNext()
			l.startCaptureOf(tokens.INSERT)
			return tokens.Token{tokens.INSERT, insert, line, col}
		}

		// Block Text
		if !l.capturedBlockStart {
			start, startLine, startCol := l.scanTextFromBlockStart()

			if (start == "") {
				continue
			}

			l.capturedBlockStart = true
			return tokens.Token{tokens.BLOCK_TEXT, start, startLine, startCol}
		}

		text, textLine, textCol := l.scanWhileBlockText()
		if text == "" {
			continue
		}

		return tokens.Token{tokens.BLOCK_TEXT, text, textLine, textCol}
	}

	panic(fmt.Sprintf("Repeat position [%d]! %q (%d:%d)", l.pos, l.next, l.line, l.col))
}
