package consensus

import (
	"testing"

	"dslab.inf.usi.ch/tendermint/crypto"
)

func getNVotes(e int64, block *Block, n int) []*Message {
	votes := make([]*Message, n)
	for i := 0; i < n; i++ {
		votes[i] = NewVoteMessage(e, block.BlockID(), block.Height, int16(i))
		votes[i].Sign(crypto.GeneratePrivateKey())
	}
	return votes
}

func getNSilences(e int64, n int) []*Message {
	silences := make([]*Message, n)
	for i := 0; i < n; i++ {
		silences[i] = NewSilenceMessage(e, int16(i))
	}
	return silences
}

func TestStart(t *testing.T) {
	e := int64(5)
	//coordinator
	p := NewTestProcess(0, 3)
	c := NewAlterBFT(e, p)
	c.Start(nil, nil)
	if len(p.state.sendQueue) != 1 ||
		p.state.sendQueue[0].Type != PROPOSE ||
		p.state.sendQueue[0].Block.Height != MIN_HEIGHT ||
		p.state.sendQueue[0].Block.PrevBlockID != nil ||
		len(p.state.timeoutQueue) != 0 {
		t.Error("Process coordinator doesn't work!")
	}
	//coordinator + validCertificate!=nil
	p = NewTestProcess(0, 3)
	c = NewAlterBFT(e, p)
	b0 := NewBlock(testRandValue(1024), nil)
	blockCert := NewBlockCertificate(e-1, b0.BlockID(), b0.Height)
	blockCert.block = b0
	c.Start(blockCert, blockCert)
	if len(p.state.sendQueue) != 1 ||
		p.state.sendQueue[0].Type != PROPOSE ||
		p.state.sendQueue[0].Block.Height != MIN_HEIGHT+1 ||
		!p.state.sendQueue[0].Block.PrevBlockID.Equal(b0.BlockID()) ||
		len(p.state.timeoutQueue) != 0 {
		t.Error("Process coordinator doesn't work!")
	}

	//non coordinator
	p = NewTestProcess(1, 3)
	c = NewAlterBFT(1, p)
	c.Start(nil, nil)
	if len(p.state.sendQueue) != 0 ||
		len(p.state.timeoutQueue) == 0 || p.state.timeoutQueue[0].Type != TimeoutPropose {
		t.Error("Process coordinator doesn't work!")
	}
}

func TestValidProposal(t *testing.T) {
	e := int64(10)
	p := NewTestProcess(1, 3)
	c := NewAlterBFT(e, p)
	c.Start(nil, nil)
	b0 := NewBlock(testRandValue(1024), nil)
	proposal := NewProposeMessage(e, b0, nil, 0)
	proposal.Marshall()
	c.processProposal(proposal)
	vv := NewVoteMessage(e, b0.BlockID(), b0.Height, 0)
	vv.Marshall()
	c.ProcessMessage(vv)
	if !c.Proposals.Has(proposal.Block.BlockID()) ||
		len(p.state.sendQueue) != 3 || p.state.sendQueue[0].Type != PROPOSE ||
		p.state.sendQueue[1].Type != VOTE || p.state.sendQueue[2].Type != VOTE {
		t.Errorf("Process didn't forward(proposal + vote) or vote for the valid proposal! %v \n", len(p.state.sendQueue))
	}
	err := assertMessageEquals(p.state.sendQueue[0], proposal)
	if err != nil {
		t.Error(err)
	}
	err = assertMessageEquals(p.state.sendQueue[1], vv)
	if err != nil {
		t.Error(err)
	}
	v := NewVoteMessage(e, b0.BlockID(), b0.Height, int16(c.Process.ID()))
	err = assertMessageEquals(p.state.sendQueue[2], v)
	if err != nil {
		t.Error(err)
	}

}

func TestValidProposalAfterQuorumOfVotes(t *testing.T) {
	//Good proposal missing for block certificate after quorum of votes
	//a) ready state
	e := int64(10)
	p := NewTestProcess(1, 3)
	c := NewAlterBFT(e, p)
	b0 := NewBlock(testRandValue(1024), nil)
	proposal := NewProposeMessage(e, b0, nil, 0)
	proposal.Marshall()
	c.Start(nil, nil)
	votes := getNVotes(e, proposal.Block, 2)
	for _, m := range votes {
		c.ProcessMessage(m)
	}
	if len(p.state.sendQueue) != 1 ||
		p.state.sendQueue[0].Type != QUIT_EPOCH ||
		c.epochPhase != Locked {
		t.Error("Process did not react on Ce(Bk)")
	}
	blockCert := NewBlockCertificate(e, proposal.Block.BlockID(), proposal.Block.Height)
	for _, v := range votes {
		blockCert.AddSignature(v.Signature, v.Sender)
	}
	err := assertCertificateEquals(c.validCertificate, blockCert)
	if err != nil {
		t.Error(err)
	}
	err = assertCertificateEquals(c.lockedCertificate, blockCert)
	if err != nil {
		t.Error(err)
	}
	quitEpoch := NewQuitEpochMessage(e, blockCert)
	quitEpoch.Sender = 1
	err = assertMessageEquals(p.state.sendQueue[0], quitEpoch)
	if err != nil {
		t.Error(err)
	}
	c.ProcessMessage(proposal)
	if !c.Proposals.Has(proposal.Block.BlockID()) ||
		len(p.state.sendQueue) != 1 ||
		c.epochPhase != Locked {
		t.Errorf("Process did not commit!")
	}
}

