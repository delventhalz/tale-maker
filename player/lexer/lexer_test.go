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
            t.Fatalf("[%d] expected=%q, got=%q", i, exp, act)
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

func TestBlocks(t *testing.T) {
    input := `
> greet >
Hello!

= world =

>>     greet     >>
Hello, world!

    >>>dramatically

Hello...


    ...world.



== 世界
>>> greet unicode
Hello, 世界!
    `

    expectTokens(t, input, []tokens.Token{
        {tokens.INPUT_HEADER, ">"},
        {tokens.ARG, "greet"},
        {tokens.HEADER_END, ">"},
        {tokens.TEXT, "Hello!"},

        {tokens.STATE_HEADER, "="},
        {tokens.ARG, "world"},
        {tokens.HEADER_END, "="},

        {tokens.INPUT_HEADER, ">>"},
        {tokens.ARG, "greet"},
        {tokens.HEADER_END, ">>"},
        {tokens.TEXT, "Hello, world!"},

        {tokens.INPUT_HEADER, ">>>"},
        {tokens.ARG, "dramatically"},
        {tokens.HEADER_END, "\n"},
        {tokens.TEXT, "Hello...\n\n\n    ...world."},

        {tokens.STATE_HEADER, "=="},
        {tokens.ARG, "世界"},
        {tokens.HEADER_END, "\n"},

        {tokens.INPUT_HEADER, ">>>"},
        {tokens.ARG, "greet"},
        {tokens.ARG, "unicode"},
        {tokens.HEADER_END, "\n"},
        {tokens.TEXT, "Hello, 世界!"},

        {tokens.EOF, ""},
    })
}

func TestHeaderAtFileEnd(t *testing.T) {
    expectTokens(t, "Why do this?\n>", []tokens.Token{
        {tokens.TEXT, "Why do this?"},
        {tokens.INPUT_HEADER, ">"},
        {tokens.HEADER_END, ""},
        {tokens.EOF, ""},
    })
}

func TestHeaderEndAtFileEnd(t *testing.T) {
    expectTokens(t, "== eof", []tokens.Token{
        {tokens.STATE_HEADER, "=="},
        {tokens.ARG, "eof"},
        {tokens.HEADER_END, ""},
        {tokens.EOF, ""},
    })
}

func TestPaddedHeaderEnd(t *testing.T) {
	input := "\t> padded >   \t \nYou should trim your whitespace!"

    expectTokens(t, input, []tokens.Token{
        {tokens.INPUT_HEADER, ">"},
        {tokens.ARG, "padded"},
        {tokens.HEADER_END, ">"},
        {tokens.TEXT, "You should trim your whitespace!"},
        {tokens.EOF, ""},
    })
}

func TestPaddedText(t *testing.T) {
	input := " \t\n\n\t I love\t\n\n whitespace!!!\t\t \n\n\t \n> respond\nOkay"

    expectTokens(t, input, []tokens.Token{
        // Keep non-breaking whitespace on leading/trailing contentful lines
        {tokens.TEXT, "\t I love\t\n\n whitespace!!!\t\t "},
        {tokens.INPUT_HEADER, ">"},
        {tokens.ARG, "respond"},
        {tokens.HEADER_END, "\n"},
        {tokens.TEXT, "Okay"},
        {tokens.EOF, ""},
    })
}
