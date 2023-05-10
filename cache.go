package gocache

import (
	"sync"

	"github.com/einsier/go-cache/lru"
)

// cache is a wrapper of lru.Cache that is safe for concurrent access.
type cache struct {
	mu         sync.Mutex // add and get are both have write operation
	lru        *lru.Cache
	cacheBytes int64
}

// add adds a value to the cache.
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

// get gets a value from the cache.
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
