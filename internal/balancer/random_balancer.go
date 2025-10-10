package balancer

import (
	"math/rand"
)

type RandomLB struct {
	urls []string
	rnd  *rand.Rand
}

func NewRandomLB(urls []string, seed int64) *RandomLB {
	return &RandomLB{
		urls: urls,
		rnd:  rand.New(rand.NewSource(seed)),
	}
}

func (r *RandomLB) Pick() string {
	if len(r.urls) == 0 {
		return ""
	}
	idx := r.rnd.Intn(len(r.urls))
	return r.urls[idx]
}
