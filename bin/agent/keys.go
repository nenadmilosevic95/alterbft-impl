package main

import (
	"encoding/binary"

	"dslab.inf.usi.ch/tendermint/crypto"
)

type KeySet struct {
	PrivateKeys []crypto.PrivateKey
	PublicKeys  []crypto.PublicKey
}

func DeterministicKeySet(seed int64, numProcess int) *KeySet {
	set := &KeySet{
		PrivateKeys: make([]crypto.PrivateKey, numProcess),
		PublicKeys:  make([]crypto.PublicKey, numProcess),
	}
	// Secret: [seed (8 bytes), processId (4 bytes)]
	secret := make([]byte, 8+4)
	binary.LittleEndian.PutUint64(secret, uint64(seed))
	for id := 0; id < numProcess; id++ {
		binary.LittleEndian.PutUint32(secret[8:], uint32(id))
		set.PrivateKeys[id] = crypto.GeneratePrivateKeyFromSecret(secret)
		set.PublicKeys[id] = set.PrivateKeys[id].PubKey()
	}
	return set
}
