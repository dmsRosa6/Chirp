package core

import (
	"net"
	"strings"
	"sync"
	"time"

	"github.com/dmsRosa6/Chirp/grammar"
)

type DeadLetterEntry struct {
	Client  *Client
	Message Message
	Err     error
}

type Queue struct {
	name        string
	fullPath    string
	deadLetter  []DeadLetterEntry
	owner       *net.IPAddr
	subscribers map[string]*Client
	mu          sync.RWMutex
	dlMu        sync.Mutex
}

func NewQueue(fullPath string) *Queue {
	pathParts := strings.Split(fullPath, grammar.QUEUE_PATH_DELIMITER)
	return &Queue{
		name:        pathParts[len(pathParts)-1],
		fullPath:    fullPath,
		subscribers: make(map[string]*Client),
	}

}
func (c *Client) Send(payload []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := c.Conn.Write(payload)
	return err
}

// Flood Publish on the queue
func (q *Queue) Publish(msg Message) error {
	q.mu.RLock()
	subs := make([]*Client, 0, len(q.subscribers))
	for _, c := range q.subscribers {
		subs = append(subs, c)
	}
	q.mu.RUnlock()

	payload := []byte(msg.Body + "\n")

	var failed []DeadLetterEntry
	for _, cli := range subs {
		if err := cli.Send(payload); err != nil {
			failed = append(failed, DeadLetterEntry{Client: cli, Message: msg, Err: err})
		}
	}

	if len(failed) > 0 {
		q.dlMu.Lock()
		q.deadLetter = append(q.deadLetter, failed...)
		q.dlMu.Unlock()
	}
	return nil
}

func (q *Queue) Subscribe(c *Client) error {
	panic("Too implement")
}

func (q *Queue) Unsubscribe(c *Client) error {
	panic("Too implement")
}
