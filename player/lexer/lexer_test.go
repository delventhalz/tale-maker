package lexer

import (
    "tale/tokens"
    "testing"
)

func expectTokens(t *testing.T, input string, expected []tokens.Token) {
    lex := New(input)

    for i, exp := range expected {
        act := lex.Next()
        if act != exp {
            t.Fatalf("[%d] expected=%v, got=%v", i, exp, act)
        }
    }
}

func TestEof(t *testing.T) {
    expectTokens(t, "", []tokens.Token{
        {tokens.EOF, ""},
    })
}

func TestText(t *testing.T) {
    expectTokens(t, "Hello, world!", []tokens.Token{
        {tokens.TEXT, "Hello, world!"},
        {tokens.EOF, ""},
    })
}

func TestUnicode(t *testing.T) {
    expectTokens(t, "Hello, 世界!", []tokens.Token{
        {tokens.TEXT, "Hello, 世界!"},
        {tokens.EOF, ""},
    })
}

func TestTrim(t *testing.T) {
    expectTokens(t, "\n\n    Hello, \nworld!\n\t", []tokens.Token{
        {tokens.TEXT, "Hello, \nworld!"},
        {tokens.EOF, ""},
    })
}
