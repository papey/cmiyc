package cache

import (
	"net/http"
	"sync"
	"time"
)

type HttpCache struct {
	MaxSize      int // in bytes
	MaxEntrySize int // in bytes
	Entries      map[Key]Entry
	CurrentSize  int // in bytes
	stop         chan struct{}
	sync.RWMutex
}

type Key = string

func KeyFrom(request *http.Request) Key {
	return request.URL.String()
}

func NewCache(entries map[Key]Entry, maxSizeMiB int, maxEntrySizeMiB int) *HttpCache {
	cache := &HttpCache{
		Entries:      entries,
		MaxSize:      MiBToBytes(maxSizeMiB),
		MaxEntrySize: MiBToBytes(maxEntrySizeMiB),
		stop:         make(chan struct{}),
	}

	go cache.autoCleanup(3 * time.Minute)

	return cache
}

func NewEmptyCache(maxSizeMiB, maxEntrySizeMiB int) *HttpCache {
	emptyEntries := make(map[Key]Entry)
	return NewCache(emptyEntries, maxSizeMiB, maxEntrySizeMiB)
}

func (c *HttpCache) Cleanup() {
	close(c.stop)
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

	return entry.Clone(), true
}

func (c *HttpCache) Set(r *http.Request, resp *CachableResponse, expiresAt time.Time) {
	key, entry := NewEntry(r, resp, expiresAt)
	entrySize := len(entry.Body)

	if !c.CanStore(entry.Body) {
		return
	}

	c.Lock()
	defer c.Unlock()

	if old, exists := c.Entries[key]; exists {
		c.CurrentSize -= len(old.Body)
	}

	if c.CurrentSize+entrySize > c.MaxSize {
		return
	}

	c.Entries[key] = entry
	c.CurrentSize += entrySize
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
	entry, exists := c.Entries[key]
	if !exists {
		c.Unlock()
		return
	}

	delete(c.Entries, key)
	c.CurrentSize -= len(entry.Body)
	if c.CurrentSize < 0 {
		c.CurrentSize = 0
	}
	c.Unlock()
}

func (c *HttpCache) autoCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.Lock()
			for key, entry := range c.Entries {
				if entry.IsExpired() {
					delete(c.Entries, key)
					c.CurrentSize -= len(entry.Body)
				}
			}
			if c.CurrentSize < 0 {
				c.CurrentSize = 0
			}
			c.Unlock()

		case <-c.stop:
			return
		}
	}
}

func (c *HttpCache) HasFreeSpace(size int) bool {
	c.RLock()
	defer c.RUnlock()
	return c.CurrentSize+size <= c.MaxSize
}

func (c *HttpCache) CanStore(entry []byte) bool {
	entrySize := len(entry)
	if entrySize > c.MaxEntrySize {
		return false
	}
	return c.HasFreeSpace(entrySize)
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

func MiBToBytes(mib int) int {
	return mib * 1024 * 1024
}
