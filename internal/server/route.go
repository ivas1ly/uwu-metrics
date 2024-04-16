package server

import (
	"crypto/rsa"
	"net"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Install the gzip compressor
	// https://github.com/grpc/grpc-go/blob/a4afd4d995b0e60b9beb7b54923fa74ef97a5098/examples/features/compression/server/main.go
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/reflection"

	"github.com/ivas1ly/uwu-metrics/internal/lib/postgres"
	"github.com/ivas1ly/uwu-metrics/internal/server/entity"
	gRPCHandlers "github.com/ivas1ly/uwu-metrics/internal/server/handlers/grpc"
	handlers "github.com/ivas1ly/uwu-metrics/internal/server/handlers/http"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/checkhash"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/checkip"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/decompress"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/reqlogger"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/rsadecrypt"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/sethash"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/writesync"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/persistent"
	"github.com/ivas1ly/uwu-metrics/internal/utils/rsakeys"
	pb "github.com/ivas1ly/uwu-metrics/pkg/api/metrics"
)

const unaryInterceptorsCap = 5

type MetricsService interface {
	UpsertMetric(mType, mName, mValue string) error
	GetMetric(mType, mName string) (*int64, *float64, error)
	GetAllMetrics() entity.Metrics
	UpsertTypeMetric(metric *entity.Metric) (*entity.Metric, error)
}

// NewRouter creates a new HTTP router and adds common middlewares for all handlers.
func NewRouter(metricsService MetricsService, persistentStorage persistent.Storage,
	db *postgres.DB, cfg Config, log *zap.Logger) *chi.Mux {
	router := chi.NewRouter()

	_, trustedSubnet, err := net.ParseCIDR(cfg.TrustedSubnet)
	if err != nil {
		log.Warn("can't parse trusted subnet CIDR")
	}

	if trustedSubnet != nil {
		router.Use(checkip.New(log, trustedSubnet))
	}

	router.Use(middleware.Compress(defaultCompressLevel))
	router.Use(decompress.New(log))

	var privateKey *rsa.PrivateKey
	if cfg.PrivateKeyPath != "" {
		privateKey, err = rsakeys.PrivateKey(cfg.PrivateKeyPath)
		if err != nil {
			log.Warn("can't get private key from file", zap.Error(err))
		}
		log.Info("private key successfully loaded")
		router.Use(rsadecrypt.New(log, privateKey))
	}

	router.Use(reqlogger.New(log))
	if cfg.HashKey != "" {
		router.Use(checkhash.New(log, []byte(cfg.HashKey)))
		router.Use(sethash.New(log, []byte(cfg.HashKey)))
	}

	if cfg.StoreInterval == 0 {
		log.Info("all data will be saved synchronously", zap.Int("store interval", cfg.StoreInterval))
		router.Use(writesync.New(persistentStorage, log))
	}

	handlers.NewRoutes(router, metricsService, log)

	router.Get("/ping", handlers.PingDB(db, log))

	return router
}

func NewgRPCServer(metricsService MetricsService, persistentStorage persistent.Storage,
	cfg Config, log *zap.Logger) *grpc.Server {
	unaryInterceptors := make([]grpc.UnaryServerInterceptor, 0, unaryInterceptorsCap)
	unaryInterceptors = append(unaryInterceptors,
		reqlogger.NewInterceptor(log),
	)

	if cfg.StoreInterval == 0 {
		log.Info("all data will be saved synchronously", zap.Int("store interval", cfg.StoreInterval))
		unaryInterceptors = append(unaryInterceptors, writesync.NewInterceptor(persistentStorage, log))
	}

	server := grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
	)

	reflection.Register(server)

	pb.RegisterMetricsServiceServer(server, gRPCHandlers.NewRoutes(metricsService, log))

	return server
}
