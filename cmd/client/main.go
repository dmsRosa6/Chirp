package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"

	"github.com/dmsRosa6/Chirp/wire"
)

const (
	defaultAddr = "localhost:4222"
	prompt      = "chirp> "
)

var (
	printMu  sync.Mutex
	quitting atomic.Bool // set on a deliberate exit so readLoop stays quiet
)

// show prints a frame that arrived while the user may be typing: it
// overwrites the current prompt line, prints, and draws the prompt again.
func show(line string) {
	printMu.Lock()
	defer printMu.Unlock()
	fmt.Printf("\r\033[K%s\n%s", line, prompt)
}

func format(r wire.Reply) string {
	if r.Type == wire.PushMessage {
		return fmt.Sprintf("[%s] %s: %s", r.Args[1], r.Args[2], r.Args[3])
	}
	return r.Value
}

// readLoop prints every frame the server sends, replies and pushes alike.
// The server answers commands in order and pushes are a different frame type,
// so there is nothing to match up.
func readLoop(server *bufio.Reader) {
	for {
		reply, err := wire.ReadReply(server)
		if err != nil {
			if quitting.Load() {
				return
			}
			printMu.Lock()
			fmt.Println("\r\033[Kconnection lost:", err)
			printMu.Unlock()
			os.Exit(1)
		}
		show(format(reply))
	}
}

func main() {
	addr := defaultAddr
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("could not connect:", err)
		os.Exit(1)
	}
	defer func() {
		quitting.Store(true)
		conn.Close()
	}()

	go readLoop(bufio.NewReader(conn))

	stdin := bufio.NewScanner(os.Stdin)
	fmt.Printf("connected to %s\n", addr)
	fmt.Println("commands: PING | PUB <subject> <payload> | SUB <pattern> <sub_id> | UNSUB <sub_id>")
	fmt.Println("patterns may end in * (foo.b* matches foo.bar). Ctrl-D to quit.")
	fmt.Print(prompt)

	for stdin.Scan() {
		args := parseLine(stdin.Text())
		if len(args) == 0 {
			printMu.Lock()
			fmt.Print(prompt)
			printMu.Unlock()
			continue
		}
		if err := wire.WriteCommand(conn, args...); err != nil {
			fmt.Println("write error:", err)
			return
		}
	}
}
