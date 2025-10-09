package forwarder

import (
	"fmt"
	"net"
	"net/http"

	"github.com/papey/cmiyc/internal"
)

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

func cleanHopByHopHeaders(h http.Header) {
	for key := range HopByHopHeaderMap {
		h.Del(key)
	}
}
