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
		n.DropClient(client)
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
	return q.Add(core.Message{Body: payload, Time: time.Now().UnixNano()})
}

func (n *Node) Subscribe(c *core.Client, subject, subID string) error {
	q, err := n.qm.GetByFullPath(subject)
	if err != nil {
		return err
	}
	if !c.TrackSub(subID, &q) {
		return fmt.Errorf("sub id %s already in use", subID)
	}
	q.Subscribe(c)
	return nil
}

func (n *Node) Unsubscribe(c *core.Client, subID string) error {
	q, ok := c.UntrackSub(subID)
	if !ok {
		return fmt.Errorf("unknown sub id %s", subID)
	}
	q.Unsubscribe(c)
	return nil
}

func (n *Node) DropClient(c *core.Client) {
	for _, q := range c.DropAllSubs() {
		q.Unsubscribe(c)
	}
}
