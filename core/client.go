package core

import (
	"net"
	"sync"
)

type Client struct {
	Conn    net.Conn
	writeMu sync.Mutex
}

func NewClient(conn net.Conn) *Client {
	return &Client{Conn: conn}
}

func (c *Client) Addr() string {
	return c.Conn.RemoteAddr().String()
}
