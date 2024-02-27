package server

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	defaultHost                 = "localhost"
	defaultPort                 = "8080"
	defaultReadTimeout          = 10 * time.Second
	defaultReadHeaderTimeout    = 5 * time.Second
	defaultWriteTimeout         = 10 * time.Second
	defaultIdleTimeout          = 1 * time.Minute
	defaultShutdownTimeout      = 5 * time.Second
	defaultCompressLevel        = 5
	defaultLogLevel             = "info"
	defaultStoreInterval        = 300
	defaultFileStoragePath      = "/tmp/metrics-db.json"
	defaultFileRestore          = true
	defaultFilePerm             = 0666
	exampleDatabaseDSN          = "postgres://postgres:postgres@localhost:5432/metrics?sslmode=disable"
	defaultDatabaseConnTimeout  = 10 * time.Second
	defaultDatabaseConnAttempts = 3
	exampleKey                  = ""
	defaultPprofAddr            = "localhost:9090"
)

// Config structure contains the received information for running the application.
type Config struct {
	Endpoint        string
	FileStoragePath string
	DatabaseDSN     string
	Key             string
	StoreInterval   int
	Restore         bool
}

// NewConfig creates a new configuration depending on the method.
// It can be set either with command line flags or with environment variables.
// Environment variables take precedence over flags.
func NewConfig() Config {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	cfg := Config{}

	endpointUsage := fmt.Sprintf("HTTP server endpoint, example: %q or %q",
		net.JoinHostPort(defaultHost, defaultPort), net.JoinHostPort("", defaultPort))
	flag.StringVar(&cfg.Endpoint, "a", net.JoinHostPort(defaultHost, defaultPort), endpointUsage)

	storeIntervalUsage := fmt.Sprintf("time interval in seconds after which the server "+
		"saves all collected metrics data to disk, example: \"%d\"", defaultStoreInterval)
	si := flag.Int("i", defaultStoreInterval, storeIntervalUsage)

	fileStoragePathUsage := fmt.Sprintf("path to the file where the metrics will be read from and written to"+
		", example: %q", defaultFileStoragePath)
	flag.StringVar(&cfg.FileStoragePath, "f", defaultFileStoragePath, fileStoragePathUsage)

	fileRestoreUsage := fmt.Sprintf("load or not previously saved values from the specified file, "+
		"example: \"%t\"", defaultFileRestore)
	flag.BoolVar(&cfg.Restore, "r", defaultFileRestore, fileRestoreUsage)

	dsnUsage := fmt.Sprintf("PostgreSQL connection string, example: %q", exampleDatabaseDSN)
	flag.StringVar(&cfg.DatabaseDSN, "d", "", dsnUsage)

	hashKeyUsage := fmt.Sprintf("key for checking the request hash and "+
		"computing the response body hash, example: %q", exampleKey)
	flag.StringVar(&cfg.Key, "k", "", hashKeyUsage)

	flag.Parse()

	if endpoint := os.Getenv("ADDRESS"); endpoint != "" {
		cfg.Endpoint = endpoint
	}

	// check store interval value
	if *si < 0 {
		cfg.StoreInterval = defaultStoreInterval
	} else {
		cfg.StoreInterval = *si
	}

	if storeInterval := os.Getenv("STORE_INTERVAL"); storeInterval != "" {
		envValue, err := strconv.Atoi(storeInterval)
		if err == nil && envValue > -1 {
			cfg.StoreInterval = envValue
		}
	}

	if fileStoragePath := os.Getenv("FILE_STORAGE_PATH"); fileStoragePath != "" {
		cfg.FileStoragePath = fileStoragePath
	}

	if restore := os.Getenv("RESTORE"); restore != "" {
		envValue, err := strconv.ParseBool(restore)
		if err == nil {
			cfg.Restore = envValue
		}
	}

	if databaseDSN := os.Getenv("DATABASE_DSN"); databaseDSN != "" {
		cfg.DatabaseDSN = databaseDSN
	}

	if hashKey := os.Getenv("KEY"); hashKey != "" {
		cfg.Key = hashKey
	}

	fmt.Printf("%+v\n\n", cfg)

	return cfg
}
