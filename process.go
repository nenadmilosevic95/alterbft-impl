package tendermint

import (
	//	"fmt"

	"time"

	"dslab.inf.usi.ch/tendermint/consensus"
	"dslab.inf.usi.ch/tendermint/net"
)

/*
type DeltaStat struct {
	n         int
	epoch     int64
	proposals int
	startTime time.Time
	durations []int64
}

func NewDeltaStat(n int, epoch int) *DeltaStat {
	return &DeltaStat{
		n:         n,
		epoch:     int64(epoch),
		durations: make([]int64, n),
	}
}

func (ds *DeltaStat) Start() {
	ds.startTime = time.Now()
}

func (ds *DeltaStat) Update(sender int) {
	ds.proposals++
	updateTime := time.Now()
	ds.durations[sender] = updateTime.Sub(ds.startTime).Milliseconds()
}

func (ds *DeltaStat) Print() {
	fmt.Printf("DeltaStat(%v): %v ms\n", ds.epoch, ds.duration)
}
*/

// Process supports the execution of consensus.
type Process struct {
	id  int
	num int

	config    *Config
	transport net.Transport
	proxy     net.Proxy

	verifier      *Verifier
	timeoutTicker *consensus.TimeoutTicker
	blockchain    *consensus.Blockchain

	// Epoch window
	lastDecided int64
	lastEpoch   int64
	epochs      []consensus.Consensus

	// Parallel message signing and broadcast
	broadcastQueue chan *consensus.Message
	// Parallel message receiving and signature validation
	deliveryQueue chan *consensus.Message

	// Process statistics, periodically published to statsQueue
	stats       *Stats
	statsQueue  chan *Stats
	statsTicker <-chan time.Time

	// Sync delta statistics
	deltaStartTimes []time.Time
}

func NewProcess(id, numProcesses int, config *Config, transport net.Transport, proxy net.Proxy) *Process {
	if config == nil {
		config = DefaultConfig()
	}
	p := &Process{
		id:  id,
		num: numProcesses,

		config:    config,
		transport: transport,
		proxy:     proxy,

		timeoutTicker: consensus.NewTimeoutTicker(),

		stats:      NewStats(),
		statsQueue: make(chan *Stats, config.MessageQueuesSize),

		deltaStartTimes: make([]time.Time, config.MaxEpochToStart),
	}
	// Sign and broadcast messages in parallel
	if config.SignatureGenerationThreads > 0 {
		p.broadcastQueue = make(chan *consensus.Message, config.MessageQueuesSize)
	}
	p.deliveryQueue = make(chan *consensus.Message, config.MessageQueuesSize*numProcesses)
	// Receives and validates messages in parallel, adding them to the deliveryQueue
	p.verifier = NewVerifier(config.PublicKeys, transport.ReceiveQueue(), p.deliveryQueue)
	// Blockchain abstraction.
	p.blockchain = consensus.NewBlockchain(int(config.BlockchainSize))

	p.BootstrapEpochWindow()
	return p
}

// ID returns the process' unique identifier.
func (p *Process) ID() int {
	return p.id
}

// NumProcesses returns the total number of processes.
func (p *Process) NumProcesses() int {
	return p.num
}

// Broadcast a consensus message.
// Consensus messages are signed before being broadcast.
func (p *Process) Broadcast(message *consensus.Message) {
	if p.config.SignatureGenerationThreads > 0 {
		// Signature computed in parallel
		p.broadcastQueue <- message
	} else {
		p.signAndBroadcast(message)
	}
}

// Routine for signing and broadcasting messages.
func (p *Process) signAndBroadcastRoutine() {
	for {
		message := <-p.broadcastQueue
		p.signAndBroadcast(message)
	}
}

// Signs a consensus message with the process private key and broadcast it.
func (p *Process) signAndBroadcast(message *consensus.Message) {
	if p.config.PrivateKeys[message.Sender] != nil {
		message.Sign(p.config.PrivateKeys[message.Sender])
	}
	// Here we start meashuring delta
	if message.Type == consensus.PROPOSE {
		p.deltaStartTimes[message.Epoch] = time.Now()
	}
	//p.config.Log.Printf("Message broadcast: %v\n", message)
	p.transport.Broadcast(message.Marshall())
}

