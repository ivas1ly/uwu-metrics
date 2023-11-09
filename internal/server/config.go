package server

import (
	"flag"
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
)

type Config struct {
	Endpoint string
}

func NewConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Endpoint, "a", net.JoinHostPort(defaultHost, defaultPort), "HTTP server endpoint, "+
		"example: 'localhost:8080' or ':8080'")
	flag.Parse()

	if endpoint := os.Getenv("ADDRESS"); endpoint != "" {
		cfg.Endpoint = endpoint
	}

	return cfg
}
