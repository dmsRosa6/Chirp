package repl

import (
	"bufio"
	"strings"

	"github.com/dmsRosa6/Chirp/core"
	"github.com/dmsRosa6/Chirp/grammar"
	"github.com/dmsRosa6/Chirp/wire"
)

// Handler runs one validated command. It replies through the session; it
// never writes to the socket. A returned error ends the connection.
type Handler func(s *core.Session, args []string) error

var handlers = map[string]Handler{
	"PING":  handlePing,
	"PUB":   handlePub,
	"SUB":   handleSub,
	"UNSUB": handleUnsub,
}

func handlePing(s *core.Session, args []string) error {
	return s.Simple("PONG")
}

func handlePub(s *core.Session, args []string) error {
	subject, payload := args[0], args[1]
	if err := grammar.ValidateSubject(subject); err != nil {
		return s.Err(err.Error())
	}
	s.Broker().Publish(subject, payload)
	return s.OK()
}

func handleSub(s *core.Session, args []string) error {
	pattern, subID := args[0], args[1]
	if err := grammar.ValidatePattern(pattern); err != nil {
		return s.Err(err.Error())
	}
	if err := grammar.ValidateSubID(subID); err != nil {
		return s.Err(err.Error())
	}
	if err := s.Broker().Subscribe(s, pattern, subID); err != nil {
		return s.Err(err.Error())
	}
	return s.OK()
}

func handleUnsub(s *core.Session, args []string) error {
	if err := s.Broker().Unsubscribe(s, args[0]); err != nil {
		return s.Err(err.Error())
	}
	return s.OK()
}

// Serve is the read-eval-reply loop for one connection. It returns when the
// connection ends (io.EOF on a clean disconnect) or a reply can't be queued.
func Serve(s *core.Session, r *bufio.Reader) error {
	for {
		frame, err := wire.ReadCommand(r)
		if err != nil {
			if wire.IsProtocolError(err) {
				_ = s.Err(err.Error()) // best effort; the writer flushes it on Close
			}
			return err
		}
		if len(frame) == 0 {
			continue
		}

		spec, ok := grammar.Lookup(frame[0])
		if !ok {
			if err := s.Err("unknown command: " + strings.ToUpper(frame[0])); err != nil {
				return err
			}
			continue
		}
		args := frame[1:]
		if len(args) != spec.Args {
			if err := s.Err("usage: " + spec.Usage); err != nil {
				return err
			}
			continue
		}

		h, ok := handlers[spec.Name]
		if !ok {
			if err := s.Err("not implemented: " + spec.Name); err != nil {
				return err
			}
			continue
		}
		if err := h(s, args); err != nil {
			return err
		}
	}
}
