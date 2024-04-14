package server

import (
	"crypto/rsa"
	"net"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	pb "github.com/ivas1ly/uwu-metrics/pkg/api/metrics"
)

type MetricsService interface {
	UpsertMetric(mType, mName, mValue string) error
	GetMetric(mType, mName string) (*int64, *float64, error)
	GetAllMetrics() entity.Metrics
	UpsertTypeMetric(metric *entity.Metric) (*entity.Metric, error)
}

// NewRouter creates a new HTTP router and adds common middlewares for all handlers.
func NewRouter(metricsService MetricsService, db *postgres.DB, key string,
	privateKey *rsa.PrivateKey, trustedSubnet *net.IPNet, log *zap.Logger) *chi.Mux {
	router := chi.NewRouter()

	if trustedSubnet != nil {
		router.Use(checkip.New(log, trustedSubnet))
	}
	router.Use(middleware.Compress(defaultCompressLevel))
	router.Use(decompress.New(log))
	router.Use(rsadecrypt.New(log, privateKey))
	router.Use(reqlogger.New(log))
	if key != "" {
		router.Use(checkhash.New(log, []byte(key)))
		router.Use(sethash.New(log, []byte(key)))
	}

	handlers.NewRoutes(router, metricsService, log)

	router.Get("/ping", handlers.PingDB(db, log))

	return router
}

func NewgRPCServer(metricsService MetricsService, log *zap.Logger) *grpc.Server {
	server := grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
		grpc.UnaryInterceptor(reqlogger.NewInterceptor(log)),
	)

	reflection.Register(server)

	pb.RegisterMetricsServiceServer(server, gRPCHandlers.NewRoutes(metricsService, log))

	return server
}
