package crypto

import (
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

// Size in bytes of produced signatures: 64 bytes.
const SignatureSize = ed25519.SignatureSize

// A key used to verify signatures produced by the associated private key.
type PublicKey tmcrypto.PubKey

// A key used to produce signatures and to generate the associated public key.
type PrivateKey tmcrypto.PrivKey

// GeneratePrivateKey generates a new private key.
func GeneratePrivateKey() PrivateKey {
	return ed25519.GenPrivKey()
}

// GeneratePrivateKeyFromSecret generates a private key from a secret.
func GeneratePrivateKeyFromSecret(secret []byte) PrivateKey {
	return ed25519.GenPrivKeyFromSecret(secret)
}
