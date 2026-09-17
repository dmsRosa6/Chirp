package parser

import (
	"errors"

	"github.com/dmsRosa6/Chirp/core"
	"github.com/dmsRosa6/Chirp/lexer"
)

type Parser struct {
}

func NewParser(input string) *Parser {
	return &Parser{}
}

func (p *Parser) Parse(input string) (core.Command, error) {
	lex := lexer.NewLexer(input)
	var tok lexer.Token
	tok, hasNext := lex.Next()
	if !hasNext {
		return core.Command{}, errors.New("empty command")
	}

	if tok.Type != lexer.TokenOp {
		return core.Command{}, errors.New("unknown command: " + tok.Value)
	}

	arity, ok := core.GrammarArity[g]
	if !ok {
		return core.Command{}, errors.New("missing arity for command: " + tok.Value)
	}

	return core.Command{}, nil
}
