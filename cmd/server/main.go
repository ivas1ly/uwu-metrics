package main

import (
	"fmt"

	"github.com/ivas1ly/uwu-metrics/internal/server"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	cfg := server.NewConfig()

	server.Run(cfg)
}
