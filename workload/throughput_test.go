package workload

import (
	"testing"
	"time"
)

func TestThroughput(t *testing.T) {
	tt := new(Throughput)
	if tt.Values() != 0 {
		t.Error("Expected no value, got", tt.Values())
	}
	if tt.Duration() != 0 {
		t.Error("Expected zero duration, got", tt.Duration())
	}
	if tt.ValuesPerSecond() != 0.0 {
		t.Error("Expected zero throughput, got", tt.ValuesPerSecond())
	}
	if tt.BytesPerSecond() != 0.0 {
		t.Error("Expected zero bytes throughput, got", tt.ValuesPerSecond())
	}

	ts := time.Now()
	tt.Add(ts, 1, 100)
	if tt.Values() != 1 {
		t.Error("Expected 1 value, got", tt.Values())
	}
	if tt.Duration() != 0 {
		t.Error("Expected no duration, got", tt.Duration())
	}
	if tt.ValuesPerSecond() != 0.0 {
		t.Error("Expected no throughput, got", tt.ValuesPerSecond())
	}
	if tt.BytesPerSecond() != 0.0 {
		t.Error("Expected zero bytes throughput, got", tt.BytesPerSecond())
	}

	dur := 100 * time.Millisecond
	tt.Add(ts.Add(dur), 1, 200)
	if tt.Values() != 2 {
		t.Error("Expected 2 values, got", tt.Values())
	}
	if tt.Duration() != dur {
		t.Error("Expected duration", dur, "got", tt.Duration())
	}
	msgs := 2.0 / dur.Seconds()
	if tt.ValuesPerSecond() != msgs {
		t.Error("Expected throughput", msgs, "got", tt.ValuesPerSecond())
	}
	bytes := (100.0 + 200.0) / dur.Seconds()
	if tt.BytesPerSecond() != bytes {
		t.Error("Expected throughput", bytes, "got", tt.BytesPerSecond())
	}
}
