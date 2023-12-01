package consensus

import (
	"encoding/binary"
	"fmt"

	"dslab.inf.usi.ch/tendermint/crypto"
)

// Message types.
const (
	PROPOSE = iota
	SILENCE
	VOTE
	QUIT_EPOCH
	CERTIFICATE

	DELTA_REQUEST
	DELTA_RESPONSE
)

// Code of consensus marshalled messages.
const MessageCode = byte(0)

// Zeroed array representing an invalid signature.
var zeroSignature [crypto.SignatureSize]byte

// Encoding to translate between bytes and number fields
var encoding binary.ByteOrder = binary.LittleEndian

// Message is a generic consensus message.
type Message struct {
	Type   int16
	Epoch  int64
	Height int64

	Block       *Block
	BlockID     BlockID
	Certificate *Certificate

	Sender     int
	Signature  Signature
	Sender2    int
	Signature2 Signature

	// This sender is used only for Delta statistics
	// and it is set when a process forward the proposal message.
	SenderFwd int

	// Unexported byte version
	marshalled []byte
	// Message payload
	payload []byte
}

func NewProposeMessage(e int64, b *Block, c *Certificate, sender int16) *Message {
	return &Message{
		Type:        PROPOSE,
		Epoch:       e,
		Block:       b,
		Certificate: c,
		Sender:      int(sender),

		SenderFwd: int(sender),
	}
}

func NewSilenceMessage(e int64, sender int16) *Message {
	return &Message{
		Type:   SILENCE,
		Epoch:  e,
		Sender: int(sender),
	}
}

func NewVoteMessage(e int64, id BlockID, height int64, sender int16, sender2 int16) *Message {
	return &Message{
		Type:    VOTE,
		Epoch:   e,
		BlockID: id,
		Height:  height,
		Sender:  int(sender),
		Sender2: int(sender2),
	}
}

func NewQuitEpochMessage(e int64, c *Certificate) *Message {
	return &Message{
		Type:        QUIT_EPOCH,
		Epoch:       e,
		Certificate: c,
	}
}

func NewCertificateMessage(e int64, c *Certificate) *Message {
	return &Message{
		Type:        CERTIFICATE,
		Epoch:       e,
		Certificate: c,
	}
}

func NewDeltaRequestMessage(payload []byte, sender int) *Message {
	return &Message{
		Type:    DELTA_REQUEST,
		payload: payload,
		Sender:  sender,
	}
}

func NewDeltaResponseMessage(payload []byte, sender int) *Message {
	return &Message{
		Type:    DELTA_RESPONSE,
		payload: payload,
		Sender:  sender,
	}
}

