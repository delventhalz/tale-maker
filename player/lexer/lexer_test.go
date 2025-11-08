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
			t.Fatalf(
				"[%d] expected={%v %q %d %d}, got={%v %q %d %d}",
				i,
				exp.Type, exp.Literal, exp.Line, exp.Column,
				act.Type, act.Literal, act.Line, act.Column,
			)
		}
	}
}

func TestEof(t *testing.T) {
	expectTokens(t, "", []tokens.Token{
		{tokens.EOF, "", 1, 1},
	})
}

func TestText(t *testing.T) {
	expectTokens(t, "Hello, world!", []tokens.Token{
		{tokens.TEXT, "Hello, world!", 1, 1},
		{tokens.EOF, "", 1, 14},
	})
}

func TestUnicode(t *testing.T) {
	expectTokens(t, "Hello, 世界!", []tokens.Token{
		{tokens.TEXT, "Hello, 世界!", 1, 1},
		{tokens.EOF, "", 1, 11},
	})
}

func TestBlocks(t *testing.T) {
	input := `
> greet >
Hello!

=world=

>>     greet     >>
Hello, world!

    >>>dramatically

Hello...


    ...world.



== 世界
>>> greet unicode
Hello, 世界!
>>> greet mathematically
If 世界 > world, then greet = hello
    `

	expectTokens(t, input, []tokens.Token{
		{tokens.INPUT_HEADER, ">", 2, 1},
		{tokens.ARG, "greet", 2, 3},
		{tokens.HEADER_END, ">", 2, 9},
		{tokens.TEXT, "Hello!", 3, 1},

		{tokens.STATE_HEADER, "=", 5, 1},
		{tokens.ARG, "world", 5, 2},
		{tokens.HEADER_END, "=", 5, 7},

		{tokens.INPUT_HEADER, ">>", 7, 1},
		{tokens.ARG, "greet", 7, 8},
		{tokens.HEADER_END, ">>", 7, 18},
		{tokens.TEXT, "Hello, world!", 8, 1},

		{tokens.INPUT_HEADER, ">>>", 10, 5},
		{tokens.ARG, "dramatically", 10, 8},
		{tokens.HEADER_END, "\n", 10, 20},
		{tokens.TEXT, "Hello...\n\n\n    ...world.", 12, 1},

		{tokens.STATE_HEADER, "==", 19, 1},
		{tokens.ARG, "世界", 19, 4},
		{tokens.HEADER_END, "\n", 19, 6},

		{tokens.INPUT_HEADER, ">>>", 20, 1},
		{tokens.ARG, "greet", 20, 5},
		{tokens.ARG, "unicode", 20, 11},
		{tokens.HEADER_END, "\n", 20, 18},
		{tokens.TEXT, "Hello, 世界!", 21, 1},

		{tokens.INPUT_HEADER, ">>>", 22, 1},
		{tokens.ARG, "greet", 22, 5},
		{tokens.ARG, "mathematically", 22, 11},
		{tokens.HEADER_END, "\n", 22, 25},
		{tokens.TEXT, "If 世界 > world, then greet = hello", 23, 1},

		{tokens.EOF, "", 24, 5},
	})
}

func TestHeaderAtFileEnd(t *testing.T) {
	expectTokens(t, "Why do this?\n>", []tokens.Token{
		{tokens.TEXT, "Why do this?", 1, 1},
		{tokens.INPUT_HEADER, ">", 2, 1},
		{tokens.HEADER_END, "", 2, 2},
		{tokens.EOF, "", 2, 2},
	})
}

func TestHeaderEndAtFileEnd(t *testing.T) {
	expectTokens(t, "== eof", []tokens.Token{
		{tokens.STATE_HEADER, "==", 1, 1},
		{tokens.ARG, "eof", 1, 4},
		{tokens.HEADER_END, "", 1, 7},
		{tokens.EOF, "", 1, 7},
	})
}

func TestPaddedHeaderEnd(t *testing.T) {
	input := "\t> padded >   \t \nYou should trim your whitespace!"

	expectTokens(t, input, []tokens.Token{
		{tokens.INPUT_HEADER, ">", 1, 2},
		{tokens.ARG, "padded", 1, 4},
		{tokens.HEADER_END, ">", 1, 11},
		{tokens.TEXT, "You should trim your whitespace!", 2, 1},
		{tokens.EOF, "", 2, 33},
	})
}

func TestPaddedText(t *testing.T) {
	input := " \t\n\n\t I love\t\n\n whitespace!!!\t\t \n\n\t \n> respond\nOkay"

	expectTokens(t, input, []tokens.Token{
		// Keep non-breaking whitespace on leading/trailing contentful lines
		{tokens.TEXT, "\t I love\t\n\n whitespace!!!\t\t ", 3, 1},
		{tokens.INPUT_HEADER, ">", 8, 1},
		{tokens.ARG, "respond", 8, 3},
		{tokens.HEADER_END, "\n", 8, 10},
		{tokens.TEXT, "Okay", 9, 1},
		{tokens.EOF, "", 9, 5},
	})
}

