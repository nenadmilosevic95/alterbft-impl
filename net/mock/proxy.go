package mock

import (
	"dslab.inf.usi.ch/tendermint/consensus"
	"dslab.inf.usi.ch/tendermint/net"
	"dslab.inf.usi.ch/tendermint/types"
)

// Proxy is a mock implementation of net.Proxy interface.
var _ net.Proxy = new(Proxy)

type Proxy struct {
	Proposals chan []byte
	Decisions chan *net.Decision
}

// NewGossip creates a mock proxy implementation.
func NewProxy(queueSize int) *Proxy {
	return &Proxy{
		Proposals: make(chan []byte, queueSize),
		Decisions: make(chan *net.Decision, queueSize),
	}
}

// DrainQueues unblocks proposal and decision queues.
func (p *Proxy) DrainQueues() {
	for len(p.Proposals) > 0 {
		<-p.Proposals
	}
	for len(p.Decisions) > 0 {
		<-p.Decisions
	}
}

// Deliver delivers a block committed by the consensus protocol.
func (p *Proxy) Deliver(epoch int64, block *consensus.Block) {
	p.Decisions <- &net.Decision{
		Instance: uint64(block.Height),
		Value:    block.Value,
		ValueID:  net.ValueID(block.Value),
	}
}

// GetValue returns a value to be proposed in the consensus protocol.
func (p *Proxy) GetValue() types.Value {
	select {
	case value := <-p.Proposals:
		return value
	default:
		return nil
	}
}
