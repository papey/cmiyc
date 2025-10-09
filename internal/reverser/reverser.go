package reverser

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/papey/cmiyc/internal/config"
	"github.com/papey/cmiyc/internal/forwarder"
)

type Reverser struct {
	config config.Config
	client *forwarder.Client
	server *http.Server
}

func NewReverser(cfg config.Config) *Reverser {
	b := &Reverser{
		config: cfg,
		client: forwarder.NewClient(),
	}

	return b
}

func (rev *Reverser) handleRequest(w http.ResponseWriter, r *http.Request) {
	configuredRoute, found := rev.config.GetPrioritizedMatchingRoute(r.URL.Path)
	if !found {
		http.Error(w, "Route not found", http.StatusNotFound)
		return
	}

	c, ok := rev.config.GetConfigForRoute(configuredRoute)
	if !ok {
		http.Error(w, "Route configuration not found", http.StatusNotFound)
		return
	}

	err := rev.client.ProxifyAndServe(w, r, c.Backend[0].URL)
	if err != nil {
		log.Println(err)
		return
	}
}

func (rev *Reverser) Start() error {
	rev.server = &http.Server{
		Addr:    rev.config.Listen,
		Handler: http.HandlerFunc(rev.handleRequest),
	}

	log.Printf("Reverse proxy listening on %s", rev.config.Listen)
	return rev.server.ListenAndServe()
}

const gracefulWait = 15 * time.Second

func (rev *Reverser) Stop() error {
	if rev.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), gracefulWait)
	defer cancel()

	log.Println("Shutting down reverser...")
	return rev.server.Shutdown(ctx)
}
