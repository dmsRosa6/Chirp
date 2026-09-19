package core

type Queue struct {
	name        string
	fullPath    string
	subscribers map[*Client]struct{}
}

type IndexNode struct {
	isQueue bool
	queue   *Queue
}
