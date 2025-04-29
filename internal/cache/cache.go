package cache

import (
	"context"
	"time"
)

// Cacheable is an interface that models can implement to be cached
type Cacheable interface {
	// CacheKey returns the key used to store the object in cache
	CacheKey() string
}

// Cache defines the interface for cache implementations
type Cache interface {
	// Set stores a value in the cache with the given key and optional expiration
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Get retrieves a value from the cache using the given key
	Get(ctx context.Context, key string, dest interface{}) error

	// Delete removes an item from the cache using the given key
	Delete(ctx context.Context, key string) error

	// Clear removes all items from the cache
	Clear(ctx context.Context) error

	// Close closes the cache connection
	Close() error
}
