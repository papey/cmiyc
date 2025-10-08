package config

import (
	"os"
	"sort"

	"gopkg.in/yaml.v2"
)

type LoadBalancerStrategy string

const (
	Single LoadBalancerStrategy = "single"
)

type Backend struct {
	URL string `yaml:"url"`
}

type Route struct {
	LoadBalancerType LoadBalancerStrategy `yaml:"load_balancer_strategy"`
	Backend          []Backend            `yaml:"backends"`
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
