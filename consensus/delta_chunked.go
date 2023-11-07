package consensus

import (
	"fmt"
	"sync/atomic"
	"time"
)

// Consensus implement one epoch of consensus.
type DeltaChunkedProtocol struct {
	Epoch   int64   // Consensus epoch identifier
	Process Process // Auxiliary methods implementation

	chunksNumber int

	started   bool
	timeStart time.Time

	counters [5]int32
	cnt      int32
}

// NewConsensus creates a consensus instance for the provided epoch.
func NewDeltaChunkedProtocol(epoch int64, process Process, chunksNumber int) *DeltaChunkedProtocol {
	c := &DeltaChunkedProtocol{
		Epoch:        epoch,
		Process:      process,
		chunksNumber: chunksNumber,
	}
	return c
}

// Start this epoch of consensus
func (c *DeltaChunkedProtocol) Start(validCertificate *Certificate, lockedCertificate *Certificate) {
	msg := NewDeltaRequestMessage(c.Process.GetValue(), c.Process.ID())
	c.timeStart = time.Now()
	for i := 0; i < c.chunksNumber; i++ {
		go c.Process.Send(msg, c.Process.ID())
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
	go c.processMessage(message)
}

func (c *DeltaChunkedProtocol) processMessage(message *Message) {
	switch message.Type {
	case DELTA_REQUEST:
		atomic.AddInt32(&c.counters[message.Sender], 1)
		if atomic.LoadInt32(&c.counters[message.Sender]) == int32(c.chunksNumber) {
			m := NewDeltaResponseMessage(message.payload, c.Process.ID())
			for i := 0; i < c.chunksNumber; i++ {
				go c.Process.Send(m, message.Sender)
			}
			atomic.StoreInt32(&c.counters[message.Sender], 0)
		}
	case DELTA_RESPONSE:
		atomic.AddInt32(&c.cnt, 1)
		if atomic.LoadInt32(&c.cnt) == int32(c.chunksNumber) {
			duration := time.Now().Sub(c.timeStart).Milliseconds()
			fmt.Printf("%v DeltaStat: Process %v (%v) received forwarded proposal from %v (%v) in %v ms\n", time.Now(), c.Process.ID(), c.Process.ID()%5, message.Sender, message.Sender%5, duration)
			atomic.StoreInt32(&c.cnt, 0)
			nextProcess := (message.Sender + 1) % c.Process.NumProcesses()
			msg := NewDeltaRequestMessage(c.Process.GetValue(), c.Process.ID())
			c.timeStart = time.Now()
			for i := 0; i < c.chunksNumber; i++ {
				go c.Process.Send(msg, nextProcess)
			}
		}
	}
}

func (c *DeltaChunkedProtocol) ProcessTimeout(timeout *Timeout) {

}
