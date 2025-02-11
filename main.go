package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/danielbaker06072001/foreverstore/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		StorageRoot:       fmt.Sprintf("%s_network", listenAddr[1:]),
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}

	s := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer

	return s
}

func main() {
	s1 := makeServer(":5001", "")
	s2 := makeServer(":5002", ":5001")

	go func() {
		log.Fatal(s1.Start())
	}()

	time.Sleep(1 * time.Second)
	go s2.Start()
	time.Sleep(1 * time.Second)

	data := bytes.NewReader([]byte("super large data file!"))
	s2.StoreData("myprivatedata", data)

	select {}
}