func TestInValidProposal(t *testing.T) {
	// 1) From process that is not coordinator
	e := int64(15)
	p := NewTestProcess(1, 3)
	c := NewAlterBFT(e, p)
	b0 := NewBlock(testRandValue(1024), nil)
	proposal := NewProposeMessage(e, b0, nil, 0)
	proposal.Marshall()
	c.Start(nil, nil)
	// Set sender to non coordinator process
	proposal.Sender = 2
	c.ProcessMessage(proposal)
	if c.Proposals.Has(proposal.Block.BlockID()) ||
		len(p.state.sendQueue) != 0 {
		t.Errorf("Process didn't process well an invalid proposal!\nConditions: %v,%v", c.Proposals.Has(proposal.Block.BlockID()), len(p.state.sendQueue) != 0)
	}
	// 2) With certificate lower than process' locked certificate
	// again set to "valid" value
	e = 16
	p = NewTestProcess(1, 3)
	c = NewAlterBFT(e, p)

	// Create a block cert
	b0Cert := NewBlockCertificate(e-1, b0.BlockID(), b0.Height)
	votes := getNVotes(e-1, b0, 2)
	for _, v := range votes {
		b0Cert.AddSignature(v.Signature, v.Sender)
	}
	c.Start(b0Cert, b0Cert)
	// create same cert but with smaller epoch
	b0CertLower := NewBlockCertificate(e-2, b0.BlockID(), b0.Height)
	votes = getNVotes(e-2, b0, 2)
	for _, v := range votes {
		b0CertLower.AddSignature(v.Signature, v.Sender)
	}
	b1 := NewBlock(testRandValue(1024), b0)
	proposal = NewProposeMessage(e, b1, b0CertLower, 0)
	proposal.Marshall()
	c.processProposal(proposal)
	if len(p.state.sendQueue) != 0 {
		t.Error("Process accepted invalid proposal!")
	}
}

func TestProcessProposal(t *testing.T) {
	TestValidProposal(t)
	TestInValidProposal(t)
	TestValidProposalAfterQuorumOfVotes(t)
}

func TestProcessVoteMessage(t *testing.T) {
	// Majority of votes with no proposal
	e := int64(16)
	p := NewTestProcess(1, 3)
	c := NewAlterBFT(e, p)
	c.Start(nil, nil)
	b0 := NewBlock(testRandValue(1024), nil)
	votes := getNVotes(e, b0, 2)
	for _, v := range votes {
		c.ProcessMessage(v)
	}
	if len(p.state.sendQueue) != 1 || p.state.sendQueue[0].Type != QUIT_EPOCH ||
		len(p.state.timeoutQueue) != 2 ||
		p.state.timeoutQueue[0].Type != TimeoutPropose ||
		p.state.timeoutQueue[1].Type != TimeoutEquivocation {
		t.Error("Process sent a message or scheduler a timeout it shouldn't!")
	}
	// Majority of votes with proposal
	p = NewTestProcess(1, 3)
	c = NewAlterBFT(e, p)
	c.Start(nil, nil)
	b0 = NewBlock(testRandValue(1024), nil)
	// Add manually proposal
	proposal := NewProposeMessage(e, b0, nil, 0)
	proposal.Marshall()
	c.ProcessMessage(proposal)
	votes = getNVotes(e, b0, 2)
	for _, v := range votes {
		c.ProcessMessage(v)
	}
	if len(p.state.sendQueue) != 4 ||
		p.state.sendQueue[0].Type != PROPOSE ||
		p.state.sendQueue[1].Type != VOTE ||
		p.state.sendQueue[2].Type != VOTE ||
		p.state.sendQueue[3].Type != QUIT_EPOCH ||
		len(p.state.timeoutQueue) != 2 ||
		p.state.timeoutQueue[0].Type != TimeoutPropose ||
		p.state.timeoutQueue[1].Type != TimeoutEquivocation {
		t.Errorf("Process didn't process well majority of votes with proposal! %v and %v", len(p.state.sendQueue), len(p.state.timeoutQueue))
	}
	// Majority + 1 vote
	votes = getNVotes(e, b0, 1)
	c.ProcessMessage(votes[0])
	// Nothing should change
	if len(p.state.sendQueue) != 4 ||
		len(p.state.timeoutQueue) != 2 {
		t.Errorf("Process didn't process well majority of votes with proposal! %v and %v", len(p.state.sendQueue), len(p.state.timeoutQueue))
	}
}

func TestProcessSilenceMessage(t *testing.T) {
	e := int64(17)
	p := NewTestProcess(1, 3)
	c := NewAlterBFT(e, p)
	c.Start(nil, nil)
	silences := getNSilences(e, 2)
	for _, m := range silences {
		c.ProcessMessage(m)
	}
	if len(p.state.sendQueue) != 1 || p.state.sendQueue[0].Type != QUIT_EPOCH ||
		len(p.state.timeoutQueue) != 2 ||
		p.state.timeoutQueue[0].Type != TimeoutPropose ||
		p.state.timeoutQueue[1].Type != TimeoutQuitEpoch {
		t.Error("Process sent a message or scheduler a timeout it shouldn't!")
	}
	cert := NewSilenceCertificate(e)
	for _, m := range silences {
		cert.AddSignature(m.Signature, m.Sender)
	}
	quitMsg := NewQuitEpochMessage(e, cert)
	quitMsg.Sender = 1
	err := assertMessageEquals(quitMsg, p.state.sendQueue[0])
	if err != nil {
		t.Error(err)
	}
}

