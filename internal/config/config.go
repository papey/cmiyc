package config

import (
	"os"
	"sort"

	"gopkg.in/yaml.v2"
)

type LoadBalancerStrategy string

const (
	LBStrategySingle     LoadBalancerStrategy = "single"
	LBStrategyRandom     LoadBalancerStrategy = "random"
	LBStrategyRoundRobin LoadBalancerStrategy = "round_robin"
)

type Backend struct {
	URL string `yaml:"url"`
}

type CacheConfig struct {
	Enabled      bool `yaml:"enabled"`
	MaxSize      int  `yaml:"max_size"`
	MaxEntrySize int  `yaml:"max_entry_size"`
	TTL          int  `yaml:"ttl"`
}

type LBConfig struct {
	Type LoadBalancerStrategy `yaml:"strategy"`
}

type Route struct {
	LoadBalancerType LoadBalancerStrategy `yaml:"load_balancer_strategy"`
	CacheConfig      CacheConfig          `yaml:"cache"`
	LBConfig         LBConfig             `yaml:"load_balancer"`
	Backends         []Backend            `yaml:"backends"`
}

func (r *Route) ConfiguredURLs() []string {
	urls := make([]string, 0, len(r.Backends))
	for _, b := range r.Backends {
		urls = append(urls, b.URL)
	}
	return urls
}

type Config struct {
	Routes map[string]Route `yaml:"routes"`
	Listen string           `yaml:"listen"`

	prioritizedRoutes []string
}

func (c *Config) GetConfigForRoute(route string) (*Route, bool) {
	r, ok := c.Routes[route]

	return &r, ok
}

func (c *Config) GetPrioritizedMatchingRoute(route string) (string, bool) {
	for _, r := range c.prioritizedRoutes {
		if len(route) >= len(r) && route[:len(r)] == r {
			return r, true
		}
	}

	return "", false
}

func BuildConfigurationFromFile(path string) (*Config, error) {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}

	config.prioritizedRoutes = sortRoutesByLength(config.Routes)

	return &config, err
}

func NewConfig(listen string, routes map[string]Route) Config {
	return Config{
		Listen:            listen,
		Routes:            routes,
		prioritizedRoutes: sortRoutesByLength(routes),
	}
}

func sortRoutesByLength(routes map[string]Route) []string {
	sortedRoutes := make([]string, 0, len(routes))
	for r := range routes {
		sortedRoutes = append(sortedRoutes, r)
	}

	sort.Slice(sortedRoutes, func(i, j int) bool {
		lenI := len(sortedRoutes[i])
		lenJ := len(sortedRoutes[j])

		if lenI != lenJ {
			return lenI > lenJ
		}

		return sortedRoutes[i] < sortedRoutes[j]
	})

	return sortedRoutes
}
