package grammar

import "strings"

type Spec struct {
	Name  string // canonical (upper-case) command name, e.g. "PUB"
	Args  int    // exact number of arguments required after the name
	Usage string // shown to the user on a bad call, e.g. "PUB <subject> <payload>"
}

var Commands = map[string]Spec{
	"PING":   {Name: "PING", Args: 0, Usage: "PING"},
	"CREATE": {Name: "CREATE", Args: 1, Usage: "CREATE <subject>"},
	"EXISTS": {Name: "EXISTS", Args: 1, Usage: "EXISTS <subject>"},
	"PUB":    {Name: "PUB", Args: 2, Usage: "PUB <subject> <payload>"},
	"SUB":    {Name: "SUB", Args: 2, Usage: "SUB <subject> <sub_id>"},
	"UNSUB":  {Name: "UNSUB", Args: 1, Usage: "UNSUB <sub_id>"},
}

func Lookup(name string) (Spec, bool) {
	spec, ok := Commands[strings.ToUpper(name)]
	return spec, ok
}
