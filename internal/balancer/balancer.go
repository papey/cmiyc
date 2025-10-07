package balancer

import (
	"io"
	"log"
	"net/http"

	"github.com/papey/cmiyc/internal/config"
)

type Balancer struct {
	config config.Config
	client *HttpClient
}

func NewBalancer(cfg config.Config) *Balancer {
	b := &Balancer{
		config: cfg,
		client: NewHttpClient(),
	}

	return b
}

func (b *Balancer) handleRequest(w http.ResponseWriter, r *http.Request) {
	c, found := b.config.GetConfigForRoute(r.URL.Path)
	if !found {
		http.Error(w, "Route not found", http.StatusNotFound)
		return
	}

	resp, err := b.client.Proxify(r, c.Backend[0].URL)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error forwarding request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}
}

func (b *Balancer) Start() error {
	return http.ListenAndServe(b.config.Listen, http.HandlerFunc(b.handleRequest))
}