func TestQuitEpochMessageWitBlockCertificate(t *testing.T) {
	// A: QuitEpoch with BlockCert
	// Test 1: c.epochPhase == Ready

	e := int64(17)
	b := NewBlock(testRandValue(1024), nil)
	b1 := NewBlock(testRandValue(1024), nil)
	cert := testBlockCertificate(e, b, 2)
	quitEpoch := NewQuitEpochMessage(e, cert)
	p := NewTestProcess(1, 3)
	c := NewAlterBFT(e, p)
	c.Start(nil, nil)
	c.ProcessMessage(quitEpoch)
	if c.epochPhase != Locked ||
		len(p.state.sendQueue) != 1 ||
		p.state.sendQueue[0].Type != QUIT_EPOCH ||
		len(p.state.timeoutQueue) != 2 ||
		p.state.timeoutQueue[0].Type != TimeoutPropose ||
		p.state.timeoutQueue[1].Type != TimeoutEquivocation {
		t.Errorf("Process send a message or invoke timeout it shouldn't\nConditions: %v,%v,%v", c.epochPhase, len(p.state.sendQueue), len(p.state.timeoutQueue))
	}
	// Test 2: c.epochPhase == Ready + proposal received
	p = NewTestProcess(1, 3)
	c = NewAlterBFT(e, p)
	c.Start(nil, nil)
	proposal := NewProposeMessage(e, b, nil, 0)
	proposal.Marshall()
	c.processProposal(proposal)
	c.ProcessMessage(quitEpoch)
	if c.epochPhase != Locked ||
		(len(p.state.sendQueue) != 1 && len(p.state.sendQueue) != 4) ||
		p.state.sendQueue[len(p.state.sendQueue)-1].Type != QUIT_EPOCH ||
		len(p.state.timeoutQueue) != 2 ||
		p.state.timeoutQueue[0].Type != TimeoutPropose ||
		p.state.timeoutQueue[1].Type != TimeoutEquivocation {
		t.Errorf("Process send a message or invoke timeout it shouldn't\nConditions: %v,%v,%v", c.epochPhase, len(p.state.sendQueue), len(p.state.timeoutQueue))
	}
	// Test 3: c.epochPhase == epochChange + silence cert
	p = NewTestProcess(1, 3)
	c = NewAlterBFT(e, p)
	c.Start(nil, nil)
	silences := getNSilences(e, 2)
	for _, m := range silences {
		c.ProcessMessage(m)
	}
	c.ProcessMessage(quitEpoch)
	if c.epochPhase != Finished ||
		len(p.state.sendQueue) != 2 ||
		p.state.sendQueue[0].Type != QUIT_EPOCH ||
		p.state.sendQueue[1].Type != QUIT_EPOCH ||
		len(p.state.timeoutQueue) != 2 ||
		p.state.timeoutQueue[0].Type != TimeoutPropose ||
		p.state.timeoutQueue[1].Type != TimeoutQuitEpoch {
		t.Errorf("Process send a message or invoke timeout it shouldn't %v, %v, %v", c.epochPhase, len(p.state.sendQueue), len(p.state.timeoutQueue))
	}
	// Test 4: c.epochPhase == epochChange + silence cert + proposal
	p = NewTestProcess(1, 3)
	c = NewAlterBFT(e, p)
	c.Start(nil, nil)
	silences = getNSilences(e, 2)
	for _, m := range silences {
		c.ProcessMessage(m)
	}
	proposal = NewProposeMessage(e, b, nil, 0)
	proposal.Marshall()
	c.processProposal(proposal)
	c.ProcessMessage(quitEpoch)
	if c.epochPhase != Finished ||
		len(p.state.sendQueue) != 2 ||
		p.state.sendQueue[0].Type != QUIT_EPOCH ||
		p.state.sendQueue[1].Type != QUIT_EPOCH ||
		len(p.state.timeoutQueue) != 2 ||
		p.state.timeoutQueue[0].Type != TimeoutPropose ||
		p.state.timeoutQueue[1].Type != TimeoutQuitEpoch {
		t.Errorf("Process send a message or invoke timeout it shouldn't %v, %v, %v", c.epochPhase, len(p.state.sendQueue), len(p.state.timeoutQueue))
	}
	// Test 5: c.epochPhase == epochChange + equivocation
	p = NewTestProcess(1, 3)
	c = NewAlterBFT(e, p)
	c.Start(nil, nil)
	proposal = NewProposeMessage(e, b, nil, 0)
	proposal1 := NewProposeMessage(e, b1, nil, 0)
	vote := NewVoteMessage(e, proposal.Block.BlockID(), proposal.Block.Height, 0)
	vote1 := NewVoteMessage(e, proposal1.Block.BlockID(), proposal1.Block.Height, 0)
	c.ProcessMessage(vote)
	c.ProcessMessage(vote1)
	c.ProcessMessage(quitEpoch)
	if c.epochPhase != Finished ||
		len(p.state.sendQueue) != 1 ||
		p.state.sendQueue[0].Type != QUIT_EPOCH ||
		len(p.state.timeoutQueue) != 2 ||
		p.state.timeoutQueue[0].Type != TimeoutPropose ||
		p.state.timeoutQueue[1].Type != TimeoutQuitEpoch {
		t.Error("Process send a message or invoke timeout it shouldn't")
	}
	// Test 6: c.epochPhase == locked
	p = NewTestProcess(1, 3)
	c = NewAlterBFT(e, p)
	c.Start(nil, nil)
	proposal = NewProposeMessage(e, b, nil, 0)
	proposal.Marshall()
	c.processProposal(proposal)
	votes := getNVotes(e, proposal.Block, 2)
	for _, m := range votes {
		c.ProcessMessage(m)
	}
	c.ProcessMessage(quitEpoch)
	if c.epochPhase != Locked ||
		len(p.state.sendQueue) != 4 ||
		p.state.sendQueue[0].Type != PROPOSE ||
		p.state.sendQueue[1].Type != VOTE ||
		p.state.sendQueue[2].Type != VOTE ||
		p.state.sendQueue[3].Type != QUIT_EPOCH ||
		len(p.state.timeoutQueue) != 2 ||
		p.state.timeoutQueue[0].Type != TimeoutPropose ||
		p.state.timeoutQueue[1].Type != TimeoutEquivocation {
		t.Error("Process send a message or invoke timeout it shouldn't")
	}
}

