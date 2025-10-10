package balancer

import (
	"testing"
)

func TestSingleLBPick(t *testing.T) {
	lb := NewSingleLB([]string{
		"http://backend1",
		"http://backend2",
		"http://backend3",
	})

	for i := 0; i < 10; i++ {
		selected := lb.Pick()
		if selected != "http://backend1" {
			t.Errorf("expected %s, got %s", "http://backend1", selected)
		}
	}
}
