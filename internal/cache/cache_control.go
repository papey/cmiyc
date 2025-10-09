package cache

import (
	"fmt"
	"strings"
	"time"
)

type HeaderCacheControl struct {
	NoStore bool
	Private bool
	NoCache bool
	MaxAge  int64
	SMaxAge int64
}

func ParseCacheControl(cacheControlValue string) *HeaderCacheControl {
	cc := &HeaderCacheControl{
		MaxAge:  -1,
		SMaxAge: -1,
	}

	if cacheControlValue == "" {
		return cc
	}

	directives := strings.Split(cacheControlValue, ",")
	for _, raw := range directives {
		directive := strings.ToLower(strings.TrimSpace(raw))

		switch {
		case directive == "no-store":
			cc.NoStore = true
		case directive == "private":
			cc.Private = true
		case directive == "no-cache":
			cc.NoCache = true
		case strings.HasPrefix(directive, "max-age="):
			var n int64
			_, err := fmt.Sscanf(directive, "max-age=%d", &n)
			if err == nil {
				cc.MaxAge = n
			}
		case strings.HasPrefix(directive, "s-maxage="):
			var n int64
			_, err := fmt.Sscanf(directive, "s-maxage=%d", &n)
			if err == nil {
				cc.SMaxAge = n
			}
		}
	}

	return cc
}

func (cc *HeaderCacheControl) isCachable() bool {
	return !cc.NoStore && !cc.Private && (cc.MaxAge != 0 || cc.SMaxAge != 0)
}

func (cc *HeaderCacheControl) isExplicitlyCachable() bool {
	return cc.isCachable() && (cc.MaxAge > 0 || cc.SMaxAge > 0)
}

func (cc *HeaderCacheControl) TTL() (bool, time.Duration) {
	if cc.SMaxAge >= 0 {
		return true, time.Duration(cc.SMaxAge) * time.Second
	}

	if cc.MaxAge >= 0 {
		return true, time.Duration(cc.MaxAge) * time.Second
	}

	return false, 0
}
