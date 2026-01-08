package gateway

import (
	"time"

	"github.com/patrickmn/go-cache"
)

// Cache handles local in-memory caching for the gateway
type Cache struct {
	store *cache.Cache
}

// NewCache creates a new gateway cache
func NewCache(defaultExpiration, cleanupInterval time.Duration) *Cache {
	return &Cache{
		store: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Set stores a value in the cache
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	c.store.Set(key, value, duration)
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	return c.store.Get(key)
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.store.Delete(key)
}

// Flush clears all items from the cache
func (c *Cache) Flush() {
	c.store.Flush()
}
