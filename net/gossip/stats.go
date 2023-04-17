package gossip

import (
	"fmt"
)

type CacheStats struct {
	Added      int
	Duplicated int
}

func (c CacheStats) String() string {
	total := c.Added + c.Duplicated
	return fmt.Sprintf("%d, %d, %d, %.1f%%",
		total, c.Added, c.Duplicated,
		float64(c.Duplicated)/float64(total)*100.0)
}

type QueueStats struct {
	Low, High, Full int // Queue occupation
}

func (q QueueStats) Total() int {
	return q.Low + q.High + q.Full
}

func (q QueueStats) String() string {
	return fmt.Sprintf("%d, %d, %d, %d",
		q.Total(), q.Low, q.High, q.Full)
}

func (q *QueueStats) Sum(qq *QueueStats) {
	q.Low += qq.Low
	q.High += qq.High
	q.Full += qq.Full
}

type Stats struct {
	Cache       CacheStats
	BQueue      QueueStats
	DQueue      QueueStats
	RQueue      QueueStats
	SQueues     QueueStats
	Validator   ValidatorStats
	MessageLoss MessageLossStats
}

type ValidatorStats struct {
	Enqueued int
	Filtered int
}

func (v ValidatorStats) String() string {
	ratio := float64(v.Filtered) / float64(v.Enqueued) * 100.0
	return fmt.Sprintf("%d, %d, %.1f%%", v.Enqueued, v.Filtered, ratio)

}

type MessageLossStats struct {
	Received int
	Lost     int
}

func (m MessageLossStats) String() string {
	ratio := float64(m.Lost) / float64(m.Received) * 100.0
	return fmt.Sprintf("%d, %d, %.1f%%", m.Received, m.Lost, ratio)

}
