package lexer

func isLineBreak(r rune) bool {
	return r == '\n' || r == '\r' || r == '\f'
}

func isWindowsLineBreak(first, second rune) bool {
	return first == '\r' && second == '\n'
}

func isNonBreakingSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\v'
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

func isAction(r rune) bool {
	return r == '<'
}

func isActionEnd(r rune) bool {
	return r == '>'
}
