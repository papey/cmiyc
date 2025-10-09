package cache

import (
	"bytes"
	"net/http"
	"time"
)

type CachableResponse struct {
	http.ResponseWriter
	StatusCode int
	Body       *bytes.Buffer
}

func NewCachableResponse(w http.ResponseWriter) *CachableResponse {
	return &CachableResponse{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
		Body:           &bytes.Buffer{},
	}
}

func (cr *CachableResponse) Write(data []byte) (int, error) {
	cr.Body.Write(data)
	return cr.ResponseWriter.Write(data)
}

func (cr *CachableResponse) WriteHeader(statusCode int) {
	cr.StatusCode = statusCode
	cr.ResponseWriter.WriteHeader(statusCode)
}

var cachableResponseStatuses = []int{
	http.StatusOK,                   // 200
	http.StatusNonAuthoritativeInfo, // 203
	http.StatusNoContent,            // 204
	http.StatusPartialContent,       // 206
	http.StatusMultipleChoices,      // 300
	http.StatusMovedPermanently,     // 301
	http.StatusNotFound,             // 404
	http.StatusMethodNotAllowed,     // 405
	http.StatusGone,                 // 410
	http.StatusRequestURITooLong,    // 414
	http.StatusNotImplemented,       // 501
}

func isCachableStatus(status int) bool {
	for _, s := range cachableResponseStatuses {
		if status == s {
			return true
		}
	}

	return false
}

func (cr *CachableResponse) CacheTTL() (bool, time.Duration) {
	present, duration := ParseCacheControl(cr.Header().Get("Cache-Control")).TTL()
	if present {
		return true, duration
	}

	expiresHeader := cr.Header().Get("Expires")
	if expiresHeader != "" {
		if expiresTime, err := http.ParseTime(expiresHeader); err == nil {
			return true, time.Until(expiresTime)
		}
	}

	return false, 0
}

func (cr *CachableResponse) IsCachable() bool {
	return isCachableStatus(cr.StatusCode) && ParseCacheControl(cr.Header().Get("Cache-Control")).isCachable()
}

func (cr *CachableResponse) IsCachableConsideringAuth() bool {
	return isCachableStatus(cr.StatusCode) && ParseCacheControl(cr.Header().Get("Cache-Control")).isCachable()
}

func IsRequestCachable(requestMethod string) bool {
	return requestMethod == http.MethodGet
}
