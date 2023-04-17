package mock

import "dslab.inf.usi.ch/tendermint/net"

// Gossip is a mock implementation of net.Gossip interface.
type Gossip struct {
	RecvQueue chan net.Message
	SendQueue chan net.Message
}

// NewGossip creates a mock gossip implementation.
func NewGossip(queueSize int) *Gossip {
	return &Gossip{
		RecvQueue: make(chan net.Message, queueSize),
		SendQueue: make(chan net.Message, queueSize),
	}
}

// Broadcast implements net.Gossip.Broadcast()
func (g *Gossip) Broadcast(message net.Message) {
	g.SendQueue <- message
	g.RecvQueue <- message
}

// DrainQueues unblocks send and receive queues.
func (g *Gossip) DrainQueues() {
	for len(g.RecvQueue) > 0 {
		<-g.RecvQueue
	}
	for len(g.SendQueue) > 0 {
		<-g.SendQueue
	}
}

// Receive implements net.Gossip.Receive()
func (g *Gossip) Receive() net.Message {
	return <-g.RecvQueue
}

// ReceiveQueue implements net.Gossip.ReceiveQueue()
func (g *Gossip) ReceiveQueue() <-chan net.Message {
	return g.RecvQueue
}
