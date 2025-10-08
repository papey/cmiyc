package balancer

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/papey/cmiyc/internal"
)

type HttpClient struct {
	client *http.Client
}

func NewHttpClient() *HttpClient {
	return &HttpClient{
		client: &http.Client{},
	}
}

func (c *HttpClient) Proxify(r *http.Request, dest string) (*http.Response, error) {
	backendURL, proxyURL, err := buildProxyURLs(r, dest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(r.Method, proxyURL.String(), r.Body)
	if err != nil {
		return nil, err
	}

	req.Host = backendURL.Host
	req.Header = c.buildHeaders(r)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
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

func (c *HttpClient) buildHeaders(r *http.Request) http.Header {
	h := http.Header{}

	c.forwardHeaders(r.Header, h)
	c.addHeaders(r, h)

	return h
}

func (c *HttpClient) forwardHeaders(origin http.Header, dest http.Header) http.Header {
	for key := range origin {
		if _, ok := HopByHopHeaderMap[key]; ok {
			continue
		}

		for _, v := range origin.Values(key) {
			dest.Add(key, v)
		}

	}

	return dest
}

func (c *HttpClient) addHeaders(r *http.Request, h http.Header) {
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

func buildProxyURLs(r *http.Request, dest string) (*url.URL, *url.URL, error) {
	backendURL, err := url.Parse(dest)
	if err != nil {
		return nil, nil, err
	}

	return backendURL, backendURL.ResolveReference(r.URL), nil
}
