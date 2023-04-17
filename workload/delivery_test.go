package workload

import (
	"testing"
	"time"

	"dslab.inf.usi.ch/tendermint/consensus"
)

func TestDeliveryRecord(t *testing.T) {
	e1 := int64(2)
	v1 := make([]byte, 123)
	sub := NewSubmission(12, len(v1))

	t0 := time.Now()
	b1 := consensus.NewBlock(v1, nil)
	d1 := NewDelivery(e1, b1)

	if d1.Epoch != e1 {
		t.Error("Unexpected delivery epoch", e1, d1)
	}
	if d1.Height != b1.Height {
		t.Error("Unexpected delivery height", b1.Height, d1)
	}
	if d1.Size != len(v1) {
		t.Error("Unexpected delivery height", len(v1), d1)
	}
	t1 := time.Now()
	if d1.Time.Before(t0) || d1.Time.After(t1) {
		t.Error("Delivery time is in the expected interval", d1, t0, t1)
	}

	if d1.Submission != nil {
		t.Error("Unexpected associated submission", d1.Submission)
	}
	if d1.Latency() != 0 {
		t.Error("Unexpected latency without associated submission", d1.Latency())
	}

	d1.Submission = sub
	if d1.Latency() <= 0 || d1.Latency() > t1.Sub(t0) {
		t.Error("Unexpected invalid latency with associated submission",
			t0, t1, d1.Latency())
	}
}
