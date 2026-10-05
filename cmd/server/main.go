package main

import (
	"log"
	"net"

	"github.com/dmsRosa6/Chirp/node"
)

func main() {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	node := node.New(addr, nil)
	node.Start()
}
