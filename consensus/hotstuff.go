package consensus

import "fmt"

// Consensus implement one epoch of consensus.
type HotStuff struct {
	Process Process // Auxiliary methods implementation

	epochPhase int

	lockedCertificate *Certificate

	// Helpers for storing messages
	Proposals *ProposalSet
	Votes     *CertificateSet

	// Pending messages
	messages []*Message

	// Avoids duplicated timeouts
	scheduledTimeouts []bool
}

// NewConsensus creates a consensus instance for the provided epoch.
func NewHotStuff(view int64, process Process) *HotStuff {
	c := &HotStuff{
		Process: process,
	}
	c.Init()
	return c
}

// Init initializes consensus variables.
func (c *HotStuff) Init() {
	c.epochPhase = Inactive
	c.Proposals = NewProposalSet()
	c.Votes = NewCertificateSet()
	c.scheduledTimeouts = make([]bool, TimeoutQuitEpoch+1)
}

// Start this epoch of consensus
func (c *HotStuff) Start(validCertificate *Certificate, lockedCertificate *Certificate) {
	c.lockedCertificate = lockedCertificate
	c.epochPhase = Ready

	c.broadcastProposal()

	for _, m := range c.messages {
		c.ProcessMessage(m)
	}
}

// Started informs whether this epoch has been started.
func (c *HotStuff) Started() bool {
	return c.epochPhase > Inactive
}

// Stop this instance of consensus.
func (c *HotStuff) Stop() {
	c.epochPhase = Finished
}

func (c *HotStuff) GetEpoch() int64 {
	return 0
}

// ProcessMessage processes a consensus message.
//
// Contract: message belongs to this epoch of consensus.
func (c *HotStuff) ProcessMessage(message *Message) {
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
	case VOTE:
		c.processVote(message)
	}
}

func (c *HotStuff) processProposal(proposal *Message) {
	block := proposal.Block
	cert := proposal.Certificate
	// Check if the new proposal is valid!
	if !block.PrevBlockID.Equal(cert.BlockID()) || !block.Extend(c.lockedCertificate.block) {
		panic(fmt.Errorf("Equivocated block received in epoch %v\n", block.Height))
	}
	// We try to add block to the blockchain
	// if the predecessor is not present in the blockchain
	// this call will fail!
	ok := c.Process.AddBlock(block)
	if !ok {
		panic(fmt.Errorf("P%v proposal could not be added to the blockchain in epoch %v\n", c.Process.ID(), block.Height))
	}
	// Save the proposal
	c.Proposals.Add(proposal)
	// Update locked certificate
	c.lockedCertificate = cert
	c.lockedCertificate.block = c.Proposals.Get(c.lockedCertificate.blockID).Block
	// Vote for new block
	c.sendVote(VOTE, proposal.Block.BlockID(), block.Height)
	// Commit block.Height-2
	if block.Height >= 2 {
		commitBlock := c.Proposals.Get(c.lockedCertificate.block.PrevBlockID).Block
		c.Process.Decide(commitBlock.Height, commitBlock)
	}

	// Get or create a block certificate
	blockCert := c.Votes.Get(block.Height, block.BlockID(), block.Height)
	// No vote for this proposal before
	if blockCert == nil {
		blockCert = NewBlockCertificate(block.Height, block.BlockID(), block.Height)
		blockCert.block = block
		c.Votes.Add(blockCert)
		// Some vote for the proposal is received before the proposal
	} else {
		blockCert.block = block
		// We need to check whether only proposal was missing
		if blockCert.SignatureCount() == c.Process.NumProcesses()*2/3+1 {
			c.processBlockCertificate(blockCert)
		}
	}
}

func (c *HotStuff) processVote(vote *Message) {
	if c.Process.ID() != c.Process.Proposer(vote.Epoch+1) {
		panic(fmt.Errorf("Non proposer %v received vote in epoch %v\n", c.Process.ID(), vote.Epoch))
	}
	// we don't process votes in commit phase because we know that
	// process already received Ce(Bk) for some block Bk in epoch e
	blockID := vote.BlockID
	blockCert := c.Votes.Get(vote.Height, blockID, vote.Height)
	if blockCert == nil {
		blockCert = NewBlockCertificate(vote.Height, blockID, vote.Height)
		c.Votes.Add(blockCert)
	}
	ok := blockCert.AddSignature(vote.Signature, vote.Sender)
	if !ok {
		panic(fmt.Errorf("Should not receive two same votes!"))
	}
	if blockCert.block != nil && blockCert.SignatureCount() == c.Process.NumProcesses()*2/3+1 {
		c.processBlockCertificate(blockCert)
	}
}

// processBlockCertificate is called:
// when new proposer receives a quroum of votes in the previous view
// b) when a non-proposer receives a new proposal
func (c *HotStuff) processBlockCertificate(cert *Certificate) {
	if c.Process.ID() != c.Process.Proposer(cert.Epoch+1) {
		panic(fmt.Errorf("Non proposer %v received block certificate in epoch %v\n", c.Process.ID(), cert.Epoch))
	}
	c.lockedCertificate = cert
	c.broadcastProposal()
}

func (c *HotStuff) broadcastProposal() {
	var block *Block
	value := c.Process.GetValue()
	if value == nil {
		fmt.Printf("Generator returned nil in epoch\n")
		return
	}
	if c.lockedCertificate == nil {
		block = NewBlock(value, nil)
	} else {
		block = NewBlock(value, c.lockedCertificate.block)
	}
	proposal := &Message{
		Type:        PROPOSE,
		Epoch:       block.Height,
		Block:       block,
		Certificate: c.lockedCertificate,
		Sender:      c.Process.ID(),
		SenderFwd:   c.Process.ID(),
	}
	c.Process.Broadcast(proposal)
}

func (c *HotStuff) sendVote(voteType int16, blockID BlockID, epoch int64) {
	vote := &Message{
		Type:    voteType,
		Epoch:   epoch,
		BlockID: blockID,
		Sender:  c.Process.ID(),
	}
	//fmt.Printf("Process %v (%v) epoch %v vote for value %v\n", c.Process.ID(), c.Process.ID()%5, c.Epoch, vote.BlockID[0:4])
	// send vote to the proposer of the following epoch
	c.Process.Send(vote, c.Process.Proposer(epoch+1))
}

// ProcessMessage processes a consensus timeout.
//
// Contract: timeout belongs to this instance (height) of consensus.
func (c *HotStuff) ProcessTimeout(timeout *Timeout) {
}
