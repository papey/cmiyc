package cache

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewEntry(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	rec := httptest.NewRecorder()

	// Use the real CachableResponse
	resp := NewCachableResponse(rec)
	resp.WriteHeader(200)
	_, _ = resp.Write([]byte("Hello"))
	resp.Header().Set("Content-Type", "text/plain")

	expiration := time.Now().Add(time.Minute)
	key, entry := NewEntry(req, resp, expiration)

	if entry.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", entry.StatusCode)
	}

	if string(entry.Body) != "Hello" {
		t.Errorf("expected body 'Hello', got %s", string(entry.Body))
	}

	if entry.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("expected header Content-Type=text/plain, got %s", entry.Header.Get("Content-Type"))
	}

	if entry.Header.Get("X-Cache") != "HIT" {
		t.Errorf("expected header X-Cache=HIT, got %s", entry.Header.Get("X-Cache"))
	}

	if !entry.ExpiresAt.Equal(expiration) {
		t.Errorf("expected expiration %v, got %v", expiration, entry.ExpiresAt)
	}

	if key == "" {
		t.Error("expected non-empty key")
	}
}

func TestWriteResponse(t *testing.T) {
	entry := Entry{
		StatusCode: 200,
		Body:       []byte("Hello World"),
		Header:     http.Header{"X-Test": []string{"value"}, "X-Cache": []string{"HIT"}},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	err := entry.WriteResponse(rec, req)
	if err != nil {
		t.Fatalf("WriteResponse returned error: %v", err)
	}

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", res.StatusCode)
	}

	if res.Header.Get("X-Test") != "value" {
		t.Errorf("expected header X-Test=value, got %s", res.Header.Get("X-Test"))
	}

	if res.Header.Get("X-Cache") != "HIT" {
		t.Errorf("expected header X-Cache=HIT, got %s", res.Header.Get("X-Cache"))
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	if buf.String() != "Hello World" {
		t.Errorf("expected body 'Hello World', got %s", buf.String())
	}
}

func TestIsExpired(t *testing.T) {
	entry := Entry{
		ExpiresAt: time.Now().Add(-time.Second),
	}

	if !entry.IsExpired() {
		t.Error("expected entry to be expired")
	}

	entry.ExpiresAt = time.Now().Add(time.Minute)
	if entry.IsExpired() {
		t.Error("expected entry to not be expired")
	}
}
