package grammar

import "strings"

type Spec struct {
	Name  string
	Args  int
	Usage string
}

var Commands = map[string]Spec{
	"PING":  {Name: "PING", Args: 0, Usage: "PING"},
	"PUB":   {Name: "PUB", Args: 2, Usage: "PUB <subject> <payload>"},
	"SUB":   {Name: "SUB", Args: 2, Usage: "SUB <subject> <sub_id>"},
	"UNSUB": {Name: "UNSUB", Args: 1, Usage: "UNSUB <sub_id>"},
}

func Lookup(name string) (Spec, bool) {
	spec, ok := Commands[strings.ToUpper(name)]
	return spec, ok
}
