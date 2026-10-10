package core

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dmsRosa6/Chirp/wire"
)

func newSession(b *Broker) (*Session, *bytes.Buffer) {
	var buf bytes.Buffer
	return NewSession(b, &buf, "test"), &buf
}

// frames closes the session, waits for the writer to flush, and parses
// everything it wrote.
func frames(t *testing.T, s *Session, buf *bytes.Buffer) []wire.Reply {
	t.Helper()
	s.Close()
	s.Wait()

	r := bufio.NewReader(buf)
	var out []wire.Reply
	for {
		reply, err := wire.ReadReply(r)
		if errors.Is(err, io.EOF) {
			return out
		}
		if err != nil {
			t.Fatalf("ReadReply: %v", err)
		}
		out = append(out, reply)
	}
}

func mustSub(t *testing.T, b *Broker, s *Session, pattern, id string) {
	t.Helper()
	if err := b.Subscribe(s, pattern, id); err != nil {
		t.Fatalf("Subscribe(%q, %q): %v", pattern, id, err)
	}
}

func TestFanOut(t *testing.T) {
	b := NewBroker(Config{})
	const n = 3
	sessions := make([]*Session, n)
	bufs := make([]*bytes.Buffer, n)
	for i := range sessions {
		sessions[i], bufs[i] = newSession(b)
		mustSub(t, b, sessions[i], "foo.bar", "s")
	}

	if got := b.Publish("foo.bar", "hi"); got != n {
		t.Fatalf("Publish matched %d subscriptions, want %d", got, n)
	}

	for i := range sessions {
		got := frames(t, sessions[i], bufs[i])
		if len(got) != 1 {
			t.Fatalf("session %d got %d frames, want 1", i, len(got))
		}
		want := []string{"MSG", "s", "foo.bar", "hi"}
		for j := range want {
			if got[0].Type != wire.PushMessage || got[0].Args[j] != want[j] {
				t.Fatalf("session %d frame = %+v, want args %v", i, got[0], want)
			}
		}
	}
}

func TestWildcardAndLateBinding(t *testing.T) {
	b := NewBroker(Config{})
	s, buf := newSession(b)

	// Subscribe before any of these subjects have ever been published to.
	mustSub(t, b, s, "foo.ba*", "s")

	b.Publish("foo.bar", "1")
	b.Publish("foo.baz", "2")
	b.Publish("foo.x", "3")
	b.Publish("foo.bar.deep", "4") // one level only

	got := frames(t, s, buf)
	if len(got) != 2 {
		t.Fatalf("got %d frames, want 2: %+v", len(got), got)
	}
	if got[0].Args[2] != "foo.bar" || got[1].Args[2] != "foo.baz" {
		t.Fatalf("unexpected subjects/order: %+v", got)
	}
}

func TestTwoMatchingSubsInOneSessionGetTwoFrames(t *testing.T) {
	b := NewBroker(Config{})
	s, buf := newSession(b)
	mustSub(t, b, s, "foo.*", "a")
	mustSub(t, b, s, "foo.bar", "b")

	if got := b.Publish("foo.bar", "hi"); got != 2 {
		t.Fatalf("matched %d, want 2", got)
	}

	got := frames(t, s, buf)
	var ids []string
	for _, f := range got {
		ids = append(ids, f.Args[1])
	}
	sort.Strings(ids)
	if len(ids) != 2 || ids[0] != "a" || ids[1] != "b" {
		t.Fatalf("sub ids = %v, want [a b]", ids)
	}
}

func TestSubIDsAreScopedPerSession(t *testing.T) {
	b := NewBroker(Config{})
	s1, _ := newSession(b)
	s2, _ := newSession(b)
	defer s1.Close()
	defer s2.Close()

	mustSub(t, b, s1, "foo", "sub1")
	mustSub(t, b, s2, "foo", "sub1") // same id, different session: fine

	if err := b.Subscribe(s1, "bar", "sub1"); !errors.Is(err, ErrSubIDInUse) {
		t.Fatalf("duplicate id: got %v, want ErrSubIDInUse", err)
	}
}

