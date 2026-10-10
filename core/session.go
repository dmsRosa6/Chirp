package core

import (
	"bytes"
	"errors"
	"io"
	"log"
	"sync"
	"time"

	"github.com/dmsRosa6/Chirp/wire"
)

const writeTimeout = 5 * time.Second

var (
	ErrSessionClosed      = errors.New("session closed")
	ErrSubIDInUse         = errors.New("sub id in use")
	ErrNoSuchSubscription = errors.New("no such subscription")
)

// Session is one client connection.
//
// Everything bound for the socket (replies and pushed messages alike) goes
// through the out channel and is written by a single goroutine, so frames
// never interleave. Handlers never touch the socket.
type Session struct {
	id     uint64
	addr   string
	broker *Broker

	out        chan []byte
	done       chan struct{} // closed by Close
	writerDone chan struct{} // closed when the writer goroutine has exited
	closeOnce  sync.Once

	mu     sync.Mutex // guards subs and closed
	subs   map[string]*Subscription
	closed bool
}

// NewSession creates a session and starts its writer goroutine, which owns w
// from now on. If w is an io.Closer (a net.Conn) it is closed once the writer
// exits, which also unblocks a reader stuck on the same connection.
func NewSession(b *Broker, w io.Writer, addr string) *Session {
	s := &Session{
		id:         b.nextID.Add(1),
		addr:       addr,
		broker:     b,
		out:        make(chan []byte, b.cfg.OutBuffer),
		done:       make(chan struct{}),
		writerDone: make(chan struct{}),
		subs:       make(map[string]*Subscription),
	}
	go s.writeLoop(w)
	return s
}

func (s *Session) ID() uint64            { return s.id }
func (s *Session) Addr() string          { return s.addr }
func (s *Session) Broker() *Broker       { return s.broker }
func (s *Session) Done() <-chan struct{} { return s.done }

// Wait blocks until the writer goroutine has flushed what was queued and
// exited. Call it after Close.
func (s *Session) Wait() { <-s.writerDone }

// Close is idempotent. It stops accepting output, removes every
// subscription of this session from the broker, and lets the writer flush
// whatever is already queued before it exits.
func (s *Session) Close() {
	s.closeOnce.Do(func() {
		s.mu.Lock()
		s.closed = true
		subs := make([]*Subscription, 0, len(s.subs))
		for _, sub := range s.subs {
			subs = append(subs, sub)
		}
		s.subs = nil
		s.mu.Unlock()

		close(s.done)
		s.broker.detachAll(subs)
	})
}

// Push queues a frame without blocking. If the outbound buffer is full the
// consumer is too slow: the session is closed and false is returned.
func (s *Session) Push(frame []byte) bool {
	select {
	case <-s.done:
		return false
	default:
	}

	select {
	case s.out <- frame:
		return true
	default:
		log.Printf("chirp: disconnecting slow consumer %s (outbound buffer full)", s.addr)
		s.Close()
		return false
	}
}

// Reply helpers. They return an error only if the session is closed, which
// tells the read loop to stop.

func (s *Session) OK() error { return s.reply(wire.WriteOK) }

func (s *Session) Simple(str string) error {
	return s.reply(func(w io.Writer) error { return wire.WriteSimple(w, str) })
}

func (s *Session) Err(msg string) error {
	return s.reply(func(w io.Writer) error { return wire.WriteError(w, msg) })
}

func (s *Session) Bulk(str string) error {
	return s.reply(func(w io.Writer) error { return wire.WriteBulk(w, str) })
}

func (s *Session) reply(write func(io.Writer) error) error {
	var b bytes.Buffer
	if err := write(&b); err != nil {
		return err
	}
	if !s.Push(b.Bytes()) {
		return ErrSessionClosed
	}
	return nil
}

// addSub registers sub under its id. Called by the broker.
func (s *Session) addSub(sub *Subscription) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrSessionClosed
	}
	if _, ok := s.subs[sub.ID]; ok {
		return ErrSubIDInUse
	}
	s.subs[sub.ID] = sub
	return nil
}

// takeSub removes and returns the subscription with this id. Called by the broker.
func (s *Session) takeSub(id string) (*Subscription, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sub, ok := s.subs[id]
	if ok {
		delete(s.subs, id)
	}
	return sub, ok
}

func (s *Session) writeLoop(w io.Writer) {
	defer close(s.writerDone)
	defer func() {
		if c, ok := w.(io.Closer); ok {
			c.Close()
		}
	}()
	defer s.Close() // a failed write ends the session too

	for {
		select {
		case b := <-s.out:
			if !write(w, b) {
				return
			}
		case <-s.done:
			// Flush what is already queued (e.g. a final -ERR), then stop.
			for {
				select {
				case b := <-s.out:
					if !write(w, b) {
						return
					}
				default:
					return
				}
			}
		}
	}
}

func write(w io.Writer, b []byte) bool {
	// A stuck peer must not hold the writer (and its goroutine) forever.
	if d, ok := w.(interface{ SetWriteDeadline(time.Time) error }); ok {
		d.SetWriteDeadline(time.Now().Add(writeTimeout))
	}
	_, err := w.Write(b)
	return err == nil
}
