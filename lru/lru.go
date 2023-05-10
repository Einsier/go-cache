package lru

import "container/list"

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}

// Cache is a LRU cache. It is not safe for concurrent access.
type Cache struct {
	maxBytes int64                    // max memory can be used
	nbytes   int64                    // current memory used
	ll       *list.List               // double linked list
	cache    map[string]*list.Element // map key to list.Element
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

// New is the constructor of Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get look up a key
func (c *Cache) Get(key string) (value Value, ok bool) {
	if e, ok := c.cache[key]; ok {
		// move to the front of list
		c.ll.MoveToFront(e)
		kv := e.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest remove the oldest item
func (c *Cache) RemoveOldest() {
	// get the last element
	e := c.ll.Back()
	if e != nil {
		c.ll.Remove(e)
		kv := e.Value.(*entry)
		// delete from map
		delete(c.cache, kv.key)
		// update memory usage
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		// callback
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add add a key value pair
func (c *Cache) Add(key string, value Value) {
	if e, ok := c.cache[key]; ok {
		// update value
		c.ll.MoveToFront(e)
		kv := e.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		// add new key value pair
		e := c.ll.PushFront(&entry{key, value})
		c.cache[key] = e
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	// remove oldest if necessary
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Len return the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}
