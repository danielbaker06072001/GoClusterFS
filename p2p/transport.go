package p2p

// * Peer is an interface that represents the
// remote node in the network
// Example : it can be a dude or a girl that we're connecting to
type Peer interface {
	Close() error
}

// * Transport is anything that handles the communication
// between nodes mong the network. This can be of the form
// {TCP/UDP websocket, ...  }
type Transport interface {
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
