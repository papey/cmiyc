package cache

import (
	"net/http"
	"time"
)

type Entry struct {
	StatusCode int
	Body       []byte
	Header     http.Header
	ExpiresAt  time.Time
}

func NewEntry(request *http.Request, resp *CachableResponse, expiresAt time.Time) (Key, Entry) {

	return KeyFrom(request), Entry{
		StatusCode: resp.StatusCode,
		Body:       resp.Body.Bytes(),
		Header:     ensureCacheHitHeader(cloneHeader(resp.Header())),
		ExpiresAt:  expiresAt,
	}
}

func (entry *Entry) Clone() Entry {
	bodyCopy := make([]byte, len(entry.Body))
	copy(bodyCopy, entry.Body)

	return Entry{
		StatusCode: entry.StatusCode,
		Body:       bodyCopy,
		Header:     cloneHeader(entry.Header),
		ExpiresAt:  entry.ExpiresAt,
	}
}

func ensureCacheHitHeader(h http.Header) http.Header {
	h.Set("X-Cache", "HIT")
	return h
}

func (entry *Entry) WriteResponse(w http.ResponseWriter, r *http.Request) error {
	for key, values := range entry.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(entry.StatusCode)
	_, err := w.Write(entry.Body)

	return err
}

func (entry *Entry) IsExpired() bool {
	return time.Now().After(entry.ExpiresAt)
}
