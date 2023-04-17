package workload

import (
	"math"
	"time"
)

// Latency stores latency measurements.
// Computes latency average and standard deviation.
type Latency struct {
	count   int
	moment1 float64
	moment2 float64
}

// Add a latency measurement.
func (l *Latency) Add(tlatency time.Duration) {
	latency := float64(tlatency)
	l.count += 1

	// Using online Welford's algorithm from:
	// https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance
	delta1 := latency - l.moment1
	l.moment1 += delta1 / float64(l.count)
	delta2 := latency - l.moment1
	l.moment2 += delta1 * delta2
}

// Count returns the number of latency measurements.
func (l *Latency) Count() int {
	return l.count
}

// Average returns the aggregated average latency.
func (l *Latency) Average() time.Duration {
	return time.Duration(l.moment1)
}

// Stdev returns the aggregated latency standard deviation.
func (l *Latency) Stdev() time.Duration {
	if l.count < 2 {
		return 0
	}
	variance := l.moment2 / float64(l.count-1)
	return time.Duration(math.Sqrt(variance))
}
