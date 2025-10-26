package lexer

import "tale/tokens"

type Lexer struct {
	input string
	pos int
}

func New(input string) *Lexer {
	return &Lexer{input, 0}
}

func (l *Lexer) Next() tokens.Token {
	if (l.pos >= len(l.input)) {
		return tokens.Token{tokens.EOF, ""}
	}

	return tokens.Token{tokens.ILLEGAL, string(l.input[l.pos])}
}
