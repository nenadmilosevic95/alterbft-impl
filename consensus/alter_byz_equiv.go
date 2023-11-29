package consensus

import "fmt"

// Consensus implement one epoch of consensus.
type AlterBFTEquivLeader struct {
	Epoch   int64   // Consensus epoch identifier
	Process Process // Auxiliary methods implementation

	epochPhase int

	fastAlterEnabled bool

	lockedCertificate     *Certificate
	sentLockedCertificate bool

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
func NewAlterBFTEquivLeader(epoch int64, process Process, fastAlterEnabled bool) *AlterBFTEquivLeader {
	c := &AlterBFTEquivLeader{
		Epoch:            epoch,
		Process:          process,
		fastAlterEnabled: fastAlterEnabled,
	}
	c.Init()
	return c
}

// Init initializes consensus variables.
func (c *AlterBFTEquivLeader) Init() {
	c.epochPhase = Inactive
	c.Proposals = NewProposalSet()
	c.SilenceCertificate = NewSilenceCertificate(c.Epoch)
	c.Votes = NewCertificateSet()
	// TO DO: here we need to add also other sets for different models
	c.Precommits = nil
	c.Commits = nil
	c.scheduledTimeouts = make([]bool, TimeoutEpochChange+1)
	c.hasVoted = false
	c.sentLockedCertificate = false
}

// Start this epoch of consensus
func (c *AlterBFTEquivLeader) Start(lockedCertificate *Certificate, sentLockedCertificate bool) {
	c.lockedCertificate = lockedCertificate
	c.sentLockedCertificate = sentLockedCertificate
	c.epochPhase = Ready
	if c.Process.Proposer(c.Epoch) == c.Process.ID() {
		if c.Epoch == MIN_EPOCH || c.lockedCertificate.Epoch == c.Epoch-1 {
			c.broadcastTwoProposals()
		} else {
			c.scheduleTimeout(TimeoutEpochChange)
		}
	} else {
		if !c.sentLockedCertificate && c.lockedCertificate != nil {
			c.sendCertificateToLeader()
		}
		c.scheduleTimeout(TimeoutPropose)
	}
	// Process messages that process received before starting an epoch.
	for _, m := range c.messages {
		c.ProcessMessage(m)
	}
}

// Started informs whether this epoch has been started.
func (c *AlterBFTEquivLeader) Started() bool {
	return c.epochPhase > Inactive
}

// Stop this instance of consensus.
func (c *AlterBFTEquivLeader) Stop() {
	c.epochPhase = Finished
}

func (c *AlterBFTEquivLeader) GetEpoch() int64 {
	return c.Epoch
}

// ProcessMessage processes a consensus message.
//
// Contract: message belongs to this epoch of consensus.
func (c *AlterBFTEquivLeader) ProcessMessage(message *Message) {
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

func (c *AlterBFTEquivLeader) processProposal(proposal *Message) {
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
	c.tryToVote(proposal)
	// maybe this block vas missing
	c.tryToCommit()

}

func (c *AlterBFTEquivLeader) checkProposalValidity(proposal *Message) bool {
	// Check if the new proposal is valid!
	isFromProposer := proposal.Sender == c.Process.Proposer(proposal.Epoch)
	correspondToCertificate := (proposal.Certificate == nil && proposal.Block.Height == MIN_HEIGHT) ||
		(proposal.Certificate != nil && proposal.Block.PrevBlockID.Equal(proposal.Certificate.BlockID()))
	isValidProposal := isFromProposer && correspondToCertificate
	return isValidProposal
}

func (c *AlterBFTEquivLeader) tryToVote(proposal *Message) {
	// always vote
	fmt.Printf("Byzantine process %v voted for %v in epoch %v.\n", c.Process.ID(), proposal.Block.BlockID()[0:4], c.Epoch)
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
	if proposal.Certificate.RanksHigherOrEqual(c.lockedCertificate) {
		c.sentLockedCertificate = false
		c.lockedCertificate = proposal.Certificate
	}
}

func (c *AlterBFTEquivLeader) checkEquivocation() {
	if len(c.Votes.certificates) < 2 {
		return
	}

	if c.epochPhase == Ready {
		c.epochPhase = EpochChange
		fmt.Printf("Process %v epoch %v nolock+nodec+equiv value %v\n", c.Process.ID(), c.Epoch, c.lockedCertificate.BlockID()[0:4])
		if c.fastAlterEnabled {
			c.scheduleTimeout(TimeoutQuitEpoch)
		} else {
			fmt.Printf("Process %v epoch %v nolock+nodecision\n", c.Process.ID(), c.Epoch)
			c.Process.Finish(c.Epoch, c.lockedCertificate, c.sentLockedCertificate)
		}
	}
	if c.epochPhase == Locked { // process received Ce(Bk) before this one
		fmt.Printf("Process %v epoch %v lock+nodec+equiv value %v\n", c.Process.ID(), c.Epoch, c.lockedCertificate.BlockID()[0:4])
		c.epochPhase = Finished
	}
	// the byzantine process will not send equivocation certificate
	//c.Process.Forward(vote1)
	//c.Process.Forward(vote2)

}

func (c *AlterBFTEquivLeader) processVote(vote *Message) {
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
		if c.fastAlterEnabled && blockCert.SignatureCount() == c.Process.NumProcesses() && c.epochPhase == Locked {
			// decision
			c.decision = c.lockedCertificate.BlockID()
			c.epochPhase = Commit
			c.tryToCommit()
		}
	}
}

// processBlockCertificate is called when cert has quorum of signatures and proposal.
func (c *AlterBFTEquivLeader) processBlockCertificate(cert *Certificate) {
	if c.epochPhase == Locked || c.epochPhase == Commit || c.epochPhase == Finished {
		return
	}
	if cert.RanksHigherOrEqual(c.lockedCertificate) {
		c.lockedCertificate = cert
		c.sentLockedCertificate = false
	}
	if cert.Epoch == c.Epoch {
		if c.epochPhase == Ready {
			c.epochPhase = Locked
			c.scheduleTimeout(TimeoutEquivocation)
		}
		if c.epochPhase == EpochChange {
			c.epochPhase = Finished
			fmt.Printf("Process %v epoch %v nolock+block value %v\n", c.Process.ID(), c.Epoch, c.lockedCertificate.BlockID()[0:4])
		}
		// Whenever we receive Ce(Bk) in epoch e we can finish epoch e and start epoch e+1
		c.broadcastQuitEpoch(cert)
		c.sentLockedCertificate = true
		c.Process.Finish(c.Epoch, c.lockedCertificate, c.sentLockedCertificate)
	}
}

func (c *AlterBFTEquivLeader) processSilence(silence *Message) {
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

func (c *AlterBFTEquivLeader) processSilenceCertificate(cert *Certificate) {
	if c.epochPhase == Ready { // this is the first certificate process has received
		c.epochPhase = EpochChange
		//fmt.Printf("Process %v in epoch %v didn't lock!\n", c.Process.ID(), c.Epoch)
		c.broadcastQuitEpoch(cert)
		if c.fastAlterEnabled {
			c.scheduleTimeout(TimeoutQuitEpoch)
		} else {
			fmt.Printf("Process %v epoch %v nolock+nodecision\n", c.Process.ID(), c.Epoch)
			c.Process.Finish(c.Epoch, c.lockedCertificate, c.sentLockedCertificate)
		}
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

func (c *AlterBFTEquivLeader) processQuitEpoch(quitEpoch *Message) {
	cert := quitEpoch.Certificate
	messages := cert.ReconstructMessages(c.Process.Proposer(c.Epoch))
	for _, m := range messages {
		c.ProcessMessage(m)
	}
}

// ProcessMessage processes a consensus timeout.
//
// Contract: timeout belongs to this instance (height) of consensus.
func (c *AlterBFTEquivLeader) ProcessTimeout(timeout *Timeout) {
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
	case TimeoutEpochChange:
		c.processTimeoutEpochChange()
	}
}

func (c *AlterBFTEquivLeader) processTimeoutPropose() {
	c.scheduledTimeouts[TimeoutPropose] = false
	if c.epochPhase == Ready && c.hasVoted == false {
		c.broadcastSilence()
	}
}

func (c *AlterBFTEquivLeader) processTimeoutEquivocation() {

	c.scheduledTimeouts[TimeoutEquivocation] = false
	if c.epochPhase == Locked {
		// decision
		fmt.Println("no-fast-path-decision")
		c.decision = c.lockedCertificate.BlockID()
		c.epochPhase = Commit
		c.tryToCommit()
	}
}

func (c *AlterBFTEquivLeader) tryToCommit() {
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

func (c *AlterBFTEquivLeader) processTimeoutQuitEpoch() {
	c.scheduledTimeouts[TimeoutQuitEpoch] = false
	if c.epochPhase == EpochChange {
		c.epochPhase = Finished
		fmt.Printf("Process %v epoch %v nolock+nodecision\n", c.Process.ID(), c.Epoch)
		c.Process.Finish(c.Epoch, c.lockedCertificate, c.sentLockedCertificate)
	}
}

func (c *AlterBFTEquivLeader) processTimeoutEpochChange() {
	c.scheduledTimeouts[TimeoutEpochChange] = false
	c.broadcastTwoProposals()
}

func (c *AlterBFTEquivLeader) broadcastTwoProposals() {
	value1 := c.Process.GetValue()
	value2 := c.Process.GetValue()
	if value1 == nil || value2 == nil {
		fmt.Printf("Generator returned nil in epoch %v\n", c.Epoch)
		return
	}
	var prevBlockID BlockID
	var height int64 = MIN_HEIGHT
	if c.lockedCertificate != nil {
		prevBlockID = c.lockedCertificate.BlockID()
		height = c.lockedCertificate.Height + 1
	}
	block1 := &Block{
		Value:       value1,
		Height:      height,
		PrevBlockID: prevBlockID,
	}
	block2 := &Block{
		Value:       value2,
		Height:      height,
		PrevBlockID: prevBlockID,
	}
	proposal1 := &Message{
		Type:        PROPOSE,
		Epoch:       c.Epoch,
		Block:       block1,
		Certificate: c.lockedCertificate,
		Sender:      c.Process.ID(),
		SenderFwd:   c.Process.ID(),
	}
	proposal2 := &Message{
		Type:        PROPOSE,
		Epoch:       c.Epoch,
		Block:       block2,
		Certificate: c.lockedCertificate,
		Sender:      c.Process.ID(),
		SenderFwd:   c.Process.ID(),
	}
	c.sendToTheFirstHalf(proposal1)
	c.sendToTheSecondHalf(proposal2)
}

func (c *AlterBFTEquivLeader) sendToTheFirstHalf(proposal *Message) {
	for i := 0; i < c.Process.NumProcesses()/2; i++ {
		c.Process.Send(proposal, i)
	}
}

func (c *AlterBFTEquivLeader) sendToTheSecondHalf(proposal *Message) {
	for i := c.Process.NumProcesses() / 2; i < c.Process.NumProcesses(); i++ {
		c.Process.Send(proposal, i)
	}
}

func (c *AlterBFTEquivLeader) broadcastVote(voteType int16, block *Block) {
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

func (c *AlterBFTEquivLeader) broadcastSilence() {
	silence := &Message{
		Type:   SILENCE,
		Epoch:  c.Epoch,
		Sender: c.Process.ID(),
	}
	//fmt.Printf("Process %v (%v) epoch %v sent silence.\n", c.Process.ID(), c.Process.ID()%5, c.Epoch)
	c.Process.Broadcast(silence)

}

func (c *AlterBFTEquivLeader) broadcastQuitEpoch(certificate *Certificate) {
	quitEpoch := &Message{
		Type:        QUIT_EPOCH,
		Epoch:       certificate.Epoch,
		Certificate: certificate,
		Sender:      c.Process.ID(),
	}
	c.Process.Broadcast(quitEpoch)
}

func (c *AlterBFTEquivLeader) sendCertificateToLeader() {
	quitEpoch := &Message{
		Type:        QUIT_EPOCH,
		Epoch:       c.lockedCertificate.Epoch,
		Certificate: c.lockedCertificate,
		Sender:      c.Process.ID(),
	}
	c.Process.Send(quitEpoch, c.Process.Proposer(c.Epoch))
}

// Schedule a timeout for the epoch phase, if not already scheduled.
func (c *AlterBFTEquivLeader) scheduleTimeout(timeoutType int16) {
	if !c.scheduledTimeouts[timeoutType] {
		duration := c.Process.TimeoutPropose(c.Epoch)
		if timeoutType == TimeoutEquivocation {
			duration = c.Process.TimeoutEquivocation(c.Epoch)
		} else if timeoutType == TimeoutQuitEpoch {
			duration = c.Process.TimeoutQuitEpoch(c.Epoch)
		} else if timeoutType == TimeoutEpochChange {
			duration = c.Process.TimeoutEpochChange(c.Epoch)
		}
		c.Process.Schedule(&Timeout{
			Type:     timeoutType,
			Epoch:    c.Epoch,
			Duration: duration,
		})
		c.scheduledTimeouts[timeoutType] = true
	}
}
