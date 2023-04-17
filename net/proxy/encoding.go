package proxy

import (
	"encoding/binary"

	"dslab.inf.usi.ch/tendermint/net"
)

var encoding = binary.LittleEndian

func DecodeDecision(message []byte) *net.Decision {
	return &net.Decision{
		Instance: encoding.Uint64(message[0:8]),
		ValueID:  encoding.Uint64(message[8:16]),
	}
}

func EncodeDecision(decision *net.Decision) []byte {
	message := make([]byte, 16)
	encoding.PutUint64(message[0:8], decision.Instance)
	encoding.PutUint64(message[8:16], decision.ValueID)
	return message
}
