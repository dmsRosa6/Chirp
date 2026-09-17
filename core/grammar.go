package core

type Grammar int

const (
	PUB Grammar = iota
	SUB
	PING
)

var GrammarOps = map[string]Grammar{
	"PUB":  PUB,
	"SUB":  SUB,
	"PING": PING,
}

var GrammarArity = map[Grammar]int{
	PUB:  1, // subject; payload is everything after
	SUB:  2, // subject, sub_id — no trailing payload
	PING: 0,
}

const SEPARATOR = " "
