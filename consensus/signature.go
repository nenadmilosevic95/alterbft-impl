package consensus

import (
	"bytes"
	"encoding/base64"

	"dslab.inf.usi.ch/tendermint/crypto"
)

const SignatureSize = crypto.SignatureSize

// Signature represents a signature of a message.
type Signature []byte

func (s Signature) Equal(ss Signature) bool {
	return bytes.Equal(s, ss)
}

func (s Signature) ByteSize() int {
	return SignatureSize
}

// String returns string representation of a signature.
func (s Signature) String() string {
	return base64.StdEncoding.EncodeToString(s)
}

// MarshallTo marshall signature to the buffer.
func (s Signature) MarshallTo(buffer []byte) int {
	copy(buffer[:], s)
	return SignatureSize
}

// SignatureFromBytes unmarshall signature from a buffer.
func SignatureFromBytes(buffer []byte) Signature {
	return buffer[:SignatureSize]
}
