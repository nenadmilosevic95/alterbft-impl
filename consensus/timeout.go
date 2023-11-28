package consensus

import (
	"log"
	"time"
)

const (
	TimeoutPropose = iota
	TimeoutEquivocation
	TimeoutQuitEpoch
	TimeoutEpochChange
)

// Timeout is a generic consensus timeout.
type Timeout struct {
	Type     int16
	Epoch    int64
	Duration time.Duration

	deadline time.Time
}

// TimeoutTicker is a timer that schedules and triggers timeouts.
type TimeoutTicker struct {
	In  chan *Timeout
	Out chan *Timeout

	timer    *time.Timer
	timeouts []*Timeout
}

// NewTimeoutTicker returns an instance of Timeout Ticker.
func NewTimeoutTicker() *TimeoutTicker {
	t := &TimeoutTicker{
		In:    make(chan *Timeout, 100),
		Out:   make(chan *Timeout, 100),
		timer: time.NewTimer(0),
	}
	return t
}

// Start the timeout scheduling thread in background.
func (t *TimeoutTicker) Start() {
	go t.run()
}

func (t *TimeoutTicker) resetTimer() time.Time {
	// Next deadline is the lowest scheduled deadline
	now := time.Now()
	nextDeadline := now.Add(time.Minute)
	for _, ti := range t.timeouts {
		if ti != nil && ti.deadline.Before(nextDeadline) {
			nextDeadline = ti.deadline
		}
	}

	// Stop existing timer and set it to nextDeadline
	if !t.timer.Stop() {
		select {
		case <-t.timer.C:
		default: // Should not block
		}

	}
	t.timer.Reset(nextDeadline.Sub(now))
	return nextDeadline
}

func (t *TimeoutTicker) run() {
	nextDeadline := t.resetTimer()
	for {
		select {
		// Scheduled timeout
		case ti := <-t.In:
			if ti.deadline.IsZero() {
				ti.deadline = time.Now().Add(ti.Duration)
			}
			t.timeouts = append(t.timeouts, ti)
			if ti.deadline.Before(nextDeadline) {
				nextDeadline = t.resetTimer()
			}

		// Timer has expired
		case now := <-t.timer.C:
			// Trigger expired timeouts
			for i := range t.timeouts {
				if t.timeouts[i] == nil ||
					t.timeouts[i].deadline.After(now) {
					continue
				}
				// FIXME: we should not block here
				select {
				case t.Out <- t.timeouts[i]:
				default:
					log.Println("failed to publish timeout", t.timeouts[i])
				}
				t.timeouts[i] = nil
			}
			// Clean up prefix of nil (i.e., trigged) timeouts
			for len(t.timeouts) > 0 && t.timeouts[0] == nil {
				t.timeouts = t.timeouts[1:]
			}
			nextDeadline = t.resetTimer()
		}
	}
}
