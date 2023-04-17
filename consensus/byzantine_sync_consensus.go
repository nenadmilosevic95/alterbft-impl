package consensus

/*
import (
	"fmt"
	"math/rand"
	"time"
)

// import "fmt"
const NumberOfZones = 5

const (
	Virginia  = iota // epoch hasn't started yet or has finished
	SaoPaolo         // phase before any of the certificates is observed
	Stockholm        // phase after Ce(Bk) certificate is observed
	Singapore        // phase after Ce(SILENCE) or Ce(EQUIV) is observed
	Syndey
)

// Consensus implement one epoch of consensus.
type ByzantineSyncConsensus struct {
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

	proposalAccepted bool

	// Ids of byzantine nodes
	byzantines map[int]bool
	byzTime    int

	// Most distant zones
	distantZones []int

	// Type attack
	attackType string
}

// NewConsensus creates a consensus instance for the provided epoch.
// NewConsensus creates a consensus instance for the provided epoch.
func NewByzantineSyncConsensus(epoch int64, process Process, byzantines map[int]bool, byzTime int, attackType string) *ByzantineSyncConsensus {
	c := &ByzantineSyncConsensus{
		Epoch:      epoch,
		Process:    process,
		byzantines: byzantines,
		byzTime:    byzTime,
		attackType: attackType,
	}
	c.Init()
	return c
}

// Init initializes consensus variables.
func (c *ByzantineSyncConsensus) Init() {
	c.epochPhase = Inactive
	c.Proposals = NewProposalSet()
	c.SilenceCertificate = NewSilenceCertificate(c.Epoch)
	c.Votes = NewCertificateSet()
	// TO DO: here we need to add also other sets for different models
	c.Precommits = nil
	c.Commits = nil
	c.scheduledTimeouts = make([]bool, TimeoutQuitEpoch+1)
	c.proposalAccepted = false
	// Set distant zones' numbers for each zone
	c.distantZones = make([]int, NumberOfZones)
	c.distantZones[Virginia] = Singapore
	c.distantZones[SaoPaolo] = Singapore
	c.distantZones[Stockholm] = Syndey
	c.distantZones[Singapore] = SaoPaolo
	c.distantZones[Syndey] = SaoPaolo

}

// Start this epoch of consensus
func (c *ByzantineSyncConsensus) Start(validCertificate *Certificate, lockedCertificate *Certificate) {
	c.validCertificate = validCertificate
	c.lockedCertificate = lockedCertificate
	c.initialLockedCertificate = lockedCertificate
	c.epochPhase = Ready
	// Here we launch the attack!
	c.launchAttack()
	for _, m := range c.messages {
		c.ProcessMessage(m)
	}
}

// Started informs whether this epoch has been started.
func (c *ByzantineSyncConsensus) Started() bool {
	return c.epochPhase > Inactive
}

// Stop this instance of consensus.
func (c *ByzantineSyncConsensus) Stop() {
	c.epochPhase = Inactive
}

func (c *ByzantineSyncConsensus) GetEpoch() int64 {
	return c.Epoch
}

func (c *ByzantineSyncConsensus) launchEquivocationAttack() {
	if !c.isByzantineProposer(c.Epoch) {
		return
	}
	p, q := c.getPandQ(c.Process.Proposer(c.Epoch), c.byzantines, c.Process.NumProcesses())
	closestByzantineProcessToQ := c.closestByzantineProcess(q, c.byzantines, c.Process.NumProcesses())
	if c.Process.ID() == closestByzantineProcessToQ {
		proposal := c.generateProposal(c.Process.Proposer(c.Epoch))
		if proposal == nil {
			fmt.Printf("Generator returned nil to byz in epoch %v\n", c.Epoch)
			return
		}
		msgs := c.getVoteMessages(c.byzantines, proposal.Block.BlockID())
		fmt.Printf("Byzantine %v sent prop+votes for first proposal to %v in epoch %v with %v leader %v\n", c.Process.ID(), q, c.Epoch, c.isByzantineProposer(c.Epoch), c.Process.Proposer(c.Epoch))
		c.Process.Send(proposal, q)
		for _, m := range msgs {
			c.Process.Send(m, q)
		}
	}
	closestByzantineProcessToP := c.closestByzantineProcess(p, c.byzantines, c.Process.NumProcesses())
	if c.Process.ID() == closestByzantineProcessToP {
		proposal := c.generateProposal(c.Process.Proposer(c.Epoch))
		if proposal == nil {
			fmt.Printf("Generator returned nil to byz in epoch %v\n", c.Epoch)
			return
		}
		msgs := c.getVoteMessages(c.byzantines, proposal.Block.BlockID())
		time.Sleep(time.Duration(c.byzTime) * time.Millisecond)
		fmt.Printf("Byzantine %v sent prop+votes for second proposal to %v in epoch %v with %v leader %v\n", c.Process.ID(), p, c.Epoch, c.isByzantineProposer(c.Epoch), c.Process.Proposer(c.Epoch))
		c.Process.Send(proposal, p)
		for _, m := range msgs {
			c.Process.Send(m, p)
		}
	}
}

func (c *ByzantineSyncConsensus) launchSilenceAttack() {
	p, q := c.getPandQ(c.Process.Proposer(c.Epoch), c.byzantines, c.Process.NumProcesses())
	closestByzantineProcessToQ := c.closestByzantineProcess(q, c.byzantines, c.Process.NumProcesses())
	if c.Process.ID() == closestByzantineProcessToQ {
		msgs := c.getSilenceMessages(c.byzantines)
		fmt.Printf("Byzantine %v sent silences in epoch %v with %v leader %v\n", c.Process.ID(), c.Epoch, c.isByzantineProposer(c.Epoch), c.Process.Proposer(c.Epoch))
		for _, m := range msgs {
			c.Process.Send(m, q)
		}
	}
	closestByzantineProcessToP := c.closestByzantineProcess(p, c.byzantines, c.Process.NumProcesses())
	if c.isByzantineProposer(c.Epoch) && c.Process.ID() == closestByzantineProcessToP {
		proposal := c.generateProposal(c.Process.Proposer(c.Epoch))
		if proposal == nil {
			fmt.Printf("Generator returned nil to byz in epoch %v\n", c.Epoch)
			return
		}
		msgs := c.getVoteMessages(c.byzantines, proposal.Block.BlockID())
		time.Sleep(time.Duration(c.byzTime) * time.Millisecond)
		fmt.Printf("Byzantine %v sent prop+votes in epoch %v with %v leader %v\n", c.Process.ID(), c.Epoch, c.isByzantineProposer(c.Epoch), c.Process.Proposer(c.Epoch))
		c.Process.Send(proposal, p)
		for _, m := range msgs {
			c.Process.Send(m, p)
		}
	}
}

func (c *ByzantineSyncConsensus) launchAttack() {
	switch c.attackType {
	case "equiv":
		c.launchEquivocationAttack()
	case "silence":
		c.launchSilenceAttack()
	default:
		return
	}
}

func (c *ByzantineSyncConsensus) getPandQ(leader int, byzantines map[int]bool, n int) (int, int) {
	var p, q int
	if byzantines[leader] == true {
		p, q = c.twoMostDistantProcesses(byzantines, n)
	} else {
		p = leader
		q = c.furthestProcess(p, byzantines, n)
	}
	return p, q
}

func (c *ByzantineSyncConsensus) twoMostDistantProcesses(byzantines map[int]bool, n int) (int, int) {
	p := c.getHonestProcessInZone(SaoPaolo, byzantines, n)
	q := c.furthestProcess(p, byzantines, n)
	return p, q
}

func (c *ByzantineSyncConsensus) getHonestProcessInZone(zone int, byzantines map[int]bool, n int) int {
	id := zone
	for byzantines[id] == true && id < n {
		id += NumberOfZones
	}
	return id
}

func (c *ByzantineSyncConsensus) getSilenceMessages(byzantines map[int]bool) []*Message {
	silences := make([]*Message, len(byzantines))
	i := 0
	for byzID, _ := range byzantines {
		silences[i] = NewSilenceMessage(c.Epoch, int16(byzID))
		i++
	}
	return silences
}

func (c *ByzantineSyncConsensus) closestByzantineProcess(q int, byzantines map[int]bool, n int) int {
	zone := q % NumberOfZones
	id := zone
	for byzantines[id] == false && id < n {
		id += NumberOfZones
	}
	return id
}

func GenerateByzantines(f, n int, seed int64) map[int]bool {
	byzantines := getRandomByzantines(f, n, seed)
	for checkByzantines(byzantines, n) == false {
		seed++
		byzantines = getRandomByzantines(f, n, seed)
	}
	return byzantines
}

func getRandomByzantines(f, n int, seed int64) map[int]bool {
	src := rand.NewSource(seed)
	randGen := rand.New(src)
	byzantines := make(map[int]bool, f)
	for len(byzantines) != f {
		id := randGen.Intn(n - 1)
		if id != 0 {
			byzantines[id] = true
		}
	}
	return byzantines
}

func checkByzantines(byzantines map[int]bool, n int) bool {
	return checkByzantineExistInZone(Virginia, byzantines, n) &&
		checkByzantineExistInZone(SaoPaolo, byzantines, n) &&
		checkByzantineExistInZone(Stockholm, byzantines, n) &&
		checkByzantineExistInZone(Singapore, byzantines, n) &&
		checkByzantineExistInZone(Syndey, byzantines, n)
}

func checkByzantineExistInZone(zone int, byzantines map[int]bool, n int) bool {
	for id := zone; id < n; id += NumberOfZones {
		if byzantines[id] == true {
			return true
		}
	}
	return false
}

func (c *ByzantineSyncConsensus) furthestProcess(p int, byzantines map[int]bool, n int) int {
	zone := p % NumberOfZones
	distantZone := c.distantZones[zone]
	id := distantZone
	for byzantines[id] == true && id < n {
		id += NumberOfZones
	}
	return id
}

func (c *ByzantineSyncConsensus) getVoteMessages(byzantines map[int]bool, blockID BlockID) []*Message {
	votes := make([]*Message, len(byzantines))
	i := 0
	for byzID, _ := range byzantines {
		votes[i] = NewVoteMessage(c.Epoch, blockID, int16(byzID))
		i++
	}
	return votes
}

func (c *ByzantineSyncConsensus) generateProposal(proposer int) *Message {
	var block *Block
	value := c.Process.GetValue()
	if value == nil {
		return nil
	}
	if c.validCertificate != nil {
		block = NewBlock(value, c.validCertificate.block)
	} else {
		block = NewBlock(value, nil)
	}
	proposal := &Message{
		Type:        PROPOSE,
		Epoch:       c.Epoch,
		Block:       block,
		Certificate: c.validCertificate,
		Sender:      proposer,
	}
	return proposal
}

func (c *ByzantineSyncConsensus) isByzantineProposer(epoch int64) bool {
	proposer := c.Process.Proposer(epoch)
	return c.byzantines[proposer]
}

// ProcessMessage processes a consensus message.
//
// Contract: message belongs to this epoch of consensus.
func (c *ByzantineSyncConsensus) ProcessMessage(message *Message) {
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

func (c *ByzantineSyncConsensus) processProposal(proposal *Message) {
	block := proposal.Block
	cert := proposal.Certificate
	// Check if proposal has already been processed!
	if c.Proposals.Has(block.BlockID()) {
		return
	}
	// Check if the new proposal is valid!
	if !block.PrevBlockID.Equal(cert.BlockID()) ||
		proposal.Sender != c.Process.Proposer(proposal.Epoch) {
		//fmt.Printf("Invalid proposal. Conditions: %v and %v\n", c.Process.Proposer(c.Epoch) != proposal.Sender, !block.PrevBlockID.Equal(cert.BlockID()))
		return
	}
	// We try to add block to the blockchain
	// if the predecessor is not present in the blockchain
	// this call will fail!
	ok := c.Process.AddBlock(block)
	if !ok {
		fmt.Printf("P%v: proposal could not be added to the blockchain in epoch %v\n", c.Process.ID(), c.Epoch)
		return
	}
	if proposal.Epoch != c.Epoch {
		return
	}
	// Save the proposal
	c.Proposals.Add(proposal)
	// Check if this is the first proposal
	if c.Proposals.Count() == 1 {
		proposer := c.Process.Proposer(proposal.Epoch)
		if c.attackType == "silence" && c.isByzantineProposer(c.Epoch) == false && c.closestByzantineProcess(proposer, c.byzantines, c.Process.NumProcesses()) == c.Process.ID() {
			msgs := c.getVoteMessages(c.byzantines, proposal.Block.BlockID())
			fmt.Printf("Byzantine %v sent votes in epoch %v with %v leader %v\n", c.Process.ID(), c.Epoch, c.isByzantineProposer(c.Epoch), c.Process.Proposer(c.Epoch))
			for _, m := range msgs {
				c.Process.Send(m, proposer)
			}
		}
		c.proposalAccepted = true
	} else {
		fmt.Printf("Byzantine process %v received two proposals in epoch %v\n", c.Process.ID(), c.Epoch)
	}
	// Get or create a block certificate
	blockCert := c.Votes.Get(c.Epoch, block.BlockID())
	// No vote for this proposal before
	if blockCert == nil {
		blockCert = NewBlockCertificate(c.Epoch, block.BlockID())
		blockCert.block = block
		c.Votes.Add(blockCert)
		// Some vote for the proposal is received before the proposal
	} else {
		blockCert.block = block
		// We need to check whether only proposal was missing
		if blockCert.SignatureCount() > c.Process.NumProcesses()/2 {
			c.processBlockCertificate(blockCert)
		}
	}
}

func (c *ByzantineSyncConsensus) processVote(vote *Message) {
	// we don't process votes in commit phase because we know that
	// process already received Ce(Bk) for some block Bk in epoch e
	if c.epochPhase == Ready || c.epochPhase == EpochChange {
		blockID := vote.BlockID
		blockCert := c.Votes.Get(c.Epoch, blockID)
		if blockCert == nil {
			blockCert = NewBlockCertificate(c.Epoch, blockID)
			c.Votes.Add(blockCert)
		}
		ok := blockCert.AddSignature(vote.Signature, vote.Sender)
		if !ok {
			return
		}
		if blockCert.block != nil && blockCert.SignatureCount() > c.Process.NumProcesses()/2 {
			c.processBlockCertificate(blockCert)
		}
	}
}

// processBlockCertificate is called when cert has quorum of signatures and proposal.
func (c *ByzantineSyncConsensus) processBlockCertificate(cert *Certificate) {
	if c.epochPhase == Commit {
		return
	}
	c.validCertificate = cert
	if c.epochPhase == Ready {
		c.epochPhase = Commit
		c.lockedCertificate = cert
		c.scheduleTimeout(TimeoutEquivocation)
	}
	if c.epochPhase == EpochChange {
		c.epochPhase = Finished
		fmt.Printf("Byzantine process %v epoch %v nolock+block value %v\n", c.Process.ID(), c.Epoch, c.validCertificate.BlockID()[0:4])
	}
	// Whenever we receive Ce(Bk) in epoch e we can finish epoch e and start epoch e+1
	//c.broadcastQuitEpoch(cert)
	c.Process.Finish(c.Epoch, c.validCertificate, c.lockedCertificate, nil)
}

func (c *ByzantineSyncConsensus) processSilence(silence *Message) {
	// we process Silence messages only in phases Ready and Commit,
	// we don't need to process Silence messages in EpochChange phase
	// because we know that we already received Ce(SILENCE) or Ce(EQUIV)
	if c.epochPhase == Ready || c.epochPhase == Commit {
		ok := c.SilenceCertificate.AddSignature(silence.Signature, silence.Sender)
		if !ok {
			return
		}
		if c.SilenceCertificate.SignatureCount() > c.Process.NumProcesses()/2 {
			c.processSilenceCertificate(c.SilenceCertificate)
		}
	}
}

func (c *ByzantineSyncConsensus) processSilenceCertificate(cert *Certificate) {
	if c.epochPhase == Ready { // this is the first certificate process has received
		c.epochPhase = EpochChange
		//fmt.Printf("Process %v in epoch %v didn't lock!\n", c.Process.ID(), c.Epoch)
		//c.broadcastQuitEpoch(cert)
		c.scheduleTimeout(TimeoutQuitEpoch)
		//c.Process.Decide(c.Epoch, nil)
		return
	}
	if c.epochPhase == Commit {
		c.epochPhase = Finished
		fmt.Printf("Byzantine process %v epoch %v lock+nodecision value %v\n", c.Process.ID(), c.Epoch, c.lockedCertificate.BlockID()[0:4])
		//c.Process.Decide(c.Epoch, nil)
		return
	}
}

func (c *ByzantineSyncConsensus) processQuitEpoch(quitEpoch *Message) {
	cert := quitEpoch.Certificate
	if cert.Type == BLOCK_CERT {
		blockCert := c.Votes.Get(c.Epoch, cert.BlockID())
		if blockCert == nil {
			blockCert = NewBlockCertificate(cert.Epoch, cert.BlockID())
			c.Votes.Add(blockCert)
		}
		for sender, signature := range cert.Signatures {
			ok := blockCert.AddSignature(signature, sender)
			if !ok {
				continue
			}
			if blockCert.SignatureCount() > c.Process.NumProcesses()/2 {
				if blockCert.block != nil {
					c.processBlockCertificate(blockCert)
				} else {
					break
				}
			}
		}
	}
	if cert.Type == SILENCE_CERT {
		c.processSilenceCertificate(cert)
	}
}

func (c *ByzantineSyncConsensus) processPrecommit(proposal *Message) {
	// TO DO: we don't need this for SyncTendermint
}

func (c *ByzantineSyncConsensus) processCommit(proposal *Message) {
	// TO DO: we don't need this for SyncTendermint
}

// ProcessMessage processes a consensus timeout.
//
// Contract: timeout belongs to this instance (height) of consensus.
func (c *ByzantineSyncConsensus) ProcessTimeout(timeout *Timeout) {
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

func (c *ByzantineSyncConsensus) processTimeoutPropose() {
	c.scheduledTimeouts[TimeoutPropose] = false
	if c.proposalAccepted == false {
		//c.broadcastSilence()
	}
}

func (c *ByzantineSyncConsensus) processTimeoutEquivocation() {

	c.scheduledTimeouts[TimeoutEquivocation] = false
	if c.epochPhase == Commit {
		// decision
		c.epochPhase = Finished
		decision := c.lockedCertificate.block
		fmt.Printf("Byzantine process %v epoch %v lock+decision value %v\n", c.Process.ID(), c.Epoch, decision.BlockID()[0:4])
		c.Process.Decide(c.Epoch, decision)
	}
}

func (c *ByzantineSyncConsensus) processTimeoutQuitEpoch() {
	c.scheduledTimeouts[TimeoutQuitEpoch] = false
	if c.epochPhase == EpochChange {
		c.epochPhase = Finished
		fmt.Printf("Byzantine process %v epoch %v nolock+noblock\n", c.Process.ID(), c.Epoch)
		c.Process.Finish(c.Epoch, c.validCertificate, c.lockedCertificate, nil)
	}
}

// Schedule a timeout for the epoch phase, if not already scheduled.
func (c *ByzantineSyncConsensus) scheduleTimeout(timeoutType int16) {
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
*/
