package tokens

type TokenType uint8

type Token struct {
	Type TokenType
	Literal string
}

const (
	ILLEGAL TokenType = iota
	EOF

	// Identifiers and values
	NAME
	TEXT
	NUMBER
	FLAG
	ARG

	// Delimiters
	INPUT_HEADER
	STATE_HEADER
	ACTION_START
	ACTION_END
	ENCLOSING_ACTION_START
	INSERT_START
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
	case ILLEGAL: return "Illegal Token"
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
	case ACTION_START: return "Action Start"
	case ACTION_END: return "Action End"
	case ENCLOSING_ACTION_START: return "Enclosing Action Start"
	case INSERT_START: return "Insert Start"
	case INSERT_END: return "Insert End"
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
