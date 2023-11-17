package server

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	defaultHost              = "localhost"
	defaultPort              = "8080"
	defaultReadTimeout       = 10 * time.Second
	defaultReadHeaderTimeout = 5 * time.Second
	defaultWriteTimeout      = 10 * time.Second
	defaultIdleTimeout       = 1 * time.Minute
	defaultLogLevel          = "info"
)

type Config struct {
	Endpoint string
}

func NewConfig() *Config {
	cfg := &Config{}

	endpointUsage := fmt.Sprintf("HTTP server endpoint, example: %q or %q",
		net.JoinHostPort(defaultHost, defaultPort), net.JoinHostPort("", defaultPort))
	flag.StringVar(&cfg.Endpoint, "a", net.JoinHostPort(defaultHost, defaultPort), endpointUsage)
	flag.Parse()

	if endpoint := os.Getenv("ADDRESS"); endpoint != "" {
		cfg.Endpoint = endpoint
	}

	return cfg
}
