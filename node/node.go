package node

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"

	"github.com/dmsRosa6/Chirp/core"
	"github.com/dmsRosa6/Chirp/repl"
)

// Node listens for connections and gives each one a session. It owns no
// message routing; that is the broker's job.
type Node struct {
	addr   net.Addr
	broker *core.Broker
}

func New(addr net.Addr) *Node {
	return &Node{addr: addr, broker: core.NewBroker(core.Config{})}
}

// Start listens on the node's address and serves until the listener fails.
func (n *Node) Start() error {
	ln, err := net.Listen(n.addr.Network(), n.addr.String())
	if err != nil {
		return err
	}
	return n.Serve(ln)
}

// Serve accepts connections on ln until it is closed.
func (n *Node) Serve(ln net.Listener) error {
	log.Println("chirp listening on", ln.Addr())
	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			log.Println("accept:", err)
			continue
		}
		go n.handleConn(conn)
	}
}

func (n *Node) handleConn(conn net.Conn) {
	// NewSession starts the writer goroutine, which owns conn from here on
	// and closes it when the session ends.
	s := core.NewSession(n.broker, conn, conn.RemoteAddr().String())
	defer s.Close()

	if err := repl.Serve(s, bufio.NewReader(conn)); err != nil && !errors.Is(err, io.EOF) {
		log.Printf("connection %s closed: %v", s.Addr(), err)
	}
}
