package forwarder

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/papey/cmiyc/internal"
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
	resp, err := c.Proxify(r, dest)
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

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	return err
}

func (c *Client) Proxify(r *http.Request, dest string) (*http.Response, error) {
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

var HopByHopHeaderMap = map[string]struct{}{
	"Connection":          {},
	"Keep-Alive":          {},
	"Proxy-Authenticate":  {},
	"Proxy-Authorization": {},
	"TE":                  {},
	"Trailer":             {},
	"Transfer-Encoding":   {},
	"Upgrade":             {},
}

func (c *Client) buildRequestHeaders(r *http.Request) http.Header {
	h := c.forwardHeaders(r.Header)
	c.addProxyHeaders(r, h)

	return h
}

func (c *Client) forwardHeaders(origin http.Header) http.Header {
	h := origin.Clone()

	cleanHopByHopHeaders(h)

	return h
}

func (c *Client) addProxyHeaders(r *http.Request, h http.Header) {
	h.Add("Via", internal.VersionedName())
	h.Set("X-Forwarded-Host", r.Host)

	if clientIP, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if prior := h.Get("X-Forwarded-For"); prior != "" {
			h.Set("X-Forwarded-For", fmt.Sprintf("%s, %s", prior, clientIP))
		} else {
			h.Set("X-Forwarded-For", clientIP)
		}
	}

	if r.TLS != nil {
		h.Set("X-Forwarded-Proto", "https")
	} else {
		h.Set("X-Forwarded-Proto", "http")
	}
}

func buildURLs(r *http.Request, dest string) (*url.URL, *url.URL, error) {
	backendURL, err := url.Parse(dest)
	if err != nil {
		return nil, nil, err
	}

	return backendURL, backendURL.ResolveReference(r.URL), nil
}

func cleanHopByHopHeaders(h http.Header) {
	for key := range HopByHopHeaderMap {
		h.Del(key)
	}
}
