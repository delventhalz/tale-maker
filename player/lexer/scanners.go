package lexer

import (
	"unicode/utf8"
)

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

func (l *Lexer) scanStartQuote() (string, int, int) {
	line, col := l.line, l.col

	if !isAnyQuote(l.current) || l.atEndOfFile() {
		return "", line, col
	}

	quoteStart := l.current
	l.advance()

	if isPaddedStartQuote(quoteStart, l.current) {
		quote := string(quoteStart) + string(l.current)
		l.advance()
		return quote, line, col
	}

	return string(quoteStart), line, col
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

func (l *Lexer) scanWhileWord() (string, int, int) {
	line, col := l.line, l.col

	if !isWordStart(l.current) {
		return "", line, col
	}

	word, _, _ := l.scanWhile(isWord)
	return word, line, col
}

func (l *Lexer) scanWhileNumberLiteral() (string, int, int) {
	line, col := l.line, l.col

	if (!isNumberStart(l.current)) {
		return "", line, col
	}

	number := ""

	if (isMinus(l.current)) {
		minus, _, _ := l.scanNext()
		number += minus
	}

	integer, _, _ := l.scanWhile(isNumber)
	number += integer

	if (!isDot(l.current)) {
		return number, line, col
	}

	dot, _, _ := l.scanNext();
	number += dot

	fraction, _, _ := l.scanWhile(isDigit)
	number += fraction

	return number, line, col
}

func (l *Lexer) scanWhileTextLiteral() (string, int, int) {
	startQuote, line, col := l.scanStartQuote()

	if startQuote == "" {
		return startQuote, line, col
	}

	if (!isPaddedQuoteStart(startQuote)) {
		isEndQuote := getEndQuoteTest(startQuote)
		text, _, _ := l.scanUntil(isEndQuote)
		l.advance()
		return text, line, col
	}

	isPadding, isEndQuote := getPaddedEndQuoteTests(startQuote);
	text := ""

	for {
		next, _, _ := l.scanUntil(isPadding)
		text += next

		padding, _, _ := l.scanNext()

		if (isEndQuote(l.current)) {
			l.advance();
			return text, line, col
		}

		text += padding
	}
}

// Entirety of each contentful line is captured (including enclosed empty lines)
// but empty lines before and after text content is dropped
func (l *Lexer) scanWhileTextBlock() (string, int, int) {
	line, col := l.line, l.col
	padding := ""
	text := ""

	for !l.atEndOfFile() {
		linePadding, _, _ := l.scanWhile(isNonBreakingSpace)

		switch {
		// Line is a block header, capture up to end of last line
		case isHeader(l.current):
			return text, line, col

		// First non-whitespace is an action or insert, capture padding if preceded by non-empty line
		case isAction(l.current), isInsert(l.current):
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
			lineEnd, _, _ := l.scanUntil(isAnyOf(isLineBreak, isAction, isInsert))
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
