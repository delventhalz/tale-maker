package lexer

import (
	"strings"
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

func (l *Lexer) advanceTo(loc int) {
	l.pos = loc
	l.read()
}

func (l *Lexer) advance() {
	l.advanceTo(l.nextPos)
}

func (l *Lexer) startCaptureOf(t tokens.TokenType) {
	l.captureStack = append(l.captureStack, t)
}

func (l *Lexer) endCurrentCapture() {
	if (len(l.captureStack) > 0) {
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

func (l *Lexer) getRawHeaderAt(loc int) string {
	// Block headers must always be on a new line
	if r, _ := utf8.DecodeLastRuneInString(l.input[:loc]); loc > 0 && !isLineBreak(r) {
		return ""
	}

	header := ""
	hasHeaderToken := false

	for _, r := range l.input[loc:] {
		if !hasHeaderToken && isNonBreakingSpace(r) {
			header += string(r)
		} else if isHeader(r) {
			hasHeaderToken = true
			header += string(r)
		} else {
			break
		}
	}

	if (hasHeaderToken) {
		return header
	}

	return ""
}

func (l *Lexer) getRawHeaderEnd() string {
	headerEnd := ""

	for _, r := range l.input[l.pos:] {
		switch {
		case isNonBreakingSpace(r):
			headerEnd += string(r)
		case isHeader(r):
			headerEnd += string(r)
		case isLineBreak(r):
			headerEnd += string(r)
			return headerEnd
		default:
			return ""
		}
	}

	return ""
}

func (l *Lexer) nextWhitespaceAt() int {
	for i, r := range l.input[l.nextPos:] {
		if isWhitespace(r) {
			return l.nextPos + i
		}
	}

	return len(l.input)
}

func (l *Lexer) nextBlockAt() int {
	for i := range l.input[l.nextPos:] {
		loc := l.nextPos + i
		if l.getRawHeaderAt(loc) != "" {
			return loc
		}
	}

	return len(l.input)
}

func (l *Lexer) captureNext() string {
	next := l.current
	l.advance()
	return string(next)
}

func (l *Lexer) captureArg() string {
	end := l.nextWhitespaceAt()
	rawArg := l.input[l.pos:end]
	l.advanceTo(end)
	return strings.TrimSpace(rawArg)
}

func (l *Lexer) captureText() string {
	end := l.nextBlockAt()
	rawText := l.input[l.pos:end]
	l.advanceTo(end)
	return strings.TrimSpace(rawText)
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.read()
	return l
}

func (l *Lexer) Next() tokens.Token {
	// EOF
	if (l.pos >= len(l.input)) {
		return tokens.Token{tokens.EOF, ""}
	}

	// Block Headers
	if l.isCapturingAny(tokens.INPUT_HEADER, tokens.STATE_HEADER) {
		if h := l.getRawHeaderEnd(); h != "" {
			l.advanceTo(l.pos + len(h))
			l.endCurrentCapture()

			// Retain whitespace if end is a single line break character
			headerEnd := h
			if (len(headerEnd) > 1) {
				headerEnd = strings.TrimSpace(headerEnd)
			}

			return tokens.Token{tokens.HEADER_END, headerEnd}
		}

		arg := l.captureArg()
		return tokens.Token{tokens.ARG, arg}
	}

	if h := l.getRawHeaderAt(l.pos); h != "" {
		l.advanceTo(l.pos + len(h))
		header := strings.TrimSpace(h)

		if r, _ := utf8.DecodeRuneInString(header); isStateHeader(r) {
			l.startCaptureOf(tokens.STATE_HEADER)
			return tokens.Token{tokens.STATE_HEADER, header}
		}

		l.startCaptureOf(tokens.INPUT_HEADER)
		return tokens.Token{tokens.INPUT_HEADER, header}
	}

	// Text
	text := l.captureText()

	if (text == "") {
		return l.Next()
	}

	return tokens.Token{tokens.TEXT, text}
}
