package main

import (
	"fmt"

	"github.com/ivas1ly/uwu-metrics/internal/agent"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n\n", buildCommit)

	cfg := agent.NewConfig()

	agent.Run(cfg)
}
