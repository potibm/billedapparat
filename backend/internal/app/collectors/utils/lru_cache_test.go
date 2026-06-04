package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLRUCache(t *testing.T) {
	c := NewLRUCache[string](2)
	assert.NotNil(t, c)
	assert.Equal(t, CacheStats{Size: 0}, c.Stats())
}

func TestLRUCache_GetMiss(t *testing.T) {
	c := NewLRUCache[string](2)

	val, ok := c.Get("missing")
	assert.False(t, ok)
	assert.Equal(t, "", val)
	assert.Equal(t, int64(1), c.Stats().Misses)
}

func TestLRUCache_SetAndGet(t *testing.T) {
	c := NewLRUCache[string](2)

	c.Set("key1", "value1")
	val, ok := c.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", val)

	stats := c.Stats()
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)
	assert.Equal(t, 1, stats.Size)
	assert.Equal(t, int64(0), stats.Evictions)
}

func TestLRUCache_UpdateExistingKey(t *testing.T) {
	c := NewLRUCache[string](2)

	c.Set("key1", "value1")
	c.Set("key1", "value2")

	val, ok := c.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value2", val)
	assert.Equal(t, 1, c.Stats().Size)
}

func TestLRUCache_Eviction(t *testing.T) {
	c := NewLRUCache[string](2)

	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Set("key3", "value3")

	stats := c.Stats()
	assert.Equal(t, 2, stats.Size)
	assert.Equal(t, int64(1), stats.Evictions)

	_, ok := c.Get("key1")
	assert.False(t, ok)

	val, ok := c.Get("key2")
	assert.True(t, ok)
	assert.Equal(t, "value2", val)

	val, ok = c.Get("key3")
	assert.True(t, ok)
	assert.Equal(t, "value3", val)
}

func TestLRUCache_LRUOrder(t *testing.T) {
	c := NewLRUCache[string](2)

	c.Set("key1", "value1")
	c.Set("key2", "value2")

	// Access key1 to make it recently used
	c.Get("key1")

	// Add key3, should evict key2 (least recently used)
	c.Set("key3", "value3")

	_, ok := c.Get("key2")
	assert.False(t, ok)

	val, ok := c.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", val)

	val, ok = c.Get("key3")
	assert.True(t, ok)
	assert.Equal(t, "value3", val)
}

func TestLRUCache_SetExistingMovesToFront(t *testing.T) {
	c := NewLRUCache[string](2)

	c.Set("key1", "value1")
	c.Set("key2", "value2")

	// Updating key1 should move it to front
	c.Set("key1", "updated")

	// Adding key3 should now evict key2
	c.Set("key3", "value3")

	_, ok := c.Get("key2")
	assert.False(t, ok)

	val, ok := c.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "updated", val)
}

func TestLRUCache_Stats(t *testing.T) {
	c := NewLRUCache[string](1)

	c.Set("key1", "value1")
	c.Get("key1")
	c.Get("missing")
	c.Set("key2", "value2") // evicts key1

	stats := c.Stats()
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
	assert.Equal(t, int64(1), stats.Evictions)
	assert.Equal(t, 1, stats.Size)
}

func TestLRUCache_ConcurrentAccess(t *testing.T) {
	c := NewLRUCache[string](100)

	// Just ensure no race detector complaints
	for i := 0; i < 100; i++ {
		go func(i int) {
			c.Set("key", "value")
			c.Get("key")
			c.Stats()
		}(i)
	}
}
