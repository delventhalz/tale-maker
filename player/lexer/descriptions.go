package lexer

import (
	"fmt"
	"tale/tokens"
)

func isEof(r rune) bool {
	return r == 0
}

func isLineBreak(r rune) bool {
	return r == '\n' || r == '\r' || r == '\f'
}

func isWindowsBreakStart(r rune) bool {
	return r == '\r'
}

func isWindowsBreakEnd(r rune) bool {
	return r == '\n'
}

func isEndOfLine(r rune) bool {
	return isEof(r) || isLineBreak(r)
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
	return r == '{'
}

func isEnclosingMarker(r rune) bool {
	return r == '/'
}

func getActionToken(action string) tokens.TokenType {
	switch action {
	case "{/":
		return tokens.ENCLOSING_ACTION
	case "{":
		return tokens.ACTION
	default:
		panic(fmt.Sprintf("Unknown action %q!", action))
	}
}

func isActionEnd(r rune) bool {
	return r == '}'
}

func isComment(r rune) bool {
	return r == '{'
}

func isCommentMarker(r rune) bool {
	return r == '!'
}

func isCommentEnd(r rune) bool {
	return r == '}'
}

func isOperator(r rune) bool {
	return r == ':' ||
		r == '+' ||
		r == '-' ||
		r == '*' ||
		r == '/' ||
		r == '%' ||
		r == '>' ||
		r == '<' ||
		r == '(' ||
		r == ')'
}

func isEqualsableOperator(r rune) bool {
	return r == '>' || r == '<'
}

func isEqualsMarker(r rune) bool {
	return r == '='
}

func getOperatorToken(operator string) tokens.TokenType {
	switch operator {
	case ":":
		return tokens.COLON
	case "+":
		return tokens.PLUS
	case "-":
		return tokens.MINUS
	case "*":
		return tokens.MULTIPLY
	case "/":
		return tokens.DIVIDE
	case "%":
		return tokens.REMAINDER
	case ">":
		return tokens.GT
	case "<":
		return tokens.LT
	case ">=":
		return tokens.GTE
	case "<=":
		return tokens.LTE
	case "(":
		return tokens.PAREN
	case ")":
		return tokens.PAREN_END
	default:
		return tokens.INVALID
	}
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isDot(r rune) bool {
	return r == '.'
}

func isNumberStart(r rune) bool {
	return isDigit(r) || isDot(r)
}

func isNumber(r rune) bool {
	return isDigit(r) || r == ',' || r == '_'
}

func isWordStart(r rune) bool {
	return r == '_' ||
		(r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r > 127 && !isAnyQuote(r))
}

func isWord(r rune) bool {
	return isDigit(r) || isWordStart(r)
}

func isFlag(word string) bool {
	return word == "yes" ||
		word == "no" ||
		word == "on" ||
		word == "off" ||
		word == "true" ||
		word == "false"
}

func getWordToken(word string) tokens.TokenType {
	switch word {
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
	case "it":
		return tokens.IT
	default:
		if (isFlag(word)) {
			return tokens.FLAG
		}

		return tokens.NAME
	}
}

func isEscapeStart(r rune) bool {
	return r == '\\'
}

func getEscaped(r rune) string {
	switch r {
	case 's':
		return " "
	case 'n':
		return "\n"
	case 'r':
		return "\r"
	case 't':
		return "\t"
	case 'v':
		return "\v"
	case 'f':
		return "\f"
	default:
		return string(r)
	}
}

func isAnyQuote(r rune) bool {
	return r == '"' ||
		r == '\'' ||
		r == '`' ||
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

func isQuote(r rune) bool {
	return r == '"'
}

func isSingleQuote(r rune) bool {
	return r == '\''
}

func isBacktick(r rune) bool {
	return r == '`'
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

func isPaddableStartQuote(r rune) bool {
	return r == '«' || r == '‹'
}

func isQuotePadding(r rune) bool {
	return r == ' '
}

func getEndQuoteTest(r rune) func(rune) bool {
	switch r {
	case '"':
		return isQuote
	case '\'':
		return isSingleQuote
	case '`':
		return isBacktick
	case '“', '”':
		return isRightQuote
	case '‘', '’':
		return isSingleRightQuote
	case '„':
		return isCurlyQuote
	case '‚':
		return isSingleCurlyQuote
	case '«':
		return isRightAngleQuote
	case '‹':
		return isSingleRightAngleQuote
	case '»':
		return isLeftAngleQuote
	case '›':
		return isSingleLeftAngleQuote
	default:
		panic(fmt.Sprintf("Unknown start quote %q!", r))
	}
}
