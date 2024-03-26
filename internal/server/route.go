package server

import (
	"crypto/rsa"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/lib/postgres"
	"github.com/ivas1ly/uwu-metrics/internal/server/handlers"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/checkhash"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/decompress"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/reqlogger"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/rsadecrypt"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/sethash"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/memory"
)

// NewRouter creates a new HTTP router and adds common middlewares for all handlers.
func NewRouter(ms memory.Storage, db *postgres.DB, key string, privateKey *rsa.PrivateKey, log *zap.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Compress(defaultCompressLevel))
	// an error occurs here, can't use "middleware" package name for my own middlewares
	router.Use(decompress.New(log))
	router.Use(rsadecrypt.New(log, privateKey))
	router.Use(reqlogger.New(log))
	if key != "" {
		router.Use(checkhash.New(log, []byte(key)))
		router.Use(sethash.New(log, []byte(key)))
	}

	handlers.NewRoutes(router, ms, log)

	router.Get("/ping", handlers.PingDB(db, log))

	return router
}
