package consensus

import (
	"fmt"
	"time"
)

// Consensus implement one epoch of consensus.
type DeltaProtocol struct {
	Epoch   int64   // Consensus epoch identifier
	Process Process // Auxiliary methods implementation

	started   bool
	timeStart time.Time
}

// NewConsensus creates a consensus instance for the provided epoch.
func NewDeltaProtocol(epoch int64, process Process) *DeltaProtocol {
	c := &DeltaProtocol{
		Epoch:   epoch,
		Process: process,
	}
	return c
}

// Start this epoch of consensus
func (c *DeltaProtocol) Start(validCertificate *Certificate, lockedCertificate *Certificate) {
	msg := NewDeltaRequestMessage(c.Process.GetValue(), c.Process.ID())
	c.timeStart = time.Now()
	c.Process.Send(msg, c.Process.ID())
}

// Started informs whether this epoch has been started.
func (c *DeltaProtocol) Started() bool {
	return c.started
}

// Stop this instance of consensus.
func (c *DeltaProtocol) Stop() {
	c.started = false
}

func (c *DeltaProtocol) GetEpoch() int64 {
	return c.Epoch
}

// ProcessMessage processes a consensus message.
//
// Contract: message belongs to this epoch of consensus.
func (c *DeltaProtocol) ProcessMessage(message *Message) {
	switch message.Type {
	case DELTA_REQUEST:
		m := NewDeltaResponseMessage(message.payload, c.Process.ID())
		c.Process.Send(m, message.Sender)
	case DELTA_RESPONSE:
		duration := time.Now().Sub(c.timeStart).Milliseconds()
		fmt.Printf("DeltaStat: Process %v (%v) received forwarded proposal from %v (%v) in %v ms\n", c.Process.ID(), c.Process.ID()%5, message.Sender, message.Sender%5, duration)
		nextProcess := (message.Sender + 1) % c.Process.NumProcesses()
		msg := NewDeltaRequestMessage(c.Process.GetValue(), c.Process.ID())
		c.timeStart = time.Now()
		c.Process.Send(msg, nextProcess)
	}

}

func (c *DeltaProtocol) ProcessTimeout(timeout *Timeout) {

}
