package dota

import (
	"encoding/json"
	"sync"
	"time"
)

const (
	playerCacheTTL        = 2 * time.Minute
	recentMatchesCacheTTL = 45 * time.Second
	matchCacheTTL         = 10 * time.Minute
)

type cacheEntry struct {
	data    []byte
	expires time.Time
}

// Entries are immutable JSON snapshots, so callers can sort slices and change
// returned models without changing the next caller's result.
type responseCache struct {
	mu                   sync.Mutex
	entries              map[string]cacheEntry
	bytes                int
	maxEntries, maxBytes int
	now                  func() time.Time
}

func newResponseCache(maxEntries, maxBytes int, now func() time.Time) *responseCache {
	return &responseCache{entries: make(map[string]cacheEntry), maxEntries: maxEntries, maxBytes: maxBytes, now: now}
}

func (c *responseCache) get(key string) ([]byte, bool) {
	if c == nil {
		return nil, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if !c.now().Before(entry.expires) {
		c.remove(key)
		return nil, false
	}
	return entry.data, true
}

// remove must be called with the mutex held.
func (c *responseCache) remove(key string) {
	c.bytes -= len(c.entries[key].data)
	delete(c.entries, key)
}

func (c *responseCache) put(key string, data []byte, ttl time.Duration) {
	if c == nil || ttl <= 0 || c.maxEntries <= 0 || len(data) > c.maxBytes {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	for existing, entry := range c.entries {
		if existing == key || !now.Before(entry.expires) {
			c.remove(existing)
		}
	}
	for len(c.entries) >= c.maxEntries || c.bytes+len(data) > c.maxBytes {
		var oldest string
		var expiry time.Time
		for existing, entry := range c.entries {
			if expiry.IsZero() || entry.expires.Before(expiry) {
				oldest, expiry = existing, entry.expires
			}
		}
		c.remove(oldest)
	}
	c.entries[key] = cacheEntry{data: data, expires: now.Add(ttl)}
	c.bytes += len(data)
}

func cachedGet[T any](c *Client, path string, ttl time.Duration, validate func(*T) error) (*T, error) {
	var result T
	if data, ok := c.cache.get(path); ok {
		if err := json.Unmarshal(data, &result); err == nil {
			return &result, nil
		}
	}
	if err := c.get(path, &result); err != nil {
		return nil, err
	}
	if validate != nil {
		if err := validate(&result); err != nil {
			return nil, err
		}
	}
	if c.cache != nil {
		if data, err := json.Marshal(result); err == nil {
			c.cache.put(path, data, ttl)
		}
	}
	return &result, nil
}
