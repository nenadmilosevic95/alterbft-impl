package tendermint

import (
	"fmt"

	"dslab.inf.usi.ch/tendermint/consensus"
)

func (p *Process) BootstrapEpochWindow() {
	p.lastDecided = -1
	p.lastEpoch = -1
	p.epochs = make([]consensus.Consensus, p.config.MaxActiveEpochs)
}

// StartNewEpoch creates and starts a new epoch of consensus
func (p *Process) StartNewEpoch(validCertificate *consensus.Certificate, lockedCertificate *consensus.Certificate, oldCertificate *consensus.Certificate) {
	activeEpochs := p.lastEpoch - p.lastDecided
	if activeEpochs >= p.config.MaxActiveEpochs {
		fmt.Errorf("Epoch window is not big enough: %v > %v!", activeEpochs, p.config.MaxActiveEpochs)
		return
	}
	p.lastEpoch++
	if p.config.MaxEpochToStart > 0 && p.lastEpoch >= p.config.MaxEpochToStart {
		p.config.Log.Println("Not starting epoch", p.lastEpoch,
			"MaxEpochToStart:", p.config.MaxEpochToStart)
		return
	}
	/*
		// Here we init DeltaStat.
		if p.Proposer(p.lastEpoch) == p.ID() {
			p.deltaStat[p.lastEpoch] = NewDeltaStat(p.num, int(p.lastEpoch))
		}*/
	//p.config.Log.Printf("Epoch %v started with %v\n and %v.\n", p.lastEpoch, validCertificate, lockedCertificate)
	index := p.lastEpoch % p.config.MaxActiveEpochs
	if p.epochs[index] == nil || p.epochs[index].GetEpoch() != p.lastEpoch {
		p.epochs[index] = p.CreateNewEpoch(p.lastEpoch)
	}
	p.epochs[index].Start(validCertificate, lockedCertificate)
	p.stats.InstanceStarted()
}

func (p *Process) CreateNewEpoch(epoch int64) consensus.Consensus {
	switch p.config.Model {
	case "sync":
		return consensus.NewAlterBFT(epoch, p)
	case "delta":
		return consensus.NewDeltaProtocol(epoch, p)
	case "slow":
		if p.config.Byzantines[p.ID()] {
			//return consensus.NewByzantineSyncConsensus(epoch, p, p.config.Byzantines, p.config.ByzTime, p.config.ByzAttack)
			return consensus.NewAlterBFTSlowLeader()
		} else {
			return consensus.NewAlterBFT(epoch, p)
		}
	case "equiv":
		if p.config.Byzantines[p.ID()] {
			//return consensus.NewByzantineSyncConsensus(epoch, p, p.config.Byzantines, p.config.ByzTime, p.config.ByzAttack)
			return consensus.NewAlterBFTEquivLeader(epoch, p)
		} else {
			return consensus.NewAlterBFT(epoch, p)
		}
	}
	return nil
}

// FinishEpoch finishes epoch and stop all active epochs before this one.
func (p *Process) FinishEpoch(epoch int64) bool {
	if epoch > p.lastDecided {
		for i := epoch; i > p.lastDecided; i-- {
			index := i % p.config.MaxActiveEpochs
			p.epochs[index].Stop()
			//p.config.Log.Printf("Epoch %v finished.\n", epoch)
		}
		p.lastDecided = epoch
		return true
	}
	return false
}

func (p *Process) GetConsensusEpoch(epoch int64) consensus.Consensus {
	index := epoch % p.config.MaxActiveEpochs
	if p.epochs[index] != nil && p.epochs[index].GetEpoch() == epoch {
		return p.epochs[index]
	} else {
		//p.config.Log.Printf("Message received before epoch %v is started\n", epoch)
		p.epochs[index] = p.CreateNewEpoch(epoch)
	}
	return p.epochs[index]
}
