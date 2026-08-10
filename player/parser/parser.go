package parser

import (
	"log"
	"os"
	"tale/lexer"
	"tale/blocks"
	"tale/tokens"
)

type Parser struct {
	path string
	input string
	lexer *lexer.Lexer
	next tokens.Token
	blockCount uint
}

func (p *Parser) advance() {
	p.next = p.lexer.Next()
}

func (p *Parser) atBlockEnd() bool {
	return p.next.Type == tokens.EOF ||
		p.next.Type == tokens.INPUT_HEADER ||
		p.next.Type == tokens.STATE_HEADER
}

func (p *Parser) atNestedBlockStart(depth int) bool {
	return p.next.Type == tokens.INPUT_HEADER &&
		depth > 0 &&
		depth < len(p.next.Literal)
}

func (p *Parser) parseInputHeader() []blocks.Expression {
	var header []blocks.Expression
	p.advance()

	for p.next.Type != tokens.HEADER_END {
		header = append(header, blocks.Expression{Token: p.next})
		p.advance()
	}

	p.advance()
	return header
}

func New(absTalePath string) *Parser {
	taleBytes, err := os.ReadFile(absTalePath)
	if err != nil {
		log.Fatal(err)
	}

	input := string(taleBytes)

	p := &Parser{
		path: absTalePath,
		input: input,
		lexer: lexer.New(input),
		blockCount: 0,
	}

	p.next = p.lexer.Next()
	return p
}

func (p *Parser) Next() blocks.Block {
	var block blocks.Block
	depth := 0

	if p.next.Type == tokens.EOF {
		block.Type = blocks.END_OF_BLOCKS
		return block
	}

	if p.next.Type == tokens.INPUT_HEADER {
		depth = len(p.next.Literal)
		block.Type = blocks.INPUT
		block.Header = p.parseInputHeader()
	}

	if block.Type == 0 && p.blockCount == 0 {
		block.Type = blocks.START
	}
	p.blockCount += 1

	for !p.atBlockEnd() {
		switch p.next.Type {
		case tokens.TEXT:
			block.Body = append(block.Body, blocks.BodyNode{p.next.Literal})
		}
		p.advance()
	}

	for p.atNestedBlockStart(depth) {
		block.ChildBlocks = append(block.ChildBlocks, p.Next())
	}
	return block
}
