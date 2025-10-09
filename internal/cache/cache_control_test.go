package cache

import (
	"testing"
	"time"
)

func TestParseCacheControl(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected HeaderCacheControl
	}{
		{
			name:   "empty header",
			header: "",
			expected: HeaderCacheControl{
				MaxAge:  -1,
				SMaxAge: -1,
			},
		},
		{
			name:   "no-store directive",
			header: "no-store",
			expected: HeaderCacheControl{
				NoStore: true,
				MaxAge:  -1,
				SMaxAge: -1,
			},
		},
		{
			name:   "private directive",
			header: "private",
			expected: HeaderCacheControl{
				Private: true,
				MaxAge:  -1,
				SMaxAge: -1,
			},
		},
		{
			name:   "no-cache directive",
			header: "no-cache",
			expected: HeaderCacheControl{
				NoCache: true,
				MaxAge:  -1,
				SMaxAge: -1,
			},
		},
		{
			name:   "max-age and s-maxage",
			header: "max-age=3600, s-maxage=7200",
			expected: HeaderCacheControl{
				MaxAge:  3600,
				SMaxAge: 7200,
			},
		},
		{
			name:   "mixed directives with spaces",
			header: "  private,  max-age=120 , s-maxage=300 ",
			expected: HeaderCacheControl{
				Private: true,
				MaxAge:  120,
				SMaxAge: 300,
			},
		},
		{
			name:   "invalid numbers ignored",
			header: "max-age=abc, s-maxage=def",
			expected: HeaderCacheControl{
				MaxAge:  -1,
				SMaxAge: -1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := ParseCacheControl(tt.header)
			if cc.NoStore != tt.expected.NoStore {
				t.Errorf("NoStore = %v, expected %v", cc.NoStore, tt.expected.NoStore)
			}
			if cc.Private != tt.expected.Private {
				t.Errorf("Private = %v, expected %v", cc.Private, tt.expected.Private)
			}
			if cc.NoCache != tt.expected.NoCache {
				t.Errorf("NoCache = %v, expected %v", cc.NoCache, tt.expected.NoCache)
			}
			if cc.MaxAge != tt.expected.MaxAge {
				t.Errorf("MaxAge = %d, expected %d", cc.MaxAge, tt.expected.MaxAge)
			}
			if cc.SMaxAge != tt.expected.SMaxAge {
				t.Errorf("SMaxAge = %d, expected %d", cc.SMaxAge, tt.expected.SMaxAge)
			}
		})
	}
}

func TestIsCachable(t *testing.T) {
	tests := []struct {
		name     string
		cc       HeaderCacheControl
		expected bool
	}{
		{"default (no flags)", HeaderCacheControl{MaxAge: -1, SMaxAge: -1}, true},
		{"no-store", HeaderCacheControl{NoStore: true}, false},
		{"private", HeaderCacheControl{Private: true}, false},
		{"max-age=0", HeaderCacheControl{MaxAge: 0}, false},
		{"s-maxage=0", HeaderCacheControl{SMaxAge: 0}, false},
		{"max-age>0", HeaderCacheControl{MaxAge: 100}, true},
		{"s-maxage>0", HeaderCacheControl{SMaxAge: 200}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cc.isCachable()
			if got != tt.expected {
				t.Errorf("isCachable() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestIsExplicitlyCachable(t *testing.T) {
	tests := []struct {
		name     string
		cc       HeaderCacheControl
		expected bool
	}{
		{"no max-age", HeaderCacheControl{MaxAge: -1, SMaxAge: -1}, false},
		{"max-age=0", HeaderCacheControl{MaxAge: 0}, false},
		{"max-age=100", HeaderCacheControl{MaxAge: 100}, true},
		{"s-maxage=200", HeaderCacheControl{SMaxAge: 200}, true},
		{"private with max-age", HeaderCacheControl{Private: true, MaxAge: 100}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cc.isExplicitlyCachable()
			if got != tt.expected {
				t.Errorf("isExplicitlyCachable() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestTTL(t *testing.T) {
	tests := []struct {
		name     string
		cc       HeaderCacheControl
		expected bool
		duration time.Duration
	}{
		{"both unset", HeaderCacheControl{MaxAge: -1, SMaxAge: -1}, false, 0},
		{"max-age=60", HeaderCacheControl{MaxAge: 60, SMaxAge: -1}, true, 60 * time.Second},
		{"s-maxage=120", HeaderCacheControl{MaxAge: -1, SMaxAge: 120}, true, 120 * time.Second},
		{"s-maxage takes priority", HeaderCacheControl{MaxAge: 60, SMaxAge: 120}, true, 120 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, ttl := tt.cc.TTL()
			if ok != tt.expected {
				t.Errorf("TTL() ok = %v, expected %v", ok, tt.expected)
			}
			if ttl != tt.duration {
				t.Errorf("TTL() duration = %v, expected %v", ttl, tt.duration)
			}
		})
	}
}
