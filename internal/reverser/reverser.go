package reverser

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/papey/cmiyc/internal/balancer"
	"github.com/papey/cmiyc/internal/cache"
	"github.com/papey/cmiyc/internal/config"
	"github.com/papey/cmiyc/internal/forwarder"
)

type Reverser struct {
	config config.Config
	client *forwarder.Client
	server *http.Server
	caches map[string]*cache.HttpCache
	lbs    map[string]balancer.Balancer
}

func NewReverser(cfg config.Config) *Reverser {
	caches := make(map[string]*cache.HttpCache)
	lbs := make(map[string]balancer.Balancer)

	for k, c := range cfg.Routes {
		if c.CacheConfig.Enabled {
			caches[k] = cache.NewEmptyCache(c.CacheConfig.MaxSize, c.CacheConfig.MaxEntrySize)
		}

		switch c.LBConfig.Type {
		case config.LBStrategySingle:
			lbs[k] = balancer.NewSingleLB(c.ConfiguredURLs())
		default:
			fmt.Printf("Unknown load balancer strategy %s for route %s, defaulting to single", c.LBConfig.Type, k)
			lbs[k] = balancer.NewSingleLB(c.ConfiguredURLs())
		}

	}

	if len(lbs) != len(cfg.Routes) {
		log.Fatalf("Some routes do not have a load balancer configured")
	}

	r := &Reverser{
		config: cfg,
		client: forwarder.NewClient(),
		caches: caches,
		lbs:    lbs,
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

	lb, exists := rev.getLbforRoute(matchingRoute)
	if !exists {
		http.Error(w, "Load balancer not found for route", http.StatusInternalServerError)
		return
	}

	resp := cache.NewCachableResponse(w)

	if !c.CacheConfig.Enabled {
		err := rev.proxyDirect(resp, r, lb.Pick())
		if err != nil {
			log.Println(err)
			return
		}

		return
	}

	routeCache, exists := rev.getCacheForRoute(matchingRoute)
	if !exists {
		err := rev.proxyDirect(resp, r, lb.Pick())
		if err != nil {
			log.Println(err)
			return
		}

		return
	}

	err := rev.proxyCache(resp, r, c, lb.Pick(), routeCache)
	if err != nil {
		log.Println(err)
	}
}

func (rev *Reverser) proxyDirect(resp *cache.CachableResponse, r *http.Request, baseURL string) error {
	if err := rev.client.ProxifyAndServe(resp, r, baseURL); err != nil {
		return err
	}

	return nil
}

func (rev *Reverser) proxyCache(resp *cache.CachableResponse, r *http.Request, rc *config.Route, baseURL string, routeCache *cache.HttpCache) error {
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

	err := rev.client.ProxifyAndServe(resp, r, baseURL)
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

func (rev *Reverser) getLbforRoute(route string) (balancer.Balancer, bool) {
	lb, exists := rev.lbs[route]
	return lb, exists
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
