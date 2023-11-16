package main

import (
	"github.com/ivas1ly/uwu-metrics/internal/server"
)

func main() {
	cfg := server.NewConfig()

	server.Run(cfg)
}