// MessageFromBytes parses a message from a byte array.
// The provided byte array is retained and should not be externally re-used.
func MessageFromBytes(buffer []byte) *Message {
	mType := int16(buffer[1])
	var epoch int64
	var height int64
	if mType != QUIT_EPOCH && mType != DELTA_REQUEST && mType != DELTA_RESPONSE {
		epoch = int64(encoding.Uint64(buffer[2:]))
	}
	index := 10
	var block *Block = nil
	var blockID BlockID = nil
	var certificate *Certificate = nil
	var payload []byte
	var sender2 int16
	var signature2 Signature
	switch mType {
	case PROPOSE:
		n := int32(encoding.Uint32(buffer[index:]))
		index += 4
		end := index + int(n)
		blockHeightIndex := index
		block = BlockFromBytes(buffer[index:end])
		index = end
		if buffer[index] == 1 {
			index += 1
			end := len(buffer) - SignatureSize - 4
			certificate = CertificateFromBytes(buffer[index:end])
			index = end
		} else {
			index += 1
		}
		// payload is epoch+blockID
		payload = make([]byte, 16+BlockIDSize)
		copy(payload[:8], buffer[2:10])
		copy(payload[8:16], buffer[blockHeightIndex:blockHeightIndex+8])
		block.BlockID().MarshallTo(payload[16:])
	case VOTE:
		height = int64(encoding.Uint64(buffer[index:]))
		index += 8
		blockID = BlockIDFromBytes(buffer[index:])
		index += BlockIDSize
		payload = buffer[2:index]
		sender2 = int16(encoding.Uint16(buffer[index:]))
		index += 2
		signature2 = SignatureFromBytes(buffer[index:])
		index += SignatureSize
	case SILENCE:
		payload = buffer[2:index]
	case QUIT_EPOCH:
		certificate = CertificateFromBytes(buffer[2:])
		epoch = certificate.Epoch
	case CERTIFICATE:
		certificate = CertificateFromBytes(buffer[index:])
		index += certificate.ByteSize()
	case DELTA_REQUEST, DELTA_RESPONSE:
		payload = buffer[2 : len(buffer)-2]
		index = len(buffer) - 2
	}
	var sender int16
	if mType != QUIT_EPOCH && mType != CERTIFICATE {
		sender = int16(encoding.Uint16(buffer[index:]))
		index += 2
	}
	var senderFwd int16
	if mType == PROPOSE {
		senderFwd = int16(encoding.Uint16(buffer[index:]))
		index += 2
	}
	var signature Signature
	if mType != QUIT_EPOCH && mType != CERTIFICATE && mType != DELTA_REQUEST && mType != DELTA_RESPONSE {
		signature = SignatureFromBytes(buffer[index:])
	}

	return &Message{
		Type:        mType,
		Epoch:       epoch,
		Height:      height,
		Block:       block,
		BlockID:     blockID,
		Certificate: certificate,

		Sender:     int(sender),
		Signature:  signature,
		Sender2:    int(sender2),
		Signature2: signature2,

		SenderFwd: int(senderFwd),

		marshalled: buffer,
		payload:    payload,
	}
}

// String returns string representation of a certificate.
func (m *Message) String() string {
	return fmt.Sprintf("Type: %d\nEpoch: %v\nBlock:%v\nBlockID:%v\nCertificate:%v\nSender:%v\nSignature:%v\n", m.Type, m.Epoch, m.Block, m.BlockID, m.Certificate, m.Sender, m.Signature)
}

// ByteSize returns the size of the bytes encoded version of the message.
func (m *Message) ByteSize() int {
	switch m.Type {
	case PROPOSE:
		if m.Certificate == nil {
			return 19 + m.Block.ByteSize() + SignatureSize
		} else {
			return 19 + m.Block.ByteSize() + m.Certificate.ByteSize() + SignatureSize
		}
	case SILENCE:
		return 12 + SignatureSize
	case VOTE:
		return 22 + BlockIDSize + 2*SignatureSize
	case QUIT_EPOCH:
		return 2 + m.Certificate.ByteSize()
	case CERTIFICATE:
		return 2 + m.Certificate.ByteSize() + 8
	case DELTA_REQUEST, DELTA_RESPONSE:
		return 2 + len(m.payload) + 2
	default:
		return 0
	}
}

// Payload returns message payload, it is safer to access payload through this method.
func (m *Message) Payload() []byte {
	if m.payload == nil {
		m.Marshall()
	}
	return m.payload
}

// Marshal serialises a consensus message to an array of bytes.
func (m *Message) Marshall() []byte {
	if m.marshalled == nil {
		m.marshalled = make([]byte, m.ByteSize())
		m.MarshallTo(m.marshalled)
	}
	return m.marshalled
}

// This message is only called before process forwards the proposal message.
func (m *Message) setFwdSender(sender int) {
	index := m.ByteSize() - SignatureSize - 2
	encoding.PutUint16(m.marshalled[index:], uint16(sender))
}

