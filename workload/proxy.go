package workload

import (
	"dslab.inf.usi.ch/tendermint/consensus"
	"dslab.inf.usi.ch/tendermint/net"
	"dslab.inf.usi.ch/tendermint/types"
)

// Generator implements net.Proxy
var _ net.Proxy = new(Generator)

// Deliver delivers a block committed by the consensus protocol.
//
// This method extracts the delivery data, which is added to a delivery queue.
func (g *Generator) Deliver(epoch int64, block *consensus.Block) {
	delivery := NewDelivery(epoch, block)
	select {
	case g.deliveryQueue <- delivery:
	default:
		// If the delivery queue is full, the delivery is dropped
		g.log.Println("Dropping delivery data of height", delivery.Height)
	}
}

// GetValue returns a value to be proposed in the consensus protocol.
//
// The produced value encodes submission data, including the ID of this
// generator and the time it was produced.
func (g *Generator) GetValue() types.Value {
	select {
	case value := <-g.values:
		NewSubmission(g.id, len(value)).Write(value)
		return value
	default:
		// If the values queue is empty, a nil value is returned
		return nil
	}
}