func TestQuitEpochMessageWitSilenceCertificate(t *testing.T) {
	e := int64(10)
	b := NewBlock(testRandValue(1024), nil)
	proposal := NewProposeMessage(e, b, nil, 0)
	// B: QuitEpoch with SilenceCert
	// Test 1: c.epochPhase == Ready
	cert := NewSilenceCertificate(e)
	silences := getNSilences(e, 2)
	for _, s := range silences {
		cert.AddSignature(s.Signature, s.Sender)
	}
	quitEpoch := NewQuitEpochMessage(e, cert)
	p := NewTestProcess(1, 3)
	c := NewAlterBFT(e, p)
	c.Start(nil, nil)
	c.ProcessMessage(quitEpoch)
	if c.epochPhase != EpochChange ||
		len(p.state.sendQueue) != 1 ||
		p.state.sendQueue[0].Type != QUIT_EPOCH ||
		len(p.state.timeoutQueue) != 2 ||
		p.state.timeoutQueue[0].Type != TimeoutPropose ||
		p.state.timeoutQueue[1].Type != TimeoutQuitEpoch {
		t.Errorf("Process send a message or invoke timeout it shouldn't: %v %v %v", c.epochPhase, len(p.state.sendQueue), len(p.state.timeoutQueue))
	}
	// Test 2: c.epochPhase == EpochChange
	p = NewTestProcess(1, 3)
	c = NewAlterBFT(e, p)
	c.Start(nil, nil)
	// We can set here manually because silence certificate don't add any logic
	c.epochPhase = EpochChange
	c.ProcessMessage(quitEpoch)
	if c.epochPhase != EpochChange ||
		len(p.state.sendQueue) != 0 ||
		len(p.state.timeoutQueue) != 1 ||
		p.state.timeoutQueue[0].Type != TimeoutPropose {
		t.Error("Process send a message or invoke timeout it shouldn't")
	}
	// Test 3: c.epochPhase == Commit
	p = NewTestProcess(1, 3)
	c = NewAlterBFT(e, p)
	c.Start(nil, nil)
	proposal = NewProposeMessage(e, b, nil, 0)
	proposal.Marshall()
	c.processProposal(proposal)
	votes := getNVotes(e, proposal.Block, 2)
	for _, m := range votes {
		c.ProcessMessage(m)
	}
	c.ProcessMessage(quitEpoch)
	if c.epochPhase != Finished ||
		len(p.state.sendQueue) != 4 ||
		len(p.state.timeoutQueue) != 2 ||
		p.state.timeoutQueue[0].Type != TimeoutPropose ||
		p.state.timeoutQueue[1].Type != TimeoutEquivocation {
		t.Error("Process send a message or invoke timeout it shouldn't")
	}

}

func TestProcessPrecommitMessage(t *testing.T) {
	// TO DO
}

func TestProcessCommitMessage(t *testing.T) {
	// TO DO
}

