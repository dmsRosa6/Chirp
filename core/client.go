package core

import (
	"net"
	"sync"
)

type Client struct {
	conn net.Conn
	mu   sync.Mutex
	subs map[string]*Queue
}

func NewClient(conn net.Conn) *Client {
	return &Client{conn: conn, subs: make(map[string]*Queue)}
}

func (c *Client) TrackSub(id string, q *Queue) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.subs[id]; ok {
		return false
	}
	c.subs[id] = q
	return true
}

func (c *Client) UntrackSub(id string) (*Queue, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	q, ok := c.subs[id]
	delete(c.subs, id)
	return q, ok
}

func (c *Client) DropAllSubs() []*Queue {
	c.mu.Lock()
	defer c.mu.Unlock()
	qs := make([]*Queue, 0, len(c.subs))
	for _, q := range c.subs {
		qs = append(qs, q)
	}
	c.subs = make(map[string]*Queue)
	return qs
}
