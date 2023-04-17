package consensus

import (
	"fmt"

	"dslab.inf.usi.ch/tendermint/crypto"
)

// Certificate types.
const (
	INVALID_CERT = iota
	SILENCE_CERT
	BLOCK_CERT
)

// MessageSignatureSize is the byte size of the signature portion of a message.
const MessageSignatureSize = SignatureSize + 2 // Sender (uint16)

// A Certificate aggregates identical messages signed by multiple senders.
type Certificate struct {
	// Signed payload
	Type    int16   // Type of the certificate (BLOCK or SILENCE)
	Epoch   int64   // Epoch in which the certificate was created.
	blockID BlockID // BlockID of a block we have a block certificate of.
	Height  int64
	// This field is used only locally, it is not marshalled and it needs to be
	// set explicitly when we update validCertificate or lockedCertificate.
	// It is used when creating new block based on the validCertificate.
	block *Block

	// Signatures
	Signatures map[int]Signature

	// A certificate has the same payload signed by multiple replicas.
	// It is set when unmarshalling the message.
	payload []byte

	// Unexported byte version
	marshalled []byte
}

// NewBlockCertificate creates a block certificate with no signers.
// This method fills the fields that are common to the certified VOTE messages.
func NewBlockCertificate(epoch int64, blockID BlockID, height int64) *Certificate {
	return &Certificate{
		Type:       BLOCK_CERT,
		Epoch:      epoch,
		blockID:    blockID,
		Height:     height,
		Signatures: make(map[int]Signature),
	}
}

// NewSilenceCertificate creates a silence certificate with no signers.
// This method fills the fields that are common to the certified SILENCE messages.
func NewSilenceCertificate(epoch int64) *Certificate {
	return &Certificate{
		Type:  SILENCE_CERT,
		Epoch: epoch,

		Signatures: make(map[int]Signature),
	}
}

func (c *Certificate) Equal(cc *Certificate) bool {
	if c == nil || cc == nil {
		return c == cc
	}
	if !c.blockID.Equal(cc.blockID) ||
		c.Epoch != cc.Epoch || c.Type != cc.Type ||
		len(c.Signatures) != len(cc.Signatures) {
		return false
	}
	for i, _ := range c.Signatures {
		if !c.Signatures[i].Equal(cc.Signatures[i]) {
			return false
		}
	}
	return true
}

// AddSignature adds a signature to the certificate.
// The signature should sign the common fields (the payload) of the certificate.
// FIXME: the certificate signatures are shared with the certified messages.
func (c *Certificate) AddSignature(s Signature, sender int) bool {
	// FIXME: check if the message matched the certificate?
	if _, ok := c.Signatures[sender]; !ok {
		c.Signatures[sender] = s
		return true
	}
	return false
}

func (c *Certificate) RanksHigherOrEqual(cc *Certificate) bool {
	if c != nil && cc == nil {
		return true
	}
	if c == nil && cc != nil {
		return false
	}
	if c == nil && cc == nil {
		return true
	}
	return c.Epoch >= cc.Epoch

}

func (c *Certificate) BlockID() BlockID {
	if c != nil {
		return c.blockID
	}
	return nil
}

func (c *Certificate) SignatureCount() int {
	return len(c.Signatures)
}

// String returns string representation of a certificate.
func (c *Certificate) String() string {
	return fmt.Sprintf("Type: %d\nEpoch: %v\nBlockID:%v\nSignatures:%v\n", c.Type, c.Epoch, c.blockID, c.Signatures)
}

// ByteSize returns the size of the bytes encoded version of the certificate.
// The certificate is composed by a constant-size payload portion, plus a
// variable number of pairs signer and signature signing the same payload.
func (c *Certificate) ByteSize() int {
	payloadSize := 10
	if c.Type == BLOCK_CERT {
		payloadSize += BlockIDSize + 8
	}
	numSignatures := len(c.Signatures)
	return payloadSize + numSignatures*MessageSignatureSize
}

// Marshal serialises the certificate to an array of bytes.
func (c *Certificate) Marshall() []byte {
	if c.marshalled == nil {
		c.marshalled = make([]byte, c.ByteSize())
		c.MarshallTo(c.marshalled)
	}
	return c.marshalled
}

