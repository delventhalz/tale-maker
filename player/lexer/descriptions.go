package lexer

import (
	"fmt"
	"strings"
	"tale/tokens"
)

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

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isMinus(r rune) bool {
	return r == '-'
}

func isDot(r rune) bool {
	return r == '.'
}

func isNumberStart(r rune) bool {
	return isDigit(r) || isMinus(r) || isDot(r)
}

func isNumber(r rune) bool {
	return isDigit(r) || r == ',' || r == '_'
}

func isWordStart(r rune) bool {
	return r == '_' ||
		(r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r > 'z' && !isAnyQuote(r))
}

func isWord(r rune) bool {
	return isDigit(r) || isWordStart(r)
}

func isFlag(word string) bool {
	lower := strings.ToLower(word)
	return lower == "yes" ||
		lower == "no" ||
		lower == "on" ||
		lower == "off" ||
		lower == "true" ||
		lower == "false"
}

func getWordToken(word string) tokens.TokenType {
	switch strings.ToLower(word) {
	case "is":
		return tokens.IS
	case "has":
		return tokens.HAS
	case "in":
		return tokens.IN
	case "of":
		return tokens.OF
	case "with":
		return tokens.WITH
	case "and":
		return tokens.AND
	case "or":
		return tokens.OR
	case "not":
		return tokens.NOT
	default:
		if (isFlag(word)) {
			return tokens.FLAG
		}

		return tokens.NAME
	}
}

func isAnyQuote(r rune) bool {
	return r == '"' ||
		r == '\'' ||
		r == '“' ||
		r == '‘' ||
		r == '„' ||
		r == '‚' ||
		r == '”' ||
		r == '’' ||
		r == '«' ||
		r == '‹' ||
		r == '»' ||
		r == '›'
}

func isPaddedStartQuote(first, second rune) bool {
	return second == ' ' && (first == '«' || first == '‹')
}

func isQuote(r rune) bool {
	return r == '"'
}

func isSingleQuote(r rune) bool {
	return r == '\''
}

func isRightQuote(r rune) bool {
	return r == '”'
}

func isSingleRightQuote(r rune) bool {
	return r == '’'
}

func isCurlyQuote(r rune) bool {
	return r == '“' || r == '”'
}

func isSingleCurlyQuote(r rune) bool {
	return r == '‘' || r == '’'
}

func isQuotePadding(r rune) bool {
	return r == ' '
}

func isLeftAngleQuote(r rune) bool {
	return r == '«'
}

func isSingleLeftAngleQuote(r rune) bool {
	return r == '‹'
}

func isRightAngleQuote(r rune) bool {
	return r == '»'
}

func isSingleRightAngleQuote(r rune) bool {
	return r == '›'
}

func getEndQuoteTest(startQuote string) func(rune) bool {
	switch startQuote {
	case "\"":
		return isQuote
	case "'":
		return isSingleQuote
	case "“", "”":
		return isRightQuote
	case "‘", "’":
		return isSingleRightQuote
	case "„":
		return isCurlyQuote
	case "‚":
		return isSingleCurlyQuote
	case "«":
		return isRightAngleQuote
	case "‹":
		return isSingleRightAngleQuote
	case "»":
		return isLeftAngleQuote
	case "›":
		return isSingleLeftAngleQuote
	default:
		panic(fmt.Sprintf("Unknown start quote %q!", startQuote))
	}
}

func getPaddedEndQuoteTests(paddedStartQuote string) (func(rune) bool, func(rune) bool) {
	switch paddedStartQuote {
	case "« ":
		return isQuotePadding, isRightAngleQuote
	case "‹ ":
		return isQuotePadding, isSingleRightAngleQuote
	default:
		panic(fmt.Sprintf("Unknown padded start quote %q!", paddedStartQuote))
	}
}

func isPaddedLeftAngleQuote(quote string) bool {
	return quote == "« "
}

func isSinglePaddedLeftAngleQuote(quote string) bool {
	return quote == "‹ "
}

func isPaddedRightAngleQuote(quote string) bool {
	return quote == " »"
}

func isSinglePaddedRightAngleQuote(quote string) bool {
	return quote == " ›"
}

func isPaddedQuoteStart(quote string) bool {
	return isPaddedLeftAngleQuote(quote) || isSinglePaddedLeftAngleQuote(quote)
}
