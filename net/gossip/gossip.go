package gossip

import (
	"sync/atomic"
	"time"

	"dslab.inf.usi.ch/tendermint/net"
	"dslab.inf.usi.ch/tendermint/net/libp2p"

	"math/rand"
)

var _ net.Gossip = new(Gossip)
var _ net.Transport = new(Gossip)

var DefaultQueueSize = 32

var BroadcastQueueSize int = 32
var DeliveryQueueSize int = 32

var RecvQueueSize int = 32
var RecvQueueDrop bool = false

var SendQueuesBatchMax int = 0
var SendQueuesBatchMin int = 0

var SendQueuesSize int = 32
var SendQueuesDrop bool = false

var MaxPayloadSize int = 2048

var StatsInterval = time.Second

type Gossip struct {
	Host  *libp2p.Host
	Peers *PeersTable

	// Input queues
	BroadcastQueue *MessageQueue // Messages to be broadcast.
	PeerRecvQueue  *MessageQueue // Messages received from active peers.

	// Output queues
	DeliveryQueue  *DeliveryQueue  // Messages to be delivered.
	PeerSendQueues []*MessageQueue // Messages to send to active peers.

	// Same content as PeerSendQueues, but possibly with nil entries
	PeerSendQueuesByID []*MessageQueue

	// Suport for unicasts
	UnicastQueue *MessageQueue

	// Peer management
	Neighbors []*Peer // List of peers active for gossip.

	Cache *MessageCache

	Log     net.Log
	Network *Network

	Validator ValidatorBuilder
	VFiltered uint32

	statsQueue    chan *Stats
	statsInterval time.Duration

	// Message Loss stats
	msgLossRand  *rand.Rand
	msgLossRate  float64
	msgsReceived int
	msgsLost     int
}

func NewGossipTransport(host *libp2p.Host, log net.Log, msgLossRate float64) *Gossip {
	peers := NewPeersTable()
	transport := &Gossip{
		Log:   log,
		Host:  host,
		Peers: peers,

		Network: NewNetwork(host, peers),
		Cache:   NewMessageCache(),

		BroadcastQueue: NewMessageQueue(BroadcastQueueSize, false),
		DeliveryQueue:  NewDeliveryQueue(DeliveryQueueSize, false),
		PeerRecvQueue:  NewMessageQueue(RecvQueueSize, RecvQueueDrop),

		msgLossRand: rand.New(rand.NewSource(time.Now().UnixNano())),
		msgLossRate: msgLossRate,

		statsQueue: make(chan *Stats, 32),
	}
	go transport.gossipMainLoop()
	return transport
}

// Implements the 'tendermint/net/Gossip' interface
func (g *Gossip) Broadcast(message net.Message) {
	gmessage := &Message{
		Sender:  uint16(g.Host.ID),
		Message: message,
		from:    g.Host.ID,
	}
	g.BroadcastQueue.Add(gmessage)
}

// Implements the 'tendermint/net/Gossip' interface
func (g *Gossip) Receive() net.Message {
	return <-g.DeliveryQueue.Chan()
}

// Implements the 'tendermint/net/Gossip' interface
func (g *Gossip) ReceiveQueue() <-chan net.Message {
	return g.DeliveryQueue.Chan()
}

func (g *Gossip) StatsQueue() chan *Stats {
	return g.statsQueue
}

func (g *Gossip) addNeighbor(peer *Peer) {
	g.Log.Println("added peer", peer.ID, peer.Chosen, peer.Addr)
	sendQueue := NewMessageQueue(SendQueuesSize, SendQueuesDrop)
	g.Neighbors = append(g.Neighbors, peer)
	g.PeerSendQueues = append(g.PeerSendQueues, sendQueue)

	for len(g.PeerSendQueuesByID) <= peer.ID {
		g.PeerSendQueuesByID = append(g.PeerSendQueuesByID, nil)
	}
	g.PeerSendQueuesByID[peer.ID] = sendQueue

	go g.receiver(peer)
	if SendQueuesBatchMax < 1 {
		go g.sender(peer, sendQueue)
		//	} else {
		//	go g.senderBatch(peer, sendQueue)
	}
}

func (g *Gossip) deactivateNeighbor(peer *Peer) {
	g.Log.Println("deactivated peer", peer.ID, peer.Chosen, peer.Addr)
	index := -1
	for i := range g.Neighbors {
		if g.Neighbors[i].ID == peer.ID {
			index = i
			break
		}
	}
	if index < 0 {
		g.Log.Println("failed to deactivate peer", peer.ID)
		return
	}

	neighbors := g.Neighbors
	sendQueues := g.PeerSendQueues
	g.Neighbors = g.Neighbors[:0]
	g.PeerSendQueues = g.PeerSendQueues[:0]
	for i := range neighbors {
		if i == index {
			continue
		}
		g.Neighbors = append(g.Neighbors, neighbors[i])
		g.PeerSendQueues = append(g.PeerSendQueues, sendQueues[i])
	}
}

