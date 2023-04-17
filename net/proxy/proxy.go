package proxy

import (
	"bufio"
	"io"

	"dslab.inf.usi.ch/tendermint/consensus"
	"dslab.inf.usi.ch/tendermint/net"
	"dslab.inf.usi.ch/tendermint/net/libp2p"
	"dslab.inf.usi.ch/tendermint/types"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

// Proxy implements net.Proxy interface.
var _ net.Proxy = new(Proxy)

var QueueSize = 32
var ProtocolID = libp2p.Protocol("/values")

type Proxy struct {
	decisionQueue chan *net.Decision
	proposalQueue chan []byte

	debug bool
	host  *libp2p.Host
	log   net.Log

	streams      []network.Stream
	streamsQueue chan network.Stream
}

func NewProxy(host *libp2p.Host, log net.Log, debug bool) *Proxy {
	proxy := &Proxy{
		host:          host,
		debug:         debug,
		log:           log,
		decisionQueue: make(chan *net.Decision, QueueSize),
		proposalQueue: make(chan []byte, QueueSize),
		streamsQueue:  make(chan network.Stream, QueueSize),
	}
	proxy.log.Prefix += " proxy"
	host.Host.SetStreamHandler(ProtocolID, func(s network.Stream) {
		proxy.streamsQueue <- s
	})
	go proxy.mainLoop()
	return proxy
}

// Deliver delivers a block committed by the consensus protocol.
//
// This method extracts the delivery data, which is added to the decisions queue.
func (p *Proxy) Deliver(epoch int64, block *consensus.Block) {
	p.decisionQueue <- &net.Decision{
		Instance: uint64(block.Height),
		Value:    block.Value,
		ValueID:  net.ValueID(block.Value),
	}
}

// GetValue returns a value to be proposed in the consensus protocol.
func (p *Proxy) GetValue() types.Value {
	select {
	case value := <-p.proposalQueue:
		return value
	default:
		return nil
	}
}

func (p *Proxy) Debug(enable bool) {
	p.debug = enable
}

func (p *Proxy) mainLoop() {
	for {
		select {
		case stream := <-p.streamsQueue:
			p.addClient(stream)

		case decision := <-p.decisionQueue:
			if p.debug {
				p.log.Println("decision",
					decision.ValueID)
			}
			if len(p.streams) > 0 {
				p.broadcastDecision(decision)
			}
		}
	}
}

func (p *Proxy) addClient(stream network.Stream) {
	// Listen to proposed values
	go p.receiver(stream)

	// Register as destination of decisions
	p.streams = append(p.streams, stream)

	p.log.Println("client added", len(p.streams)-1,
		addrInfoFromStream(stream))
}

func (p *Proxy) broadcastDecision(decision *net.Decision) {
	// Send decision to all registered clients
	message := EncodeDecision(decision)
	for index, stream := range p.streams {
		if stream == nil {
			continue
		}
		_, err := stream.Write(message)
		if err != nil {
			p.removeClient(index)
		}
	}
}

func (p *Proxy) removeClient(index int) {
	// Close stream with remote client
	p.streams[index].Reset()
	p.log.Println("client removed", index,
		addrInfoFromStream(p.streams[index]))
	// Just unsed the stream slot
	p.streams[index] = nil
}

func (p *Proxy) receiver(stream network.Stream) {
	var err error
	header := make([]byte, 4)
	reader := bufio.NewReader(stream)
	for {
		_, err = io.ReadFull(reader, header)
		if err != nil {
			break
		}

		size := int(encoding.Uint32(header))
		value := make([]byte, size)

		_, err = io.ReadFull(reader, value)
		if err != nil {
			break
		}

		if p.debug {
			p.log.Println("proposal", net.ValueID(value))
		}

		// Propose value for consensus
		p.proposalQueue <- value
	}
}

func addrInfoFromStream(stream network.Stream) peer.AddrInfo {
	return peer.AddrInfo{
		ID: stream.Conn().RemotePeer(),
		Addrs: []multiaddr.Multiaddr{
			stream.Conn().RemoteMultiaddr(),
		},
	}
}
