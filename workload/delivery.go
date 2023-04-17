package workload

import (
	"time"

	"dslab.inf.usi.ch/tendermint/consensus"
	"dslab.inf.usi.ch/tendermint/types"
)

// Delivery records the delivery of a value.
type Delivery struct {
	// Consensus data.
	Epoch   int64
	Height  int64
	Value   types.Value
	BlockID consensus.BlockID

	// Performance data.
	Size int
	Time time.Time

	// Associated submission, if any.
	Submission *Submission
}

// NewDelivery creates a Delivery from a delivered block.
func NewDelivery(epoch int64, block *consensus.Block) *Delivery {
	return &Delivery{
		Epoch:   epoch,
		Height:  block.Height,
		Value:   block.Value,
		BlockID: block.BlockID(),

		Size: len(block.Value),
		Time: time.Now(),
	}
}

// Latency returns the latency of the delivered value, if there is an
// associated Submission.
func (d *Delivery) Latency() time.Duration {
	if d.Submission == nil {
		return time.Duration(0)
	}
	return d.Time.Sub(d.Submission.Time)
}