// MarshallTo writes a certificate in bytes to a buffer and returns its size.
// The buffer is assumed to have enough space to store the encoded certificate.
func (c *Certificate) MarshallTo(buffer []byte) (payloadSize int) {
	// 1. Payload portion
	buffer[0] = MessageCode
	buffer[1] = byte(c.Type)                        // 2 bytes
	encoding.PutUint64(buffer[2:], uint64(c.Epoch)) // 8 bytes
	index := 10
	if c.Type == BLOCK_CERT {
		encoding.PutUint64(buffer[index:], uint64(c.Height))
		index += 8
		c.BlockID().MarshallTo(buffer[index:]) // BlockIDSize
		index += BlockIDSize
	}
	// 2. Number of signatures * MessageSignatureSize bytes
	for sender, signature := range c.Signatures {
		encoding.PutUint16(buffer[index:], uint16(sender))
		signature.MarshallTo(buffer[index+2:])
		index += MessageSignatureSize
	}
	return c.ByteSize()
}

// Payload returns the certificate payload, encoded into bytes.
// The payload contains the common fields of the messages added to the certificate.
func (c *Certificate) Payload() []byte {
	c.Marshall()
	payloadSize := 10
	if c.Type == BLOCK_CERT {
		payloadSize += BlockIDSize + 8
	}
	return c.marshalled[:payloadSize]
}

// CertificateFromBytes parses a certificate from a byte array.
// The provided byte array is retained and should not be externally re-used.
func CertificateFromBytes(buffer []byte) *Certificate {
	certificate := new(Certificate)
	certificate.marshalled = buffer
	// 1. Payload portion
	certificate.Type = int16(buffer[1])                    // 1 byte
	certificate.Epoch = int64(encoding.Uint64(buffer[2:])) // 8 bytes
	payloadSize := 10
	if certificate.Type == BLOCK_CERT {
		certificate.Height = int64(encoding.Uint64(buffer[payloadSize:]))
		payloadSize += 8
		certificate.blockID = BlockIDFromBytes(buffer[payloadSize:]) // BlockIDSize
		payloadSize += BlockIDSize
	}
	// 2. Number of signatures * MessageSignatureSize bytes
	numSignatures := (len(buffer) - payloadSize) / MessageSignatureSize
	certificate.Signatures = make(map[int]Signature)
	for count := 0; count < numSignatures; count++ {
		index := payloadSize + count*MessageSignatureSize
		sender := int(encoding.Uint16(buffer[index:]))
		certificate.Signatures[sender] = SignatureFromBytes(buffer[index+2:])
	}
	return certificate
}

func (c *Certificate) ReconstructMessage(sender int) *Message {
	var message *Message
	if signature, ok := c.Signatures[sender]; ok {
		if c.Type == BLOCK_CERT {
			message = NewVoteMessage(c.Epoch, c.BlockID(), c.Height, int16(sender))
		}
		if c.Type == SILENCE_CERT {
			message = NewSilenceMessage(c.Epoch, int16(sender))
		}
		message.Signature = signature
	}
	return message
}

// ReconstructMessages reconstructs the messages aggregated by the certificate.
// FIXME: the reconstructed messages have fields shared with this certificate.
func (c *Certificate) ReconstructMessages() []*Message {
	messages := make([]*Message, len(c.Signatures))
	i := 0
	var message *Message
	for sender, signature := range c.Signatures {
		if c.Type == BLOCK_CERT {
			message = NewVoteMessage(c.Epoch, c.BlockID(), c.Height, int16(sender))

		}
		if c.Type == SILENCE_CERT {
			message = NewSilenceMessage(c.Epoch, int16(sender))
		}
		message.Signature = signature
		messages[i] = message
		i++
	}
	return messages
}

func (c *Certificate) GetCryptoSignatures() []*crypto.Signature {
	n := len(c.Signatures)
	signatures := make([]*crypto.Signature, n)
	i := 0
	for sender, signature := range c.Signatures {
		signatures[i] = crypto.NewSignature(sender, c.Payload(), signature)
		i++
	}
	return signatures
}
