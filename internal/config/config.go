package config

import (
	"os"

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
}

func (c *Config) GetConfigForRoute(route string) (*Route, bool) {
	r, ok := c.Routes[route]

	return &r, ok
}

func ParseConfigurationFile(path string) (*Config, error) {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}

	return &config, err
}
