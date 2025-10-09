package cache

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWriteAndWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	cr := NewCachableResponse(rec)

	data := []byte("hello world")
	n, err := cr.Write(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(data) {
		t.Errorf("expected %d bytes written, got %d", len(data), n)
	}
	if got := cr.Body.String(); got != "hello world" {
		t.Errorf("expected Body buffer to contain 'hello world', got %q", got)
	}

	cr.WriteHeader(http.StatusNotFound)
	if cr.StatusCode != http.StatusNotFound {
		t.Errorf("expected StatusCode 404, got %d", cr.StatusCode)
	}
}

func TestIsCachableStatus(t *testing.T) {
	validStatuses := []int{
		http.StatusOK, http.StatusNonAuthoritativeInfo, http.StatusNoContent,
		http.StatusPartialContent, http.StatusMultipleChoices, http.StatusMovedPermanently,
		http.StatusNotFound, http.StatusMethodNotAllowed, http.StatusGone,
		http.StatusRequestURITooLong, http.StatusNotImplemented,
	}

	for _, code := range validStatuses {
		if !isCachableStatus(code) {
			t.Errorf("expected %d to be cachable", code)
		}
	}

	invalidStatuses := []int{201, 202, 500, 302, 403}
	for _, code := range invalidStatuses {
		if isCachableStatus(code) {
			t.Errorf("expected %d to NOT be cachable", code)
		}
	}
}

func TestCacheTTLWithCacheControl(t *testing.T) {
	rec := httptest.NewRecorder()
	cr := NewCachableResponse(rec)

	cr.Header().Set("Cache-Control", "max-age=60")
	ok, _ := cr.CacheTTL()

	if !ok {
		t.Fatalf("expected TTL to be present")
	}
}

func TestCacheTTLWithExpires(t *testing.T) {
	rec := httptest.NewRecorder()
	cr := NewCachableResponse(rec)

	expires := time.Now().Add(2 * time.Minute).UTC().Format(http.TimeFormat)
	cr.Header().Set("Expires", expires)

	ok, _ := cr.CacheTTL()
	if !ok {
		t.Fatalf("expected TTL to be present from Expires header")
	}
}

func TestCacheTTLNoCacheHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	cr := NewCachableResponse(rec)

	ok, ttl := cr.CacheTTL()
	if ok {
		t.Errorf("expected TTL to be absent, got ok=true")
	}
	if ttl != 0 {
		t.Errorf("expected TTL=0, got %v", ttl)
	}
}

func TestIsCachableConsideringAuth(t *testing.T) {
	rec := httptest.NewRecorder()
	cr := NewCachableResponse(rec)
	cr.StatusCode = http.StatusOK
	cr.Header().Set("Cache-Control", "max-age=120")

	if !cr.IsCachableConsideringAuth() {
		t.Errorf("expected IsCachableConsideringAuth() to return true for cachable header")
	}
}
