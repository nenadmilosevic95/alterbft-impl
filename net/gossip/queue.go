package gossip

type MessageQueue struct {
	channel chan *Message
	drop    bool
	stats   QueueStats
}

func NewMessageQueue(size int, drop bool) *MessageQueue {
	return &MessageQueue{
		drop:    drop,
		channel: make(chan *Message, size),
	}
}

func (q *MessageQueue) Add(message *Message) {
	select {
	case q.channel <- message:
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
		q.channel <- message
	}

}

func (q *MessageQueue) Chan() <-chan *Message {
	return q.channel
}

func (q *MessageQueue) Next() *Message {
	return <-q.channel
}

func (q *MessageQueue) Retrieve(batch []*Message) int {
	var index int
	batch[index] = <-q.channel
	for index = 1; index < len(batch) && len(q.channel) > 0; index++ {
		batch[index] = <-q.channel
	}
	return index
}

func (q *MessageQueue) Stats() QueueStats {
	return q.stats
}
