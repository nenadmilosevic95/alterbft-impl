package gossip

import (
	"math/rand"
	"time"

	"dslab.inf.usi.ch/tendermint/net"
	"dslab.inf.usi.ch/tendermint/net/libp2p"
)

func NewUnicastTransport(host *libp2p.Host, log net.Log, msgLossRate float64) *Gossip {
	peers := NewPeersTable()
	transport := &Gossip{
		Log:   log,
		Host:  host,
		Peers: peers,

		Cache:   NewMessageCache(),
		Network: NewNetwork(host, peers),

		BroadcastQueue: NewMessageQueue(BroadcastQueueSize, false),
		DeliveryQueue:  NewDeliveryQueue(DeliveryQueueSize, false),
		PeerRecvQueue:  NewMessageQueue(RecvQueueSize, RecvQueueDrop),
		UnicastQueue:   NewMessageQueue(BroadcastQueueSize, false),

		msgLossRand: rand.New(rand.NewSource(time.Now().UnixNano())),
		msgLossRate: msgLossRate,

		statsQueue: make(chan *Stats, 32),
	}
	go transport.unicastMainLoop()
	return transport
}

func (g *Gossip) Send(message net.Message, pids ...int) {
	gmessage := &Message{
		Sender:  uint16(g.Host.ID),
		Message: message,
		to:      pids,
	}
	g.UnicastQueue.Add(gmessage)
}

func (g *Gossip) sendUnicastMessage(message *Message) {
	for _, pid := range message.to {
		if pid == g.Host.ID { // do not sent, but delivery
			g.DeliveryQueue.Add(message)
			continue
		}
		if pid < 0 || pid >= len(g.PeerSendQueuesByID) {
			g.Log.Println("Invalid destination", pid,
				"for message", message.Message)
			continue
		}
		if g.PeerSendQueuesByID[pid] == nil {
			g.Log.Println("Not connected to destination",
				pid, "of message", message.Message)
			continue
		}
		g.PeerSendQueuesByID[pid].Add(message)
	}
}

func (g *Gossip) unicastMainLoop() {
	ticker := time.Tick(StatsInterval)
	for {
		select {
		case message := <-g.UnicastQueue.Chan():
			g.sendUnicastMessage(message)

		case message := <-g.BroadcastQueue.Chan():
			g.DeliveryQueue.Add(message)
			for i := range g.PeerSendQueues { // Broadcast
				g.PeerSendQueues[i].Add(message)
			}

		case message := <-g.PeerRecvQueue.Chan():
			// here we add message loss
			g.msgsReceived++
			if g.msgLossRate > 0 {
				flip := g.msgLossRand.Float64() * 100
				if flip < g.msgLossRate {
					g.msgsLost++
					continue
				}
			}
			g.DeliveryQueue.Add(message)

		case peer := <-g.Peers.Active:
			if peer.Active {
				g.addNeighbor(&peer)
			} else {
				g.deactivateNeighbor(&peer)
			}

		case <-ticker:
			g.statsReport()
		}
	}
}
