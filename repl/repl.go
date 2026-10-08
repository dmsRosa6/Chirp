package repl

import (
	"bufio"
	"strings"

	"github.com/dmsRosa6/Chirp/grammar"
	"github.com/dmsRosa6/Chirp/wire"
)

type Handler func(s *Session, args []string) error

var handlers = map[string]Handler{
	"PING":   handlePing,
	"CREATE": handleCreate,
	"EXISTS": handleExists,
	"PUB":    handlePub,
	"SUB":    handleSub,
	"UNSUB":  handleUnsub,
}

func handlePing(s *Session, args []string) error {
	return wire.WriteSimple(s.W, "PONG")
}

func handleCreate(s *Session, args []string) error {
	if err := s.Broker.CreateQueue(args[0]); err != nil {
		return wire.WriteError(s.W, err.Error())
	}
	return wire.WriteOK(s.W)
}

func handleExists(s *Session, args []string) error {
	if s.Broker.QueueExists(args[0]) {
		return wire.WriteSimple(s.W, "1")
	}
	return wire.WriteSimple(s.W, "0")
}

func handlePub(s *Session, args []string) error {
	if err := s.Broker.Publish(args[0], args[1]); err != nil {
		return wire.WriteError(s.W, err.Error())
	}
	return wire.WriteOK(s.W)
}

func handleSub(s *Session, args []string) error {
	if err := s.Broker.Subscribe(s.Client, args[0]); err != nil {
		return wire.WriteError(s.W, err.Error())
	}
	return wire.WriteOK(s.W)
}

func handleUnsub(s *Session, args []string) error {
	if err := s.Broker.Unsubscribe(s.Client, args[0]); err != nil {
		return wire.WriteError(s.W, err.Error())
	}
	return wire.WriteOK(s.W)
}

// Serve is the read-eval-print loop.
func Serve(r *bufio.Reader, s *Session) error {
	for {
		frame, err := wire.ReadCommand(r)
		if err != nil {
			return err // includes io.EOF on clean disconnect
		}
		if len(frame) == 0 {
			continue
		}
		name, args := strings.ToUpper(frame[0]), frame[1:]

		spec, ok := grammar.Lookup(name)
		if !ok {
			if err := wire.WriteError(s.W, "unknown command: "+name); err != nil {
				return err
			}
			continue
		}
		if len(args) != spec.Args {
			if err := wire.WriteError(s.W, "usage: "+spec.Usage); err != nil {
				return err
			}
			continue
		}

		h, ok := handlers[name]
		if !ok {
			if err := wire.WriteError(s.W, "not implemented: "+name); err != nil {
				return err
			}
			continue
		}
		if err := h(s, args); err != nil {
			return err
		}
	}
}
