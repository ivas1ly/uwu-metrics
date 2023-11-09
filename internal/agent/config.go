package agent

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultPollInterval   = 2
	defaultReportInterval = 10
	defaultEndpointHost   = "localhost:8080"
	defaultClientTimeout  = 3 * time.Second
)

type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	EndpointHost   string
}

func NewConfig() *Config {
	cfg := &Config{}

	endpointHostUsage := fmt.Sprintf("HTTP server report endpoint, example: %q", defaultEndpointHost)
	flag.StringVar(&cfg.EndpointHost, "a", defaultEndpointHost, endpointHostUsage)

	reportIntervalUsage := fmt.Sprintf("frequency of sending metrics to the server, example: %q",
		defaultReportInterval)
	ri := flag.Int("r", defaultReportInterval, reportIntervalUsage)

	pollIntervalUsage := fmt.Sprintf("frequency of polling metrics from the runtime package, example: %q",
		defaultPollInterval)
	pi := flag.Int("p", defaultPollInterval, pollIntervalUsage)

	flag.Parse()

	// check report interval value
	if *ri <= 0 {
		cfg.ReportInterval = defaultReportInterval * time.Second
	} else {
		cfg.ReportInterval = time.Duration(*ri) * time.Second
	}

	// check poll interval value
	if *pi <= 0 {
		cfg.PollInterval = defaultPollInterval * time.Second
	} else {
		cfg.PollInterval = time.Duration(*pi) * time.Second
	}

	if endpointHost := os.Getenv("ADDRESS"); endpointHost != "" {
		cfg.EndpointHost = endpointHost
	}

	if reportInterval := os.Getenv("REPORT_INTERVAL"); reportInterval != "" {
		envValue, err := strconv.Atoi(reportInterval)
		if err == nil && envValue > 0 {
			cfg.ReportInterval = time.Duration(envValue) * time.Second
		}
	}

	if pollInterval := os.Getenv("POLL_INTERVAL"); pollInterval != "" {
		envValue, err := strconv.Atoi(pollInterval)
		if err == nil && envValue > 0 {
			cfg.PollInterval = time.Duration(envValue) * time.Second
		}
	}

	return cfg
}
