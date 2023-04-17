package net

import (
	"hash/fnv"
	"time"

	"dslab.inf.usi.ch/tendermint/consensus"
	"dslab.inf.usi.ch/tendermint/types"
)

// Decision is a decided value.
type Decision struct {
	Instance  uint64
	Value     []byte
	ValueID   uint64
	Timestamp time.Time
}

// ValueID computes an unique ID for a value.
func ValueID(value []byte) uint64 {
	h := fnv.New64a()
	h.Write(value)
	return h.Sum64()
}

type Proxy interface {
	// Deliver delivers a block committed by the consensus protocol.
	Deliver(epoch int64, block *consensus.Block)

	// GetValue returns a value to be proposed in the consensus protocol.
	GetValue() types.Value
}
