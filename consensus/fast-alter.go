package consensus

import "fmt"

// Consensus implement one epoch of consensus.
type FastAlterBFT struct {
	Epoch   int64   // Consensus epoch identifier
	Process Process // Auxiliary methods implementation

	epochPhase int

	initialLockedCertificate *Certificate
	lockedCertificate        *Certificate
	validCertificate         *Certificate

	// Helpers for storing messages
	Proposals          *ProposalSet
	SilenceCertificate *Certificate
	Votes              *CertificateSet
	Precommits         *CertificateSet
	Commits            *CertificateSet

	// Pending messages
	messages []*Message

	// Avoids duplicated timeouts
	scheduledTimeouts []bool

	hasVoted bool
	decision BlockID
}

// NewConsensus creates a consensus instance for the provided epoch.
func NewFastAlterBFT(epoch int64, process Process) *FastAlterBFT {
	c := &FastAlterBFT{
		Epoch:   epoch,
		Process: process,
	}
	c.Init()
	return c
}

// Init initializes consensus variables.
func (c *FastAlterBFT) Init() {
	c.epochPhase = Inactive
	c.Proposals = NewProposalSet()
	c.SilenceCertificate = NewSilenceCertificate(c.Epoch)
	c.Votes = NewCertificateSet()
	// TO DO: here we need to add also other sets for different models
	c.Precommits = nil
	c.Commits = nil
	c.scheduledTimeouts = make([]bool, TimeoutQuitEpoch+1)
	c.hasVoted = false
}

// Start this epoch of consensus
func (c *FastAlterBFT) Start(validCertificate *Certificate, lockedCertificate *Certificate) {
	c.validCertificate = validCertificate
	c.lockedCertificate = lockedCertificate
	c.initialLockedCertificate = lockedCertificate
	c.epochPhase = Ready
	if c.Process.Proposer(c.Epoch) == c.Process.ID() {
		c.broadcastProposal()
	} else {
		c.scheduleTimeout(TimeoutPropose)
	}
	// Process messages that process received before starting an epoch.
	for _, m := range c.messages {
		c.ProcessMessage(m)
	}
}

// Started informs whether this epoch has been started.
func (c *FastAlterBFT) Started() bool {
	return c.epochPhase > Inactive
}

// Stop this instance of consensus.
func (c *FastAlterBFT) Stop() {
	c.epochPhase = Finished
}

func (c *FastAlterBFT) GetEpoch() int64 {
	return c.Epoch
}

// ProcessMessage processes a consensus message.
//
// Contract: message belongs to this epoch of consensus.
func (c *FastAlterBFT) ProcessMessage(message *Message) {
	if c.epochPhase == Inactive {
		//fmt.Printf("Message received while in Inactive phase: %v\n", message)
		c.messages = append(c.messages, message)
		return
	}
	if c.epochPhase == Finished {
		return
	}
	switch message.Type {
	case PROPOSE:
		c.processProposal(message)
	case SILENCE:
		c.processSilence(message)
	case VOTE:
		c.processVote(message)
	case QUIT_EPOCH:
		c.processQuitEpoch(message)
	}
}

func (c *FastAlterBFT) processProposal(proposal *Message) {
	// Check if proposal has already been processed!
	if c.Proposals.Has(proposal.Block.BlockID()) {
		return
	}
	if c.checkProposalValidity(proposal) == false {
		fmt.Printf("Invalid proposal.")
		return
	}
	// Try to add new block to the blockchain!
	ok := c.Process.AddBlock(proposal.Block)
	if !ok {
		fmt.Printf("P%v proposal could not be added to the blockchain in epoch %v\n", c.Process.ID(), c.Epoch)
		return
	}
	// Save the proposal
	if proposal.Epoch == c.Epoch {
		c.Proposals.Add(proposal)
	}
	c.tryToVote()
	c.tryToCommit()
}

func (c *FastAlterBFT) checkProposalValidity(proposal *Message) bool {
	// Check if the new proposal is valid!
	isFromProposer := proposal.Sender == c.Process.Proposer(proposal.Epoch)
	correspondToCertificate := (proposal.Certificate == nil && proposal.Block.Height == MIN_HEIGHT) ||
		(proposal.Certificate != nil && proposal.Block.PrevBlockID.Equal(proposal.Certificate.BlockID()))
	isValidProposal := isFromProposer && correspondToCertificate
	return isValidProposal
}

func (c *FastAlterBFT) tryToVote() {
	if c.hasVoted || c.Process.Proposer(c.Epoch) == c.Process.ID() ||
		c.epochPhase == EpochChange || c.Proposals.Count() != 1 {
		return
	}

	proposal := c.Proposals.proposals[0]
	shouldVote := proposal.Certificate.RanksHigherOrEqual(c.initialLockedCertificate) &&
		c.Process.ExtendValidChain(proposal.Block)

	if shouldVote {
		fmt.Printf("Honest process %v voted for %v in epoch %v.\n", c.Process.ID(), proposal.Block.BlockID()[0:4], c.Epoch)
		proposal.setFwdSender(c.Process.ID())
		c.Process.Forward(proposal)
		proposerVote := NewVoteMessage(proposal.Epoch, proposal.Block.BlockID(), proposal.Block.Height, int16(proposal.Sender), int16(proposal.Sender))
		proposerVote.Signature = proposal.Signature
		proposerVote.Signature2 = proposal.Signature
		c.processVote(proposerVote)
		vote := NewVoteMessage(proposal.Epoch, proposal.Block.BlockID(), proposal.Block.Height, int16(c.Process.ID()), int16(proposal.Sender))
		vote.Signature2 = proposal.Signature
		c.Process.Broadcast(vote)
		c.hasVoted = true
	}
}

