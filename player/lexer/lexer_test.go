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
		{tokens.NAME, "greet", 2, 3},
		{tokens.HEADER_END, ">", 2, 9},
		{tokens.TEXT, "Hello!", 3, 1},

		{tokens.STATE_HEADER, "=", 5, 1},
		{tokens.NAME, "world", 5, 2},
		{tokens.HEADER_END, "=", 5, 7},

		{tokens.INPUT_HEADER, ">>", 7, 1},
		{tokens.NAME, "greet", 7, 8},
		{tokens.HEADER_END, ">>", 7, 18},
		{tokens.TEXT, "Hello, world!", 8, 1},

		{tokens.INPUT_HEADER, ">>>", 10, 5},
		{tokens.NAME, "dramatically", 10, 8},
		{tokens.HEADER_END, "\n", 10, 20},
		{tokens.TEXT, "Hello...\n\n\n    ...world.", 12, 1},

		{tokens.STATE_HEADER, "==", 19, 1},
		{tokens.NAME, "世界", 19, 4},
		{tokens.HEADER_END, "\n", 19, 6},

		{tokens.INPUT_HEADER, ">>>", 20, 1},
		{tokens.NAME, "greet", 20, 5},
		{tokens.NAME, "unicode", 20, 11},
		{tokens.HEADER_END, "\n", 20, 18},
		{tokens.TEXT, "Hello, 世界!", 21, 1},

		{tokens.INPUT_HEADER, ">>>", 22, 1},
		{tokens.NAME, "greet", 22, 5},
		{tokens.NAME, "mathematically", 22, 11},
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
		{tokens.NAME, "eof", 1, 4},
		{tokens.HEADER_END, "", 1, 7},
		{tokens.EOF, "", 1, 7},
	})
}

func TestPaddedHeaderEnd(t *testing.T) {
	input := "\t> padded >   \t \nYou should trim your whitespace!"

	expectTokens(t, input, []tokens.Token{
		{tokens.INPUT_HEADER, ">", 1, 2},
		{tokens.NAME, "padded", 1, 4},
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
		{tokens.NAME, "respond", 8, 3},
		{tokens.HEADER_END, "\n", 8, 10},
		{tokens.TEXT, "Okay", 9, 1},
		{tokens.EOF, "", 9, 5},
	})
}

func TestCarriageReturns(t *testing.T) {
	expectTokens(t, "\rLet's...\rgo!\r\r> cheer\rRa", []tokens.Token{
		{tokens.TEXT, "Let's...\rgo!", 2, 1},
		{tokens.INPUT_HEADER, ">", 5, 1},
		{tokens.NAME, "cheer", 5, 3},
		{tokens.HEADER_END, "\r", 5, 8},
		{tokens.TEXT, "Ra", 6, 1},
		{tokens.EOF, "", 6, 3},
	})
}

func TestWindowsLineBreaks(t *testing.T) {
	expectTokens(t, "\r\nU wut...\r\nm8?\r\n\r\n> hit\r\nnvm", []tokens.Token{
		{tokens.TEXT, "U wut...\r\nm8?", 2, 1},
		{tokens.INPUT_HEADER, ">", 5, 1},
		{tokens.NAME, "hit", 5, 3},
		{tokens.HEADER_END, "\r\n", 5, 6},
		{tokens.TEXT, "nvm", 6, 1},
		{tokens.EOF, "", 6, 4},
	})
}

func TestWeirdLineBreaks(t *testing.T) {
	expectTokens(t, "\n\rWhere did you get...\n\rthis file?\r\r\n\n> reply\n\rshhh", []tokens.Token{
		{tokens.TEXT, "Where did you get...\n\rthis file?", 3, 1},
		{tokens.INPUT_HEADER, ">", 8, 1},
		{tokens.NAME, "reply", 8, 3},
		{tokens.HEADER_END, "\n", 8, 8},
		{tokens.TEXT, "shhh", 10, 1},
		{tokens.EOF, "", 10, 5},
	})
}

