package balancer

import (
	"fmt"
	"sync/atomic"
)

type RRBalancer struct {
	urls  []string
	index atomic.Int64
}

func NewRRBalancer(urls []string) *RRBalancer {
	return &RRBalancer{urls: urls}
}

func (r *RRBalancer) Pick() string {
	n := int64(len(r.urls))

	i := r.index.Add(1) - 1
	fmt.Println(i % n)

	return r.urls[i%n]
}
