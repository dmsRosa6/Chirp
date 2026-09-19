package repl

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/dmsRosa6/Chirp/grammar"
	"github.com/dmsRosa6/Chirp/wire"
)

func TestHandlersMatchGrammar(t *testing.T) {
	for name := range grammar.Commands {
		if _, ok := handlers[name]; !ok {
			t.Errorf("grammar defines %s but no handler is registered", name)
		}
	}
	for name := range handlers {
		if _, ok := grammar.Commands[name]; !ok {
			t.Errorf("handler registered for %s but grammar doesn't define it", name)
		}
	}
}

func TestServeDispatch(t *testing.T) {
	var in bytes.Buffer
	wire.WriteCommand(&in, "PING")
	wire.WriteCommand(&in, "PUB", "foo", "hi")
	wire.WriteCommand(&in, "SUB", "foo", "sub1")
	wire.WriteCommand(&in, "UNSUB", "sub1")
	wire.WriteCommand(&in, "BOGUS")

	var out bytes.Buffer
	if err := Serve(bufio.NewReader(&in), &out); err == nil {
		t.Fatal("expected EOF at end of input")
	}

	got := out.String()
	for _, want := range []string{"+PONG", "+OK", "+OK", "+OK", "-ERR unknown command: BOGUS"} {
		if !strings.Contains(got, want) {
			t.Fatalf("reply stream %q missing %q", got, want)
		}
	}
}

func TestBadArgCount(t *testing.T) {
	var in, out bytes.Buffer
	wire.WriteCommand(&in, "SUB", "onlyone")
	Serve(bufio.NewReader(&in), &out)
	if !strings.HasPrefix(out.String(), "-ERR usage: SUB") {
		t.Fatalf("got %q", out.String())
	}
}
