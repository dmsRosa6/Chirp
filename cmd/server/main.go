package main

import (
	"bufio"
	"log"
	"net"

	"github.com/dmsRosa6/Chirp/repl"
)

func main() {
	ln, err := net.Listen("tcp", ":4222") // NATS's default port, felt fitting
	if err != nil {
		log.Fatal(err)
	}
	log.Println("chirp listening on :4222")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept:", err)
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			if err := repl.Serve(bufio.NewReader(c), c); err != nil {
				log.Println("connection closed:", err)
			}
		}(conn)
	}
}
