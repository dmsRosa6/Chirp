package node

import (
	"bufio"
	"net"
	"testing"
	"time"

	"github.com/dmsRosa6/Chirp/wire"
)

type testClient struct {
	t    *testing.T
	conn net.Conn
	r    *bufio.Reader
}

func startNode(t *testing.T) (*Node, string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	n := New(ln.Addr())
	go n.Serve(ln)
	t.Cleanup(func() { ln.Close() })
	return n, ln.Addr().String()
}

func dial(t *testing.T, addr string) *testClient {
	t.Helper()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return &testClient{t: t, conn: conn, r: bufio.NewReader(conn)}
}

func (c *testClient) send(args ...string) {
	c.t.Helper()
	if err := wire.WriteCommand(c.conn, args...); err != nil {
		c.t.Fatal(err)
	}
}

func (c *testClient) read() wire.Reply {
	c.t.Helper()
	c.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	reply, err := wire.ReadReply(c.r)
	if err != nil {
		c.t.Fatalf("ReadReply: %v", err)
	}
	return reply
}

// do sends a command and expects a plain +OK.
func (c *testClient) ok(args ...string) {
	c.t.Helper()
	c.send(args...)
	if r := c.read(); r.Type != wire.SimpleString || r.Value != "OK" {
		c.t.Fatalf("%v: reply = %+v, want +OK", args, r)
	}
}

func (c *testClient) expectPush(subID, subject, payload string) {
	c.t.Helper()
	r := c.read()
	if r.Type != wire.PushMessage {
		c.t.Fatalf("got %+v, want a push", r)
	}
	want := []string{"MSG", subID, subject, payload}
	for i := range want {
		if r.Args[i] != want[i] {
			c.t.Fatalf("push = %v, want %v", r.Args, want)
		}
	}
}

func (c *testClient) expectSilence() {
	c.t.Helper()
	c.conn.SetReadDeadline(time.Now().Add(150 * time.Millisecond))
	if r, err := wire.ReadReply(c.r); err == nil {
		c.t.Fatalf("unexpected frame: %+v", r)
	}
}

func TestPing(t *testing.T) {
	_, addr := startNode(t)
	c := dial(t, addr)
	c.send("PING")
	if r := c.read(); r.Value != "PONG" {
		t.Fatalf("reply = %+v", r)
	}
}

func TestSubscriberReceivesExactlyOneMessage(t *testing.T) {
	_, addr := startNode(t)
	a, b := dial(t, addr), dial(t, addr)

	a.ok("SUB", "foo.*", "s1")
	b.ok("PUB", "foo.bar", "hi there")

	a.expectPush("s1", "foo.bar", "hi there")
	a.expectSilence()
}

func TestWildcardDelivery(t *testing.T) {
	_, addr := startNode(t)
	a, b := dial(t, addr), dial(t, addr)

	a.ok("SUB", "foo.ba*", "s1")
	b.ok("PUB", "foo.bar", "1")
	b.ok("PUB", "foo.x", "ignored")
	b.ok("PUB", "foo.baz", "2")

	a.expectPush("s1", "foo.bar", "1")
	a.expectPush("s1", "foo.baz", "2")
	a.expectSilence()
}

func TestSubjectsNeedNoCreation(t *testing.T) {
	// Publishing to a subject nobody has seen is fine, and a subscription made
	// beforehand picks it up.
	_, addr := startNode(t)
	a, b := dial(t, addr), dial(t, addr)

	a.ok("SUB", "brand.new.subject", "s1")
	b.ok("PUB", "brand.new.subject", "hello")
	a.expectPush("s1", "brand.new.subject", "hello")
}

func TestUnsubStopsDelivery(t *testing.T) {
	_, addr := startNode(t)
	a, b := dial(t, addr), dial(t, addr)

	a.ok("SUB", "foo", "s1")
	a.ok("UNSUB", "s1")
	b.ok("PUB", "foo", "x")
	a.expectSilence()
}

func TestDisconnectRemovesSubscriptions(t *testing.T) {
	n, addr := startNode(t)
	a := dial(t, addr)
	a.ok("SUB", "foo", "s1")
	a.ok("SUB", "bar.*", "s2")
	if got := n.broker.SubscriptionCount(); got != 2 {
		t.Fatalf("SubscriptionCount = %d, want 2", got)
	}

	a.conn.Close()

	deadline := time.Now().Add(2 * time.Second)
	for n.broker.SubscriptionCount() != 0 {
		if time.Now().After(deadline) {
			t.Fatalf("SubscriptionCount = %d, want 0 after disconnect", n.broker.SubscriptionCount())
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestProtocolErrorIsReportedThenDisconnected(t *testing.T) {
	_, addr := startNode(t)
	c := dial(t, addr)

	c.conn.Write([]byte("garbage\r\n"))
	if r := c.read(); r.Type != wire.ErrorReply {
		t.Fatalf("reply = %+v, want an error", r)
	}
	c.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := wire.ReadReply(c.r); err == nil {
		t.Fatal("expected the server to close the connection")
	}
}

func TestPayloadsWithSpacesAndNewlines(t *testing.T) {
	_, addr := startNode(t)
	a, b := dial(t, addr), dial(t, addr)

	payload := "line one\nline two   with   spaces\r\nand a CRLF"
	a.ok("SUB", "foo", "s1")
	b.ok("PUB", "foo", payload)
	a.expectPush("s1", "foo", payload)
}
