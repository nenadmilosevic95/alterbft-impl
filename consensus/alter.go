package consensus

import "fmt"

//import "fmt"

const MIN_HEIGHT = 0
const MIN_EPOCH = 0

// EpochPhase: phase of an epoch of consensus.
const (
	Inactive    = iota // epoch hasn't started yet or has finished
	Ready              // phase before any of the certificates is observed
	Locked             // phase after Ce(Bk) certificate is observed
	Commit             // phase after deciding but before receiving all the blocks
	EpochChange        // phase after Ce(SILENCE) or Ce(EQUIV) is observed
	Finished
)

// Consensus implement one epoch of consensus.
type AlterBFT struct {
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
func NewAlterBFT(epoch int64, process Process) *AlterBFT {
	c := &AlterBFT{
		Epoch:   epoch,
		Process: process,
	}
	c.Init()
	return c
}

// Init initializes consensus variables.
func (c *AlterBFT) Init() {
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
func (c *AlterBFT) Start(validCertificate *Certificate, lockedCertificate *Certificate) {
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
func (c *AlterBFT) Started() bool {
	return c.epochPhase > Inactive
}

// Stop this instance of consensus.
func (c *AlterBFT) Stop() {
	c.epochPhase = Finished
}

func (c *AlterBFT) GetEpoch() int64 {
	return c.Epoch
}

// ProcessMessage processes a consensus message.
//
// Contract: message belongs to this epoch of consensus.
func (c *AlterBFT) ProcessMessage(message *Message) {
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

func (c *AlterBFT) processProposal(proposal *Message) {
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

func (c *AlterBFT) checkProposalValidity(proposal *Message) bool {
	// Check if the new proposal is valid!
	isFromProposer := proposal.Sender == c.Process.Proposer(proposal.Epoch)
	correspondToCertificate := (proposal.Certificate == nil && proposal.Block.Height == MIN_HEIGHT) ||
		(proposal.Certificate != nil && proposal.Block.PrevBlockID.Equal(proposal.Certificate.BlockID()))
	isValidProposal := isFromProposer && correspondToCertificate
	return isValidProposal
}

func (c *AlterBFT) tryToVote() {
	if c.hasVoted || c.Process.Proposer(c.Epoch) == c.Process.ID() ||
		c.epochPhase != Ready || c.Proposals.Count() != 1 {
		return
	}

	proposal := c.Proposals.proposals[0]
	shouldVote := proposal.Certificate.RanksHigherOrEqual(c.initialLockedCertificate) &&
		c.Process.ExtendValidChain(proposal.Block)

	if shouldVote {
		fmt.Printf("Honest process %v voted for %v in epoch %v.\n", c.Process.ID(), proposal.Block.BlockID()[0:4], c.Epoch)
		proposal.setFwdSender(c.Process.ID())
		c.Process.Forward(proposal)
		proposerVote := NewVoteMessage(proposal.Epoch, proposal.Block.BlockID(), proposal.Block.Height, int16(proposal.Sender))
		proposerVote.Signature = proposal.Signature
		c.processVote(proposerVote)
		c.Process.Forward(proposerVote)
		c.broadcastVote(VOTE, proposal.Block)
		c.hasVoted = true
	}
}

func (c *AlterBFT) checkEquivocation() {
	if len(c.Votes.certificates) < 2 {
		return
	}
	proposerID := c.Process.Proposer(c.Epoch)
	equivocationDetected := false
	proposerVotes := make([]*Message, 2)
	for _, c := range c.Votes.certificates {
		vote := c.ReconstructMessage(proposerID)
		if vote != nil {
			proposerVotes = append(proposerVotes, vote)
		}
		// Equivocation detected.
		if len(proposerVotes) > 1 {
			equivocationDetected = true
			break
		}
	}
	if !equivocationDetected {
		return
	}
	if c.epochPhase == Ready {
		c.epochPhase = EpochChange
		c.scheduleTimeout(TimeoutQuitEpoch)
	}
	if c.epochPhase == Locked { // process received Ce(Bk) before this one
		fmt.Printf("Process %v epoch %v lock+nodec+equiv value %v\n", c.Process.ID(), c.Epoch, c.validCertificate.BlockID()[0:4])
		c.epochPhase = Finished
	}
	for _, vote := range proposerVotes {
		c.Process.Forward(vote)
	}
}

func (c *AlterBFT) processVote(vote *Message) {
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
		// Check equivocation!
		if vote.Sender == c.Process.Proposer(c.Epoch) {
			c.checkEquivocation()
		}
		if blockCert.SignatureCount() > c.Process.NumProcesses()/2 {
			c.processBlockCertificate(blockCert)
		}
	}
}

// processBlockCertificate is called when cert has quorum of signatures and proposal.
func (c *AlterBFT) processBlockCertificate(cert *Certificate) {
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

func (c *AlterBFT) processSilence(silence *Message) {
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

func (c *AlterBFT) processSilenceCertificate(cert *Certificate) {
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

func (c *AlterBFT) processQuitEpoch(quitEpoch *Message) {
	cert := quitEpoch.Certificate
	messages := cert.ReconstructMessages()
	for _, m := range messages {
		c.ProcessMessage(m)
	}
}

// ProcessMessage processes a consensus timeout.
//
// Contract: timeout belongs to this instance (height) of consensus.
func (c *AlterBFT) ProcessTimeout(timeout *Timeout) {
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

func (c *AlterBFT) processTimeoutPropose() {
	c.scheduledTimeouts[TimeoutPropose] = false
	if c.epochPhase == Ready && c.hasVoted == false {
		c.broadcastSilence()
	}
}

func (c *AlterBFT) processTimeoutEquivocation() {

	c.scheduledTimeouts[TimeoutEquivocation] = false
	if c.epochPhase == Locked {
		// decision
		c.decision = c.lockedCertificate.BlockID()
		c.epochPhase = Commit
		c.tryToCommit()
	}
}

func (c *AlterBFT) tryToCommit() {
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

func (c *AlterBFT) processTimeoutQuitEpoch() {
	c.scheduledTimeouts[TimeoutQuitEpoch] = false
	if c.epochPhase == EpochChange {
		c.epochPhase = Finished
		fmt.Printf("Process %v epoch %v nolock+nodecision\n", c.Process.ID(), c.Epoch)
		c.Process.Finish(c.Epoch, c.validCertificate, c.lockedCertificate, nil)
	}
}

func (c *AlterBFT) broadcastProposal() {
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

func (c *AlterBFT) broadcastVote(voteType int16, block *Block) {
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

func (c *AlterBFT) broadcastSilence() {
	silence := &Message{
		Type:   SILENCE,
		Epoch:  c.Epoch,
		Sender: c.Process.ID(),
	}
	//fmt.Printf("Process %v (%v) epoch %v sent silence.\n", c.Process.ID(), c.Process.ID()%5, c.Epoch)
	c.Process.Broadcast(silence)

}

func (c *AlterBFT) broadcastQuitEpoch(certificate *Certificate) {
	quitEpoch := &Message{
		Type:        QUIT_EPOCH,
		Epoch:       c.Epoch,
		Certificate: certificate,
		Sender:      c.Process.ID(),
	}
	c.Process.Broadcast(quitEpoch)
}

// Schedule a timeout for the epoch phase, if not already scheduled.
func (c *AlterBFT) scheduleTimeout(timeoutType int16) {
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
