package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCPPeer represent the remote node over TCP establised connection.
type TCPPeer struct {
	// conn is the underlying connection of the peer
	conn net.Conn

	// if we dial and retrieve a connection => outbound = true
	// if we accept and retrieve a connection => outbound = false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransport struct {
	listenAddress string
	listener      net.Listener
	handshakeFunc HandshakeFunc

	// * it is common in go to put a mutex on top of the thing we wanted to protect
	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NOPHandshake(any) error {
	return nil
}

func NewTCPTransport(listenerAdddr string) *TCPTransport {
	return &TCPTransport{
		handshakeFunc: NOPHandshake,
		listenAddress: listenerAdddr,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() error {

	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP error: %v", err)
		}

		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)

	fmt.Printf("New incoming connection %+v\n", peer)
}
