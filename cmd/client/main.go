package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/dmsRosa6/Chirp/wire"
)

func splitLine(line string) []string {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil
	}
	if strings.EqualFold(fields[0], "PUB") && len(fields) >= 3 {
		rest := strings.SplitN(line, fields[1], 2)[1]
		return []string{fields[0], fields[1], strings.TrimSpace(rest)}
	}
	return fields
}

func main() {
	addr := "localhost:4222"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("could not connect:", err)
		os.Exit(1)
	}
	defer conn.Close()

	server := bufio.NewReader(conn)
	stdin := bufio.NewScanner(os.Stdin)

	fmt.Printf("connected to %s. type a command (PING, PUB, SUB, UNSUB), Ctrl-D to quit.\n", addr)
	for {
		fmt.Print("chirp> ")
		if !stdin.Scan() {
			return
		}
		args := splitLine(stdin.Text())
		if len(args) == 0 {
			continue
		}

		if err := wire.WriteCommand(conn, args...); err != nil {
			fmt.Println("write error:", err)
			return
		}
		reply, err := wire.ReadReply(server)
		if err != nil {
			fmt.Println("connection lost:", err)
			return
		}
		fmt.Println(reply.Value)
	}
}
