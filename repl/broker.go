package repl

import (
	"io"

	"github.com/dmsRosa6/Chirp/core"
)

type Broker interface {
	CreateQueue(subject string) error
	QueueExists(subject string) bool
	Publish(subject, payload string) error
	Subscribe(c *core.Client, subject string) error
	Unsubscribe(c *core.Client, subject string) error
}

type Session struct {
	W      io.Writer
	Client *core.Client
	Broker Broker
}
