package cache

import (
	"sync"
	"time"
)

// Item represents a cached item
type Item struct {
	Value      interface{}
	Expiration time.Time
}

// Cache is a simple in-memory cache with TTL support
type Cache struct {
	items map[string]*Item
	mu    sync.RWMutex
	ttl   time.Duration
}

// New creates a new cache with the specified TTL
func New(ttl time.Duration) *Cache {
	c := &Cache{
		items: make(map[string]*Item),
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go c.cleanup()

	return c
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if item has expired
	if time.Now().After(item.Expiration) {
		return nil, false
	}

	return item.Value, true
}

// Set adds an item to the cache with the default TTL
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.ttl)
}

// SetWithTTL adds an item to the cache with a custom TTL
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &Item{
		Value:      value,
		Expiration: time.Now().Add(ttl),
	}
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*Item)
}

// Size returns the number of items in the cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// cleanup removes expired items periodically
func (c *Cache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.Expiration) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// GetOrSet retrieves an item from cache, or sets it if not present
func (c *Cache) GetOrSet(key string, fn func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if val, ok := c.Get(key); ok {
		return val, nil
	}

	// Not in cache, compute value
	val, err := fn()
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.Set(key, val)
	return val, nil
}