package cache

import (
	"net/http"
	"time"
)

type Entry struct {
	StatusCode  int
	Body        []byte
	Header      http.Header
	LocalTimeFn func() time.Time
	ExpiresAt   time.Time
}

func NewEntry(request *http.Request, resp *CachableResponse, expiresAt time.Time) (Key, Entry) {
	return KeyFrom(request), Entry{
		StatusCode: resp.StatusCode,
		Body:       resp.Body.Bytes(),
		Header:     cloneHeader(resp.Header()),
		ExpiresAt:  expiresAt,
	}
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
