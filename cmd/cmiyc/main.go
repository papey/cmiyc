package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/papey/cmiyc/internal/balancer"
	"github.com/papey/cmiyc/internal/config"
)

func main() {
	configPath := parseArgs()

	conf, err := config.ParseConfigurationFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to parse configuration file: %v", err)
	}

	log.Printf("Configuration loaded: %s", *configPath)

	bal := balancer.NewBalancer(*conf)

	log.Printf("Starting balancer on %s", conf.Listen)
	err = bal.Start()
	if err != nil {
		log.Fatalf("Failed to start balancer: %v", err)
	}
}

func parseArgs() *string {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")

	flag.Usage = func() {
		fmt.Println("Usage of cmiyc")
		flag.PrintDefaults()
	}

	flag.Parse()

	return configPath
}
