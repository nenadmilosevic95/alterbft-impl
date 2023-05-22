package consensus

type AlterBFTSlowLeader struct {
	started bool
}

// NewConsensus creates a consensus instance for the provided epoch.
func NewAlterBFTSlowLeader() *AlterBFTSlowLeader {
	c := &AlterBFTSlowLeader{}
	return c
}

// Start this epoch of consensus
func (c *AlterBFTSlowLeader) Start(validCertificate *Certificate, lockedCertificate *Certificate) {
	c.started = true
}

// Started informs whether this epoch has been started.
func (c *AlterBFTSlowLeader) Started() bool {
	return c.started
}

// Stop this instance of consensus.
func (c *AlterBFTSlowLeader) Stop() {
	c.started = false
}

func (c *AlterBFTSlowLeader) GetEpoch() int64 {
	return 0
}

// ProcessMessage processes a consensus message.
//
// Contract: message belongs to this epoch of consensus.
func (c *AlterBFTSlowLeader) ProcessMessage(message *Message) {
}

func (c *AlterBFTSlowLeader) processProposal(proposal *Message) {

}

// ProcessMessage processes a consensus timeout.
//
// Contract: timeout belongs to this instance (height) of consensus.
func (c *AlterBFTSlowLeader) ProcessTimeout(timeout *Timeout) {

}
