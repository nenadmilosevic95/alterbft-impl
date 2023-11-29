package consensus

import (
	"time"
)

// Process supports the execution of consensus.
type Process interface {
	// ID returns the process id.
	ID() int

	// NumProcesses returns the total number of processes.
	NumProcesses() int

	// Broadcast a consensus message.
	Broadcast(message *Message)

	// Forward a consensus message.
	Forward(message *Message)

	// Send a consensus message to a set of peers.
	Send(message *Message, ids ...int)

	// Schedule a consensus timeout.
	Schedule(timeout *Timeout)

	// Proposer returns the ID of the proposer of a round of consensus.
	Proposer(epoch int64) int

	// GetBlock returns a set of txs to propose as a new block.
	GetValue() []byte

	// AddBlock try to add a block to the blockchain, returns success result.
	AddBlock(block *Block) bool

	// Extend check if bb extend b.
	ExtendValidChain(b *Block) bool

	// IsEquivocatedBlock check if block is equivocated block.
	IsEquivocatedBlock(block *Block) bool

	// Decide in an epoch of consensus.
	Decide(epoch int64, block *Block)

	// Finish an epoch of consensus.
	Finish(epoch int64, lockedCertificate *Certificate, sentLockedCertificate bool)

	// TimeoutPropose returns the timeout duration within which proposer should propose.
	TimeoutPropose(epoch int64) time.Duration

	// TimeoutEquivocation returns the timeout duration we need to detect equivocation.
	TimeoutEquivocation(epoch int64) time.Duration

	// TimeoutQuitEpoch returns the timeout duration of QuitEpochStep.
	TimeoutQuitEpoch(epoch int64) time.Duration

	// TimeoutEpochChange returns the timeout duration of timeout needed to learn the highest locked certificate.
	TimeoutEpochChange(epoch int64) time.Duration
}
