package lexer

func (l *Lexer) atEndOfFile() bool {
	return l.pos >= len(l.input)
}

func (l *Lexer) atEndOfLine() bool {
	return l.atEndOfFile() || isLineBreak(l.next)
}

func (l *Lexer) scanNext() (string, int, int) {
	line, col := l.line, l.col

	if l.atEndOfFile() {
		return "", line, col
	}

	next := l.next
	l.advance()
	return string(next), line, col
}

func (l *Lexer) scanLineBreak() (string, int, int) {
	line, col := l.line, l.col

	if !isLineBreak(l.next) {
		return "", line, col
	}

	breakStart, _, _ := l.scanNext()

	if isScanningWindowsLineBreak(breakStart, l.next) {
		breakEnd, _, _ := l.scanNext()
		return breakStart + breakEnd, line, col
	}

	return breakStart, line, col
}

func (l *Lexer) scanStartAction() (string, int, int) {
	line, col := l.line, l.col

	if !isAction(l.next) {
		return "", line, col
	}

	action, _, _ := l.scanNext();

	if !isEnclosingMarker(l.next) {
		return action, line, col
	}

	marker, _, _ := l.scanNext()

	return action + marker, line, col;
}

func (l *Lexer) scanStartQuote() (string, int, int) {
	line, col := l.line, l.col

	if !isAnyQuote(l.next) || l.atEndOfFile() {
		return "", line, col
	}

	quoteStart, _, _ := l.scanNext()

	if isScanningPaddedStartQuote(quoteStart, l.next) {
		quoteEnd, _, _ := l.scanNext()
		return quoteStart + quoteEnd, line, col
	}

	return quoteStart, line, col
}

func (l *Lexer) scanEscape() (string, int, int) {
	line, col := l.line, l.col

	if !isEscapeStart(l.next) {
		return "", line, col
	}

	escapeStart, _, _ := l.scanNext()

	// Escape full two-character windows line breaks
	if isLineBreak(l.next) {
		lineBreak, _, _ := l.scanLineBreak()
		return escapeStart + lineBreak, line, col
	}

	escaped, _, _ := l.scanNext()
	return escapeStart + escaped, line, col
}

func (l *Lexer) scanWhile(test func (rune) bool) (string, int, int) {
	line, col := l.line, l.col
	scanned := ""

	// End when current rune fails test
	for !l.atEndOfFile() && test(l.next) {
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
	startQuote, line, col := l.scanStartQuote()

	if startQuote == "" {
		return startQuote, line, col
	}

	if (!isPaddedQuoteStart(startQuote)) {
		isEndQuote := getEndQuoteTest(startQuote)
		text, _, _ := l.scanUntil(isEndQuote)
		endQuote, _, _ := l.scanNext();
		return startQuote + text + endQuote, line, col
	}

	isPadding, isEndQuote := getPaddedEndQuoteTests(startQuote);
	text := ""

	for {
		next, _, _ := l.scanUntil(isPadding)
		text += next

		padding, _, _ := l.scanNext()
		text += padding

		if (isEndQuote(l.next)) {
			endQuote, _, _ := l.scanNext();
			return startQuote + text + endQuote, line, col
		}

	}
}

// Entirety of each contentful line is captured (including enclosed empty lines)
// but empty lines before and after text content is dropped
func (l *Lexer) scanWhileBlockText() (string, int, int) {
	line, col := l.line, l.col
	padding := ""
	text := ""

	for !l.atEndOfFile() {
		linePadding, _, _ := l.scanWhile(isNonBreakingSpace)

		switch {
		// Line is a block header, capture up to end of last line
		case isHeader(l.next):
			return text, line, col

		// First non-whitespace is an action or insert, capture padding if
		// not the block start or preceded by non-empty line
		case isAction(l.next), isInsert(l.next):
			if text != "" || l.capturedBlockStart {
				text += padding + linePadding
			}
			return text, line, col

		case isEscapeStart(l.next):
			escape, _, _ := l.scanEscape()
			text += padding + linePadding + escape
			padding = ""

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
			lineEnd, _, _ := l.scanUntil(isAnyOf(isEscapeStart, isLineBreak, isAction, isInsert))
			text += padding + linePadding + lineEnd

			padding = ""
			if isLineBreak(l.next) {
				lineBreak, _, _ := l.scanLineBreak()
				padding += lineBreak
			}
		}
	}

	return text, line, col
}
