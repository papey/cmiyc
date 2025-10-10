package cache

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

var rawTestDate = "2025-12-12T12:12:12Z"
var testTime, _ = time.Parse(time.RFC3339, rawTestDate)
var ttl = 1 * time.Minute

var dummyEntry = Entry{
	StatusCode: 200,
	Body:       []byte("Hello, World!"),
	Header:     nil,
	ExpiresAt:  testTime.Add(ttl),
}

func newEmptyConfiguredCache(t *testing.T) *HttpCache {
	c := NewEmptyCache(1, 1) // 1 MiB max, 1 MiB per-entry max for test
	t.Cleanup(func() { c.Cleanup() })
	return c
}

func TestNewEmptyCache(t *testing.T) {
	c := newEmptyConfiguredCache(t)
	if c.Entries == nil {
		t.Error("HttpCache entries map is not initialized")
	}
}

func TestCacheSetAndGet(t *testing.T) {
	c := newEmptyConfiguredCache(t)

	rec := httptest.NewRecorder()
	resp := NewCachableResponse(rec)
	resp.WriteHeader(dummyEntry.StatusCode)
	resp.Write(dummyEntry.Body)

	req := &http.Request{URL: &url.URL{Path: "testKey"}}
	c.Set(req, resp, testTime.Add(ttl))

	if len(c.Entries) != 1 {
		t.Error("HttpCache entry was not set correctly")
	}

	entry := c.Entries["testKey"]
	if entry.StatusCode != dummyEntry.StatusCode {
		t.Error("HttpCache entry has incorrect StatusCode")
	}
	if !bytes.Equal(entry.Body, dummyEntry.Body) {
		t.Error("HttpCache entry has incorrect Body")
	}
	if !entry.ExpiresAt.Equal(dummyEntry.ExpiresAt) {
		t.Error("HttpCache entry has incorrect ExpiresAt")
	}

	got, exists := c.Get(req)
	if !exists {
		t.Error("HttpCache entry was not found by Get")
	}
	if got.StatusCode != dummyEntry.StatusCode {
		t.Error("HttpCache Get returned entry with incorrect StatusCode")
	}
	if !bytes.Equal(got.Body, dummyEntry.Body) {
		t.Error("HttpCache Get returned entry with incorrect Body")
	}
	if !got.ExpiresAt.Equal(dummyEntry.ExpiresAt) {
		t.Error("HttpCache Get returned entry with incorrect ExpiresAt")
	}
}

func TestHasFreeSpace(t *testing.T) {
	c := newEmptyConfiguredCache(t)

	size := MiBToBytes(1) // 1 MiB
	if !c.HasFreeSpace(size) {
		t.Error("Expected cache to have free space for 1 MiB")
	}

	c.CurrentSize = c.MaxSize
	if c.HasFreeSpace(1) {
		t.Error("Expected cache to be full and not have free space")
	}
}

func TestCanStore(t *testing.T) {
	c := newEmptyConfiguredCache(t)

	// entry within limits (1 MiB)
	entry := make([]byte, MiBToBytes(1))
	if !c.CanStore(entry) {
		t.Error("Expected CanStore to return true for entry within limits")
	}

	// entry exceeds per-entry max (1.5 MiB)
	tooLargeEntry := make([]byte, MiBToBytes(2))
	if c.CanStore(tooLargeEntry) {
		t.Error("Expected CanStore to return false for entry exceeding MaxEntrySize")
	}

	// cache almost full
	c.CurrentSize = c.MaxSize - 512 // bytes
	entrySmall := make([]byte, 1024)
	if c.CanStore(entrySmall) {
		t.Error("Expected CanStore to return false when cache does not have enough free space")
	}
}

func TestInvalidate(t *testing.T) {
	c := newEmptyConfiguredCache(t)

	rec := httptest.NewRecorder()
	resp := NewCachableResponse(rec)
	resp.WriteHeader(dummyEntry.StatusCode)
	resp.Write(dummyEntry.Body)

	req := &http.Request{URL: &url.URL{Path: "testKey"}}
	c.Set(req, resp, testTime.Add(ttl))

	c.Invalidate(req)
	if _, exists := c.Get(req); exists {
		t.Error("Expected entry to be removed after Invalidate")
	}
}

func TestServeIfPresent(t *testing.T) {
	c := newEmptyConfiguredCache(t)

	rec := httptest.NewRecorder()
	resp := NewCachableResponse(rec)
	resp.WriteHeader(dummyEntry.StatusCode)
	resp.Write(dummyEntry.Body)

	req := &http.Request{URL: &url.URL{Path: "testKey"}}
	c.Set(req, resp, testTime.Add(ttl))

	w := httptest.NewRecorder()
	found, err := c.ServeIfPresent(w, req)
	if !found || err != nil {
		t.Error("ServeIfPresent should serve the cached entry")
	}
	if body := w.Body.Bytes(); !bytes.Equal(body, dummyEntry.Body) {
		t.Error("ServeIfPresent wrote incorrect body")
	}
}
