package lexer

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
	line, col := l.line, l.col

	if !isLineBreak(l.next) {
		return "", line, col
	}

	if (isWindowsBreakStart(l.next) && isWindowsBreakEnd(l.peek)) {
		lineBreak, _, _ := l.scanPeek()
		return lineBreak, line, col
	}

	lineBreak, _, _ := l.scanNext()
	return lineBreak, line, col
}

func (l *Lexer) scanStartAction() (string, int, int) {
	line, col := l.line, l.col

	if !isAction(l.next) {
		return "", line, col
	}

	if (isEnclosingMarker(l.peek)) {
		action, _, _ := l.scanPeek();
		return action, line, col
	}

	action, _, _ := l.scanNext();
	return action, line, col;
}

func (l *Lexer) scanEscape() (string, int, int) {
	line, col := l.line, l.col

	if !isEscapeStart(l.next) {
		return "", line, col
	}

	// Escape full two-character windows line breaks
	if isLineBreak(l.peek) {
		escapeStart, _, _ := l.scanNext()
		lineBreak, _, _ := l.scanLineBreak()
		return escapeStart + lineBreak, line, col
	}

	escape, _, _ := l.scanPeek()
	return escape, line, col
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
	line, col := l.line, l.col

	if !isWordStart(l.next) {
		return "", line, col
	}

	word, _, _ := l.scanWhile(isWord)
	return word, line, col
}

func (l *Lexer) scanWhileNumberLiteral() (string, int, int) {
	line, col := l.line, l.col

	if (!isNumberStart(l.next)) {
		return "", line, col
	}

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

	if !isAnyQuote(l.next) {
		return "", line, col
	}

	endTest := getEndQuoteTest(l.next)

	if (isPaddableStartQuote(l.next) && isQuotePadding(l.peek)) {
		quote, _, _ := l.scanPeek()
		text, _, _ := l.scanUntilSequence(isQuotePadding, endTest)
		endQuote, _, _ := l.scanPeek()
		return quote + text + endQuote, line, col
	}

	quote, _, _ := l.scanNext()
	text, _, _ := l.scanUntil(endTest)
	endQuote, _, _ := l.scanNext()

	return quote + text + endQuote, line, col
}

// Run after the first text in a block has been captured.
// Drops empty lines from end if followed by new block header.
func (l *Lexer) scanWhileBlockText() (string, int, int) {
	line, col := l.line, l.col
	padding := ""
	text := ""


	for !isEof(l.next) {
		linePadding, _, _ := l.scanWhile(isNonBreakingSpace)

		switch {
		// Hit a block header, capture up to end of last non-empty line
		case isHeader(l.next):
			return text, line, col

		// Hit an action or insert, capture all text and padding
		case isAction(l.next), isInsert(l.next):
			return text + padding + linePadding, line, col

		case isEndOfLine(l.next):
			lineBreak, _, _ := l.scanLineBreak()
			padding += linePadding + lineBreak

		case isEscapeStart(l.next):
			escape, _, _ := l.scanEscape()
			text += padding + linePadding + escape
			padding = ""

		default:
			lineEnd, _, _ := l.scanUntil(isAnyOf(isEscapeStart, isLineBreak, isAction, isInsert))
			text += padding + linePadding + lineEnd
			padding = ""
		}
	}

	return text, line, col
}

// Run if no text has been captured for a block yet. Drops empty lines
// both at the start and end (if end is followed by a header)
func (l *Lexer) scanTextFromBlockStart() (string, int, int) {
	line, col := l.line, l.col
	padding := ""

	// Skip initial empty lines
	for isWhitespace(l.next) {
		linePadding, paddingLine, paddingCol := l.scanWhile(isNonBreakingSpace)

		if isLineBreak(l.next) {
			l.scanLineBreak()
			line, col = l.line, l.col
		} else {
			padding = linePadding
			line, col = paddingLine, paddingCol
		}
	}

	text, _, _ := l.scanWhileBlockText()
	if (text == "") {
		return "", line, col
	}

	return padding + text, line, col
}
