package core

type Command struct {
	grammar Grammar
	args    []string
}

func NewCommand(g Grammar, args []string) Command {
	return Command{grammar: g, args: args}
}

func (c Command) Args() []string {
	return c.args
}

func (c Command) Command() Grammar {
	return c.grammar
}