// MarshallTo encodes a message into bytes and writes them to a buffer.
// The buffer is assumed to have enough space to store the encoded message.
func (m *Message) MarshallTo(buffer []byte) {
	var n int
	buffer[0] = MessageCode
	buffer[1] = byte(m.Type) // 2 bytes
	if m.Type != QUIT_EPOCH && m.Type != DELTA_REQUEST && m.Type != DELTA_RESPONSE {
		encoding.PutUint64(buffer[2:], uint64(m.Epoch))
	}
	index := 10
	switch m.Type {
	case PROPOSE:
		blockSize := m.Block.ByteSize()
		encoding.PutUint32(buffer[index:], uint32(blockSize))
		index += 4
		n = m.Block.MarshallTo(buffer[index:])
		index += n
		if m.Certificate != nil {
			copy(buffer[index:], []byte{1})
			index += 1
			end := len(buffer) - SignatureSize - 4
			if m.Certificate.marshalled != nil {
				copy(buffer[index:end], m.Certificate.marshalled)
			} else {
				m.Certificate.MarshallTo(buffer[index:end])
			}
			index = end
		} else {
			copy(buffer[index:], []byte{0})
			index += 1
		}
		// payload is epoch+blockID
		m.payload = make([]byte, 16+BlockIDSize)
		encoding.PutUint64(m.payload[0:], uint64(m.Epoch))
		encoding.PutUint64(m.payload[8:], uint64(m.Block.Height))
		m.Block.BlockID().MarshallTo(m.payload[16:])
	case VOTE:
		encoding.PutUint64(buffer[index:], uint64(m.Height))
		index += 8
		n = m.BlockID.MarshallTo(buffer[index:])
		index += n
		m.payload = buffer[2:index]
		encoding.PutUint16(buffer[index:], uint16(m.Sender2))
		index += 2
		n := m.Signature2.MarshallTo(buffer[index:])
		index += n
		//m.payload = make([]byte, 8+BlockIDSize)
		//encoding.PutUint64(m.payload[0:], uint64(m.Epoch))
		//m.BlockID.MarshallTo(m.payload[8:])
	case SILENCE:
		m.payload = buffer[2:index]
	case QUIT_EPOCH:
		n = m.Certificate.MarshallTo(buffer[2:])
	case CERTIFICATE:
		n = m.Certificate.MarshallTo(buffer[index:])
		index += m.Certificate.ByteSize()
	case DELTA_REQUEST, DELTA_RESPONSE:
		copy(buffer[2:], m.payload)
		index = 2 + len(m.payload)
	}
	if m.Type != QUIT_EPOCH && m.Type != CERTIFICATE {
		encoding.PutUint16(buffer[index:], uint16(m.Sender))
		index += 2
	}
	if m.Type == PROPOSE {
		encoding.PutUint16(buffer[index:], uint16(m.SenderFwd))
	}
}

// Sign the message with the provided private key.
// The message is first encoded into bytes, from which the signature is computed.
// The computed signature bytes then becomes the suffix of the byte-encoded message.
func (m *Message) Sign(key crypto.PrivateKey) {
	if m.Type == QUIT_EPOCH || m.Type == CERTIFICATE || m.Type == DELTA_REQUEST || m.Type == DELTA_RESPONSE {
		return
	}
	message := m.Marshall() // [message bytes : message signature]
	sigIndex := len(message) - SignatureSize
	sig, _ := key.Sign(m.Payload())
	m.Signature = SignatureFromBytes(sig)
	m.Signature.MarshallTo(message[sigIndex:])
}

// VerifySignature verifies the message signature with the provided public key.
func (m *Message) VerifySignature(key crypto.PublicKey) bool {
	return key.VerifySignature(m.Payload(), m.Signature)
}

// GetCryptoSigntures returns all signatures of a message.
func (m *Message) GetCryptoSignatures() []*crypto.Signature {
	var sigs []*crypto.Signature
	if m.Signature != nil {
		sigs = append(sigs, crypto.NewSignature(m.Sender, m.Payload(), m.Signature))
	}
	if m.Signature2 != nil {
		sigs = append(sigs, crypto.NewSignature(m.Sender2, m.Payload(), m.Signature2))
	}
	if m.Certificate != nil {
		sigs = append(sigs, m.Certificate.GetCryptoSignatures()...)
	}
	return sigs
}
