package balancer_test

import (
	"testing"

	"github.com/papey/cmiyc/internal/balancer"
)

func TestRRBalancerPickWithURLs(t *testing.T) {
	urls := []string{
		"http://backend1.local",
		"http://backend2.local",
		"http://backend3.local",
	}
	rr := balancer.NewRRBalancer(urls)

	expected := []string{
		"http://backend1.local",
		"http://backend2.local",
		"http://backend3.local",
		"http://backend1.local",
		"http://backend2.local",
		"http://backend3.local",
	}

	for i, exp := range expected {
		got := rr.Pick()
		if got != exp {
			t.Errorf("Pick #%d: expected %q, got %q", i+1, exp, got)
		}
	}
}
