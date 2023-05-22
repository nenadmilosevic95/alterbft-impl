package consensus

import "fmt"

// Consensus implement one epoch of consensus.
type AlterBFTEquivLeader struct {
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
func NewAlterBFTEquivLeader(epoch int64, process Process) *AlterBFTEquivLeader {
	c := &AlterBFTEquivLeader{
		Epoch:   epoch,
		Process: process,
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
	c.scheduledTimeouts = make([]bool, TimeoutQuitEpoch+1)
	c.hasVoted = false
}

// Start this epoch of consensus
func (c *AlterBFTEquivLeader) Start(validCertificate *Certificate, lockedCertificate *Certificate) {
	c.validCertificate = validCertificate
	c.lockedCertificate = lockedCertificate
	c.initialLockedCertificate = lockedCertificate
	c.epochPhase = Ready
	if c.Process.Proposer(c.Epoch) == c.Process.ID() {
		c.broadcastTwoProposals()
	} else {
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
	// Save the proposal
	if proposal.Epoch == c.Epoch {
		c.Proposals.Add(proposal)
	}
	c.vote(proposal)
}

func (c *AlterBFTEquivLeader) vote(proposal *Message) {
	fmt.Printf("Byzantine process %v voted for %v in epoch %v.\n", c.Process.ID(), proposal.Block.BlockID()[0:4], c.Epoch)
	proposal.setFwdSender(c.Process.ID())
	c.Process.Forward(proposal)
	proposerVote := NewVoteMessage(proposal.Epoch, proposal.Block.BlockID(), proposal.Block.Height, int16(proposal.Sender))
	proposerVote.Signature = proposal.Signature
	c.processVote(proposerVote)
	c.Process.Forward(proposerVote)
	c.broadcastVote(VOTE, proposal.Block)
}

func (c *AlterBFTEquivLeader) processVote(vote *Message) {
	// we don't process votes in locked phase because we know that
	// process already received Ce(Bk) for some block Bk in epoch e
	if c.epochPhase != Commit {
		blockCert := c.Votes.Get(c.Epoch, vote.BlockID, vote.Height)
		if blockCert == nil {
			blockCert = NewBlockCertificate(c.Epoch, vote.BlockID, vote.Height)
			c.Votes.Add(blockCert)
		}
		ok := blockCert.AddSignature(vote.Signature, vote.Sender)
		if !ok {
			return
		}
		if blockCert.SignatureCount() > c.Process.NumProcesses()/2 {
			c.processBlockCertificate(blockCert)
		}
	}
}

// processBlockCertificate is called when cert has quorum of signatures and proposal.
func (c *AlterBFTEquivLeader) processBlockCertificate(cert *Certificate) {
	if c.epochPhase == Locked || c.epochPhase == Commit || c.epochPhase == Finished {
		return
	}
	c.validCertificate = cert
	if c.epochPhase == Ready {
		c.epochPhase = Locked
		c.lockedCertificate = cert
		//c.scheduleTimeout(TimeoutEquivocation)
	}
	if c.epochPhase == EpochChange {
		c.epochPhase = Finished
		//fmt.Printf("Process %v epoch %v nolock+block value %v\n", c.Process.ID(), c.Epoch, c.validCertificate.BlockID()[0:4])
	}
	// Whenever we receive Ce(Bk) in epoch e we can finish epoch e and start epoch e+1
	c.broadcastQuitEpoch(cert)
	c.Process.Finish(c.Epoch, c.validCertificate, c.lockedCertificate, nil)

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
		c.scheduleTimeout(TimeoutQuitEpoch)
		//c.Process.Decide(c.Epoch, nil)
		return
	}
	if c.epochPhase == Locked {
		c.epochPhase = Finished
		//fmt.Printf("Process %v epoch %v lock+nodecision value %v\n", c.Process.ID(), c.Epoch, c.lockedCertificate.BlockID()[0:4])
		//c.Process.Decide(c.Epoch, nil)
		return
	}
}

func (c *AlterBFTEquivLeader) processQuitEpoch(quitEpoch *Message) {
	cert := quitEpoch.Certificate
	messages := cert.ReconstructMessages()
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
	case TimeoutQuitEpoch:
		c.processTimeoutQuitEpoch()
	}
}

func (c *AlterBFTEquivLeader) processTimeoutPropose() {
	c.scheduledTimeouts[TimeoutPropose] = false
	if c.epochPhase == Ready && c.hasVoted == false {
		c.broadcastSilence()
	}
}

func (c *AlterBFTEquivLeader) processTimeoutQuitEpoch() {
	c.scheduledTimeouts[TimeoutQuitEpoch] = false
	if c.epochPhase == EpochChange {
		c.epochPhase = Finished
		fmt.Printf("Process %v epoch %v nolock+noblock\n", c.Process.ID(), c.Epoch)
		c.Process.Finish(c.Epoch, c.validCertificate, c.lockedCertificate, nil)
	}
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
	if c.validCertificate != nil {
		prevBlockID = c.validCertificate.BlockID()
		height = c.validCertificate.Height + 1
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
		Certificate: c.validCertificate,
		Sender:      c.Process.ID(),
		SenderFwd:   c.Process.ID(),
	}
	proposal2 := &Message{
		Type:        PROPOSE,
		Epoch:       c.Epoch,
		Block:       block2,
		Certificate: c.validCertificate,
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
		Epoch:       c.Epoch,
		Certificate: certificate,
		Sender:      c.Process.ID(),
	}
	c.Process.Broadcast(quitEpoch)
}

// Schedule a timeout for the epoch phase, if not already scheduled.
func (c *AlterBFTEquivLeader) scheduleTimeout(timeoutType int16) {
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
