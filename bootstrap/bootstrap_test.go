package bootstrap

import (
	"bytes"
	"testing"
)

func TestBootstrapPeriodicMessages(t *testing.T) {
	pid := 1
	b := NewBootstrap(pid, 3)
	if b.Active() || b.Done() {
		t.Error("Process has not a quorum but is active/done", b.Active(), b.Done())
	}

	m1 := b.ProcessTick()
	if m1 == nil {
		t.Fatal("Bootstrap return nil message on clock tick")
	}
	if m1.Sender() != pid {
		t.Error("Bootstrap message has wrong sender", m1, pid)
	}
	if m1.Active() {
		t.Error("Unexpected active bootstrap message", m1)
	}
	z1 := m1.Marshall()
	if b.Active() || b.Done() {
		t.Error("Process has not a quorum but is active/done", b.Active(), b.Done())
	}

	m2 := b.ProcessTick()
	if m2 == nil {
		t.Fatal("Bootstrap return nil message on clock tick")
	}
	if m2.Sender() != pid {
		t.Error("Bootstrap message has wrong sender", m2, pid)
	}
	if m2.Active() {
		t.Error("Unexpected active bootstrap message", m2)
	}
	z2 := m2.Marshall()
	if b.Active() || b.Done() {
		t.Error("Process has not a quorum but is active/done", b.Active(), b.Done())
	}

	if bytes.Equal(z1, z2) {
		t.Error("Messages broadcast in two clock ticks are equal", z1, z2)
	}

	b.ProcessMessage(NewMessage(0, 1, false))
	b.ProcessMessage(NewMessage(3, 2, false))
	b.ProcessMessage(NewMessage(1, 1, false)) // Should be active from now

	m3 := b.ProcessTick()
	if m3 == nil {
		t.Fatal("Bootstrap return nil message on clock tick")
	}
	if m3.Sender() != pid {
		t.Error("Bootstrap message has wrong sender", m3, pid)
	}
	if !m3.Active() {
		t.Error("Expected active bootstrap message", m3)
	}
	z3 := m3.Marshall()

	if !b.Active() {
		t.Error("Process is expected to be active after a quorum")
	}
	if b.Done() {
		t.Error("Process is not expected to be done in the protocol")
	}

	if bytes.Equal(z1, z3) {
		t.Error("Messages broadcast in two clock ticks are equal", z1, z3)
	}
	if bytes.Equal(z2, z3) {
		t.Error("Messages broadcast in two clock ticks are equal", z2, z3)
	}
}

func TestBootstrapMessagesFromQuorum(t *testing.T) {
	pid := 1
	b := NewBootstrap(pid, 3)

	m := b.ProcessMessage(NewMessage(0, 1, false))
	if m != nil {
		t.Error("Unexpected message, quorum 1, got", m)
	}
	if b.Active() || b.Done() {
		t.Error("Process has not a quorum but is active/done", b.Active(), b.Done())
	}

	m = b.ProcessMessage(NewMessage(0, 2, true)) // duplicated
	if m != nil {
		t.Error("Unexpected message, quorum 1, got", m)
	}
	if b.Active() || b.Done() {
		t.Error("Process has not a quorum but is active/done", b.Active(), b.Done())
	}

	m = b.ProcessMessage(NewMessage(1, 3, false))
	if m != nil {
		t.Error("Unexpected message, quorum 2, got", m)
	}
	if b.Active() || b.Done() {
		t.Error("Process has not a quorum but is active/done", b.Active(), b.Done())
	}

	m = b.ProcessMessage(NewMessage(0, 3, true)) // duplicated
	if m != nil {
		t.Error("Unexpected message, quorum 2, got", m)
	}
	if b.Active() || b.Done() {
		t.Error("Process has not a quorum but is active/done", b.Active(), b.Done())
	}

	m = b.ProcessMessage(NewMessage(1, 3, false)) // duplicated'
	if m != nil {
		t.Error("Unexpected message, quorum 2, got", m)
	}
	if b.Active() || b.Done() {
		t.Error("Process has not a quorum but is active/done", b.Active(), b.Done())
	}

	m = b.ProcessMessage(NewMessage(2, 1, false))
	if m == nil {
		t.Fatal("Unexpected nil message, quorum 3")
	}
	if m.Sender() != pid {
		t.Error("Bootstrap message has wrong sender", m, pid)
	}
	if !m.Active() {
		t.Error("Expected active bootstrap message", m)
	}

	if !b.Active() {
		t.Error("Process is expected to be active after a quorum")
	}
	if b.Done() {
		t.Error("Process is not expected to be done in the protocol")
	}

	m = b.ProcessMessage(NewMessage(2, 2, true)) // duplicated
	if m != nil {
		t.Error("Unexpected message after quorum, got", m)
	}
	if !b.Active() {
		t.Error("Process is expected to be active after a quorum")
	}
	if b.Done() {
		t.Error("Process is not expected to be done in the protocol")
	}

	m = b.ProcessMessage(NewMessage(4, 1, false))
	if m != nil {
		t.Error("Unexpected message after quorum, got", m)
	}
	if !b.Active() {
		t.Error("Process is expected to be active after a quorum")
	}
	if b.Done() {
		t.Error("Process is not expected to be done in the protocol")
	}
}

func TestBootstrapActiveDone(t *testing.T) {
	for i, test := range []struct {
		messages []*Message
		active   bool
		done     bool
	}{
		{
			[]*Message{},
			false,
			false,
		},
		{
			[]*Message{
				NewMessage(0, 1, false),
				NewMessage(1, 1, false),
			},
			false,
			false,
		},
		{
			[]*Message{
				NewMessage(0, 1, false),
				NewMessage(1, 1, false),
				NewMessage(2, 1, false),
			},
			true,
			false,
		},
		{
			[]*Message{
				NewMessage(0, 1, true),
				NewMessage(1, 1, true),
				NewMessage(2, 1, true),
			},
			true,
			true,
		},
		{
			[]*Message{
				NewMessage(0, 1, true),
				NewMessage(1, 1, false),
				NewMessage(2, 1, true),
			},
			true,
			false,
		},
	} {
		b := NewBootstrap(1, 3)
		for _, m := range test.messages {
			b.ProcessMessage(m)
		}
		if b.Active() != test.active {
			t.Error(i, "Expected active", test.active, "got", b.Active())
		}
		if b.Done() != test.done {
			t.Error(i, "Expected done", test.done, "got", b.Done())
		}
	}
}
