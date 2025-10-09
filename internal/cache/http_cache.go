package cache

import (
	"net/http"
	"sync"
	"time"
)

type HttpCache struct {
	Entries map[Key]Entry
	sync.RWMutex
}

type Key = string

func KeyFrom(request *http.Request) Key {
	return request.URL.String()
}

func NewCache(entries map[Key]Entry) *HttpCache {
	cache := &HttpCache{
		Entries: entries,
	}

	go cache.autoCleanup(3 * time.Minute)

	return cache
}

func NewEmptyCache() *HttpCache {
	emtpyEntries := make(map[Key]Entry)
	return NewCache(emtpyEntries)
}

func (c *HttpCache) Get(r *http.Request) (Entry, bool) {
	key := KeyFrom(r)
	c.RLock()
	entry, exists := c.Entries[key]
	c.RUnlock()

	if !exists {
		return Entry{}, false
	}

	if entry.IsExpired() {
		c.delete(key)
		return Entry{}, false
	}

	return entry, exists
}

func (c *HttpCache) Set(r *http.Request, resp *CachableResponse, expiresAt time.Time) {
	key, entry := NewEntry(r, resp, expiresAt)
	c.Lock()
	c.Entries[key] = entry
	c.Unlock()
}

func (c *HttpCache) Invalidate(r *http.Request) {
	key := KeyFrom(r)
	c.delete(key)
}

func (c *HttpCache) ServeIfPresent(w http.ResponseWriter, r *http.Request) (bool, error) {
	entry, found := c.Get(r)
	if !found {
		return false, nil
	}

	return true, entry.WriteResponse(w, r)
}

func (c *HttpCache) delete(key string) {
	c.Lock()
	delete(c.Entries, key)
	c.Unlock()
}

func cloneHeader(h http.Header) http.Header {
	out := make(http.Header, len(h))
	for k, vv := range h {
		copied := make([]string, len(vv))
		copy(copied, vv)
		out[k] = copied
	}
	return out
}

func (c *HttpCache) autoCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		c.Lock()
		for key, entry := range c.Entries {
			if entry.IsExpired() {
				c.delete(key)
			}
		}
		c.Unlock()
	}
}
