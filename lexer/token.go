package lexer

type TokenType int

const (
	TokenOp      TokenType = iota // PUB, SUB, UNSUB, PING, PONG
	TokenArg                      // subject, sub_id
	TokenPayload                  // rest-of-line, only present for PUB, etc
	TokenEOL
)

type Token struct {
	Type  TokenType
	Value string
}
