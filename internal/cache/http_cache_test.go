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

func TestNewEmptyCache(t *testing.T) {
	c := NewEmptyCache()

	if c.Entries == nil {
		t.Error("HttpCache entries map is not initialized")
	}
}

func TestCacheSet(t *testing.T) {
	c := NewEmptyCache()

	rec := httptest.NewRecorder()

	resp := NewCachableResponse(rec)

	resp.WriteHeader(dummyEntry.StatusCode)
	resp.Write(dummyEntry.Body)

	c.Set(&http.Request{URL: &url.URL{Path: "testKey"}}, resp, testTime.Add(ttl))

	if len(c.Entries) != 1 {
		t.Error("HttpCache entry was not set correctly")
	}

	if c.Entries["testKey"].StatusCode != dummyEntry.StatusCode {
		t.Error("HttpCache entry has incorrect StatusCode")
	}

	if !bytes.Equal(c.Entries["testKey"].Body, dummyEntry.Body) {
		t.Error("HttpCache entry has incorrect Body")
	}

	if c.Entries["testKey"].ExpiresAt != dummyEntry.ExpiresAt {
		t.Error("HttpCache entry has incorrect ExpiresAt")
	}
}

func TestCacheGet(t *testing.T) {
	c := NewCache(map[Key]Entry{"testKey": dummyEntry})

	e, exists := c.Get(&http.Request{URL: &url.URL{Path: "testKey"}})
	if !exists {
		t.Error("HttpCache entry was not found")
	}

	if e.StatusCode != dummyEntry.StatusCode {
		t.Error("HttpCache entry has incorrect StatusCode")
	}

	if !bytes.Equal(e.Body, dummyEntry.Body) {
		t.Error("HttpCache entry has incorrect Body")
	}

	if e.ExpiresAt != dummyEntry.ExpiresAt {
		t.Error("HttpCache entry has incorrect ExpiresAt")
	}
}
