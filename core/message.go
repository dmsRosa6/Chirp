package core

import "net"

type Message struct {
	Body   string
	Time   int64
	Origin net.Addr
}
