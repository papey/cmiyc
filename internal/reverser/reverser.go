package reverser

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/papey/cmiyc/internal/cache"
	"github.com/papey/cmiyc/internal/config"
	"github.com/papey/cmiyc/internal/forwarder"
)

type Reverser struct {
	config config.Config
	client *forwarder.Client
	server *http.Server
	caches map[string]*cache.HttpCache
}

func NewReverser(cfg config.Config) *Reverser {
	caches := make(map[string]*cache.HttpCache)

	for k, _ := range cfg.Routes {
		caches[k] = cache.NewEmptyCache()
	}

	r := &Reverser{
		config: cfg,
		client: forwarder.NewClient(),
		caches: caches,
	}

	return r
}

func (rev *Reverser) handleRequest(w http.ResponseWriter, r *http.Request) {
	matchingRoute, found := rev.config.GetPrioritizedMatchingRoute(r.URL.Path)
	if !found {
		http.Error(w, "Route not found", http.StatusNotFound)
		return
	}

	c, ok := rev.config.GetConfigForRoute(matchingRoute)
	if !ok {
		http.Error(w, "Route configuration not found", http.StatusInternalServerError)
		return
	}

	routeCache, cacheExists := rev.getCacheForRoute(matchingRoute)
	if !cacheExists {
		log.Printf("Cache for route %s not found", matchingRoute)
	}

	isRequestCachable := cache.IsRequestCachable(r.Method) && cacheExists
	if isRequestCachable {
		served, err := routeCache.ServeIfPresent(w, r)
		if served {
			return
		}

		if err != nil {
			log.Println("Error serving from cache:", err)
		}
	}

	resp := cache.NewCachableResponse(w)
	err := rev.client.ProxifyAndServe(resp, r, c.Backends[0].URL)
	if err != nil {
		log.Println(err)
		return
	}

	contextAllowsCaching := isRequestCachable && ((withoutAuthorizationHeader(r) && resp.IsCachable()) || resp.IsCachableConsideringAuth())
	if contextAllowsCaching {
		cacheDuration := cacheDurationWithFallback(resp, 5*time.Minute)
		routeCache.Set(r, resp, time.Now().Add(cacheDuration))
	}
}

func (rev *Reverser) getCacheForRoute(route string) (*cache.HttpCache, bool) {
	c, exists := rev.caches[route]
	return c, exists
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

func withoutAuthorizationHeader(r *http.Request) bool {
	return r.Header.Get("Authorization") == ""
}

func cacheDurationWithFallback(resp *cache.CachableResponse, fallback time.Duration) time.Duration {
	specified, value := resp.CacheTTL()
	if specified {
		return value
	}

	return fallback
}
