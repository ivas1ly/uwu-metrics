package server

import (
	"flag"
	"net"
	"os"
)

const (
	defaultHost = "localhost"
	defaultPort = "8080"
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