// DOVDE sam stigao
func TestDecisionBlock(t *testing.T) {
	// Test 1 - proposal received before timeoutEquivocation expired
	//Create a process that is not coordinator so with id!=0
	e := int64(7)
	p := NewTestProcess(2, 3)
	c := NewAlterBFT(e, p)
	c.Start(nil, nil)
	if len(p.state.timeoutQueue) == 0 || p.state.timeoutQueue[0].Type != TimeoutPropose {
		t.Error("Process didn't schedule timeoutPropose!")
	}
	// Get Proposal
	b0 := NewBlock(testRandValue(1024), nil)
	proposal := NewProposeMessage(e, b0, nil, 0)
	// Get n votes
	votes := getNVotes(e, proposal.Block, 2)
	// Process proposal and n votes for it
	proposal.Marshall()
	c.ProcessMessage(proposal)
	if len(p.state.sendQueue) != 3 {
		t.Error("Process did not vote!")
	}
	for _, m := range votes {
		c.ProcessMessage(m)
	}
	if len(p.state.sendQueue) != 4 || c.epochPhase != Locked ||
		len(p.state.timeoutQueue) != 2 || p.state.timeoutQueue[1].Type != TimeoutEquivocation {
		t.Error("Process didn't schedule timeoutEquivocation!")
	}

	// This scenario should trigger timeoutEquivocation and when it expires a process should commit
	tEq := p.state.timeoutQueue[1]
	// Simulate timeoutEquivocation expiring
	c.ProcessTimeout(tEq)

	if c.epochPhase != Finished ||
		p.state.decision == nil ||
		p.state.validCertificate == nil ||
		p.state.lockedCertificate == nil {
		t.Errorf("Process didn't decide: %v %v \n", c.epochPhase, p.state.decision)
	}

	// Test 2 - proposal received after timeoutEquivocation expired
	//Create a process that is not coordinator so with id!=0
	e = int64(7)
	p = NewTestProcess(2, 3)
	c = NewAlterBFT(e, p)
	c.Start(nil, nil)
	if len(p.state.timeoutQueue) == 0 || p.state.timeoutQueue[0].Type != TimeoutPropose {
		t.Error("Process didn't schedule timeoutPropose!")
	}
	// Get Proposal
	b0 = NewBlock(testRandValue(1024), nil)
	proposal = NewProposeMessage(e, b0, nil, 0)
	// Get n votes
	votes = getNVotes(e, proposal.Block, 2)
	// Process proposal and n votes for it
	proposal.Marshall()
	if len(p.state.sendQueue) != 0 {
		t.Error("Process voted before proposer's vote!")
	}
	for _, m := range votes {
		c.ProcessMessage(m)
	}
	if len(p.state.sendQueue) != 1 || c.epochPhase != Locked ||
		len(p.state.timeoutQueue) != 2 || p.state.timeoutQueue[1].Type != TimeoutEquivocation {
		t.Error("Process didn't schedule timeoutEquivocation!")
	}

	// This scenario should trigger timeoutEquivocation and when it expires a process should commit
	tEq = p.state.timeoutQueue[1]
	// Simulate timeoutEquivocation expiring
	c.ProcessTimeout(tEq)

	if c.epochPhase != Commit ||
		c.decision == nil ||
		p.state.validCertificate == nil ||
		p.state.lockedCertificate == nil {
		t.Errorf("Process didn't decide: %v %v \n", c.epochPhase, c.decision)
	}

	c.ProcessMessage(proposal)
	if c.epochPhase != Finished ||
		len(p.state.sendQueue) != 1 {
		t.Errorf("Process didn't decide: %v %v \n", c.epochPhase, p.state.decision)
	}
}

func TestNoDecisionSilence(t *testing.T) {
	// Create a process that is not coordinator so with id!=0
	e := int64(3)
	p := NewTestProcess(2, 3)
	c := NewAlterBFT(e, p)
	c.Start(nil, nil)
	if len(p.state.timeoutQueue) == 0 || p.state.timeoutQueue[0].Type != TimeoutPropose {
		t.Error("Process didn't schedule timeoutPropose!")
	}
	// Simulate timeoutPropose expiring
	tPropose := p.state.timeoutQueue[0]
	c.ProcessTimeout(tPropose)
	if len(p.state.sendQueue) != 1 || p.state.sendQueue[0].Type != SILENCE {
		t.Error("Process didn't blame the leader when timeoutPropose expired!")
	}
	// Get blames
	blames := getNSilences(e, 2)
	for _, b := range blames {
		c.ProcessMessage(b)
	}
	if len(p.state.timeoutQueue) != 2 || p.state.timeoutQueue[1].Type != TimeoutQuitEpoch {
		t.Error("Process didn't schedule timeoutQuitEpoch!")
	}
	if len(p.state.sendQueue) != 2 ||
		p.state.sendQueue[1].Type != QUIT_EPOCH ||
		p.state.sendQueue[1].Certificate.Type != SILENCE_CERT {
		t.Error("Process didn't send a quit epoch message with silence certificate!")
	}
	// This scenario should trigger timeoutEpochChange and when it expires a process should finish epoch
	tEpCh := p.state.timeoutQueue[1]
	// Simulate timeoutEquivocation expiring
	c.ProcessTimeout(tEpCh)

	if c.epochPhase != Finished ||
		p.state.decision != nil ||
		p.state.validCertificate != nil ||
		p.state.lockedCertificate != nil {
		t.Error("Process didn't finish epoch and decide nil!")
	}
}

func TestNoDecisionEquivocation(t *testing.T) {
	// Create a process that is not coordinator so with id!=0
	e := int64(4)
	p := NewTestProcess(2, 3)
	c := NewAlterBFT(e, p)
	c.Start(nil, nil)
	if len(p.state.timeoutQueue) == 0 || p.state.timeoutQueue[0].Type != TimeoutPropose {
		t.Error("Process didn't schedule timeoutPropose!")
	}

	// Get proposals
	b0 := NewBlock(testRandValue(1024), nil)
	b1 := NewBlock(testRandValue(1024), nil)
	proposals := []*Message{NewProposeMessage(e, b0, nil, 0), NewProposeMessage(e, b1, nil, 0)}
	votes := []*Message{NewVoteMessage(e, b0.BlockID(), b0.Height, int16(0)), NewVoteMessage(e, b1.BlockID(), b1.Height, int16(0))}
	for _, m := range proposals {
		m.Marshall()
		c.ProcessMessage(m)
	}
	if c.epochPhase != Ready ||
		len(p.state.sendQueue) != 3 {
		t.Error("Process detected equivocationion without votes!")
	}
	if len(p.state.timeoutQueue) != 1 || p.state.timeoutQueue[0].Type != TimeoutPropose {
		t.Error("Process schedule timeoutEquivocation!")
	}

	for _, m := range votes {
		c.ProcessMessage(m)
	}

	if c.epochPhase != EpochChange ||
		len(p.state.timeoutQueue) != 2 || p.state.timeoutQueue[1].Type != TimeoutQuitEpoch {
		t.Error("Process did not detect equivocation!")
	}

	// This scenario should trigger timeoutEpochChange and when it expires a process should finish epoch
	tEpCh := p.state.timeoutQueue[1]
	// Simulate timeoutEquivocation expiring
	c.ProcessTimeout(tEpCh)

	if c.epochPhase != Finished ||
		p.state.decision != nil ||
		(p.state.validCertificate != nil && !p.state.validCertificate.Equal(proposals[0].Certificate)) ||
		p.state.lockedCertificate != nil {
		t.Error("Process didn't finish epoch and decide nil!")
	}
}

