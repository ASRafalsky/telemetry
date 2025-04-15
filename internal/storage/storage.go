package storage

import (
	"context"
	"sync"
)

// MemStorage non-blocking kv storage.
type MemStorage[K comparable, V any] struct {
	mx      sync.RWMutex
	storage map[K]V
}

// New creates new MemStorage unit.
func New[K comparable, V any]() *MemStorage[K, V] {
	m := MemStorage[K, V]{
		storage: make(map[K]V),
	}
	return &m
}

// Set sets value with key.
func (m *MemStorage[K, V]) Set(k K, v V) {
	m.mx.Lock()
	defer m.mx.Unlock()

	m.storage[k] = v
}

// Get returns value and true from the MemStorage if it exists, or empty value and false.
func (m *MemStorage[K, V]) Get(k K) (V, bool) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	if v, ok := m.storage[k]; ok {
		return v, true
	}
	var v V
	return v, false
}

// Delete deletes entry by the key.
func (m *MemStorage[K, V]) Delete(k K) {
	m.mx.Lock()
	defer m.mx.Unlock()

	delete(m.storage, k)
}

// Size returns number of items in the MemStorage.
func (m *MemStorage[K, V]) Size() int {
	return len(m.storage)
}

func (m *MemStorage[K, V]) ForEach(ctx context.Context, fn func(k K, v V) error) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	for k, v := range m.storage {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}
