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
	configuredRoute, found := b.config.GetPrioritizedMatchingRoute(r.URL.Path)
	if !found {
		http.Error(w, "Route not found", http.StatusNotFound)
		return
	}

	c, ok := b.config.GetConfigForRoute(configuredRoute)
	if !ok {
		http.Error(w, "Route configuration not found", http.StatusNotFound)
		return
	}

	err := b.client.ProxifyAndServe(w, r, c.Backend[0].URL)
	if err != nil {
		log.Println(err)
		return
	}
}

func (b *Balancer) Start() error {
	return http.ListenAndServe(b.config.Listen, http.HandlerFunc(b.handleRequest))
}
