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
	defaultLogLevel       = "info"
	exampleKey            = ""
	defaultRateLimit      = 1
	defaultPprofAddr      = "localhost:9091"
	examplePublicKeyPath  = "./cmd/agent/public_key.pem"
)

// Config structure contains the received information for running the application.
type Config struct {
	EndpointHost   string
	Key            string
	PublicKeyPath  string
	PollInterval   time.Duration
	ReportInterval time.Duration
	RateLimit      int
}

// NewConfig creates a new configuration depending on the method.
// It can be set either with command line flags or with environment variables.
// Environment variables take precedence over flags.
func NewConfig() Config {
	cfg := Config{}

	endpointHostUsage := fmt.Sprintf("HTTP server report endpoint, example: %q", defaultEndpointHost)
	flag.StringVar(&cfg.EndpointHost, "a", defaultEndpointHost, endpointHostUsage)

	reportIntervalUsage := fmt.Sprintf("frequency of sending metrics to the server, example: %q",
		defaultReportInterval)
	ri := flag.Int("r", defaultReportInterval, reportIntervalUsage)

	pollIntervalUsage := fmt.Sprintf("frequency of polling metrics from the runtime package, example: %q",
		defaultPollInterval)
	pi := flag.Int("p", defaultPollInterval, pollIntervalUsage)

	hashKeyUsage := fmt.Sprintf("key for signing the request body hash, example: %q", exampleKey)
	flag.StringVar(&cfg.Key, "k", "", hashKeyUsage)

	rateLimitUsage := fmt.Sprintf("number of concurrent requests to the metrics server, example: %q",
		defaultRateLimit)
	flag.IntVar(&cfg.RateLimit, "l", defaultRateLimit, rateLimitUsage)

	publicKeyPathUsage := fmt.Sprintf("path to the file with rsa public key, example: %s",
		examplePublicKeyPath)
	flag.StringVar(&cfg.PublicKeyPath, "crypto-key", "", publicKeyPathUsage)

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

	if hashKey := os.Getenv("KEY"); hashKey != "" {
		cfg.Key = hashKey
	}

	if rateLimit := os.Getenv("RATE_LIMIT"); rateLimit != "" {
		envValue, err := strconv.Atoi(rateLimit)
		if err == nil && envValue > 0 {
			cfg.RateLimit = envValue
		}
	}

	if publicKey := os.Getenv("CRYPTO_KEY"); publicKey != "" {
		cfg.PublicKeyPath = publicKey
	}

	fmt.Printf("%+v\n\n", cfg)

	return cfg
}
