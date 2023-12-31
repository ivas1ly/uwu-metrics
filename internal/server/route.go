package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/lib/postgres"
	"github.com/ivas1ly/uwu-metrics/internal/server/handlers"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/decompress"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/reqlogger"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/memory"
)

func NewRouter(ms memory.Storage, db *postgres.DB, log *zap.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Compress(defaultCompressLevel))
	// an error occurs here, can't use "middleware" package name for my own middlewares
	router.Use(decompress.New(log))
	router.Use(reqlogger.New(log))

	handlers.NewRoutes(router, ms, log)

	router.Get("/ping", handlers.PingDB(db, log))

	router.Handle("/*", handlers.NotFound(log))

	return router
}
