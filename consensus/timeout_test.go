package consensus

import (
	"testing"
	"time"
)

func assertTimeoutEquals(t *testing.T, to, expected *Timeout) {
	t.Helper()
	if to == nil {
		t.Error("Unexpected nil timeout")
		return
	}
	if to.Type != expected.Type {
		t.Error("Expected timeout type", expected.Type, "got", to)
	}
	if to.Epoch != expected.Epoch {
		t.Error("Expected timeout epoch", expected.Epoch, "got", to)
	}
}

func TestTimeoutTicker(t *testing.T) {
	tt := NewTimeoutTicker()
	tt.Start()

	// Two subsequent timestamps
	ts1 := &Timeout{TimeoutPropose, 3, 10 * time.Millisecond, time.Time{}}
	ts2 := &Timeout{TimeoutEquivocation, 3, 15 * time.Millisecond, time.Time{}}
	tt.In <- ts1
	tt.In <- ts2

	select {
	case tts := <-tt.Out:
		now := time.Now()
		if tts != ts1 {
			t.Error("Expected", ts1, "got", tts)
		}
		if now.Before(ts1.deadline) {
			t.Error("Received at", now, "deadline", ts1.deadline)
		}
	case <-time.After(15 * time.Millisecond):
		t.Error("Expected timeout at", ts1.deadline, "nothing by", time.Now())
	}

	select {
	case tts := <-tt.Out:
		now := time.Now()
		if tts != ts2 {
			t.Error("Expected", ts2, "got", tts)
		}
		if now.Before(ts2.deadline) {
			t.Error("Received at", now, "deadline", ts2.deadline)
		}
	case <-time.After(10 * time.Millisecond):
		t.Error("Expected timeout at", ts2.deadline, "nothing by", time.Now())
	}

	// Two timestamps, the second should trigger before the first
	ts1 = &Timeout{TimeoutPropose, 3, 20 * time.Millisecond, time.Time{}}
	ts2 = &Timeout{TimeoutPropose, 3, 10 * time.Millisecond, time.Time{}}
	tt.In <- ts1
	tt.In <- ts2

	select {
	case tts := <-tt.Out:
		now := time.Now()
		if tts != ts2 {
			t.Error("Expected", ts2, "got", tts)
		}
		if now.Before(ts2.deadline) {
			t.Error("Received at", now, "deadline", ts2.deadline)
		}
	case <-time.After(15 * time.Millisecond):
		t.Error("Expected timeout at", ts2.deadline, "nothing by", time.Now())
	}

	select {
	case tts := <-tt.Out:
		now := time.Now()
		if tts != ts1 {
			t.Error("Expected", ts1, "got", tts)
		}
		if now.Before(ts1.deadline) {
			t.Error("Received at", now, "deadline", ts2.deadline)
		}
	case <-time.After(15 * time.Millisecond):
		t.Error("Expected timeout at", ts1.deadline, "nothing by", time.Now())
	}
}
