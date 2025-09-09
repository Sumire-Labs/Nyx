package utils

import (
	"container/list"
	"sync"
	"time"
)

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Key       string
	Value     interface{}
	ExpiresAt time.Time
	Element   *list.Element
}

// LRUCache implements a thread-safe LRU cache with TTL
type LRUCache struct {
	capacity int
	ttl      time.Duration
	items    map[string]*CacheItem
	order    *list.List
	mutex    sync.RWMutex
	stats    CacheStats
}

// CacheStats tracks cache performance metrics
type CacheStats struct {
	Hits        int64
	Misses      int64
	Evictions   int64
	Expirations int64
	Size        int
}

// NewLRUCache creates a new LRU cache with specified capacity and TTL
func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		ttl:      ttl,
		items:    make(map[string]*CacheItem),
		order:    list.New(),
	}
}

// Get retrieves an item from the cache
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	item, exists := c.items[key]
	if !exists {
		c.stats.Misses++
		return nil, false
	}

	// Check if item has expired
	if time.Now().After(item.ExpiresAt) {
		c.removeItem(item)
		c.stats.Expirations++
		c.stats.Misses++
		return nil, false
	}

	// Move to front (most recently used)
	c.order.MoveToFront(item.Element)
	c.stats.Hits++
	return item.Value, true
}

// Set adds or updates an item in the cache
func (c *LRUCache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Update existing item
	if item, exists := c.items[key]; exists {
		item.Value = value
		item.ExpiresAt = time.Now().Add(c.ttl)
		c.order.MoveToFront(item.Element)
		return
	}

	// Add new item
	item := &CacheItem{
		Key:       key,
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
	item.Element = c.order.PushFront(item)
	c.items[key] = item
	c.stats.Size++

	// Evict least recently used if over capacity
	if len(c.items) > c.capacity {
		c.evictLRU()
	}
}

// Delete removes an item from the cache
func (c *LRUCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if item, exists := c.items[key]; exists {
		c.removeItem(item)
	}
}

// Clear removes all items from the cache
func (c *LRUCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*CacheItem)
	c.order = list.New()
	c.stats.Size = 0
}

// GetStats returns current cache statistics
func (c *LRUCache) GetStats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	stats := c.stats
	stats.Size = len(c.items)
	return stats
}

// Cleanup removes expired items from the cache
func (c *LRUCache) Cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	var toRemove []*CacheItem

	// Collect expired items
	for _, item := range c.items {
		if now.After(item.ExpiresAt) {
			toRemove = append(toRemove, item)
		}
	}

	// Remove expired items
	for _, item := range toRemove {
		c.removeItem(item)
		c.stats.Expirations++
	}
}

// StartCleanupRoutine starts a background goroutine to clean up expired items
func (c *LRUCache) StartCleanupRoutine(interval time.Duration) chan<- bool {
	stop := make(chan bool, 1)
	
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.Cleanup()
			case <-stop:
				return
			}
		}
	}()

	return stop
}

// evictLRU removes the least recently used item
func (c *LRUCache) evictLRU() {
	if c.order.Len() == 0 {
		return
	}

	oldest := c.order.Back()
	if oldest != nil {
		item := oldest.Value.(*CacheItem)
		c.removeItem(item)
		c.stats.Evictions++
	}
}

// removeItem removes an item from both the map and the list
func (c *LRUCache) removeItem(item *CacheItem) {
	delete(c.items, item.Key)
	c.order.Remove(item.Element)
	c.stats.Size--
}

// HitRate returns the cache hit rate (0.0 - 1.0)
func (c *LRUCache) HitRate() float64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	total := c.stats.Hits + c.stats.Misses
	if total == 0 {
		return 0.0
	}
	return float64(c.stats.Hits) / float64(total)
}