func TestNoDecisionBlockPlusSilence(t *testing.T) {
	//Create a process that is not coordinator so with id!=0
	e := int64(5)
	p := NewTestProcess(2, 3)
	c := NewAlterBFT(e, p)
	c.Start(nil, nil)
	if len(p.state.timeoutQueue) == 0 || p.state.timeoutQueue[0].Type != TimeoutPropose {
		t.Error("Process didn't schedule timeoutPropose!")
	}
	// Get Proposal
	b0 := NewBlock(testRandValue(1024), nil)
	proposal := NewProposeMessage(e, b0, nil, 0)
	// Get n votes
	votes := getNVotes(e, proposal.Block, 2)
	// Process proposal and n votes for it
	proposal.Marshall()
	c.ProcessMessage(proposal)
	if len(p.state.sendQueue) != 3 {
		t.Error("Process did not vote without vote!")
	}
	for _, m := range votes {
		c.ProcessMessage(m)
	}
	if len(p.state.timeoutQueue) != 2 || p.state.timeoutQueue[1].Type != TimeoutEquivocation {
		t.Error("Process didn't schedule timeoutEquivocation!")
	}
	if len(p.state.sendQueue) != 4 ||
		p.state.sendQueue[0].Type != PROPOSE ||
		p.state.sendQueue[1].Type != VOTE ||
		p.state.sendQueue[2].Type != VOTE ||
		p.state.sendQueue[3].Type != QUIT_EPOCH ||
		p.state.sendQueue[3].Certificate.Type != BLOCK_CERT {
		t.Error("Process didn't send a quit epoch message with block certificate!")
	}
	//Before timeoutEquivocation expires it receives Silence certificate
	blames := getNSilences(e, 2)
	for _, b := range blames {
		c.ProcessMessage(b)
	}
	//This should not trigger any timeout nor produce any more messages
	if len(p.state.sendQueue) != 4 || len(p.state.timeoutQueue) != 2 {
		t.Error("Wrong message sent or timeout scheduled!")
	}

	// This scenario should trigger timeoutEquivocation and when it expires a process should not commit
	tEq := p.state.timeoutQueue[1]
	// Simulate timeoutEquivocation expiring
	c.ProcessTimeout(tEq)

	if c.epochPhase != Finished ||
		p.state.decision != nil ||
		p.state.validCertificate == nil ||
		p.state.lockedCertificate == nil {
		t.Error("Process decide and it should not!")
	}
}

func TestNoDecisionBlockPlusEquivocation(t *testing.T) {
	//Create a process that is not coordinator so with id!=0
	e := int64(7)
	p := NewTestProcess(2, 3)
	c := NewAlterBFT(e, p)
	c.Start(nil, nil)
	if len(p.state.timeoutQueue) == 0 || p.state.timeoutQueue[0].Type != TimeoutPropose {
		t.Error("Process didn't schedule timeoutPropose!")
	}
	// Get Proposal
	b0 := NewBlock(testRandValue(1024), nil)
	proposal := NewProposeMessage(e, b0, nil, 0)
	// Get n votes
	votes := getNVotes(e, proposal.Block, 2)
	// Process proposal and n votes for it
	proposal.Marshall()
	c.ProcessMessage(proposal)

	for _, m := range votes {
		c.ProcessMessage(m)
	}
	if len(p.state.sendQueue) != 4 ||
		p.state.sendQueue[0].Type != PROPOSE ||
		p.state.sendQueue[1].Type != VOTE ||
		p.state.sendQueue[2].Type != VOTE ||
		p.state.sendQueue[3].Type != QUIT_EPOCH ||
		p.state.sendQueue[3].Certificate.Type != BLOCK_CERT {
		t.Error("Process didn't forward or vote for a proposal!")
	}
	if len(p.state.timeoutQueue) != 2 || p.state.timeoutQueue[1].Type != TimeoutEquivocation {
		t.Error("Process didn't schedule timeoutEquivocation!")
	}

	//Before timeoutEquivocation expires it receives another proposal
	b1 := NewBlock(testRandValue(1024), nil)
	secondVote := NewVoteMessage(e, b1.BlockID(), b1.Height, 0)
	c.ProcessMessage(secondVote)
	//This should not trigger any timeout nor produce any more messages
	if c.epochPhase != Finished || len(p.state.sendQueue) != 4 || len(p.state.timeoutQueue) != 2 {
		t.Error("Wrong message sent or timeout scheduled!")
	}
	// This scenario should trigger timeoutEquivocation and when it expires a process should not commit
	tEq := p.state.timeoutQueue[1]
	// Simulate timeoutEquivocation expiring
	c.ProcessTimeout(tEq)
	if c.epochPhase != Finished ||
		p.state.decision != nil ||
		p.state.validCertificate == nil ||
		p.state.lockedCertificate == nil {
		t.Errorf("Process decide while it should not! %v %v\n", c.epochPhase, p.state.decision)
	}
}

