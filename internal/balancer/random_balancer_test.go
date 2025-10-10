package balancer

import (
	"testing"
)

func TestRandomLBPickDeterministic(t *testing.T) {
	urls := []string{
		"http://backend1.local",
		"http://backend2.local",
		"http://backend3.local",
	}

	lb := NewRandomLB(urls, 42)

	expectedSequence := []string{
		"http://backend3.local",
		"http://backend3.local",
		"http://backend3.local",
		"http://backend1.local",
		"http://backend2.local",
	}

	for i, expected := range expectedSequence {
		got := lb.Pick()
		if got != expected {
			t.Errorf("step %d: expected %s, got %s", i, expected, got)
		}
	}
}
