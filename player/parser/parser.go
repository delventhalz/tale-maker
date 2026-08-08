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
	peek tokens.Token
	atStart bool
}

func (p *Parser) advance() {
	p.next = p.peek
	p.peek = p.lexer.Next()
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
		atStart: true,
	}

	p.next = p.lexer.Next()
	p.peek = p.lexer.Next()

	return p
}

func (p *Parser) Next() blocks.Block {
	var block blocks.Block

	if p.next.Type == tokens.EOF {
		block.Type = blocks.END_OF_BLOCKS
		return block
	}

	if p.atStart {
		block.Type = blocks.START
		p.atStart = false
	}

	for {
		switch p.next.Type {
		case tokens.TEXT:
			block.Body = append(block.Body, blocks.BodyNode{p.next.Literal})

		case tokens.EOF:
			p.advance()
			return block
		}

		p.advance()
	}
}
