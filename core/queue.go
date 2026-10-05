package core

import (
	"net"
	"strings"

	"github.com/dmsRosa6/Chirp/grammar"
)

type DeadLetterEntry struct {
	Client  *Client
	Message Message
}

type Queue struct {
	name        string
	fullPath    string
	deadLetter  []DeadLetterEntry
	owner       *net.IPAddr //this is here for the future when a node is a owner of a queue
	subscribers map[*Client]struct{}
}

func NewQueue(fullPath string) *Queue {
	pathParts := strings.Split(fullPath, grammar.QUEUE_PATH_DELIMITER)
	name := pathParts[len(pathParts)-1]
	return &Queue{
		name:        name,
		fullPath:    fullPath,
		owner:       nil,
		subscribers: make(map[*Client]struct{}),
	}
}

// TODO
func (q *Queue) Add(msg Message) error {
	panic("Too implement")
}

func (q *Queue) Subscribe(c *Client) error {
	panic("Too implement")
}

func (q *Queue) Unsubscribe(c *Client) error {
	panic("Too implement")
}