func TestNoDecisionSilencePlusBlock(t *testing.T) {
	// Create a process that is not coordinator so with id!=0
	e := int64(4)
	p := NewTestProcess(2, 3)
	c := NewAlterBFT(e, p)
	c.Start(nil, nil)
	if len(p.state.timeoutQueue) == 0 || p.state.timeoutQueue[0].Type != TimeoutPropose {
		t.Error("Process didn't schedule timeoutPropose!")
	}
	// Simulate timeoutPropose expiring
	tPropose := p.state.timeoutQueue[0]
	c.ProcessTimeout(tPropose)
	if len(p.state.sendQueue) != 1 || p.state.sendQueue[0].Type != SILENCE {
		t.Error("Process didn't blame the leader when timeoutPropose expired!")
	}
	// Get blames
	silences := getNSilences(e, 2)
	for _, b := range silences {
		c.ProcessMessage(b)
	}
	if len(p.state.timeoutQueue) != 2 || p.state.timeoutQueue[1].Type != TimeoutQuitEpoch {
		t.Error("Process didn't schedule timeoutQuitEpoch!")
	}
	if len(p.state.sendQueue) != 2 ||
		p.state.sendQueue[1].Type != QUIT_EPOCH ||
		p.state.sendQueue[1].Certificate.Type != SILENCE {
		t.Error("Process didn't send a quit epoch message with silence certificate!")
	}
	// Before timeoutEpoch change expires process receives Ce(Bk)
	b0 := NewBlock(testRandValue(1024), nil)
	proposal := NewProposeMessage(e, b0, nil, 0)
	c.ProcessMessage(proposal)
	if len(p.state.sendQueue) != 2 {
		t.Error("Process forward a proposal and maybe it sent a vote!")
	}
	votes := getNVotes(e, proposal.Block, 2)
	for _, m := range votes {
		c.ProcessMessage(m)
	}
	if len(p.state.sendQueue) != 3 ||
		p.state.sendQueue[2].Type != QUIT_EPOCH ||
		p.state.sendQueue[2].Certificate.Type != BLOCK_CERT {
		t.Error("Process didn't send a quit epoch message with block certificate!")
	}
	if len(p.state.timeoutQueue) != 2 {
		t.Error("Prcesses started timeoutEquivocation even though it shouldn't!")
	}
	if c.epochPhase != Finished {
		t.Error("Process didn't finish epoch!")
	}
	// This scenario should trigger timeoutEpochChange and when it expires a process will already be in inactive step
	tEpCh := p.state.timeoutQueue[1]
	// Simulate timeoutEquivocation expiring
	c.ProcessTimeout(tEpCh)

	if c.epochPhase != Finished ||
		p.state.decision != nil ||
		p.state.validCertificate == nil ||
		p.state.lockedCertificate != nil {
		t.Error("Process didn't finish epoch and decide nil!")
	}
}

func TestNoDecisionSilencePlusEquivocation(t *testing.T) {
	// Create a process that is not coordinator so with id!=0
	e := int64(3)
	p := NewTestProcess(2, 3)
	c := NewAlterBFT(e, p)
	c.Start(nil, nil)
	if len(p.state.timeoutQueue) == 0 || p.state.timeoutQueue[0].Type != TimeoutPropose {
		t.Error("Process didn't schedule timeoutPropose!")
	}
	// Simulate timeoutPropose expiring
	tPropose := p.state.timeoutQueue[0]
	c.ProcessTimeout(tPropose)
	if len(p.state.sendQueue) != 1 || p.state.sendQueue[0].Type != SILENCE {
		t.Error("Process didn't blame the leader when timeoutPropose expired!")
	}
	// Get blames
	silences := getNSilences(e, 2)
	for _, b := range silences {
		c.ProcessMessage(b)
	}
	if len(p.state.timeoutQueue) != 2 || p.state.timeoutQueue[1].Type != TimeoutQuitEpoch {
		t.Error("Process didn't schedule timeoutQuitEpoch!")
	}
	if len(p.state.sendQueue) != 2 ||
		p.state.sendQueue[1].Type != QUIT_EPOCH ||
		p.state.sendQueue[1].Certificate.Type != SILENCE {
		t.Error("Process didn't send a quit epoch message with silence certificate!")
	}
	// Before timeoutEpoch change expires process receives equivocation
	b0 := NewBlock(testRandValue(1024), nil)
	b1 := NewBlock(testRandValue(1024), nil)
	proposals := []*Message{NewProposeMessage(e, b0, nil, 0), NewProposeMessage(e, b1, nil, 0)}
	votes := []*Message{NewVoteMessage(e, b0.BlockID(), b0.Height, 0), NewVoteMessage(e, b1.BlockID(), b1.Height, 0)}
	for _, m := range votes {
		c.ProcessMessage(m)
	}
	if len(p.state.sendQueue) > 2 || len(p.state.timeoutQueue) > 2 {
		t.Errorf("Process reacted on equivocation that happened after seeing silence certificate!")
	}

	// This scenario should trigger timeoutEpochChange and when it expires a process should finish epoch
	tEpCh := p.state.timeoutQueue[1]
	// Simulate timeoutEquivocation expiring
	c.ProcessTimeout(tEpCh)

	if c.epochPhase != Finished ||
		p.state.decision != nil ||
		(p.state.validCertificate != nil && !p.state.validCertificate.Equal(proposals[0].Certificate)) ||
		p.state.lockedCertificate != nil {
		t.Error("Process didn't finish epoch and decide nil!")
	}
}

