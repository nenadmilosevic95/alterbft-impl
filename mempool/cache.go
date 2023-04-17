// Originally COPIED from github.com/tendermint/tendermint/internal/mempool package
package mempool

import (
	"container/list"

	"dslab.inf.usi.ch/tendermint/types"
)

// LRUValueCache maintains a LRU cache of value ids.
type LRUValueCache struct {
	size     int
	cacheMap map[types.ValueKey]*list.Element
	list     *list.List
}

func NewLRUValueCache(cacheSize int) *LRUValueCache {
	return &LRUValueCache{
		size:     cacheSize,
		cacheMap: make(map[types.ValueKey]*list.Element, cacheSize),
		list:     list.New(),
	}
}

// GetList returns the underlying linked-list that backs the LRU cache. Note,
// this should be used for testing purposes only!
func (c *LRUValueCache) GetList() *list.List {
	return c.list
}

func (c *LRUValueCache) Reset() {
	c.cacheMap = make(map[types.ValueKey]*list.Element, c.size)
	c.list.Init()
}

func (c *LRUValueCache) Push(key types.ValueKey) bool {
	moved, ok := c.cacheMap[key]
	if ok {
		c.list.MoveToBack(moved)
		return false
	}

	if c.list.Len() >= c.size {
		front := c.list.Front()
		if front != nil {
			frontKey := front.Value.(types.ValueKey)
			delete(c.cacheMap, frontKey)
			c.list.Remove(front)
		}
	}

	e := c.list.PushBack(key)
	c.cacheMap[key] = e

	return true
}

func (c *LRUValueCache) Remove(key types.ValueKey) {
	e := c.cacheMap[key]
	delete(c.cacheMap, key)

	if e != nil {
		c.list.Remove(e)
	}
}
