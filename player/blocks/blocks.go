package blocks

type Block struct {
	Type BlockType
	Body []BodyNode
	ChildBlocks []Block
}

type BlockType uint8

const (
	INVALID BlockType = iota
	END_OF_BLOCKS
	START
	INPUT
	STATE
)

func (bt BlockType) String() string {
	switch bt {
	case INVALID: return "Invalid Block"
	case END_OF_BLOCKS: return "End of Blocks"
	case START: return "Start Block"
	case INPUT: return "Input Block"
	case STATE: return "State Block"
	default: return "Invalid Block Value!"
	}
}

type BodyNode struct {
	Text string
}
