package lexer

type Lexer struct {
	input     string
	pos       int
	opEmitted bool
}

func NewLexer(line string) *Lexer {
	return &Lexer{input: line}
}

func (l *Lexer) skipSpaces() {
	for l.pos < len(l.input) && l.input[l.pos] == ' ' {
		l.pos++
	}
}

// Next returns the next whitespace-delimited word. The first word is
// tagged TokenOp; every word after that is tagged TokenArg. The lexer
// has no idea what these words mean — that's the parser's job.
func (l *Lexer) Next() (Token, bool) {
	l.skipSpaces()
	if l.pos >= len(l.input) {
		return Token{Type: TokenEOL}, false
	}

	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != ' ' {
		l.pos++
	}
	val := l.input[start:l.pos]

	if !l.opEmitted {
		l.opEmitted = true
		return Token{Type: TokenOp, Value: val}, true
	}
	return Token{Type: TokenArg, Value: val}, true
}

// Remainder returns whatever is left of the input, unsplit and with leading spaces trimmed.
func (l *Lexer) Remainder() string {
	l.skipSpaces()
	if l.pos >= len(l.input) {
		return ""
	}
	rest := l.input[l.pos:]
	l.pos = len(l.input)
	return rest
}
