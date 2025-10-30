package tokens

type TokenType uint8

type Token struct {
	Type TokenType
	Literal string
}

const (
	INVALID TokenType = iota
	EOF

	// Identifiers and values
	NAME
	TEXT
	NUMBER
	FLAG
	ARG // Either NAME or TEXT depending on parser context, may end up removing

	// Delimiters
	INPUT_HEADER
	STATE_HEADER
	HEADER_END
	ACTION
	ENCLOSING_ACTION
	ACTION_END
	INSERT
	INSERT_END
	QUOTE

	// Keywords
	IS
	HAS
	AND
	OR
	NOT
	UNKNOWN
)

func (tt TokenType) String() string {
	switch tt {
	case INVALID: return "Invalid Token"
	case EOF: return "End of File"

	// Identifiers and values
	case NAME: return "Name"
	case TEXT: return "Text"
	case NUMBER: return "Number"
	case FLAG: return "Flag"
	case ARG: return "Argument"

	// Delimiters
	case INPUT_HEADER: return "Input Header"
	case STATE_HEADER: return "State Header"
	case HEADER_END: return "End of Header"
	case ACTION: return "Action"
	case ENCLOSING_ACTION: return "Enclosing Action"
	case ACTION_END: return "End of Action"
	case INSERT: return "Insert"
	case INSERT_END: return "End of Insert"
	case QUOTE: return "Quote"

	// Keywords
	case IS: return "Keyword: is"
	case HAS: return "Keyword: has"
	case AND: return "Keyword: and"
	case OR: return "Keyword: or"
	case NOT: return "Keyword: not"
	case UNKNOWN: return "Keyword: unknown"

	default: return "Invalid Token Value!"
	}
}