func TestNoDecisionEquivocationPlusSilence(t *testing.T) {
	// Create a process that is not coordinator so with id!=0
	e := int64(95)
	p := NewTestProcess(2, 3)
	c := NewAlterBFT(e, p)
	c.Start(nil, nil)
	if len(p.state.timeoutQueue) == 0 || p.state.timeoutQueue[0].Type != TimeoutPropose {
		t.Error("Process didn't schedule timeoutPropose!")
	}

	// Get proposals
	b0 := NewBlock(testRandValue(1024), nil)
	b1 := NewBlock(testRandValue(1024), nil)
	proposals := []*Message{NewProposeMessage(e, b0, nil, 0), NewProposeMessage(e, b1, nil, 0)}
	votes := []*Message{NewVoteMessage(e, b0.BlockID(), b0.Height, 0), NewVoteMessage(e, b1.BlockID(), b1.Height, 0)}
	for _, m := range votes {
		c.ProcessMessage(m)
	}
	if len(p.state.sendQueue) != 0 {
		t.Error("Process forward equivocated vites!")
	}
	if len(p.state.timeoutQueue) != 2 || p.state.timeoutQueue[1].Type != TimeoutQuitEpoch {
		t.Error("Process didn't schedule timeoutQuitEpoch!")
	}

	// Before timeoutEpochChange expires process receives silence certificate
	silences := getNSilences(e, 2)
	for _, b := range silences {
		c.ProcessMessage(b)
	}
	//This should not trigger any timeout nor produce any more messages
	if len(p.state.sendQueue) != 0 || len(p.state.timeoutQueue) != 2 {
		t.Error("Wrong message sent or timeout scheduled!")
	}
	// This scenario should trigger timeoutEpochChange and when it expires a process should finish epoch
	tEpCh := p.state.timeoutQueue[1]
	// Simulate timeoutEquivocation expiring
	c.ProcessTimeout(tEpCh)

	if c.epochPhase != Finished ||
		p.state.decision != nil ||
		(p.state.validCertificate != nil && !p.state.validCertificate.Equal(proposals[0].Certificate)) ||
		p.state.lockedCertificate != nil {
		t.Error("Process didn't finish epoch and decide nil!")
	}
}

func TestNoDecisionEquvocationPlusBlock(t *testing.T) {
	// Create a process that is not coordinator so with id!=0
	e := int64(45)
	p := NewTestProcess(2, 3)
	c := NewAlterBFT(e, p)
	c.Start(nil, nil)
	if len(p.state.timeoutQueue) == 0 || p.state.timeoutQueue[0].Type != TimeoutPropose {
		t.Error("Process didn't schedule timeoutPropose!")
	}

	// Get proposals
	b0 := NewBlock(testRandValue(1024), nil)
	b1 := NewBlock(testRandValue(1024), nil)
	proposals := []*Message{NewProposeMessage(e, b0, nil, 0), NewProposeMessage(e, b1, nil, 0)}
	votes := []*Message{NewVoteMessage(e, b0.BlockID(), b0.Height, 0), NewVoteMessage(e, b1.BlockID(), b1.Height, 0)}
	for _, m := range votes {
		c.ProcessMessage(m)
	}

	if len(p.state.sendQueue) != 0 {
		t.Error("Process  forward equivocated proposals or  vote for one of them!")
	}
	if len(p.state.timeoutQueue) != 2 || p.state.timeoutQueue[1].Type != TimeoutQuitEpoch {
		t.Error("Process didn't schedule timeoutQuitEpoch!")
	}

	// Before timeoutEpochChange expires process receives block certificate
	votes = getNVotes(e, proposals[0].Block, 2)
	for _, m := range votes {
		c.ProcessMessage(m)
	}
	if len(p.state.sendQueue) != 1 ||
		p.state.sendQueue[0].Type != QUIT_EPOCH ||
		p.state.sendQueue[0].Certificate.Type != BLOCK_CERT {
		t.Error("Process didn't send a quit epoch message with block certificate!")
	}
	if len(p.state.timeoutQueue) != 2 {
		t.Error("Prcesses started timeoutEquivocation even though it shouldn't!")
	}
	if c.epochPhase != Finished {
		t.Error("Process didn't finish epoch!")
	}
	votes = getNVotes(e, proposals[1].Block, 2)
	for _, m := range votes {
		c.ProcessMessage(m)
	}
	// Second block certificate should not trigger any timeout nor produce any more messages
	if len(p.state.sendQueue) != 1 || len(p.state.timeoutQueue) != 2 {
		t.Error("Wrong message sent or timeout scheduled!")
	}
	// This scenario should trigger timeoutEpochChange and when it expires a process should finish epoch
	tEpCh := p.state.timeoutQueue[1]
	// Simulate timeoutEquivocation expiring
	c.ProcessTimeout(tEpCh)

	if c.epochPhase != Finished ||
		p.state.decision != nil ||
		p.state.validCertificate == nil ||
		p.state.lockedCertificate != nil {
		t.Error("Process didn't finish epoch and decide nil!")
	}
}
