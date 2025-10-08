package balancer

import (
	"log"
	"net/http"

	"github.com/papey/cmiyc/internal/config"
	"github.com/papey/cmiyc/internal/forwarder"
)

type Balancer struct {
	config config.Config
	client *forwarder.Client
}

func NewBalancer(cfg config.Config) *Balancer {
	b := &Balancer{
		config: cfg,
		client: forwarder.NewClient(),
	}

	return b
}

func (b *Balancer) handleRequest(w http.ResponseWriter, r *http.Request) {
	c, found := b.config.GetConfigForRoute(r.URL.Path)
	if !found {
		http.Error(w, "Route not found", http.StatusNotFound)
		return
	}

	err := b.client.ProxifyAndServe(w, r, c.Backend[0].URL)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error forwarding request", http.StatusBadGateway)
		return
	}
}

func (b *Balancer) Start() error {
	return http.ListenAndServe(b.config.Listen, http.HandlerFunc(b.handleRequest))
}
