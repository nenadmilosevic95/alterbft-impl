package tendermint

import (
	"fmt"
	"log"
	"time"

	"dslab.inf.usi.ch/tendermint/bootstrap"
	"dslab.inf.usi.ch/tendermint/consensus"
)

// Bootstrap runs the bootstrap protocol to initialize the network.
//
// This methods returns when this process has been able to exchange messages
// with every process in the network, meaning that the network is connected.
func (p *Process) Bootstrap() {
	// Start the verifier to handle potentially early consensus messages,
	// which will then be buffered in the deliveryQueue.
	p.verifier.Start()
	ticker := time.Tick(p.config.BootstrapTickInterval)
	protocol := bootstrap.NewBootstrap(p.id, p.num)
	message := protocol.ProcessTick()
	for !protocol.Done() {
		// Broadcast produced broadcast messages, if any
		if message != nil {
			p.transport.Broadcast(message.Marshall())
		}

		select {
		case rawMessage := <-p.verifier.Skipped():
			if rawMessage.Code() == bootstrap.MessageCode {
				message = protocol.ProcessMessage(
					bootstrap.NewMessageFromBytes(rawMessage))
			} else {
				// Should not happen, as consensus messages are
				// not dropped by the verifier
				log.Println("Bootstrap dropping message", rawMessage)
				message = nil
			}
		case <-ticker:
			message = protocol.ProcessTick()
		}

	}
}

// MainLoop runs the main routine of a process.
func (p *Process) MainLoop() {
	// Start threads for signing and broadcasting messages
	for i := 0; i < p.config.SignatureGenerationThreads; i++ {
		go p.signAndBroadcastRoutine()
	}
	// Start the verifier routines of signature verification.
	p.verifier.Start()
	p.timeoutTicker.Start()
	// FIXME: handle the initialization of the first instance/epoch
	p.StartNewEpoch(nil, nil, nil)
	if p.config.StatsPublishingInterval > 0 {
		p.statsTicker = time.Tick(p.config.StatsPublishingInterval)
	}
	for {
		select {
		// p.deliveryQueue is fed by the Verifier
		case message := <-p.deliveryQueue:
			//p.config.Log.Printf("Message received: %v\n", message)
			p.stats.MessageReceived(message.Type)
			p.processConsensusMessage(message)

		case timeout := <-p.timeoutTicker.Out:
			//p.config.Log.Printf("Timeout received: %v\n", timeout)
			p.processConsensusTimeout(timeout)

		case <-p.statsTicker:
			p.publishAndResetStats()
		}

	}
}

// Deliver the message to the associated consensus instance, if present.
func (p *Process) processConsensusMessage(message *consensus.Message) {
	// Here we update the deltaStat
	//if p.Proposer(message.Epoch) == p.ID() && message.Type == consensus.PROPOSE {
	//startTime := p.deltaStartTimes[message.Epoch]
	//duration := time.Now().Sub(startTime).Milliseconds()
	//p.config.Log.Printf("DeltaStat: In epoch %v process %v (%v) received forwarded proposal from %v (%v) in %v ms\n", message.Epoch, p.ID(), p.ID()%5, message.SenderFwd, message.SenderFwd%5, duration)
	//}
	fmt.Printf("Message received %v %v %v\n", message.Type, message.Epoch, message.Sender)
	epoch := p.GetConsensusEpoch(message.Epoch)
	if epoch != nil {
		epoch.ProcessMessage(message)
	}
	// Special case send propose messages to the current epoch
	//if message.Type == consensus.PROPOSE && message.Epoch < p.lastEpoch &&
	//	p.epochs[p.lastEpoch] != nil {
	//	p.epochs[p.lastEpoch].ProcessMessage(message)
	//}
}

// Deliver the timeout to the associated consensus instance, if present.
func (p *Process) processConsensusTimeout(timeout *consensus.Timeout) {
	epoch := p.GetConsensusEpoch(timeout.Epoch)
	if epoch != nil {
		epoch.ProcessTimeout(timeout)
	}
}

// Publish stats to StatsQueue, discarding them if the queue is full.
func (p *Process) publishAndResetStats() {
	select {
	case p.statsQueue <- p.stats:
	default:
	}
	p.stats = NewStats()
}