func TestCarriageReturns(t *testing.T) {
	expectTokens(t, "\rLet's...\rgo!\r\r> cheer\rRa", []tokens.Token{
		{tokens.TEXT, "Let's...\rgo!", 2, 1},
		{tokens.INPUT_HEADER, ">", 5, 1},
		{tokens.ARG, "cheer", 5, 3},
		{tokens.HEADER_END, "\r", 5, 8},
		{tokens.TEXT, "Ra", 6, 1},
		{tokens.EOF, "", 6, 3},
	})
}

func TestWindowsLineBreaks(t *testing.T) {
	expectTokens(t, "\r\nU wut...\r\nm8?\r\n\r\n> hit\r\nnvm", []tokens.Token{
		{tokens.TEXT, "U wut...\r\nm8?", 2, 1},
		{tokens.INPUT_HEADER, ">", 5, 1},
		{tokens.ARG, "hit", 5, 3},
		{tokens.HEADER_END, "\r\n", 5, 6},
		{tokens.TEXT, "nvm", 6, 1},
		{tokens.EOF, "", 6, 4},
	})
}

func TestWeirdLineBreaks(t *testing.T) {
	expectTokens(t, "\n\rWhere did you get...\n\rthis file?\r\r\n\n> reply\n\rshhh", []tokens.Token{
		{tokens.TEXT, "Where did you get...\n\rthis file?", 3, 1},
		{tokens.INPUT_HEADER, ">", 8, 1},
		{tokens.ARG, "reply", 8, 3},
		{tokens.HEADER_END, "\n", 8, 8},
		{tokens.TEXT, "shhh", 10, 1},
		{tokens.EOF, "", 10, 5},
	})
}

func TestActions(t *testing.T) {
	input := `<name Test>
<name player Tester><place player test>

> run >
<set test_passed>
The test passes!<win_game>

> abort >
The test... <set test_failed> 😞 failed.
<

  lose

       >

Sorry.

`

	expectTokens(t, input, []tokens.Token{
		{tokens.ACTION, "<", 1, 1},
		{tokens.NAME, "name", 1, 2},
		{tokens.ARG, "Test", 1, 7},
		{tokens.ACTION_END, ">", 1, 11},

		{tokens.ACTION, "<", 2, 1},
		{tokens.NAME, "name", 2, 2},
		{tokens.ARG, "player", 2, 7},
		{tokens.ARG, "Tester", 2, 14},
		{tokens.ACTION_END, ">", 2, 20},

		{tokens.ACTION, "<", 2, 21},
		{tokens.NAME, "place", 2, 22},
		{tokens.ARG, "player", 2, 28},
		{tokens.ARG, "test", 2, 35},
		{tokens.ACTION_END, ">", 2, 39},

		{tokens.INPUT_HEADER, ">", 4, 1},
		{tokens.ARG, "run", 4, 3},
		{tokens.HEADER_END, ">", 4, 7},

		{tokens.ACTION, "<", 5, 1},
		{tokens.NAME, "set", 5, 2},
		{tokens.ARG, "test_passed", 5, 6},
		{tokens.ACTION_END, ">", 5, 17},

		{tokens.TEXT, "The test passes!", 6, 1},
		{tokens.ACTION, "<", 6, 17},
		{tokens.NAME, "win_game", 6, 18},
		{tokens.ACTION_END, ">", 6, 26},

		{tokens.INPUT_HEADER, ">", 8, 1},
		{tokens.ARG, "abort", 8, 3},
		{tokens.HEADER_END, ">", 8, 9},

		{tokens.TEXT, "The test... ", 9, 1},
		{tokens.ACTION, "<", 9, 13},
		{tokens.NAME, "set", 9, 14},
		{tokens.ARG, "test_failed", 9, 18},
		{tokens.ACTION_END, ">", 9, 29},
		{tokens.TEXT, " 😞 failed.\n", 9, 30},

		{tokens.ACTION, "<", 10, 1},
		{tokens.NAME, "lose", 12, 3},
		{tokens.ACTION_END, ">", 14, 8},
		{tokens.TEXT, "\n\nSorry.", 14, 9},

		{tokens.EOF, "", 18, 1},
	})
}

func TestFileEndInAction(t *testing.T) {
	expectTokens(t, "<_>", []tokens.Token{
		{tokens.ACTION, "<", 1, 1},
		{tokens.NAME, "_", 1, 2},
		{tokens.ACTION_END, ">", 1, 3},
		{tokens.EOF, "", 1, 4},
	})
	expectTokens(t, "<set", []tokens.Token{
		{tokens.ACTION, "<", 1, 1},
		{tokens.NAME, "set", 1, 2},
		{tokens.EOF, "", 1, 5},
	})
	expectTokens(t, "<set\n\n\n", []tokens.Token{
		{tokens.ACTION, "<", 1, 1},
		{tokens.NAME, "set", 1, 2},
		{tokens.EOF, "", 4, 1},
	})
}
