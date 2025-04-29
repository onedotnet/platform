package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// MockCache is a simple in-memory implementation of Cache for testing
type MockCache struct {
	data  map[string][]byte
	mutex sync.RWMutex
}

// NewMockCache creates a new mock cache for testing
func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string][]byte),
	}
}

// Set stores a value in the mock cache
func (m *MockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	m.data[key] = data
	return nil
}

// Get retrieves a value from the mock cache
func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	data, ok := m.data[key]
	if !ok {
		return ErrKeyNotFound{key: key}
	}

	return json.Unmarshal(data, dest)
}

// Delete removes an item from the mock cache
func (m *MockCache) Delete(ctx context.Context, key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.data, key)
	return nil
}

// Clear removes all items from the mock cache
func (m *MockCache) Clear(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.data = make(map[string][]byte)
	return nil
}

// Close is a no-op for the mock cache
func (m *MockCache) Close() error {
	return nil
}

// ErrKeyNotFound is returned when a key is not found in the cache
type ErrKeyNotFound struct {
	key string
}

func (e ErrKeyNotFound) Error() string {
	return "key not found: " + e.key
}
