package forwarder

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	client *http.Client
}

func NewClient() *Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,

		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,

		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 20 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,

		ForceAttemptHTTP2: true,
	}

	return &Client{
		client: &http.Client{
			Transport: transport,
			Timeout:   60 * time.Second,
		},
	}
}

func (c *Client) ProxifyAndServe(w http.ResponseWriter, r *http.Request, dest string) error {
	resp, err := c.proxify(r, dest)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = fmt.Fprintf(w, "Proxy error: %v", err)
		return err
	}
	defer resp.Body.Close()

	cleanHopByHopHeaders(resp.Header)

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	addCacheHeaderOnCachableRequests(r.Method, w.Header())

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	return err
}

func (c *Client) proxify(r *http.Request, dest string) (*http.Response, error) {
	backendURL, proxyURL, err := buildURLs(r, dest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(r.Method, proxyURL.String(), r.Body)
	if err != nil {
		return nil, err
	}

	req.Header = c.buildRequestHeaders(r)
	req.Host = backendURL.Host

	return c.client.Do(req)
}

func addCacheHeaderOnCachableRequests(method string, h http.Header) {
	if method == http.MethodHead || method == http.MethodGet {
		h.Set("X-Cache", "MISS")
	}
}

func buildURLs(r *http.Request, dest string) (*url.URL, *url.URL, error) {
	backendURL, err := url.Parse(dest)
	if err != nil {
		return nil, nil, err
	}

	return backendURL, backendURL.ResolveReference(r.URL), nil
}
