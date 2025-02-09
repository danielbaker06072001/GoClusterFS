package main

import (
	"fmt"
	"log"

	"github.com/danielbaker06072001/foreverstore/p2p"
)

func OnPeer(peer p2p.Peer) error {
	peer.Close()
	fmt.Println("doing some logic outside of the TCPTransport")
	return nil
}

func main() {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}
	tr := p2p.NewTCPTransport(tcpOpts)

	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("%+v\n", msg)
		}
	}()

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatalf("Failed to start transport: %v", err)
	}
	select {}
}
