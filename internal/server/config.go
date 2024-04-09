package server

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/ivas1ly/uwu-metrics/internal/utils/flags"
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
	defaultFileRestore          = false
	defaultFilePerm             = 0666
	exampleDatabaseDSN          = "postgres://postgres:postgres@localhost:5432/metrics?sslmode=disable"
	defaultDatabaseConnTimeout  = 10 * time.Second
	defaultDatabaseConnAttempts = 3
	exampleKey                  = ""
	defaultPprofAddr            = "localhost:9090"
	examplePrivateKeyPath       = "./cmd/server/private_key.pem"
	exampleConfigPathUsage      = "./config/server.json"
)

const (
	flagEndpoint        = "a"
	flagStoreInterval   = "i"
	flagFileStoragePath = "f"
	flagFileRestore     = "r"
	flagDatabaseDSN     = "d"
	flagHashKey         = "k"
	flagPrivateKey      = "crypto-key"
)

// Config structure contains the received information for running the application.
type Config struct {
	Endpoint        string
	FileStoragePath string
	DatabaseDSN     string
	HashKey         string
	PrivateKeyPath  string
	StoreInterval   int
	Restore         bool
}

// NewConfig creates a new configuration depending on the method.
// It can be set either with command line flags or with environment variables.
// Environment variables take precedence over flags.
func NewConfig() Config {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	cfg := Config{
		Endpoint:        net.JoinHostPort(defaultHost, defaultPort),
		FileStoragePath: defaultFileStoragePath,
		DatabaseDSN:     "",
		HashKey:         "",
		PrivateKeyPath:  "",
		StoreInterval:   defaultStoreInterval,
		Restore:         false,
	}

	endpointUsage := fmt.Sprintf("HTTP server endpoint, example: %q or %q",
		net.JoinHostPort(defaultHost, defaultPort), net.JoinHostPort("", defaultPort))
	endpoint := flag.String(flagEndpoint, "", endpointUsage)

	storeIntervalUsage := fmt.Sprintf("time interval in seconds after which the server "+
		"saves all collected metrics data to disk, example: \"%d\"", defaultStoreInterval)
	storeInterval := flag.Int(flagStoreInterval, -1, storeIntervalUsage)

	fileStoragePathUsage := fmt.Sprintf("path to the file where the metrics will be read from and written to"+
		", example: %q", defaultFileStoragePath)
	fileStoragePath := flag.String(flagFileStoragePath, "", fileStoragePathUsage)

	fileRestoreUsage := fmt.Sprintf("load or not previously saved values from the specified file, "+
		"example: \"%t\"", defaultFileRestore)
	restore := flag.Bool(flagFileRestore, defaultFileRestore, fileRestoreUsage)

	dsnUsage := fmt.Sprintf("PostgreSQL connection string, example: %q", exampleDatabaseDSN)
	databaseDSN := flag.String(flagDatabaseDSN, "", dsnUsage)

	hashKeyUsage := fmt.Sprintf("key for checking the request hash and "+
		"computing the response body hash, example: %q", exampleKey)
	hashKey := flag.String(flagHashKey, "", hashKeyUsage)

	privateKeyPathUsage := fmt.Sprintf("path to the file with rsa private key, example: %s",
		examplePrivateKeyPath)
	privateKeyPath := flag.String(flagPrivateKey, "", privateKeyPathUsage)

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

	if flags.IsFlagPassed(flagEndpoint) {
		cfg.Endpoint = *endpoint
	}

	if flags.IsFlagPassed(flagFileStoragePath) {
		cfg.FileStoragePath = *fileStoragePath
	}

	if flags.IsFlagPassed(flagDatabaseDSN) {
		cfg.DatabaseDSN = *databaseDSN
	}

	if flags.IsFlagPassed(flagHashKey) {
		cfg.HashKey = *hashKey
	}

	if flags.IsFlagPassed(flagPrivateKey) {
		cfg.PrivateKeyPath = *privateKeyPath
	}

	if flags.IsFlagPassed(flagStoreInterval) {
		cfg.StoreInterval = *storeInterval
	}

	if flags.IsFlagPassed(flagFileRestore) {
		cfg.Restore = *restore
	}

	if endpoint := os.Getenv("ADDRESS"); endpoint != "" {
		cfg.Endpoint = endpoint
	}

	// check store interval value
	if *storeInterval < 0 {
		cfg.StoreInterval = defaultStoreInterval
	} else {
		cfg.StoreInterval = *storeInterval
	}

	if storeIntervalEnv := os.Getenv("STORE_INTERVAL"); storeIntervalEnv != "" {
		envValue, err := strconv.Atoi(storeIntervalEnv)
		if err == nil && envValue > -1 {
			cfg.StoreInterval = envValue
		}
	}

	if fileStoragePathEnv := os.Getenv("FILE_STORAGE_PATH"); fileStoragePathEnv != "" {
		cfg.FileStoragePath = fileStoragePathEnv
	}

	if restoreEnv := os.Getenv("RESTORE"); restoreEnv != "" {
		envValue, err := strconv.ParseBool(restoreEnv)
		if err == nil {
			cfg.Restore = envValue
		}
	}

	if databaseDSNEnv := os.Getenv("DATABASE_DSN"); databaseDSNEnv != "" {
		cfg.DatabaseDSN = databaseDSNEnv
	}

	if hashKeyEnv := os.Getenv("KEY"); hashKeyEnv != "" {
		cfg.HashKey = hashKeyEnv
	}

	if privateKeyPathEnv := os.Getenv("CRYPTO_KEY"); privateKeyPathEnv != "" {
		cfg.PrivateKeyPath = privateKeyPathEnv
	}

	fmt.Printf("\nstart application with final config: %+v\n\n", cfg)

	return cfg
}

type FileConfig struct {
	Address       string `json:"address"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	HashKey       string `json:"hash_key"`
	CryptoKey     string `json:"crypto_key"`
	StoreInterval string `json:"store_interval"`
	Restore       bool   `json:"restore"`
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

	c.Endpoint = fileConfig.Address
	c.FileStoragePath = fileConfig.StoreFile
	c.DatabaseDSN = fileConfig.DatabaseDSN
	c.HashKey = fileConfig.HashKey
	c.PrivateKeyPath = fileConfig.CryptoKey
	c.Restore = fileConfig.Restore

	seconds, err := time.ParseDuration(fileConfig.StoreInterval)
	if err != nil {
		c.StoreInterval = defaultStoreInterval
	} else {
		c.StoreInterval = int(seconds.Seconds())
	}

	fmt.Printf("config from file saved: %+v\n\n", c)

	return err
}
