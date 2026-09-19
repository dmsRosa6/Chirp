// Package repl is the read-eval-print loop: decode a command off the
// wire, check its shape against the grammar, run it, write the reply,
// repeat. It leans on two packages that don't know about each other:
// wire (how bytes are framed) and grammar (what commands exist and how
// many args they take). Handlers here only implement what a command
// does — arity is already guaranteed correct by the time a handler runs.
package repl

import (
	"bufio"
	"io"
	"strings"

	"github.com/dmsRosa6/Chirp/grammar"
	"github.com/dmsRosa6/Chirp/wire"
)

// Handler runs a command's args and writes a reply. By the time a Handler
// is called, grammar has already confirmed len(args) == spec.Args.
type Handler func(w io.Writer, args []string) error

var handlers = map[string]Handler{
	"PING":  handlePing,
	"PUB":   handlePub,
	"SUB":   handleSub,
	"UNSUB": handleUnsub,
}

func handlePing(w io.Writer, args []string) error {
	return wire.WriteSimple(w, "PONG")
}

func handlePub(w io.Writer, args []string) error {
	subject, payload := args[0], args[1]
	// TODO(Phase 2): hand off to the subscription registry.
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
