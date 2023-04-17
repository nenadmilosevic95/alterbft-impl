package mempool

import (
	"dslab.inf.usi.ch/tendermint/types"
)

// Mempool stores values to be proposed for consensus.
type Mempool struct {
	config *Config

	// Recently seen values (ids)
	cache *LRUValueCache

	// Pending values
	pendingValues map[types.ValueKey]*entry
	valuesQueue   []*entry
}

// NewMempool creates a mempool.
func NewMempool(config *Config) *Mempool {
	if config == nil {
		config = DefaultConfig()
	}
	return &Mempool{
		config:        config,
		cache:         NewLRUValueCache(config.CacheSize),
		pendingValues: make(map[types.ValueKey]*entry),
	}
}

// Add a value to the mempool.
func (m *Mempool) Add(value types.Value) {
	key := value.Key()
	if !m.cache.Push(key) {
		return // Duplicated
	}
	entry := &entry{
		value:    value,
		proposed: false,
		decided:  false,
	}
	m.pendingValues[key] = entry
	m.valuesQueue = append(m.valuesQueue, entry)
}

// Decide a block of values, removing them from the pending list.
func (m *Mempool) Decide(value types.Value) {
	block := types.ParseBlock(value)
	for _, values := range block.Values {
		key := values.Key()
		// Mark value as seen
		m.cache.Push(key)
		// Mark value as decided
		if entry := m.pendingValues[key]; entry != nil {
			entry.decided = true
		}
	}

	// Clear decided values on the head of valuesQueue
	for len(m.valuesQueue) > 0 && m.valuesQueue[0].decided {
		delete(m.pendingValues, m.valuesQueue[0].value.Key())
		m.valuesQueue = m.valuesQueue[1:]
	}
}

// GetValue returns a block of pending values to be proposed.
func (m *Mempool) GetValue() types.Value {
	block := &types.Block{}
	var byteSize int
	for _, entry := range m.valuesQueue {
		if entry.proposed || entry.decided {
			continue
		}
		byteSize += len(entry.value)
		// Break if this value would exceed the maximum size
		if byteSize > m.config.BlockMaxBytes {
			break
		}
		block.Add(entry.value)
		entry.proposed = true
	}
	return block.ToValue()
}

// An entry in the pending values queue.
type entry struct {
	value    types.Value
	proposed bool
	decided  bool
}
