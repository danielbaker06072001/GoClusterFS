package main

import (
	"log"

	"github.com/danielbaker06072001/foreverstore/p2p"
)

func main() {
	tr := p2p.NewTCPTransport(":3000")
	if err := tr.ListenAndAccept(); err != nil {
		log.Fatalf("Failed to start transport: %v", err)
	}
	select {}
}
