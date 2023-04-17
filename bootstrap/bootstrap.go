package bootstrap

// Bootstrap implements a simple network initialization protocol.
//
// Processes broadcast messages to announce themselves, and wait for receiving
// similar messages from a quorum of other processes in the same network.
//
// Upon receiving announces from a quorum of processes, this process becames
// active; this information is announced to all other processes.
//
// Upon receiving announces from a quorum of processes that report them as
// active, this process is done in the bootstrap protocol.
type Bootstrap struct {
	processID int
	quorum    int

	announceCounter int
	knownProcesses  map[int]bool
	activeProcesses map[int]bool
}

// NewBootstrap creates a new instance of the bootstrap protocol.
func NewBootstrap(processID, quorum int) *Bootstrap {
	return &Bootstrap{
		processID: processID,
		quorum:    quorum,

		knownProcesses:  make(map[int]bool),
		activeProcesses: make(map[int]bool),
	}
}

// Active returns whether the process is active in the protocol.
//
// The process becomes active after receiving messages from a quorum of
// processes.
func (b *Bootstrap) Active() bool {
	return len(b.knownProcesses) >= b.quorum
}

// Done returns whether the process is done in the protocol.
//
// The process becomes done in the protocol after receiving messages from a
// quorum of processes reporting that them are active in the protocol.
func (b *Bootstrap) Done() bool {
	return len(b.activeProcesses) >= b.quorum
}

// ProcessMessage processes a received message.
//
// The method may return a bootstrap message to announce that this process has
// changed its state in the bootstrap protocol.
func (b *Bootstrap) ProcessMessage(message *Message) *Message {
	alreadyActive := b.Active()
	sender := message.Sender()
	if b.knownProcesses[sender] == false {
		b.knownProcesses[sender] = true
	}
	if message.Active() && b.activeProcesses[sender] == false {
		b.activeProcesses[sender] = true
	}
	// Return an announce message when become active
	if !alreadyActive && b.Active() {
		b.announceCounter += 1
		return NewMessage(b.processID, b.announceCounter, true)
	}
	return nil
}

// ProcessTick processes a periodic clock tick event.
//
// The method returns a bootstrap message to announce this process.
func (b *Bootstrap) ProcessTick() *Message {
	b.announceCounter += 1
	return NewMessage(b.processID, b.announceCounter, b.Active())
}
