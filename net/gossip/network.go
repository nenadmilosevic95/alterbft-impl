package gossip

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multistream"

	"dslab.inf.usi.ch/tendermint/net/libp2p"
)

var ProtocolID = libp2p.Protocol("/g")

var ConnectionSleepInterval = 400 * time.Millisecond

type Network struct {
	Host  *libp2p.Host
	Peers *PeersTable

	// Input from libp2p's Host
	ConnsQueue   <-chan network.Conn
	StreamsQueue chan network.Stream
}

func NewNetwork(host *libp2p.Host, peers *PeersTable) *Network {
	n := &Network{
		Host:  host,
		Peers: peers,
	}
	n.ConnsQueue = host.NotifyConnections(DefaultQueueSize)
	n.StreamsQueue = make(chan network.Stream, DefaultQueueSize)
	host.Host.SetStreamHandler(ProtocolID,
		func(stream network.Stream) {
			n.StreamsQueue <- stream
		})
	rand.Seed(time.Now().UnixNano())
	go n.mainLoop()
	return n
}

func (n *Network) Connect(addr peer.AddrInfo) {
	var connect bool

	n.Peers.Lock()
	peer := n.Peers.GetByAddr(addr)
	peer.Chosen = true
	if peer.ConnS == 0 { // Not tried
		peer.ConnS = 1 // Attempting
		connect = true
	}
	peer.Comment = "Connect"
	n.Peers.Notify(peer)
	n.Peers.Unlock()

	if connect {
		go n.connect(addr)
	}
}

func (n *Network) mainLoop() {
	for {
		select {
		case conn := <-n.ConnsQueue:
			n.addConnection(conn)
		case stream := <-n.StreamsQueue:
			n.addStream(stream)
		}
	}
}

// Called whenever a new connection is notified by the Host.
func (n *Network) addConnection(conn network.Conn) {
	var openStream bool

	n.Peers.Lock()
	peer := n.Peers.GetByConn(conn)
	if peer.ConnS != 2 { // Established
		peer.Conn = conn
		peer.ConnE = nil
		peer.ConnS = 2
		if peer.SendStreamS == 0 { // Not tried
			peer.SendStreamS = 1 // Attempting
			openStream = true
		}
	}
	peer.Comment = "addConnection"
	n.Peers.Notify(peer)
	n.Peers.Unlock()

	if openStream {
		go n.openStream(addrInfoFromConn(conn))
	}
}

// Called whenever the Host accepts a new ingoing stream.
func (n *Network) addStream(stream network.Stream) {
	var handleStream bool

	n.Peers.Lock()
	peer := n.Peers.GetByStream(stream)
	if peer.RecvStreamS == 0 { // Not present
		peer.RecvStreamS = 1 // Attempting
		peer.RecvStreamE = nil
		handleStream = true
	}
	peer.Comment = "addStream"
	n.Peers.Notify(peer)
	n.Peers.Unlock()

	if handleStream {
		go n.handleStream(stream)
	}
}

func (n *Network) connect(addr peer.AddrInfo) {
	//	if n.Host.UID() < addr.ID {
	//		time.Sleep(ConnectionSleepInterval +
	//			time.Duration(rand.Intn(100))*time.Millisecond)
	//	}
	times := 0
	err := n.Host.Connect(addr)
	for err != nil && times < 10 {
		err = n.Host.Connect(addr)
		time.Sleep(time.Millisecond * 100)
		times++
	}
	if err != nil {
		panic(fmt.Errorf("Connection failed: %s", err))
	}

	n.Peers.Lock()
	peer := n.Peers.GetByAddr(addr)
	if err != nil {
		peer.ConnE = err
		peer.ConnS = 3
	}
	peer.Comment = "connect"
	n.Peers.Notify(peer)
	n.Peers.Unlock()
}

func (n *Network) handleStream(stream network.Stream) {
	// Wait for a message from the stream
	message := new(Message)
	err := message.ReadFrom(stream)

	// Register the peer's receive stream
	n.Peers.Lock()
	peer := n.Peers.GetByStream(stream)
	if peer.RecvStreamS == 1 { // Attempting
		peer.RecvStream = stream
		peer.RecvStreamE = err
		if err == nil {
			peer.RecvStreamS = 2 // Established
			// Also registers the peer's ID
			peer.ID = int(message.Sender)
		} else {
			peer.RecvStreamS = 3
		}
	}
	peer.Comment = "handleStream"
	if !peer.Active && peer.FulllyConnected() {
		n.Peers.Activate(peer)
	}
	n.Peers.Notify(peer)
	n.Peers.Unlock()
}

func (n *Network) openStream(addr peer.AddrInfo) {
	for tries := 0; tries < 5; tries++ {
		comment := "openStream NewStream"
		if tries > 0 {
			comment = fmt.Sprint(comment,
				", retry ", tries)
		}
		// Open a stream to the peer and send our host ID
		stream, err := n.Host.NewStream(addr, ProtocolID)
		if err == nil {
			message := &Message{
				Sender: uint16(n.Host.ID),
			}
			err = message.WriteTo(stream)
			comment = "openStream WriteTo"
			if tries > 0 {
				comment = fmt.Sprint(comment,
					", retry ", tries)
			}
		} else if err == multistream.ErrNotSupported {
			return
		}

		// Register the peer's send stream
		n.Peers.Lock()
		peer := n.Peers.GetByAddr(addr)
		if peer.SendStreamS == 1 || // Attempting
			peer.SendStreamS == 3 { // Error
			peer.SendStream = stream
			peer.SendStreamE = err
			if err == nil {
				peer.SendStreamS = 2 // Established
			} else {
				peer.SendStreamS = 3
			}
		}
		peer.Comment = comment
		if !peer.Active && peer.FulllyConnected() {
			n.Peers.Activate(peer)
		}
		n.Peers.Notify(peer)
		n.Peers.Unlock()
		if err == nil {
			break
		}
		time.Sleep(time.Duration(100+rand.Intn(100)) * time.Millisecond)
	}
}

func addrInfoFromConn(conn network.Conn) peer.AddrInfo {
	return peer.AddrInfo{
		ID: conn.RemotePeer(),
		Addrs: []multiaddr.Multiaddr{
			conn.RemoteMultiaddr(),
		},
	}
}
