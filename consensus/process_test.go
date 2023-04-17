package consensus

import (
	"time"
)

type ProcessState struct {
	sendQueue         []*Message
	timeoutQueue      []*Timeout
	decision          *Block
	validCertificate  *Certificate
	lockedCertificate *Certificate
	blockchain        []*Block
}

func testProcessState(sendQueue []*Message, timeoutQueue []*Timeout, decision *Block, validCertificate *Certificate, lockedCertificate *Certificate, blockchain []*Block) *ProcessState {
	return &ProcessState{sendQueue, timeoutQueue, decision, validCertificate, lockedCertificate, blockchain}
}

func NewProcessState() *ProcessState {
	return &ProcessState{}
}

// TestProcess is a Test implementation of Process interface.
type TestProcess struct {
	id, num int
	state   *ProcessState
}

func NewTestProcess(id, num int) *TestProcess {
	return &TestProcess{
		id:    id,
		num:   num,
		state: NewProcessState(),
	}
}

func (p *TestProcess) ID() int {
	return p.id
}

func (p *TestProcess) NumProcesses() int {
	return p.num
}

func (p *TestProcess) Broadcast(message *Message) {
	p.state.sendQueue = append(p.state.sendQueue, message)
}

func (p *TestProcess) Forward(message *Message) {
	p.state.sendQueue = append(p.state.sendQueue, message)
}

func (p *TestProcess) Send(message *Message, ids ...int) {
	p.state.sendQueue = append(p.state.sendQueue, message)
}

func (p *TestProcess) Schedule(timeout *Timeout) {
	p.state.timeoutQueue = append(p.state.timeoutQueue, timeout)
}

func (p *TestProcess) Proposer(epoch int64) int {
	return 0
}

func (p *TestProcess) GetValue() []byte {
	return testRandValue(1024)
}

func (p *TestProcess) AddBlock(block *Block) bool {
	p.state.blockchain = append(p.state.blockchain, block)
	return true
}

func (p *TestProcess) ExtendValidChain(b *Block) bool {
	return true
}

func (p *TestProcess) IsEquivocatedBlock(block *Block) bool {
	return false
}

func (p *TestProcess) Decide(epoch int64, block *Block) {
	p.state.decision = block
}

// Finish an epoch of consensus.
func (p *TestProcess) Finish(epoch int64, validCertificate *Certificate, lockedCertificate *Certificate, oldCertificate *Certificate) {
	p.state.validCertificate = validCertificate
	p.state.lockedCertificate = lockedCertificate
}

func (p *TestProcess) TimeoutPropose(epoch int64) time.Duration {
	// TODO:
	return time.Second
}

func (p *TestProcess) TimeoutEquivocation(epoch int64) time.Duration {
	// TODO:
	return time.Second
}

func (p *TestProcess) TimeoutQuitEpoch(epoch int64) time.Duration {
	// TODO:
	return time.Second
}

// Test methods that use TestProcess
/*
func assertEqualProcessState(t *testing.T, p, expected *ProcessState) {
	if !expected.decision.Equal(p.decision) {
		t.Errorf("Expecting decision %v, got %v", expected.decision, p.decision)
	}
	if !p.validCertificate.Equal(p.validCertificate) {
		t.Errorf("Expecting valid certificate %v, got %v", expected.validCertificate, p.validCertificate)
	}
	if !p.lockedCertificate.Equal(p.lockedCertificate) {
		t.Errorf("Expecting locked certificate %v, got %v", expected.lockedCertificate, p.lockedCertificate)
	}
	assertSendQueues(t, p.sendQueue, expected.sendQueue)
	assertTimeoutQueues(t, p.timeoutQueue, expected.timeoutQueue)
}

func assertSendQueues(t *testing.T, q, expected []*Message) {
	n := len(q)
	nn := len(expected)
	if n != nn {
		t.Errorf("Expected message queue: %v, got %v", expected, q)
	}
	if n > 0 {
		for i := 0; i < n; i++ {
			assertMessageEquals(t, expected[i], q[i])
		}
	}

}

func assertTimeoutQueues(t *testing.T, q, expected []*Timeout) {
	n := len(q)
	nn := len(expected)
	if n != nn {
		t.Errorf("Expected timeout queue: %v, got %v", expected, q)
	}
	if n > 0 {
		for i := 0; i < n; i++ {
			assertTimeoutEquals(t, q[i], expected[i])
		}
	}
}*/
