package gossip

import "dslab.inf.usi.ch/tendermint/net"

type Validator interface {
	//	Aggregate(messages []*Message) []*Message
	Validate(message net.Message) bool
}

type ValidatorBuilder interface {
	New(peerID int) Validator
}
