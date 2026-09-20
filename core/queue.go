package core

import "net"

type Queue struct {
	name        string
	fullPath    string
	owner       net.IPAddr //this is here for the future when a node is a owner of a queue
	subscribers map[*Client]struct{}
}