func TestUnsubscribe(t *testing.T) {
	b := NewBroker(Config{})
	s, _ := newSession(b)
	defer s.Close()
	mustSub(t, b, s, "foo", "a")

	if err := b.Unsubscribe(s, "nope"); !errors.Is(err, ErrNoSuchSubscription) {
		t.Fatalf("unknown id: got %v, want ErrNoSuchSubscription", err)
	}
	if err := b.Unsubscribe(s, "a"); err != nil {
		t.Fatalf("Unsubscribe: %v", err)
	}
	if got := b.Publish("foo", "x"); got != 0 {
		t.Fatalf("matched %d after unsubscribe, want 0", got)
	}
	if n := b.SubscriptionCount(); n != 0 {
		t.Fatalf("SubscriptionCount = %d, want 0", n)
	}
	// The id is free again.
	mustSub(t, b, s, "foo", "a")
}

func TestCloseRemovesSubscriptions(t *testing.T) {
	b := NewBroker(Config{})
	s, _ := newSession(b)
	mustSub(t, b, s, "foo", "a")
	mustSub(t, b, s, "bar.*", "b")
	if n := b.SubscriptionCount(); n != 2 {
		t.Fatalf("SubscriptionCount = %d, want 2", n)
	}

	s.Close()
	s.Wait()

	if n := b.SubscriptionCount(); n != 0 {
		t.Fatalf("SubscriptionCount after Close = %d, want 0", n)
	}
	if got := b.Publish("foo", "x"); got != 0 {
		t.Fatalf("publish reached %d subscriptions of a closed session", got)
	}
	if err := b.Subscribe(s, "foo", "c"); !errors.Is(err, ErrSessionClosed) {
		t.Fatalf("Subscribe on closed session: got %v, want ErrSessionClosed", err)
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	b := NewBroker(Config{})
	s, _ := newSession(b)
	s.Close()
	s.Close()
	s.Wait()
}

// blockedWriter never accepts a write until released, like a client that
// stopped reading.
type blockedWriter struct{ release chan struct{} }

func (w blockedWriter) Write(p []byte) (int, error) {
	<-w.release
	return len(p), nil
}

func TestSlowConsumerIsDisconnectedAndPublisherNeverBlocks(t *testing.T) {
	b := NewBroker(Config{OutBuffer: 4})
	w := blockedWriter{release: make(chan struct{})}
	slow := NewSession(b, w, "slow")
	mustSub(t, b, slow, "foo", "s")

	fast, fastBuf := newSession(b)
	mustSub(t, b, fast, "foo", "f")

	published := make(chan struct{})
	go func() {
		defer close(published)
		for i := 0; i < 20; i++ {
			b.Publish("foo", "x")
			time.Sleep(time.Millisecond) // lets the healthy session's writer drain
		}
	}()

	select {
	case <-published:
	case <-time.After(2 * time.Second):
		t.Fatal("publisher blocked behind a slow consumer")
	}

	select {
	case <-slow.Done():
	case <-time.After(time.Second):
		t.Fatal("slow consumer was not disconnected")
	}
	if n := b.SubscriptionCount(); n != 1 {
		t.Fatalf("SubscriptionCount = %d, want 1 (only the fast session)", n)
	}

	// The healthy subscriber is unaffected and saw every message.
	if got := frames(t, fast, fastBuf); len(got) != 20 {
		t.Fatalf("healthy subscriber got %d frames, want 20", len(got))
	}

	close(w.release)
	slow.Wait()
}

func TestConcurrentSubscribePublishClose(t *testing.T) {
	b := NewBroker(Config{})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s := NewSession(b, io.Discard, "c")
			_ = b.Subscribe(s, "c.*", "s")
			_ = b.Subscribe(s, "c.x", "t")
			for j := 0; j < 20; j++ {
				b.Publish("c.x", strings.Repeat("p", 10))
			}
			_ = b.Unsubscribe(s, "t")
			s.Close()
			s.Wait()
		}()
	}
	wg.Wait()

	if n := b.SubscriptionCount(); n != 0 {
		t.Fatalf("SubscriptionCount = %d after everyone left, want 0", n)
	}
}
