package workload

import (
	"crypto/rand"
	"io"
	"time"

	"dslab.inf.usi.ch/tendermint/net"
	"dslab.inf.usi.ch/tendermint/types"
)

// Generator implements a simple workload generator.
//
// It produces random values to propose to consensus when it is invoked and
// delivers blocks committed by the consensus protocol.
type Generator struct {
	id     int
	config *Config
	log    net.Log

	values chan types.Value

	producedValues int
	randomizer     io.Reader
	deliveryQueue  chan *Delivery
}

// NewGenerator creates a new workload generator with a config.
func NewGenerator(id int, config *Config) *Generator {
	g := &Generator{
		id:     id,
		config: config,
		log:    config.Log,

		// FIXME: crypto/rand is slower than math/rand but produces
		// actually random values, not requiring a seed.
		randomizer:    rand.Reader,
		values:        make(chan types.Value, config.ValuesQueueSize), // unbuffered
		deliveryQueue: make(chan *Delivery, config.DeliveryQueueSize),
	}
	return g
}

// ProduceValues produces values to be proposed for consensus.
//
// The production of value can be stopped using the provided channel.
func (g *Generator) ProduceValues(stopCh chan struct{}) {
	g.log.Println("Started producing values")
	var active bool = true
	for active {
		value := make(types.Value, g.config.RandomValuesBytesSize)
		g.randomizer.Read(value)
		g.producedValues += 1

		// Wait for value to be proposed
		select {
		case g.values <- value:
		case <-stopCh:
			active = false
		}
	}
	g.log.Println("Done producing values")
}

// Run the workload generation.
//
// This methods returns when either of this conditions is observed:
// - the configured maxDuration is reached
// - 100 deliveries with empty values are processed
// - a delivery with epoch config.MaxEpoch - 1 is processed
func (g *Generator) Run(maxDuration time.Duration, stopCh chan struct{}) {
	var writer = NewWriter(g.config, g.id)
	g.log.Println("Workload generator started, maximum duration:", maxDuration)
	g.log.Println("Logging deliveries to file", writer)

	var emptyDeliveries int
	var logDeliveries = true
	var startTime = time.Now()
	var ticker = time.NewTicker(time.Second)

	var latency = new(Latency)
	var lastDeliveryEpoch int64
	var lastDeliveryTime time.Time
	var throughput = new(Throughput)
	for {
		var delivery *Delivery
		var tickTime time.Time
		select {
		case delivery = <-g.deliveryQueue:
			lastDeliveryTime = delivery.Time
			lastDeliveryEpoch = delivery.Epoch
		case tickTime = <-ticker.C:
		}

		if delivery == nil { // No new delivery in this round
			if tickTime.Sub(lastDeliveryTime) > 3*time.Second {
				g.log.Println("No delivery for", tickTime.Sub(lastDeliveryTime),
					"last epoch", lastDeliveryEpoch)
			}
			if tickTime.Sub(startTime) > maxDuration {
				break
			}

		} else if delivery.Size == 0 { // Empty decision, not logged
			emptyDeliveries += 1
			if emptyDeliveries >= 100 {
				break
			}
			if emptyDeliveries >= 10 {
				logDeliveries = false
			} else {
				g.log.Println("Empty value decided at height",
					delivery.Height, "epoch", delivery.Epoch)
			}

			if g.config.MaxEpoch > 0 && delivery.Epoch == g.config.MaxEpoch-1 {
				g.log.Println("Delivered empty value from last epoch",
					delivery.Epoch)
				break
			} else if g.config.MaxEpoch > 0 && delivery.Epoch >= g.config.MaxEpoch {
				g.log.Println("Should not be delivering empty value on epoch",
					delivery.Epoch, "MaxEpoch:", g.config.MaxEpoch)
			}

		} else if logDeliveries {
			submission := SubmissionFromValue(delivery.Value)
			// Record latency if the value was produced by this instance
			if submission.Sender == g.id {
				delivery.Submission = submission
				latency.Add(delivery.Latency())
			}
			throughput.Add(delivery.Time, 1, delivery.Size)
			writer.LogDelivery(delivery)

			if g.config.MaxEpoch > 0 && delivery.Epoch == g.config.MaxEpoch-1 {
				g.log.Println("Delivered value from last epoch",
					delivery.Epoch)
				break
			} else if g.config.MaxEpoch > 0 && delivery.Epoch >= g.config.MaxEpoch {
				g.log.Println("Should not be delivering value on epoch",
					delivery.Epoch, "MaxEpoch:", g.config.MaxEpoch)
			}
		}
		if lastDeliveryTime.Sub(startTime) > maxDuration {
			break
		}
	}

	//	close(stopCh)

	ticker.Stop()
	writer.Close()

	g.log.Println("Workload generator stopped after", time.Now().Sub(startTime),
		"and", emptyDeliveries, "empty values delivered")
	g.log.Printf("Throughput: %d values in %v: %f msg/s | %f bytes/s\n",
		throughput.Values(), throughput.Duration(),
		throughput.ValuesPerSecond(), throughput.BytesPerSecond())
	g.log.Printf("Latency: %d values: %v ms +- %v ms\n",
		latency.Count(), latency.Average().Milliseconds(),
		latency.Stdev().Milliseconds())
}

// NoopRoutine consumes the delivery queue without producing any data.
func (g *Generator) NoopRoutine(interval time.Duration) {
	g.log.Println("NoopRoutine started, expected duration:", interval)
	start := time.Now()
	leave := time.After(interval)
	var active = true
	var emptyValues int
	var nonEmptyValues int
	for active {
		select {
		case delivery := <-g.deliveryQueue:
			if delivery.Size > 0 {
				nonEmptyValues += 1
			} else {
				emptyValues += 1
			}

		case <-leave:
			active = false
		}
	}
	g.log.Println("Leaving NoopRoutine after", time.Now().Sub(start),
		"and", emptyValues, "empty values", nonEmptyValues, "not empty")
}
