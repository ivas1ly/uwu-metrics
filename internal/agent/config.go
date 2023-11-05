package agent

import (
	"flag"
	"time"
)

const (
	defaultPollInterval   = 2
	defaultReportInterval = 10
	defaultEndpointURL    = "localhost:8080"
)

type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	EndpointURL    string
}

func NewConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.EndpointURL, "a", defaultEndpointURL, "HTTP server report endpoint, "+
		"example: 'localhost:8080'")
	ri := flag.Int("r", defaultReportInterval, "frequency of sending metrics "+
		"to the server, example: '10'")
	pi := flag.Int("p", defaultPollInterval, "frequency of polling metrics from "+
		"the runtime package, example: '2'")

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

	return cfg
}
