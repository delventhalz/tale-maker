package lexer

import (
    "tale/tokens"
    "testing"
)

func TestEof(t *testing.T) {
	lex := New("")
	actual := lex.Next()
	expected := tokens.Token{tokens.EOF, ""}

	if actual != expected {
		t.Fatalf("Eof Failed! expected=%v, got=%v", expected, actual)
	}
}
