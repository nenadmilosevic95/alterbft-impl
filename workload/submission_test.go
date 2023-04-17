package workload

import (
	"testing"
	"time"
)

func TestSubmissionRecord(t *testing.T) {
	t0 := time.Now()
	sender := 23
	size := 1024

	s := NewSubmission(sender, size)
	if s.Sender != sender {
		t.Error("Submission sender does not match", s, sender)
	}
	if s.Size != size {
		t.Error("Submission size does not match", s, size)
	}
	t1 := time.Now()
	if s.Time.Before(t0) || s.Time.After(t1) {
		t.Error("Submission time is in the expected interval", s, t0, t1)
	}

	value := make([]byte, size)
	s.Write(value)

	ss := SubmissionFromValue(value)
	if ss.Sender != s.Sender {
		t.Error("Unmarshalled submission sender does not match", s.Sender, ss.Sender)
	}
	if ss.Size != s.Size {
		t.Error("Unmarshalled submission size does not match", s.Size, ss.Size)
	}
	if !ss.Time.Equal(s.Time) {
		t.Error("Unmarshalled submission size does not match", s.Time, ss.Time)
	}
}
