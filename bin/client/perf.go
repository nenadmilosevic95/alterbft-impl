package main

import (
	"math"
	"time"
)

// Record stores information of a decided value.
type Record struct {
	Timestamp time.Time
	Latency   time.Duration
}

// Perf stores performance data for an experiment.
type Perf struct {
	Values uint32

	// Measurements interval
	StartTime time.Time
	EndTime   time.Time

	// Latency mean and standard deviation
	moment1 float64
	moment2 float64
}

// Add a Record to the performance data.
func (p *Perf) Add(r *Record) {
	p.Values += 1

	if r.Timestamp.After(p.EndTime) {
		p.EndTime = r.Timestamp
	}
	proposalTS := r.Timestamp.Add(-1 * r.Latency)
	if p.StartTime.IsZero() || proposalTS.Before(p.StartTime) {
		p.StartTime = proposalTS
	}

	// Using online Welford's algorithm from:
	// https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance
	delta1 := float64(r.Latency) - p.moment1
	p.moment1 += delta1 / float64(p.Values)
	delta2 := float64(r.Latency) - p.moment1
	p.moment2 += delta1 * delta2

}

// ValuesPerSec returns throughput in decided values per second.
func (p *Perf) ValuesPerSec() float64 {
	duration := p.EndTime.Sub(p.StartTime)
	return float64(p.Values) / duration.Seconds()
}

// LatencyMean returns the mean latency.
func (p *Perf) LatencyMean() time.Duration {
	return time.Duration(p.moment1)
}

// LatencyStdev returns the standard deviation of latencies.
func (p *Perf) LatencyStdev() time.Duration {
	variance := p.moment2 / float64(p.Values-1)
	return time.Duration(math.Sqrt(variance))
}

// Reset the performance data.
func (p *Perf) Reset() {
	p.Values = 0

	p.StartTime = *new(time.Time)
	p.EndTime = p.StartTime

	p.moment1 = 0.0
	p.moment2 = 0.0
}
