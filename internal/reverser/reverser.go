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

	for k, c := range cfg.Routes {
		if c.CacheConfig.Enabled {
			caches[k] = cache.NewEmptyCache(c.CacheConfig.MaxSize, c.CacheConfig.MaxEntrySize)
		}
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

	resp := cache.NewCachableResponse(w)

	if !c.CacheConfig.Enabled {
		err := rev.proxyDirect(resp, r, c)
		if err != nil {
			log.Println(err)
			return
		}

		return
	}

	routeCache, exists := rev.getCacheForRoute(matchingRoute)
	if !exists {
		err := rev.proxyDirect(resp, r, c)
		if err != nil {
			log.Println(err)
			return
		}

		return
	}

	err := rev.proxyCache(resp, r, c, routeCache)
	if err != nil {
		log.Println(err)
	}
}

func (rev *Reverser) proxyDirect(resp *cache.CachableResponse, r *http.Request, rc *config.Route) error {
	if err := rev.client.ProxifyAndServe(resp, r, rc.Backends[0].URL); err != nil {
		return err
	}

	return nil
}

func (rev *Reverser) proxyCache(resp *cache.CachableResponse, r *http.Request, rc *config.Route, routeCache *cache.HttpCache) error {
	isRequestCachable := cache.IsRequestCachable(r.Method)
	if isRequestCachable {
		served, err := routeCache.ServeIfPresent(resp.ResponseWriter, r)
		if served {
			return nil
		}

		if err != nil {
			log.Println(err)
		}
	}

	err := rev.client.ProxifyAndServe(resp, r, rc.Backends[0].URL)
	if err != nil {
		return err
	}

	contextAllowsCaching := isRequestCachable && ((withoutAuthorizationHeader(r) && resp.IsCachable()) || resp.IsCachableConsideringAuth())
	if contextAllowsCaching {
		cacheDuration := cacheDurationWithFallback(resp, time.Duration(rc.CacheConfig.TTL)*time.Second)
		routeCache.Set(r, resp, time.Now().Add(cacheDuration))
	}

	return nil
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

	for _, c := range rev.caches {
		c.Cleanup()
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
