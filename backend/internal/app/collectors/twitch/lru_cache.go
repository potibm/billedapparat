package twitch

import (
	"container/list"
	"sync"
)

type entry struct {
	key   string
	value string
}

type CacheStats struct {
	Hits      int64 `json:"hits"`
	Misses    int64 `json:"misses"`
	Evictions int64 `json:"evictions"`
	Size      int   `json:"size"`
}

type LRUCache struct {
	capacity int
	mu       sync.Mutex
	items    map[string]*list.Element
	evict    *list.List

	hits      int64
	misses    int64
	evictions int64
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		evict:    list.New(),
	}
}

func (c *LRUCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ent, ok := c.items[key]; ok {
		c.evict.MoveToFront(ent)
		c.hits++

		return ent.Value.(*entry).value, true
	}

	c.misses++

	return "", false
}

func (c *LRUCache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ent, ok := c.items[key]; ok {
		c.evict.MoveToFront(ent)
		ent.Value.(*entry).value = value

		return
	}

	ent := &entry{key, value}
	element := c.evict.PushFront(ent)
	c.items[key] = element

	if c.evict.Len() > c.capacity {
		c.removeOldest()
	}
}

func (c *LRUCache) Stats() CacheStats {
	c.mu.Lock()
	defer c.mu.Unlock()

	return CacheStats{
		Hits:      c.hits,
		Misses:    c.misses,
		Evictions: c.evictions,
		Size:      c.evict.Len(),
	}
}

func (c *LRUCache) removeOldest() {
	ent := c.evict.Back()
	if ent != nil {
		c.evict.Remove(ent)
		kv := ent.Value.(*entry)
		delete(c.items, kv.key)
		c.evictions++
	}
}