func (g *Gossip) deliverAndForward(message *Message) {
	g.DeliveryQueue.Add(message)
	for i, sendQueue := range g.PeerSendQueues {
		// Do not send the message to its sources
		if message.from == g.Neighbors[i].ID ||
			int(message.Sender) == g.Neighbors[i].ID {
			continue
		}
		sendQueue.Add(message)
	}
}

//func (g *Gossip) disaggregateReceivedMessages(message *Message) {
//	messages := consensus.DisaggregateMessage(message.Message)
//	for i := range messages {
//		messageID := messages[i].ID()
//		if g.Cache.Contains(messageID) {
//			continue // Discard duplicated messages
//		}
//		g.Cache.Add(messageID)
//		g.deliverAndForward(&Message{
//			Sender:  message.Sender,
//			Message: messages[i],
//			from:    message.from,
//		})
//	}
//}

func (g *Gossip) gossipMainLoop() {
	ticker := time.Tick(StatsInterval)
	for {
		select {
		case message := <-g.BroadcastQueue.Chan():
			g.Cache.Add(message.ID())
			g.deliverAndForward(message)

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
			messageID := message.ID()
			//			if messageID != consensus.AggregatedMessageID {
			if g.Cache.Contains(messageID) {
				break
			}
			g.Cache.Add(messageID)
			g.deliverAndForward(message)
			//			} else { // Same logic for aggregated messages
			//				g.disaggregateReceivedMessages(message)
			//			}

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

func (g *Gossip) receiver(peer *Peer) {
	var err error
	reader := peer.BufferedReader()
	for {
		message := new(Message)
		err = message.ReadFrom(reader)
		if err != nil {
			break
		}
		message.from = peer.ID
		g.PeerRecvQueue.Add(message)
	}
	g.receiverError(peer, err)
}

func (g *Gossip) receiverError(peer *Peer, err error) {
	g.Peers.Lock()
	peer = g.Peers.GetByAddr(peer.Addr)
	peer.Comment = "receiver"
	peer.RecvStreamE = err
	peer.RecvStreamS = 3
	if peer.Active {
		g.Peers.Deactivate(peer)
	}
	g.Peers.Notify(peer)
	g.Peers.Unlock()

}

func (g *Gossip) sender(peer *Peer, sendQueue *MessageQueue) {
	var err error
	var message *Message
	var validator Validator
	if g.Validator != nil {
		validator = g.Validator.New(peer.ID)
	}
	for err == nil {
		message = sendQueue.Next()
		if validator == nil || validator.Validate(message.Message) {
			err = message.WriteTo(peer.SendStream)
		} else {
			atomic.AddUint32(&g.VFiltered, 1)
		}
	}
	g.senderError(peer, err)
}

//func (g *Gossip) senderBatch(peer *Peer, sendQueue *MessageQueue) {
//	var err error
//	var messages []*Message
//	var validator Validator
//	buffer := make([]byte, HeaderSize+consensus.MessageFixSize+MaxPayloadSize)
//	if g.Validator != nil {
//		validator = g.Validator.New(peer.ID)
//	}
//	messages = make([]*Message, SendQueuesBatchMax)
//	for err == nil {
//		count := sendQueue.Retrieve(messages)
//		validated := messages[:count]
//		if validator != nil {
//			validated = validator.Aggregate(validated)
//			if count > len(validated) {
//				atomic.AddUint32(&g.VFiltered,
//					uint32(count-len(validated)))
//			}
//		}
//		for _, message := range validated {
//			n := message.MarshallTo(buffer)
//			_, err = peer.SendStream.Write(buffer[:n])
//		}
//	}
//	g.senderError(peer, err)
//}

func (g *Gossip) senderError(peer *Peer, err error) {
	g.Peers.Lock()
	peer = g.Peers.GetByAddr(peer.Addr)
	peer.Comment = "sender"
	peer.SendStreamE = err
	peer.SendStreamS = 3
	if peer.Active {
		g.Peers.Deactivate(peer)
	}
	g.Peers.Notify(peer)
	g.Peers.Unlock()
}

func (g *Gossip) statsReport() {
	var stats = &Stats{}
	stats.Cache = g.Cache.Stats()
	stats.BQueue = g.BroadcastQueue.Stats()
	stats.DQueue = g.DeliveryQueue.Stats()
	stats.RQueue = g.PeerRecvQueue.Stats()
	stats.SQueues = QueueStats{}
	for i := 0; i < len(g.PeerSendQueues); i++ {
		stats.SQueues.Sum(&g.PeerSendQueues[i].stats)
	}
	stats.Validator.Enqueued = stats.SQueues.Total()
	stats.Validator.Filtered = int(atomic.LoadUint32(&g.VFiltered))
	//adding stats about message loss
	stats.MessageLoss.Received = g.msgsReceived
	stats.MessageLoss.Lost = g.msgsLost
	select {
	case g.statsQueue <- stats:
	default: // drop
	}
}
