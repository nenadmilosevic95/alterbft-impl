package bootstrap

import (
	"encoding/binary"

	"dslab.inf.usi.ch/tendermint/net"
)

const MessageCode = byte(255)

// Message is a bootstrap protocol message.
// It announces a sender, and reports its state: active or not.
type Message struct {
	sender int
	seqnum int
	active bool
}

// NewMessage creates a bootstrap message.
func NewMessage(sender, seqnum int, active bool) *Message {
	return &Message{
		sender: sender,
		seqnum: seqnum,
		active: active,
	}
}

// Sender returns the message sender.
func (m *Message) Sender() int {
	return m.sender
}

// Active returns the active state reported by the sender.
func (m *Message) Active() bool {
	return m.active
}

// Encoding to translate between bytes and number fields
var encoding binary.ByteOrder = binary.LittleEndian

// Marshall marshalls this message into a network message.
func (m *Message) Marshall() net.Message {
	payload := make(net.Message, 2+2+2)
	payload[0] = MessageCode
	if m.Active() {
		payload[1] = 1
	}
	encoding.PutUint16(payload[2:4], uint16(m.sender))
	encoding.PutUint16(payload[4:6], uint16(m.seqnum))
	return payload
}

// NewMessageFromBytes creates a bootstrap message from a network message.
func NewMessageFromBytes(marshalled net.Message) *Message {
	var active bool
	if marshalled[1] > 0 {
		active = true
	}
	sender := int(encoding.Uint16(marshalled[2:4]))
	seqnum := int(encoding.Uint16(marshalled[4:6]))
	return NewMessage(sender, seqnum, active)
}
