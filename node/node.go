package node

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/dmsRosa6/Chirp/core"
	"github.com/dmsRosa6/Chirp/repl"
)

type Node struct {
	addr       net.Addr
	neighbours []net.Addr
	qm         core.QueueManager
}

func New(addr net.Addr, neighbours []net.Addr) *Node {
	return &Node{
		addr:       addr,
		neighbours: neighbours,
		qm:         *core.NewQueueManager(),
	}
}

func (n *Node) Start() {
	ln, err := net.Listen(n.addr.Network(), n.addr.String())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("chirp listening on", n.addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept:", err)
			continue
		}
		go n.handleConn(conn)
	}
}

func (n *Node) handleConn(conn net.Conn) {
	client := core.NewClient(conn)
	defer func() {
		conn.Close()
	}()

	s := &repl.Session{W: conn, Client: client, Broker: n}
	if err := repl.Serve(bufio.NewReader(conn), s); err != nil {
		log.Println("connection closed:", err)
	}
}

func (n *Node) CreateQueue(subject string) error {
	return n.qm.AddQueue(core.NewQueue(subject))
}

func (n *Node) QueueExists(subject string) bool {
	return n.qm.ExistsByFullPath(subject)
}

func (n *Node) Publish(subject, payload string) error {
	q, err := n.qm.GetByFullPath(subject)
	if err != nil {
		return err
	}
	return q.Publish(core.Message{Body: payload, Time: time.Now().UnixNano()})
}

func (n *Node) Subscribe(c *core.Client, subject string) error {
	q, err := n.qm.GetByFullPath(subject)
	if err != nil {
		return err
	}
	if err := q.Subscribe(c); err != nil {
		return fmt.Errorf("client %s already in use", c.Addr())
	}
	return nil
}

func (n *Node) Unsubscribe(c *core.Client, subject string) error {
	q, err := n.qm.GetByFullPath(subject)
	if err != nil {
		return err
	}
	if err := q.Unsubscribe(c); err != nil {
		return fmt.Errorf("unknown client %s")
	}

	return nil
}
