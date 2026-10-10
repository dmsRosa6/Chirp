package main

import (
	"log"
	"net"
	"os"

	"github.com/dmsRosa6/Chirp/node"
)

const defaultAddr = "localhost:4222"

func main() {
	listen := defaultAddr
	if len(os.Args) > 1 {
		listen = os.Args[1]
	}

	addr, err := net.ResolveTCPAddr("tcp", listen)
	if err != nil {
		log.Fatal(err)
	}

	if err := node.New(addr).Start(); err != nil {
		log.Fatal(err)
	}
}
