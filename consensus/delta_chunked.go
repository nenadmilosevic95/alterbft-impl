package consensus

import (
	"fmt"
	"time"
)

// Consensus implement one epoch of consensus.
type DeltaChunkedProtocol struct {
	Epoch   int64   // Consensus epoch identifier
	Process Process // Auxiliary methods implementation

	chunksNumber int

	started   bool
	timeStart time.Time

	counters [5]int
	channels [5]chan *Message
	cnt      int
}

// NewConsensus creates a consensus instance for the provided epoch.
func NewDeltaChunkedProtocol(epoch int64, process Process, chunksNumber int) *DeltaChunkedProtocol {
	c := &DeltaChunkedProtocol{
		Epoch:        epoch,
		Process:      process,
		chunksNumber: chunksNumber,
	}
	for i, _ := range c.channels {
		c.channels[i] = make(chan *Message)
		c.counters[i] = 0
	}
	return c
}

// Start this epoch of consensus
func (c *DeltaChunkedProtocol) Start(validCertificate *Certificate, lockedCertificate *Certificate) {
	c.startListeningRoutines()
	c.timeStart = time.Now()
	for i := 0; i < c.chunksNumber; i++ {
		// we need to create a new message because otherwise we can have concurrency issues while marshalling the message
		msg := NewDeltaRequestMessage(c.Process.GetValue(), c.Process.ID())
		go c.Process.Send(msg, c.Process.ID())
	}

}

func (c *DeltaChunkedProtocol) startListeningRoutines() {
	for i, ch := range c.channels {
		go c.ProcessMessageRoutine(ch, i)
	}
}

// Started informs whether this epoch has been started.
func (c *DeltaChunkedProtocol) Started() bool {
	return c.started
}

// Stop this instance of consensus.
func (c *DeltaChunkedProtocol) Stop() {
	c.started = false
}

func (c *DeltaChunkedProtocol) GetEpoch() int64 {
	return c.Epoch
}

// ProcessMessage processes a consensus message.
//
// Contract: message belongs to this epoch of consensus.
func (c *DeltaChunkedProtocol) ProcessMessage(message *Message) {
	c.channels[message.Sender] <- message
}

func (c *DeltaChunkedProtocol) ProcessMessageRoutine(messageChan chan *Message, senderID int) {
	for {
		message, more := <-messageChan
		if !more {
			fmt.Printf("Sender %d channel closed\n", senderID)
			return
		}
		if message.Sender != senderID {
			panic(fmt.Errorf("Channel %v received message from sender %v\n", senderID, message.Sender))
		}
		c.processMessage(message)
	}
}

func (c *DeltaChunkedProtocol) processMessage(message *Message) {
	switch message.Type {
	case DELTA_REQUEST:
		c.counters[message.Sender]++
		if c.counters[message.Sender] == c.chunksNumber {
			for i := 0; i < c.chunksNumber; i++ {
				m := NewDeltaResponseMessage(message.payload, c.Process.ID())
				go c.Process.Send(m, message.Sender)
			}
			c.counters[message.Sender] = 0
		}
	case DELTA_RESPONSE:
		c.cnt++
		if c.cnt == c.chunksNumber {
			duration := time.Now().Sub(c.timeStart).Milliseconds()
			fmt.Printf("%v DeltaStat: Process %v (%v) received forwarded proposal from %v (%v) in %v ms\n", time.Now(), c.Process.ID(), c.Process.ID()%5, message.Sender, message.Sender%5, duration)
			c.cnt = 0
			nextProcess := (message.Sender + 1) % c.Process.NumProcesses()
			c.timeStart = time.Now()
			for i := 0; i < c.chunksNumber; i++ {
				msg := NewDeltaRequestMessage(c.Process.GetValue(), c.Process.ID())
				go c.Process.Send(msg, nextProcess)
			}
		}
	}
}

func (c *DeltaChunkedProtocol) ProcessTimeout(timeout *Timeout) {

}
