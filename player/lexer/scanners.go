package lexer

import (
	"strings"
)

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

func (l *Lexer) skipComment() {
	atLineStart := l.atLineStart

	if isComment(l.next) && isCommentMarker(l.peek) {
		l.advance()
		l.advance()

		for !isEof(l.next) && !(isCommentMarker(l.next) && isCommentEnd(l.peek)) {
			l.advance()
		}

		l.advance()
		l.advance()
		l.atLineStart = atLineStart
	}
}

func (l *Lexer) scanNext() (string, int, int) {
	line, col := l.line, l.col
	if isEof(l.next) {
		return "", line, col
	}

	next := l.next
	l.advance()
	return string(next), line, col
}

func (l *Lexer) scanPeek() (string, int, int) {
	next, line, col := l.scanNext()
	peek, _, _ := l.scanNext();
	return next + peek, line, col
}

func (l *Lexer) scanLineBreak() (string, int, int) {
	if (isWindowsBreakStart(l.next) && isWindowsBreakEnd(l.peek)) {
		return l.scanPeek()
	}

	return l.scanNext()
}

func (l *Lexer) scanStartAction() (string, int, int) {
	if (isEnclosingMarker(l.peek)) {
		return l.scanPeek();
	}

	return l.scanNext();
}

func (l *Lexer) scanEscape() (string, int, int) {
	line, col := l.line, l.col

	// Skip full two-character windows line breaks
	if isLineBreak(l.peek) {
		atLineStart := l.atLineStart
		l.scanNext()
		l.scanLineBreak()
		l.atLineStart = atLineStart
		return "", line, col
	}

	escaped := getEscaped(l.peek)
	l.scanPeek()

	return escaped, line, col
}

func (l *Lexer) scanWhile(test func (rune) bool) (string, int, int) {
	line, col := l.line, l.col
	scanned := ""

	// End when current rune fails test
	for !isEof(l.next) && test(l.next) {
		next, _, _ := l.scanNext()
		scanned += next
	}

	return scanned, line, col
}

func (l *Lexer) scanUntil(test func (rune) bool) (string, int, int) {
	// End when current rune passes test
	return l.scanWhile(func (r rune) bool {
		return !test(r)
	})
}

func (l *Lexer) scanUntilSequence(
	nextTest func (rune) bool,
	peekTest func (rune) bool,
) (string, int, int) {
	line, col := l.line, l.col
	scanned := ""

	for !isEof(l.next) && !(nextTest(l.next) && peekTest(l.peek)) {
		next, _, _ := l.scanNext()
		scanned += next
	}

	return scanned, line, col
}

func (l *Lexer) scanWhileWord() (string, int, int) {
	word, line, col := l.scanWhile(isWord)
	return strings.ToLower(word), line, col
}

func (l *Lexer) scanWhileNumberLiteral() (string, int, int) {
	line, col := l.line, l.col
	number := ""

	if (isMinus(l.next)) {
		minus, _, _ := l.scanNext()
		number += minus
	}

	integer, _, _ := l.scanWhile(isNumber)
	number += integer

	if (!isDot(l.next)) {
		return number, line, col
	}

	dot, _, _ := l.scanNext();
	number += dot

	fraction, _, _ := l.scanWhile(isDigit)
	number += fraction

	return number, line, col
}

func (l *Lexer) scanWhileQuotedText() (string, int, int) {
	line, col := l.line, l.col

	endTest := getEndQuoteTest(l.next)

	if (isPaddableStartQuote(l.next) && isQuotePadding(l.peek)) {
		l.scanPeek()
		text, _, _ := l.scanUntilSequence(isQuotePadding, endTest)
		l.scanPeek()
		return text, line, col
	}

	l.scanNext()
	text, _, _ := l.scanUntil(endTest)
	l.scanNext()

	return text, line, col
}

// Run after the first text in a block has been captured.
// Drops empty lines from end if followed by new block header.
func (l *Lexer) scanWhileBlockText() (string, int, int) {
	line, col := l.line, l.col
	allPadding := ""
	linePadding := ""
	text := ""

	for !isEof(l.next) {
		nextPadding, _, _ := l.scanWhile(isNonBreakingSpace)
		linePadding += nextPadding

		switch {
		case isComment(l.next) && isCommentMarker(l.peek):
			l.skipComment()

		// Hit a block header, capture up to end of last non-empty line
		case l.atLineStart && isHeader(l.next):
			return text, line, col

		// Hit an action or insert, capture all text and padding
		case isAction(l.next), isInsert(l.next):
			return text + allPadding + linePadding, line, col

		case isEndOfLine(l.next):
			lineBreak, _, _ := l.scanLineBreak()
			allPadding += linePadding + lineBreak
			linePadding = ""

		case isEscapeStart(l.next):
			escape, _, _ := l.scanEscape()
			text += allPadding + linePadding + escape
			allPadding = ""
			linePadding = ""

		default:
			lineEnd, _, _ := l.scanUntil(isAnyOf(isEscapeStart, isLineBreak, isAction, isInsert))
			text += allPadding + linePadding + lineEnd
			allPadding = ""
			linePadding = ""
		}
	}

	return text, line, col
}

// Run if no text has been captured for a block yet. Drops empty lines
// both at the start and end (if end is followed by a header)
func (l *Lexer) scanTextFromBlockStart() (string, int, int) {
	line, col := l.line, l.col
	linePadding := ""

	// Skip all empty lines then scan text normally
	for !isEof(l.next) {
		nextPadding, _, _ := l.scanWhile(isNonBreakingSpace)
		linePadding += nextPadding

		switch {
		case isComment(l.next) && isCommentMarker(l.peek):
			l.skipComment()

		case isLineBreak(l.next):
			l.scanLineBreak()
			line, col = l.line, l.col
			linePadding = ""

		default:
			text, _, _ := l.scanWhileBlockText()
			if (text == "") {
				return "", line, col
			}
			return linePadding + text, line, col
		}
	}

	return "", line, col
}
