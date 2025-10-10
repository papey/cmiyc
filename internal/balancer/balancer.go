package balancer

type Balancer interface {
	Pick() string
}
