package repl

import (
	"bufio"
	"io"
	"strings"

	"github.com/dmsRosa6/Chirp/grammar"
	"github.com/dmsRosa6/Chirp/wire"
)

type Handler func(w io.Writer, args []string) error

var handlers = map[string]Handler{
	"PING":   handlePing,
	"CREATE": handleCreate,
	"EXISTS": handleExists,
	"PUB":    handlePub,
	"SUB":    handleSub,
	"UNSUB":  handleUnsub,
}

func handlePing(w io.Writer, args []string) error {
	return wire.WriteSimple(w, "PONG")
}

func handleCreate(w io.Writer, args []string) error {
	return wire.WriteSimple(w, "PONG")
}

func handleExists(w io.Writer, args []string) error {
	return wire.WriteSimple(w, "PONG")
}

func handlePub(w io.Writer, args []string) error {
	subject, payload := args[0], args[1]
	_ = subject
	_ = payload
	return wire.WriteOK(w)
}

func handleSub(w io.Writer, args []string) error {
	subject, subID := args[0], args[1]
	_ = subject
	_ = subID
	return wire.WriteOK(w)
}

func handleUnsub(w io.Writer, args []string) error {
	subID := args[0]
	_ = subID
	return wire.WriteOK(w)
}

// Serve is the read-eval-print loop.
func Serve(r *bufio.Reader, w io.Writer) error {
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
			if err := wire.WriteError(w, "unknown command: "+name); err != nil {
				return err
			}
			continue
		}
		if len(args) != spec.Args {
			if err := wire.WriteError(w, "usage: "+spec.Usage); err != nil {
				return err
			}
			continue
		}

		h, ok := handlers[name]
		if !ok {
			// grammar knows this command but no handler is wired up for it —
			// a programming error, not a user error.
			if err := wire.WriteError(w, "not implemented: "+name); err != nil {
				return err
			}
			continue
		}
		if err := h(w, args); err != nil {
			return err
		}
	}
}