func (c *FastAlterBFT) checkEquivocation() {
	if len(c.Votes.certificates) < 2 {
		return
	}
	proposerID := c.Process.Proposer(c.Epoch)
	vote1 := c.Votes.certificates[0].ReconstructMessage(proposerID, proposerID)
	vote2 := c.Votes.certificates[1].ReconstructMessage(proposerID, proposerID)

	if c.epochPhase == Ready {
		c.epochPhase = EpochChange
		fmt.Printf("Process %v epoch %v nolock+nodec+equiv value %v\n", c.Process.ID(), c.Epoch, c.validCertificate.BlockID()[0:4])
		c.scheduleTimeout(TimeoutQuitEpoch)
	}
	if c.epochPhase == Locked { // process received Ce(Bk) before this one
		fmt.Printf("Process %v epoch %v lock+nodec+equiv value %v\n", c.Process.ID(), c.Epoch, c.validCertificate.BlockID()[0:4])
		c.epochPhase = Finished
	}

	c.Process.Forward(vote1)
	c.Process.Forward(vote2)

}

func (c *FastAlterBFT) processVote(vote *Message) {
	// we don't process votes in locked phase because we know that
	// process already received Ce(Bk) for some block Bk in epoch e
	if c.epochPhase != Commit {
		blockCert := c.Votes.Get(c.Epoch, vote.BlockID, vote.Height)
		if blockCert == nil {
			blockCert = NewBlockCertificate(c.Epoch, vote.BlockID, vote.Height)
			c.Votes.Add(blockCert)
			blockCert.AddSignature(vote.Signature2, vote.Sender2)
		}
		ok := blockCert.AddSignature(vote.Signature, vote.Sender)
		if !ok {
			return
		}

		c.checkEquivocation()

		if blockCert.SignatureCount() > c.Process.NumProcesses()/2 {
			c.processBlockCertificate(blockCert)
		}

		// Fast path commit
		if blockCert.SignatureCount() == c.Process.NumProcesses() && c.epochPhase == Locked {
			// decision
			c.decision = c.lockedCertificate.BlockID()
			c.epochPhase = Commit
			c.tryToCommit()
		}
	}
}

// processBlockCertificate is called when cert has quorum of signatures and proposal.
func (c *FastAlterBFT) processBlockCertificate(cert *Certificate) {
	if c.epochPhase == Locked || c.epochPhase == Commit || c.epochPhase == Finished {
		return
	}
	c.validCertificate = cert
	if c.epochPhase == Ready {
		c.epochPhase = Locked
		c.lockedCertificate = cert
		c.scheduleTimeout(TimeoutEquivocation)
	}
	if c.epochPhase == EpochChange {
		c.epochPhase = Finished
		fmt.Printf("Process %v epoch %v nolock+block value %v\n", c.Process.ID(), c.Epoch, c.validCertificate.BlockID()[0:4])
	}
	// Whenever we receive Ce(Bk) in epoch e we can finish epoch e and start epoch e+1
	c.broadcastQuitEpoch(cert)
	c.Process.Finish(c.Epoch, c.validCertificate, c.lockedCertificate, nil)

}

func (c *FastAlterBFT) processSilence(silence *Message) {
	// we process Silence messages only in phases Ready and Locked,
	// we don't need to process Silence messages in EpochChange phase
	// because we know that we already received Ce(SILENCE) or Ce(EQUIV)
	if c.epochPhase == Ready || c.epochPhase == Locked {
		ok := c.SilenceCertificate.AddSignature(silence.Signature, silence.Sender)
		if !ok {
			return
		}
		if c.SilenceCertificate.SignatureCount() > c.Process.NumProcesses()/2 {
			c.processSilenceCertificate(c.SilenceCertificate)
		}
	}
}

func (c *FastAlterBFT) processSilenceCertificate(cert *Certificate) {
	if c.epochPhase == Ready { // this is the first certificate process has received
		c.epochPhase = EpochChange
		//fmt.Printf("Process %v in epoch %v didn't lock!\n", c.Process.ID(), c.Epoch)
		c.broadcastQuitEpoch(cert)
		c.scheduleTimeout(TimeoutQuitEpoch)
		//c.Process.Decide(c.Epoch, nil)
		return
	}
	if c.epochPhase == Locked {
		c.epochPhase = Finished
		fmt.Printf("Process %v epoch %v lock+nodec+silence value %v\n", c.Process.ID(), c.Epoch, c.lockedCertificate.BlockID()[0:4])
		//c.Process.Decide(c.Epoch, nil)
		return
	}
}

