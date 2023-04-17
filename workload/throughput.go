package workload

import "time"

// Throughput is a helper for computing throughput.
type Throughput struct {
	values int
	bytes  int

	startTime time.Time
	endTime   time.Time
}

// Add a record to the throughput computation.
// Contract: successive calls record increasing times.
func (t *Throughput) Add(time time.Time, values, byteSize int) {
	t.values += values
	t.bytes += byteSize
	if t.startTime.IsZero() {
		t.startTime = time
	} else {
		t.endTime = time
	}
}

// Duration returns the time interval of throughput measurements.
func (t *Throughput) Duration() time.Duration {
	if t.endTime.IsZero() {
		return 0
	}
	return t.endTime.Sub(t.startTime)
}

// BytesPerSecond returns throughput in bytes per second.
func (t *Throughput) BytesPerSecond() float64 {
	duration := t.Duration()
	if duration == 0 {
		return 0.0
	}
	return float64(t.bytes) / duration.Seconds()
}

// Values returns the number of values recorded.
func (t *Throughput) Values() int {
	return t.values
}

// ValuesPerSecond returns throughput in values per second.
func (t *Throughput) ValuesPerSecond() float64 {
	duration := t.Duration()
	if duration == 0 {
		return 0.0
	}
	return float64(t.values) / duration.Seconds()
}
