package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/danielbaker06072001/foreverstore/p2p"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	store  *Store
	quitch chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key string
}

// func (s *FileServer) broadcast(p *Payload) error {
// 	peers := []io.Writer{}

// 	for _, peer := range s.peers {
// 		peers = append(peers, peer)
// 	}

// 	mw := io.MultiWriter(peers...)
// 	return gob.NewEncoder(mw).Encode(p)
// }

func (s *FileServer) broadcast(msg *Message) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		if _, err := peer.Write(buf.Bytes()); err != nil {
			log.Printf("error broadcasting to %s: %v", peer.RemoteAddr(), err)
		}
	}
	return nil
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	// 1. Store the file to disk
	// 2. Boradcast the file to all known peers in the network

	buf := new(bytes.Buffer)
	msg := Message{
		Payload: MessageStoreFile{
			Key: key,
		},
	}

	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	// time.Sleep(3 * time.Second)

	// payload := []byte("THIS LARGE FILE")
	// for _, peer := range s.peers {
	// 	if err := peer.Send(payload); err != nil {
	// 		return err
	// 	}
	// }

	return nil

	// buf := new(bytes.Buffer)
	// tee := io.TeeReader(r, buf)

	// if err := s.store.Write(key, tee); err != nil {
	// 	return err
	// }

	// p := &DataMessage{
	// 	Key:  key,
	// 	Data: buf.Bytes(),
	// }

	// fmt.Println(buf.Bytes())

	// return s.broadcast(&Message{
	// 	From:    "todo",
	// 	Payload: p,
	// })
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()

	defer s.peerLock.Unlock()
	s.peers[p.RemoteAddr().String()] = p

	log.Printf("connected with remote %s", p.RemoteAddr())

	return nil
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("file server stopped due to user quit action!")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("decoding error: ", err)
				return
			}

			fmt.Printf("Payload: %+s\n", msg.Payload)

			peer, ok := s.peers[rpc.From]
			if !ok {
				panic("peer not found in no peer in peers map")
			}

			b := make([]byte, 1024)

			if _, err := peer.Read(b); err != nil {
				panic(err)
			}

			fmt.Printf("%s\n", string(b))

			peer.(*p2p.TCPPeer).Wg.Done()
		// if err := s.handleMessage(&m); err != nil {
		// 	log.Println("error handling message: ", err)
		// }

		case <-s.quitch:
			return
		}
	}
}

// func (s *FileServer) handleMessage(msg *Message) error {
// 	switch v := msg.Payload.(type) {
// 	case *DataMessage:
// 		fmt.Printf("received data message: %v\n", v)
// 	}
// 	return nil
// }

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}

		go func(addr string) {
			fmt.Printf("attempting to connect with remote %s\n", addr)
			if err := s.Transport.Dial(addr); err != nil {
				log.Printf("failed to dial to %s: %v", addr, err)
			}
		}(addr)

	}
	return nil
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.bootstrapNetwork()

	s.loop()

	return nil
}

func init() {
	// gob.Register(DataMessage{})
	gob.Register(Message{})
	gob.Register(MessageStoreFile{})
}
