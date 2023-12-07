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
	defaultDatabaseDSN          = "postgres://postgres:postgres@localhost:5432/postgres?ssmode=disable"
	defaultDatabaseConnTimeout  = 3 * time.Second
	defaultDatabaseConnAttempts = 3
)

type Config struct {
	Endpoint        string
	FileStoragePath string
	DatabaseDSN     string
	StoreInterval   int
	FileRestore     bool
}

func NewConfig() *Config {
	cfg := &Config{}

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
	flag.BoolVar(&cfg.FileRestore, "r", defaultFileRestore, fileRestoreUsage)

	dsnUsage := fmt.Sprintf("PostgreSQL connection string, example: %q", defaultDatabaseDSN)
	flag.StringVar(&cfg.DatabaseDSN, "d", "", dsnUsage)

	flag.Parse()

	if endpoint := os.Getenv("ADDRESS"); endpoint != "" {
		cfg.Endpoint = endpoint
	}

	// check store interval value
	if *si <= 0 {
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

	if fileRestore := os.Getenv("RESTORE"); fileRestore != "" {
		envValue, err := strconv.ParseBool(fileRestore)
		if err == nil {
			cfg.FileRestore = envValue
		}
	}

	if databaseDSN := os.Getenv("DATABASE_DSN"); databaseDSN != "" {
		cfg.DatabaseDSN = databaseDSN
	}

	fmt.Printf("%+v\n\n", cfg)

	return cfg
}
