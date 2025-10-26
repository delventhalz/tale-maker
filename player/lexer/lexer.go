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
}

func (l *Lexer) read() {
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.current = r
	l.nextPos = l.pos + w
}

func (l *Lexer) nextBlockStart() int {
	return len(l.input)
}

func (l *Lexer) captureText() string {
	end := l.nextBlockStart()
	text := l.input[l.pos:end]
	l.pos = end
	l.read()

	return strings.TrimSpace(text)
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.read()
	return l
}

func (l *Lexer) Next() tokens.Token {
	if (l.pos >= len(l.input)) {
		return tokens.Token{tokens.EOF, ""}
	}

	text := l.captureText()
	return tokens.Token{tokens.TEXT, text}
}