func (c *FastAlterBFT) processQuitEpoch(quitEpoch *Message) {
	cert := quitEpoch.Certificate
	messages := cert.ReconstructMessages(c.Process.Proposer(c.Epoch))
	for _, m := range messages {
		c.ProcessMessage(m)
	}
}

// ProcessMessage processes a consensus timeout.
//
// Contract: timeout belongs to this instance (height) of consensus.
func (c *FastAlterBFT) ProcessTimeout(timeout *Timeout) {
	if c.epochPhase == Finished {
		return
	}
	switch timeout.Type {
	case TimeoutPropose:
		c.processTimeoutPropose()
	case TimeoutEquivocation:
		c.processTimeoutEquivocation()
	case TimeoutQuitEpoch:
		c.processTimeoutQuitEpoch()
	}
}

func (c *FastAlterBFT) processTimeoutPropose() {
	c.scheduledTimeouts[TimeoutPropose] = false
	if c.epochPhase == Ready && c.hasVoted == false {
		c.broadcastSilence()
	}
}

func (c *FastAlterBFT) processTimeoutEquivocation() {

	c.scheduledTimeouts[TimeoutEquivocation] = false
	if c.epochPhase == Locked {
		// decision
		c.decision = c.lockedCertificate.BlockID()
		c.epochPhase = Commit
		c.tryToCommit()
	}
}

func (c *FastAlterBFT) tryToCommit() {
	if c.decision == nil || c.epochPhase != Commit {
		return
	}

	proposal := c.Proposals.Get(c.decision)
	if proposal == nil {
		return
	}

	if c.Process.ExtendValidChain(proposal.Block) {
		fmt.Printf("Process %v epoch %v lock+decision value %v\n", c.Process.ID(), c.Epoch, c.decision[0:4])
		c.epochPhase = Finished
		c.Process.Decide(c.Epoch, proposal.Block)
	}

}

func (c *FastAlterBFT) processTimeoutQuitEpoch() {
	c.scheduledTimeouts[TimeoutQuitEpoch] = false
	if c.epochPhase == EpochChange {
		c.epochPhase = Finished
		fmt.Printf("Process %v epoch %v nolock+nodecision\n", c.Process.ID(), c.Epoch)
		c.Process.Finish(c.Epoch, c.validCertificate, c.lockedCertificate, nil)
	}
}

func (c *FastAlterBFT) broadcastProposal() {
	value := c.Process.GetValue()
	if value == nil {
		fmt.Printf("Generator returned nil in epoch %v\n", c.Epoch)
		return
	}
	var prevBlockID BlockID
	var height int64 = MIN_HEIGHT
	if c.validCertificate != nil {
		prevBlockID = c.validCertificate.BlockID()
		height = c.validCertificate.Height + 1
	}
	block := &Block{
		Value:       value,
		Height:      height,
		PrevBlockID: prevBlockID,
	}
	proposal := &Message{
		Type:        PROPOSE,
		Epoch:       c.Epoch,
		Block:       block,
		Certificate: c.validCertificate,
		Sender:      c.Process.ID(),
		SenderFwd:   c.Process.ID(),
	}
	c.Process.Broadcast(proposal)
}

func (c *FastAlterBFT) broadcastVote(voteType int16, block *Block) {
	vote := &Message{
		Type:    voteType,
		Epoch:   c.Epoch,
		BlockID: block.BlockID(),
		Height:  block.Height,
		Sender:  c.Process.ID(),
	}
	//fmt.Printf("Process %v (%v) epoch %v vote for value %v\n", c.Process.ID(), c.Process.ID()%5, c.Epoch, vote.BlockID[0:4])
	c.Process.Broadcast(vote)
}

func (c *FastAlterBFT) broadcastSilence() {
	silence := &Message{
		Type:   SILENCE,
		Epoch:  c.Epoch,
		Sender: c.Process.ID(),
	}
	//fmt.Printf("Process %v (%v) epoch %v sent silence.\n", c.Process.ID(), c.Process.ID()%5, c.Epoch)
	c.Process.Broadcast(silence)

}

func (c *FastAlterBFT) broadcastQuitEpoch(certificate *Certificate) {
	quitEpoch := &Message{
		Type:        QUIT_EPOCH,
		Epoch:       c.Epoch,
		Certificate: certificate,
		Sender:      c.Process.ID(),
	}
	c.Process.Broadcast(quitEpoch)
}

// Schedule a timeout for the epoch phase, if not already scheduled.
func (c *FastAlterBFT) scheduleTimeout(timeoutType int16) {
	if !c.scheduledTimeouts[timeoutType] {
		duration := c.Process.TimeoutPropose(c.Epoch)
		if timeoutType == TimeoutEquivocation {
			duration = c.Process.TimeoutEquivocation(c.Epoch)
		} else if timeoutType == TimeoutQuitEpoch {
			duration = c.Process.TimeoutQuitEpoch(c.Epoch)
		}
		c.Process.Schedule(&Timeout{
			Type:     timeoutType,
			Epoch:    c.Epoch,
			Duration: duration,
		})
		c.scheduledTimeouts[timeoutType] = true
	}
}
