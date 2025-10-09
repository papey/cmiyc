package cache

import (
	"bytes"
	"net/http"
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

func IsCachableRequest(r *http.Request) bool {
	return r.Method == http.MethodGet
}
