package libp2p

import (
	crand "crypto/rand"
	"math/rand"

	"github.com/libp2p/go-libp2p-core/crypto"
)

func DeterministicEDSAKey(seed int64) crypto.PrivKey {
	rnd := rand.New(rand.NewSource(seed))
	priv, _, err := crypto.GenerateEd25519Key(rnd)
	if err != nil {
		return nil
	}
	return priv
}

func RandomEDSAKey() crypto.PrivKey {
	priv, _, err := crypto.GenerateEd25519Key(crand.Reader)
	if err != nil {
		return nil
	}
	return priv
}
