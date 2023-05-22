package workload

import (
	"fmt"

	"dslab.inf.usi.ch/tendermint/net"
)

// MinValuesBytesSize is the minumum size in bytes of produced values.
const MinValuesBytesSize = 28

// Config stores configurations for the workload generation.
type Config struct {
	Log net.Log

	// DeliveryQueueSize is the size of the queue storing data of delivered
	// values. Once this queue gets full, new delivery data is dropped.
	DeliveryQueueSize int

	// ValuesQueueSize is the size of the queue storing data of generated
	// values.
	ValuesQueueSize int

	// LogBufferSize is the size in bytes for the buffered log writer.
	// Data is actually logged to disk when the writer buffer is full.
	LogBufferSize int

	// LogDirectory is an existing directory to write log files.
	// The directory *must* exist.
	LogDirectory string

	// RandomValuesBytesSize is the size in bytes of produced values.
	RandomValuesBytesSize int

	// WarmupValuesCount is the number of produced values in the warm-up.
	// During the warm-up, produced values have MinValuesBytesSize bytes.
	WarmupValuesCount int

	// When set to a positive value, determines the maximum (exclusive)
	// epoch considered for load generation and performace computation.
	MaxEpoch int64
}

// DefaultConfig returns a config with default values.
func DefaultConfig() *Config {
	return &Config{
		DeliveryQueueSize:     128,
		ValuesQueueSize:       128,
		RandomValuesBytesSize: MinValuesBytesSize,

		LogBufferSize: 1024 * 1024,
		LogDirectory:  ".",

		WarmupValuesCount: 0, // No warm-up phase
		MaxEpoch:          0, // No maximum epoch
	}
}

// Validate checks the validity of the configurations.
func (c *Config) Validate() error {
	if c.RandomValuesBytesSize < MinValuesBytesSize {
		return fmt.Errorf("Invalid RandomValuesBytesSize %d < %d",
			c.RandomValuesBytesSize, MinValuesBytesSize)
	}
	return nil
}
