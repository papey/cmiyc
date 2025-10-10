package balancer

type SingleLB struct {
	urls []string
}

func NewSingleLB(urls []string) *SingleLB {
	return &SingleLB{urls: urls}
}

func (s *SingleLB) Pick() string {
	return s.urls[0]
}
