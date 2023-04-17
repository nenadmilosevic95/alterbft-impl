package mempool

// Config defines the configuration for the mempool.
type Config struct {
	// Size, in values, of the cache used to filter duplicated values.
	CacheSize int

	// Maximum size, in bytes, of the values included in a block.
	BlockMaxBytes int
}

// DefaultConfig returns a default configuration for the mempool.
func DefaultConfig() *Config {
	return &Config{
		CacheSize:     10000,
		BlockMaxBytes: 1024 * 1024, // 1MB
	}
}
