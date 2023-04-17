package crypto

import "fmt"

type Signature struct {
	ID        int
	Payload   []byte
	Signature []byte
}

func NewSignature(ID int, payload []byte, signature []byte) *Signature {
	return &Signature{
		ID:        ID,
		Payload:   payload,
		Signature: signature,
	}
}

// Key returns a fixed-size key for this signature.
func (s *Signature) Key() (key [SignatureSize]byte) {
	copy(key[:], s.Signature)
	return
}

func (s Signature) String() string {
	return fmt.Sprint("ID:", s.ID, "Payload:", s.Payload, "Signature:", s.Signature)
}
