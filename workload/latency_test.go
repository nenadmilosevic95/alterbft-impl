package workload

import (
	"testing"
	"time"
)

func TestLatency(t *testing.T) {
	l := new(Latency)
	if l.Count() != 0 {
		t.Error("Expected zero latency count, got", l.Count())
	}
	if l.Average() != 0.0 {
		t.Error("Expected zero mean latency, got", l.Average())
	}
	if l.Stdev() != 0.0 {
		t.Error("Expected zero stdev latency, got", l.Stdev())
	}

	lat1 := time.Millisecond
	l.Add(lat1)
	if l.Count() != 1 {
		t.Error("Expected one latency count, got", l.Count())
	}
	if l.Average() != lat1 {
		t.Error("Expected mean latency", lat1, "got", l.Average())
	}
	if l.Stdev() != 0.0 {
		t.Error("Expected zero stdev latency, got", l.Stdev())
	}

	l.Add(lat1)
	if l.Count() != 2 {
		t.Error("Expected one latency count, got", l.Count())
	}
	if l.Average() != lat1 {
		t.Error("Expected mean latency", lat1, "got", l.Average())
	}
	if l.Stdev() != 0.0 {
		t.Error("Expected zero stdev latency, got", l.Stdev())
	}

	lat2 := time.Millisecond + 500*time.Microsecond
	l.Add(lat2)
	if l.Count() != 3 {
		t.Error("Expected one latency count, got", l.Count())
	}
	av := (lat1 + lat1 + lat2) / 3
	if l.Average() != av {
		t.Error("Expected mean latency", av, "got", l.Average())
	}
	if l.Stdev() <= 0.0 {
		t.Error("Expected stdev latency larger than 0, got", l.Stdev())
	}
}
