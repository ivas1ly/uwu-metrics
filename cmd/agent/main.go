package main

import "github.com/ivas1ly/uwu-metrics/internal/agent"

func main() {
	cfg := agent.NewConfig()

	agent.Run(cfg)
}