func TestActions(t *testing.T) {
	input := `<name _ "Test">
<name player "Tester Alice"><place player test>

> run >
<set test_passed>
The test passes!<win_game>
<set score 9000>

> abort >
The test... <set test_passed no> 😞 failed.
<

  lose

       >

Sorry.
<set score 0.1>

`

	expectTokens(t, input, []tokens.Token{
		{tokens.ACTION, "<", 1, 1},
		{tokens.NAME, "name", 1, 2},
		{tokens.NAME, "_", 1, 7},
		{tokens.TEXT, "Test", 1, 9},
		{tokens.ACTION_END, ">", 1, 15},

		{tokens.ACTION, "<", 2, 1},
		{tokens.NAME, "name", 2, 2},
		{tokens.NAME, "player", 2, 7},
		{tokens.TEXT, "Tester Alice", 2, 14},
		{tokens.ACTION_END, ">", 2, 28},

		{tokens.ACTION, "<", 2, 29},
		{tokens.NAME, "place", 2, 30},
		{tokens.NAME, "player", 2, 36},
		{tokens.NAME, "test", 2, 43},
		{tokens.ACTION_END, ">", 2, 47},

		{tokens.INPUT_HEADER, ">", 4, 1},
		{tokens.NAME, "run", 4, 3},
		{tokens.HEADER_END, ">", 4, 7},

		{tokens.ACTION, "<", 5, 1},
		{tokens.NAME, "set", 5, 2},
		{tokens.NAME, "test_passed", 5, 6},
		{tokens.ACTION_END, ">", 5, 17},

		{tokens.TEXT, "The test passes!", 6, 1},
		{tokens.ACTION, "<", 6, 17},
		{tokens.NAME, "win_game", 6, 18},
		{tokens.ACTION_END, ">", 6, 26},

		{tokens.ACTION, "<", 7, 1},
		{tokens.NAME, "set", 7, 2},
		{tokens.NAME, "score", 7, 6},
		{tokens.NUMBER, "9000", 7, 12},
		{tokens.ACTION_END, ">", 7, 16},

		{tokens.INPUT_HEADER, ">", 9, 1},
		{tokens.NAME, "abort", 9, 3},
		{tokens.HEADER_END, ">", 9, 9},

		{tokens.TEXT, "The test... ", 10, 1},
		{tokens.ACTION, "<", 10, 13},
		{tokens.NAME, "set", 10, 14},
		{tokens.NAME, "test_passed", 10, 18},
		{tokens.FLAG, "no", 10, 30},
		{tokens.ACTION_END, ">", 10, 32},
		{tokens.TEXT, " 😞 failed.\n", 10, 33},

		{tokens.ACTION, "<", 11, 1},
		{tokens.NAME, "lose", 13, 3},
		{tokens.ACTION_END, ">", 15, 8},
		{tokens.TEXT, "\n\nSorry.\n", 15, 9},

		{tokens.ACTION, "<", 18, 1},
		{tokens.NAME, "set", 18, 2},
		{tokens.NAME, "score", 18, 6},
		{tokens.NUMBER, "0.1", 18, 12},
		{tokens.ACTION_END, ">", 18, 15},

		{tokens.EOF, "", 20, 1},
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

func _TestTextLiterals(t *testing.T) {
	input := `
<set name "Nosferatu">
<name “Dr. Jekyll and Mr. Hyde” set>
<„House of Wax“>
<set name „Invasion of the Body Snatchers”>
<set name ”Rosemary's Baby”>
<set name « The Wicker Man »>
<set name «Alien»>
<set name »The Evil Dead«>
<set name 'The Thing'>
<set name ‘Dead Alive’>
<set name ’Candyman’>
<set name ‚28 Days Later’>
<set name ‚Ju-on: The Grudge‘>
<set name ’Slither’>
<set name ‹ Let the Right One In ›>
<set name ‹It Follows›>
<set name ›1922‹>
<get_out "">
`

	expectTokens(t, input, []tokens.Token{
		{tokens.ACTION, "<", 2, 1},
		{tokens.NAME, "set", 2, 2},
		{tokens.NAME, "name", 2, 6},
		{tokens.TEXT, "Nosferatu", 2, 11},
		{tokens.ACTION_END, ">", 2, 22},

		{tokens.ACTION, "<", 3, 1},
		{tokens.NAME, "name", 3, 2},
		{tokens.TEXT, "Dr. Jekyll and Mr. Hyde", 3, 7},
		{tokens.NAME, "set", 3, 33},
		{tokens.ACTION_END, ">", 3, 36},

		{tokens.ACTION, "<", 4, 1},
		{tokens.TEXT, "House of Wax", 4, 2},
		{tokens.ACTION_END, ">", 4, 16},

		{tokens.ACTION, "<", 5, 1},
		{tokens.NAME, "set", 5, 2},
		{tokens.NAME, "name", 5, 6},
		{tokens.TEXT, "Invasion of the Body Snatchers", 5, 11},
		{tokens.ACTION_END, ">", 5, 43},

		{tokens.ACTION, "<", 6, 1},
		{tokens.NAME, "set", 6, 2},
		{tokens.NAME, "name", 6, 6},
		{tokens.TEXT, "Rosemary's Baby", 6, 28},
		{tokens.ACTION_END, ">", 6, 22},

		{tokens.ACTION, "<", 7, 1},
		{tokens.NAME, "set", 7, 2},
		{tokens.NAME, "name", 7, 6},
		{tokens.TEXT, "The Wicker Man", 7, 11},
		{tokens.ACTION_END, ">", 7, 29},

		{tokens.ACTION, "<", 8, 1},
		{tokens.NAME, "set", 8, 2},
		{tokens.NAME, "name", 8, 6},
		{tokens.TEXT, "Alien", 8, 11},
		{tokens.ACTION_END, ">", 8, 18},

		{tokens.ACTION, "<", 9, 1},
		{tokens.NAME, "set", 9, 2},
		{tokens.NAME, "name", 9, 6},
		{tokens.TEXT, "The Evil Dead", 9, 11},
		{tokens.ACTION_END, ">", 9, 26},

		{tokens.ACTION, "<", 10, 1},
		{tokens.NAME, "set", 10, 2},
		{tokens.NAME, "name", 10, 6},
		{tokens.TEXT, "The Thing", 10, 11},
		{tokens.ACTION_END, ">", 10, 22},

		{tokens.ACTION, "<", 11, 1},
		{tokens.NAME, "set", 11, 2},
		{tokens.NAME, "name", 11, 6},
		{tokens.TEXT, "Dead Alive", 11, 11},
		{tokens.ACTION_END, ">", 11, 23},

		{tokens.ACTION, "<", 12, 1},
		{tokens.NAME, "set", 12, 2},
		{tokens.NAME, "name", 12, 6},
		{tokens.TEXT, "Candyman", 12, 11},
		{tokens.ACTION_END, ">", 12, 21},

		{tokens.ACTION, "<", 13, 1},
		{tokens.NAME, "set", 13, 2},
		{tokens.NAME, "name", 13, 6},
		{tokens.TEXT, "28 Days Later", 13, 11},
		{tokens.ACTION_END, ">", 13, 26},

		{tokens.ACTION, "<", 14, 1},
		{tokens.NAME, "set", 14, 2},
		{tokens.NAME, "name", 14, 6},
		{tokens.TEXT, "Ju-on: The Grudge", 14, 11},
		{tokens.ACTION_END, ">", 14, 30},

		{tokens.ACTION, "<", 15, 1},
		{tokens.NAME, "set", 15, 2},
		{tokens.NAME, "name", 15, 6},
		{tokens.TEXT, "Slither", 15, 11},
		{tokens.ACTION_END, ">", 15, 20},

		{tokens.ACTION, "<", 16, 1},
		{tokens.NAME, "set", 16, 2},
		{tokens.NAME, "name", 16, 6},
		{tokens.TEXT, "Let the Right One In", 16, 11},
		{tokens.ACTION_END, ">", 16, 35},

		{tokens.ACTION, "<", 17, 1},
		{tokens.NAME, "set", 17, 2},
		{tokens.NAME, "name", 17, 6},
		{tokens.TEXT, "It Follows", 17, 11},
		{tokens.ACTION_END, ">", 17, 23},

		{tokens.ACTION, "<", 18, 1},
		{tokens.NAME, "set", 18, 2},
		{tokens.NAME, "name", 18, 6},
		{tokens.TEXT, "1922", 18, 11},
		{tokens.ACTION_END, ">", 18, 17},

		{tokens.ACTION, "<", 19, 1},
		{tokens.NAME, "get_out", 19, 2},
		{tokens.TEXT, "", 19, 10},
		{tokens.ACTION_END, ">", 19, 12},

		{tokens.EOF, "", 20, 1},
	})
}

func _TestNumberLiterals(t *testing.T) {
	input := `
<set score 0>
<score 0.0 set>
<1>
<set score .09>
<set score 1234567890>
<set score 0123456789>
<set score 1234567.89>
<set score 1,234,567.89>
<set score 1_234_567.89>
`

	expectTokens(t, input, []tokens.Token{
		{tokens.ACTION, "<", 2, 1},
		{tokens.NAME, "set", 2, 2},
		{tokens.NAME, "score", 2, 6},
		{tokens.NUMBER, "0", 2, 12},
		{tokens.ACTION_END, ">", 2, 13},

		{tokens.ACTION, "<", 3, 1},
		{tokens.NAME, "score", 3, 2},
		{tokens.NUMBER, "0.0", 3, 8},
		{tokens.NAME, "set", 3, 12},
		{tokens.ACTION_END, ">", 3, 15},

		{tokens.ACTION, "<", 4, 1},
		{tokens.NUMBER, "1", 4, 2},
		{tokens.ACTION_END, ">", 4, 3},

		{tokens.ACTION, "<", 5, 1},
		{tokens.NAME, "set", 5, 2},
		{tokens.NAME, "score", 5, 6},
		{tokens.NUMBER, ".09", 5, 12},
		{tokens.ACTION_END, ">", 5, 15},

		{tokens.ACTION, "<", 6, 1},
		{tokens.NAME, "set", 6, 2},
		{tokens.NAME, "score", 6, 6},
		{tokens.NUMBER, "1234567890", 6, 12},
		{tokens.ACTION_END, ">", 6, 22},

		{tokens.ACTION, "<", 7, 1},
		{tokens.NAME, "set", 7, 2},
		{tokens.NAME, "score", 7, 6},
		{tokens.NUMBER, "0123456789", 7, 12},
		{tokens.ACTION_END, ">", 7, 22},

		{tokens.ACTION, "<", 8, 1},
		{tokens.NAME, "set", 8, 2},
		{tokens.NAME, "score", 8, 6},
		{tokens.NUMBER, "1234567.89", 8, 12},
		{tokens.ACTION_END, ">", 8, 22},

		{tokens.ACTION, "<", 9, 1},
		{tokens.NAME, "set", 9, 2},
		{tokens.NAME, "score", 9, 6},
		{tokens.NUMBER, "1,234,567.89", 9, 12},
		{tokens.ACTION_END, ">", 9, 24},

		{tokens.ACTION, "<", 10, 1},
		{tokens.NAME, "set", 10, 2},
		{tokens.NAME, "score", 10, 6},
		{tokens.NUMBER, "1_234_567.89", 10, 12},
		{tokens.ACTION_END, ">", 10, 24},

		{tokens.EOF, "", 11, 1},
	})
}

func _TestFlagLiterals(t *testing.T) {
	input := `
<set lights on>
<set off lights>
<yes>
<set lights no>
<set lights true>
<set lights false>
`

	expectTokens(t, input, []tokens.Token{
		{tokens.ACTION, "<", 2, 1},
		{tokens.NAME, "set", 2, 2},
		{tokens.NAME, "lights", 2, 6},
		{tokens.FLAG, "on", 2, 13},
		{tokens.ACTION_END, ">", 2, 15},

		{tokens.ACTION, "<", 3, 1},
		{tokens.NAME, "set", 3, 2},
		{tokens.FLAG, "off", 3, 6},
		{tokens.NAME, "lights", 3, 10},
		{tokens.ACTION_END, ">", 3, 16},

		{tokens.ACTION, "<", 4, 1},
		{tokens.FLAG, "yes", 4, 2},
		{tokens.ACTION_END, ">", 4, 5},

		{tokens.ACTION, "<", 5, 1},
		{tokens.NAME, "set", 5, 2},
		{tokens.NAME, "lights", 5, 6},
		{tokens.FLAG, "no", 5, 13},
		{tokens.ACTION_END, ">", 5, 15},

		{tokens.ACTION, "<", 6, 1},
		{tokens.NAME, "set", 6, 2},
		{tokens.NAME, "lights", 6, 6},
		{tokens.FLAG, "true", 6, 13},
		{tokens.ACTION_END, ">", 6, 17},

		{tokens.ACTION, "<", 7, 1},
		{tokens.NAME, "set", 7, 2},
		{tokens.NAME, "lights", 7, 6},
		{tokens.FLAG, "false", 7, 13},
		{tokens.ACTION_END, ">", 7, 18},

		{tokens.EOF, "", 8, 1},
	})
}

func _TestKeywords(t *testing.T) {
	input := `
= room has door and door of room is broken =

>> any >>
<set player in detention>
What did you do to the door!?

=== player with teacher or player is not chastised ===
Think about what you've done!

=== repeat ===
Must we go over this again?
`

	expectTokens(t, input, []tokens.Token{
		{tokens.STATE_HEADER, "=", 2, 1},
		{tokens.NAME, "room", 2, 3},
		{tokens.HAS, "has", 2, 8},
		{tokens.NAME, "door", 2, 12},
		{tokens.AND, "and", 2, 17},
		{tokens.NAME, "door", 2, 21},
		{tokens.OF, "of", 2, 26},
		{tokens.NAME, "room", 2, 29},
		{tokens.IS, "is", 2, 34},
		{tokens.NAME, "broken", 2, 37},
		{tokens.HEADER_END, "=", 2, 44},

		{tokens.INPUT_HEADER, ">>", 4, 1},
		{tokens.HAS, "any", 4, 4},
		{tokens.HEADER_END, ">>", 4, 8},

		{tokens.ACTION, "<", 5, 1},
		{tokens.NAME, "set", 5, 2},
		{tokens.NAME, "player", 5, 6},
		{tokens.IN, "in", 5, 13},
		{tokens.NAME, "detention", 5, 16},
		{tokens.ACTION_END, ">", 5, 25},

		{tokens.TEXT, "What did you do to the door!?", 6, 1},

		{tokens.STATE_HEADER, "===", 8, 1},
		{tokens.NAME, "player", 8, 5},
		{tokens.WITH, "with", 8, 12},
		{tokens.NAME, "teacher", 8, 17},
		{tokens.OR, "or", 8, 25},
		{tokens.NAME, "player", 8, 28},
		{tokens.IS, "is", 8, 35},
		{tokens.NOT, "not", 8, 38},
		{tokens.NAME, "chastised", 8, 42},
		{tokens.HEADER_END, "===", 8, 52},

		{tokens.TEXT, "Think about what you've done!", 9, 1},

		{tokens.STATE_HEADER, "===", 11, 1},
		{tokens.NAME, "repeat", 11, 5},
		{tokens.HEADER_END, "===", 11, 12},

		{tokens.TEXT, "Must we go over this again?", 12, 1},

		{tokens.EOF, "", 13, 1},
	})
}
