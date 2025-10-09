package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/papey/cmiyc/internal/config"
	"github.com/papey/cmiyc/internal/reverser"
)

func main() {
	configPath := parseArgs()

	conf, err := config.BuildConfigurationFromFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to parse configuration file: %v", err)
	}

	log.Printf("Configuration loaded: %s", *configPath)

	r := reverser.NewReverser(*conf)
	done := setupGracefulShutdown(r)

	log.Printf("Starting reverser on %s", conf.Listen)
	err = r.Start()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Failed to start reverser: %v", err)
	}

	<-done
	log.Println("Reverser stopped gracefully")
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

func setupGracefulShutdown(r *reverser.Reverser) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		if err := r.Stop(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Failed to stop reverser: %v", err)
		}
		close(done)
	}()

	return done
}
