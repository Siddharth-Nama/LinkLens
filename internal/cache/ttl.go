package cache

import (
	"sync"
	"time"
)

type entry[T any] struct {
	value     T
	expiresAt time.Time
}

type TTL[T any] struct {
	mu    sync.RWMutex
	ttl   time.Duration
	items map[string]entry[T]
}

func NewTTL[T any](ttl time.Duration) *TTL[T] {
	return &TTL[T]{
		ttl:   ttl,
		items: make(map[string]entry[T]),
	}
}

func (c *TTL[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		var zero T
		return zero, false
	}
	if time.Now().After(item.expiresAt) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		var zero T
		return zero, false
	}
	return item.value, true
}

func (c *TTL[T]) Set(key string, value T) {
	c.mu.Lock()
	c.items[key] = entry[T]{value: value, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}
