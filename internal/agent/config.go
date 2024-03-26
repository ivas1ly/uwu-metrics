package agent

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/ivas1ly/uwu-metrics/internal/utils/flags"
)

const (
	defaultPollInterval    = 2
	defaultReportInterval  = 10
	defaultEndpointHost    = "localhost:8080"
	defaultClientTimeout   = 3 * time.Second
	defaultLogLevel        = "info"
	exampleKey             = ""
	defaultRateLimit       = 1
	defaultPprofAddr       = "localhost:9091"
	examplePublicKeyPath   = "./cmd/agent/public_key.pem"
	defaultFilePerm        = 0666
	exampleConfigPathUsage = "./config/agent.json"
)

const (
	flagEndpointHost   = "a"
	flagReportInterval = "r"
	flagPollInterval   = "p"
	flagHashKey        = "k"
	flagRateLimit      = "l"
	flagPublicKey      = "crypto-key"
)

// Config structure contains the received information for running the application.
type Config struct {
	EndpointHost   string
	HashKey        string
	PublicKeyPath  string
	PollInterval   time.Duration
	ReportInterval time.Duration
	RateLimit      int
}

// NewConfig creates a new configuration depending on the method.
// It can be set either with command line flags or with environment variables.
// Environment variables take precedence over flags.
func NewConfig() Config {
	cfg := Config{
		EndpointHost:   defaultEndpointHost,
		HashKey:        "",
		PublicKeyPath:  "",
		PollInterval:   defaultPollInterval,
		ReportInterval: defaultReportInterval,
		RateLimit:      defaultRateLimit,
	}

	endpointHostUsage := fmt.Sprintf("HTTP server report endpoint, example: %q", defaultEndpointHost)
	endpointHost := flag.String(flagEndpointHost, "", endpointHostUsage)

	reportIntervalUsage := fmt.Sprintf("frequency of sending metrics to the server, example: %q",
		defaultReportInterval)
	reportInterval := flag.Int(flagReportInterval, -1, reportIntervalUsage)

	pollIntervalUsage := fmt.Sprintf("frequency of polling metrics from the runtime package, example: %q",
		defaultPollInterval)
	pollInterval := flag.Int(flagPollInterval, -1, pollIntervalUsage)

	hashKeyUsage := fmt.Sprintf("key for signing the request body hash, example: %q", exampleKey)
	hashKey := flag.String(flagHashKey, "", hashKeyUsage)

	rateLimitUsage := fmt.Sprintf("number of concurrent requests to the metrics server, example: %q",
		defaultRateLimit)
	rateLimit := flag.Int(flagRateLimit, 0, rateLimitUsage)

	publicKeyPathUsage := fmt.Sprintf("path to the file with rsa public key, example: %s",
		examplePublicKeyPath)
	publicKeyPath := flag.String(flagPublicKey, "", publicKeyPathUsage)

	var configPath string
	configPathUsage := fmt.Sprintf(", example: %s", exampleConfigPathUsage)
	flag.StringVar(&configPath, "config", "", configPathUsage)

	flag.Parse()

	if config := os.Getenv("CONFIG"); config != "" {
		configPath = config
	}

	if configPath != "" {
		err := cfg.GetConfigFromFile(configPath)
		if err != nil {
			fmt.Printf("can't get config from file: %s", err.Error())
		}
	}

	if flags.IsFlagPassed(flagEndpointHost) {
		cfg.EndpointHost = *endpointHost
	}

	if flags.IsFlagPassed(flagReportInterval) {
		cfg.ReportInterval = time.Duration(*reportInterval) * time.Second
	}

	if flags.IsFlagPassed(flagPollInterval) {
		cfg.PollInterval = time.Duration(*pollInterval) * time.Second
	}

	if flags.IsFlagPassed(flagHashKey) {
		cfg.HashKey = *hashKey
	}

	if flags.IsFlagPassed(flagPublicKey) {
		cfg.PublicKeyPath = *publicKeyPath
	}

	if flags.IsFlagPassed(flagRateLimit) {
		cfg.RateLimit = *rateLimit
	}

	// check report interval value
	if *reportInterval <= 0 {
		cfg.ReportInterval = defaultReportInterval * time.Second
	} else {
		cfg.ReportInterval = time.Duration(*reportInterval) * time.Second
	}

	// check poll interval value
	if *pollInterval <= 0 {
		cfg.PollInterval = defaultPollInterval * time.Second
	} else {
		cfg.PollInterval = time.Duration(*pollInterval) * time.Second
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
		cfg.HashKey = hashKey
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

	fmt.Printf("\nstart application with final config: %+v\n\n", cfg)

	return cfg
}

type FileConfig struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	HashKey        string `json:"hash_key"`
	CryptoKey      string `json:"crypto_key"`
	RateLimit      int    `json:"rate_limit"`
}

func (c *Config) GetConfigFromFile(filePath string) error {
	var fileConfig FileConfig

	file, err := os.OpenFile(filePath, os.O_RDONLY, defaultFilePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&fileConfig); err != nil {
		return err
	}

	fmt.Printf("config file: %+v\n", fileConfig)

	c.EndpointHost = fileConfig.Address
	c.HashKey = fileConfig.HashKey
	c.PublicKeyPath = fileConfig.CryptoKey
	c.RateLimit = fileConfig.RateLimit

	pollIntervalDuration, err := time.ParseDuration(fileConfig.PollInterval)
	if err != nil {
		c.PollInterval = defaultPollInterval
	} else {
		c.PollInterval = pollIntervalDuration
	}

	reportIntervalDuration, err := time.ParseDuration(fileConfig.ReportInterval)
	if err != nil {
		c.ReportInterval = defaultReportInterval
	} else {
		c.ReportInterval = reportIntervalDuration
	}

	fmt.Printf("config from file saved: %+v\n\n", c)

	return err
}
