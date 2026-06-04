package utils

import (
	"container/list"
	"sync"
)

type entry[V any] struct {
	key   string
	value V
}

type CacheStats struct {
	Hits      int64 `json:"hits"`
	Misses    int64 `json:"misses"`
	Evictions int64 `json:"evictions"`
	Size      int   `json:"size"`
}

type LRUCache[V any] struct {
	capacity int
	mu       sync.Mutex
	items    map[string]*list.Element
	evict    *list.List

	hits      int64
	misses    int64
	evictions int64
}

func NewLRUCache[V any](capacity int) *LRUCache[V] {
	return &LRUCache[V]{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		evict:    list.New(),
	}
}

func (c *LRUCache[V]) Get(key string) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ent, ok := c.items[key]; ok {
		c.evict.MoveToFront(ent)
		c.hits++

		return ent.Value.(*entry[V]).value, true
	}

	c.misses++

	var zero V
	return zero, false
}

func (c *LRUCache[V]) Set(key string, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ent, ok := c.items[key]; ok {
		c.evict.MoveToFront(ent)
		ent.Value.(*entry[V]).value = value

		return
	}

	ent := &entry[V]{key, value}
	element := c.evict.PushFront(ent)
	c.items[key] = element

	if c.evict.Len() > c.capacity {
		c.removeOldest()
	}
}

func (c *LRUCache[V]) Stats() CacheStats {
	c.mu.Lock()
	defer c.mu.Unlock()

	return CacheStats{
		Hits:      c.hits,
		Misses:    c.misses,
		Evictions: c.evictions,
		Size:      c.evict.Len(),
	}
}

func (c *LRUCache[V]) removeOldest() {
	ent := c.evict.Back()
	if ent != nil {
		c.evict.Remove(ent)
		kv := ent.Value.(*entry[V])
		delete(c.items, kv.key)
		c.evictions++
	}
}