// Forward a consensus message.
// The message is already signed by its original sender.
func (p *Process) Forward(message *consensus.Message) {
	//p.config.Log.Printf("Message forwarded: %v\n", message)
	p.transport.Broadcast(message.Marshall())
}

// Send a consensus message.
// The message is already signed by its original sender.
func (p *Process) Send(message *consensus.Message, ids ...int) {
	//p.config.Log.Printf("Message forwarded: %v\n", message)
	if p.config.PrivateKeys[message.Sender] != nil {
		message.Sign(p.config.PrivateKeys[message.Sender])
	}
	p.transport.Send(message.Marshall(), ids...)
}

// Schedule a consensus timeout.
func (p *Process) Schedule(timeout *consensus.Timeout) {
	if !p.config.ScheduleTimeouts {
		return
	}
	select {
	case p.timeoutTicker.In <- timeout:
	default:
		p.config.Log.Println("failed to schedule timeout", timeout)
	}
}

// Decide in an epoch of consensus.
func (p *Process) Decide(epoch int64, block *consensus.Block) {
	if p.config.Model == "hot-stuff" {
		blocks := p.blockchain.Commit(block)
		for i := len(blocks) - 1; i >= 0; i-- {
			p.proxy.Deliver(epoch, blocks[i])
			if len(blocks[i].Value) > 0 {
				p.stats.InstanceDelivered(1)
			} else {
				p.stats.InstanceDelivered(0)
			}
		}
	} else {
		if p.FinishEpoch(epoch) {
			blocks := p.blockchain.Commit(block)
			for i := len(blocks) - 1; i >= 0; i-- {
				p.proxy.Deliver(epoch, blocks[i])
				if len(blocks[i].Value) > 0 {
					p.stats.InstanceDelivered(1)
				} else {
					p.stats.InstanceDelivered(0)
				}
				//p.config.Log.Printf("Block delivered in epoch %v\n", epoch)
			}
		}
	}
}

// Finish an epoch of consensus, this should start new epoch.
func (p *Process) Finish(epoch int64, lockedCertificate *consensus.Certificate, sentLockedCertificate bool) {
	if p.lastEpoch == epoch {
		p.StartNewEpoch(lockedCertificate, sentLockedCertificate)
	}
}

// Proposer returns the ID of the proposer of a height and round of consensus.
func (p *Process) Proposer(epoch int64) int {
	var proposerID = epoch % int64(p.num)
	return int(proposerID) // TODO: mind overflow
}

// AddBlock add a possible decision block to the blockchain.
func (p *Process) AddBlock(block *consensus.Block) bool {
	return p.blockchain.AddBlock(block)
}

// Extend check if bb extend b.
func (p *Process) ExtendValidChain(b *consensus.Block) bool {
	return p.blockchain.ExtendValidChain(b)
}

func (p *Process) IsEquivocatedBlock(block *consensus.Block) bool {
	//return p.blockchain.IsEquivocatedBlock(block)
	return false
}

// GetBlock returns a set of txs to propose as a new block.
func (p *Process) GetValue() []byte {
	return p.proxy.GetValue()
}

// TimeoutPropose returns the timeout duration within which proposer should propose.
func (p *Process) TimeoutPropose(epoch int64) time.Duration {
	return p.config.TimeoutSmallDelta + p.config.TimeoutBigDelta
}

// TimeoutEquivocation returns the timeout duration we need to detect equivocation.
func (p *Process) TimeoutEquivocation(epoch int64) time.Duration {
	return 2 * p.config.TimeoutSmallDelta
}

// TimeoutQuitEpoch returns the timeout duration of QuitEpochStep.
func (p *Process) TimeoutQuitEpoch(epoch int64) time.Duration {
	return 2 * p.config.TimeoutSmallDelta
}

// TimeoutEpochChange returns the timeout duration of timeout needed to learn the highest locked certificate.
func (p *Process) TimeoutEpochChange(epoch int64) time.Duration {
	return 2 * p.config.TimeoutSmallDelta
}

// StatsQueue returns the queue to which stats are periodically published.
func (p *Process) StatsQueue() chan *Stats {
	return p.statsQueue
}
