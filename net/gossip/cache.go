package gossip

import (
	"github.com/hashicorp/golang-lru/simplelru"
)

var LRUCacheSize int = 1024

type MessageCache struct {
	cache simplelru.LRUCache

	stats CacheStats
}

func NewMessageCache() *MessageCache {
	cache, _ := simplelru.NewLRU(LRUCacheSize, nil)
	return &MessageCache{
		cache: cache,
	}
}

func (c *MessageCache) Add(messageID interface{}) bool {
	c.stats.Added += 1
	return c.cache.Add(messageID, nil)
}

func (c *MessageCache) Contains(messageID interface{}) bool {
	contains := c.cache.Contains(messageID)
	if contains {
		c.stats.Duplicated += 1
	}
	return contains
}

func (c *MessageCache) Stats() CacheStats {
	return c.stats
}
