package gossip

import (
	"encoding/binary"
	"io"

	"dslab.inf.usi.ch/tendermint/net"
)

const HeaderSize = 6 // uint16 + uint32

var encoding = binary.LittleEndian

type Message struct {
	Sender  uint16
	Message net.Message

	// Not exported
	from int
	to   []int // Support for unicasts
}

func (m *Message) ID() net.MessageID {
	return m.Message.ID()
}

func (m *Message) Header() []byte {
	var header [HeaderSize]byte
	encoding.PutUint16(header[0:2], m.Sender)
	encoding.PutUint32(header[2:6], uint32(len(m.Message)))
	return header[:]
}

func (m *Message) ReadFrom(r io.Reader) error {
	var header [HeaderSize]byte
	_, err := io.ReadFull(r, header[:])
	if err != nil {
		return err
	}
	m.Sender = encoding.Uint16(header[0:2])
	size := int(encoding.Uint32(header[2:6]))
	if size > 0 {
		payload := make([]byte, size)
		_, err = io.ReadFull(r, payload)
		if err == nil {
			m.Message = payload
		}
	}
	return err
}

func (m *Message) WriteTo(w io.Writer) error {
	_, err := w.Write(m.Header())
	if err == nil && len(m.Message) > 0 {
		_, err = w.Write(m.Message)
	}
	return err
}
