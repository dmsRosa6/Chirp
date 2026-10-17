package grammar

import "strings"

type Spec struct {
	Name    string // canonical (upper-case) command name
	MinArgs int    // min arguments required after the name
	MaxArgs int    // max arguments required after the name
	Usage   string // shown to the user on a bad call
}

var Commands = map[string]Spec{
	"PING":    {Name: "PING", MinArgs: 0, MaxArgs: 0, Usage: "PING"},
	"PUB":     {Name: "PUB", MinArgs: 2, MaxArgs: 2, Usage: "PUB <pattern> <payload>"},
	"CHANNEL": {Name: "PUB", MinArgs: 2, MaxArgs: 2, Usage: "PUB <subject> <payload>"},
	"SUB":     {Name: "SUB", MinArgs: 2, MaxArgs: 3, Usage: "SUB <pattern> <sub_id> (<group_id>)?"},
	"UNSUB":   {Name: "UNSUB", MinArgs: 2, MaxArgs: 2, Usage: "UNSUB <pattern> <sub_id>"},
	"GET":     {Name: "GET", MinArgs: 1, MaxArgs: 1, Usage: "GET <subject>"},
}

func Lookup(name string) (Spec, bool) {
	spec, ok := Commands[strings.ToUpper(name)]
	return spec, ok
}

const (
	SubjectDelimiter = "."
	SubjectWildcard  = "*"
)
