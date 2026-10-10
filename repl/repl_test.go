package repl

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/dmsRosa6/Chirp/core"
	"github.com/dmsRosa6/Chirp/wire"
)

// session is a test harness: a real broker and session whose writer goroutine
// writes into a buffer.
type session struct {
	t   *testing.T
	b   *core.Broker
	s   *core.Session
	out *bytes.Buffer
}

func newSession(t *testing.T, b *core.Broker) *session {
	if b == nil {
		b = core.NewBroker(core.Config{})
	}
	out := &bytes.Buffer{}
	return &session{t: t, b: b, s: core.NewSession(b, out, "test"), out: out}
}

// run feeds the commands to Serve (which stops at end of input), waits for
// the writer to drain, and returns everything that was written.
func (h *session) run(cmds ...[]string) (string, error) {
	var in bytes.Buffer
	for _, c := range cmds {
		if err := wire.WriteCommand(&in, c...); err != nil {
			h.t.Fatal(err)
		}
	}
	return h.serve(&in)
}

func (h *session) serve(in *bytes.Buffer) (string, error) {
	err := Serve(h.s, bufio.NewReader(in))
	h.s.Close()
	h.s.Wait()
	return h.out.String(), err
}

func cmd(args ...string) []string { return args }

func TestPing(t *testing.T) {
	out, _ := newSession(t, nil).run(cmd("PING"), cmd("ping"))
	if want := "+PONG\r\n+PONG\r\n"; out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}

func TestUnknownCommand(t *testing.T) {
	out, _ := newSession(t, nil).run(cmd("BOGUS"))
	if want := "-ERR unknown command: BOGUS\r\n"; out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}

func TestRemovedCommandsAreUnknown(t *testing.T) {
	out, _ := newSession(t, nil).run(cmd("CREATE", "foo"), cmd("EXISTS", "foo"))
	if !strings.Contains(out, "unknown command: CREATE") || !strings.Contains(out, "unknown command: EXISTS") {
		t.Fatalf("got %q", out)
	}
}

func TestWrongArgCountShowsUsage(t *testing.T) {
	out, _ := newSession(t, nil).run(cmd("PUB", "foo"), cmd("SUB", "foo"), cmd("UNSUB"))
	want := "-ERR usage: PUB <subject> <payload>\r\n" +
		"-ERR usage: SUB <pattern> <sub_id>\r\n" +
		"-ERR usage: UNSUB <sub_id>\r\n"
	if out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}

func TestPubWithNoSubscribersIsOK(t *testing.T) {
	out, _ := newSession(t, nil).run(cmd("PUB", "never.seen", "hello"))
	if want := "+OK\r\n"; out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}

func TestPubRejectsBadSubject(t *testing.T) {
	for _, subject := range []string{"foo.*", "a..b", ".a", "has space"} {
		out, _ := newSession(t, nil).run(cmd("PUB", subject, "x"))
		if !strings.HasPrefix(out, "-ERR ") {
			t.Errorf("PUB %q: got %q, want an error", subject, out)
		}
	}
}

func TestSubRejectsBadPatternAndID(t *testing.T) {
	for _, tc := range [][]string{
		{"SUB", "a*.b", "s1"},
		{"SUB", "a..b", "s1"},
		{"SUB", "foo", ""},
	} {
		out, _ := newSession(t, nil).run(tc)
		if !strings.HasPrefix(out, "-ERR ") {
			t.Errorf("%v: got %q, want an error", tc, out)
		}
	}
}

func TestSubDuplicateIDAndUnsub(t *testing.T) {
	out, _ := newSession(t, nil).run(
		cmd("SUB", "foo", "s1"),
		cmd("SUB", "bar", "s1"),
		cmd("UNSUB", "s1"),
		cmd("UNSUB", "s1"),
	)
	want := "+OK\r\n" +
		"-ERR sub id in use\r\n" +
		"+OK\r\n" +
		"-ERR no such subscription\r\n"
	if out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}

func TestSubscriberReceivesPublishFromAnotherSession(t *testing.T) {
	b := core.NewBroker(core.Config{})
	sub := newSession(t, b)
	pub := newSession(t, b)

	// Drive the subscriber through its handler; its session stays open.
	if err := handleSub(sub.s, []string{"foo.b*", "s1"}); err != nil {
		t.Fatal(err)
	}
	if out, _ := pub.run(cmd("PUB", "foo.bar", "hello world")); out != "+OK\r\n" {
		t.Fatalf("PUB reply = %q", out)
	}

	sub.s.Close()
	sub.s.Wait()

	r := bufio.NewReader(sub.out)
	ok, err := wire.ReadReply(r)
	if err != nil || ok.Type != wire.SimpleString || ok.Value != "OK" {
		t.Fatalf("first frame = %+v, %v; want +OK", ok, err)
	}
	msg, err := wire.ReadReply(r)
	if err != nil || msg.Type != wire.PushMessage {
		t.Fatalf("second frame = %+v, %v; want a push", msg, err)
	}
	want := []string{"MSG", "s1", "foo.bar", "hello world"}
	for i := range want {
		if msg.Args[i] != want[i] {
			t.Fatalf("push args = %v, want %v", msg.Args, want)
		}
	}
}

func TestProtocolErrorGetsAnErrReplyAndEndsTheLoop(t *testing.T) {
	h := newSession(t, nil)
	out, err := h.serve(bytes.NewBufferString("not-a-frame\r\n"))
	if !wire.IsProtocolError(err) {
		t.Fatalf("err = %v, want a protocol error", err)
	}
	if !strings.HasPrefix(out, "-ERR protocol error") {
		t.Fatalf("out = %q, want an -ERR protocol error reply", out)
	}
}
