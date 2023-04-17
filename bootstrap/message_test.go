package bootstrap

import "testing"

func TestMessageMarshalling(t *testing.T) {
	m0 := NewMessage(0, 0, false)
	nm := m0.Marshall() // net.Message

	if nm.Code() != MessageCode {
		t.Error("Unexpected marshalled message code", nm.Code())
	}

	m := NewMessageFromBytes(nm)
	if m == nil {
		t.Fatal("Failed to unmarshall message", nm, "expected", m0)
	}

	if m.Sender() != m0.Sender() {
		t.Error("Unmarshalled message sender differs", m, m0.Sender())
	}

	if m.seqnum != m0.seqnum {
		t.Error("Unmarshalled message seqnum differs", m, m0.seqnum)
	}
	if m.Active() != m0.Active() {
		t.Error("Unmarshalled message active differs", m, m0.Active())
	}

	m1 := NewMessage(100, 77, true)
	nm = m1.Marshall() // net.Message

	if nm.Code() != MessageCode {
		t.Error("Unexpected marshalled message code", nm.Code())
	}

	m = NewMessageFromBytes(nm)
	if m == nil {
		t.Fatal("Failed to unmarshall message", m1, nm)
	}

	if m.Sender() != m1.Sender() {
		t.Error("Unmarshalled message sender differs", m, m1.Sender())
	}

	if m.seqnum != m1.seqnum {
		t.Error("Unmarshalled message seqnum differs", m, m1.seqnum)
	}

	if m.Active() != m1.Active() {
		t.Error("Unmarshalled message active differs", m, m1.Active())
	}
}
