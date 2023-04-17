package gossip

import "dslab.inf.usi.ch/tendermint/net"

type DeliveryQueue struct {
	channel chan net.Message
	drop    bool
	stats   QueueStats
}

func NewDeliveryQueue(size int, drop bool) *DeliveryQueue {
	return &DeliveryQueue{
		drop:    drop,
		channel: make(chan net.Message, size),
	}
}

func (q *DeliveryQueue) Add(message *Message) {
	select {
	case q.channel <- message.Message:
		if len(q.channel) < cap(q.channel)/2 {
			q.stats.Low += 1
		} else {
			q.stats.High += 1
		}
	default:
		q.stats.Full += 1
		if q.drop {
			break // message is dropped
		}
		q.channel <- message.Message
	}

}

func (q *DeliveryQueue) Chan() <-chan net.Message {
	return q.channel
}

func (q *DeliveryQueue) Stats() QueueStats {
	return q.stats
}
