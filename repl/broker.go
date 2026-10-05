package repl

import (
	"io"

	"github.com/dmsRosa6/Chirp/core"
)

type Broker interface {
	CreateQueue(subject string) error
	QueueExists(subject string) bool
	Publish(subject, payload string) error
	Subscribe(c *core.Client, subject, subID string) error
	Unsubscribe(c *core.Client, subID string) error
	DropClient(c *core.Client)
}

type Session struct {
	W      io.Writer
	Client *core.Client
	Broker Broker
}
