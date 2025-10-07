package balancer

import (
	"fmt"
	"net/http"
)

type HttpClient struct {
	client *http.Client
}

func NewHttpClient() *HttpClient {
	return &HttpClient{
		client: &http.Client{},
	}
}

type Request struct {
	Method string
	Host   string
	URL    string
	Header http.Header
	Body   []byte
}

type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

func (c *HttpClient) Proxify(r *http.Request, dest string) (*http.Response, error) {
	req, err := http.NewRequest(r.Method, fmt.Sprintf("%s/%s", dest, r.URL.Path), r.Body)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
