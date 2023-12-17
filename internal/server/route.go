package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/lib/postgres"
	"github.com/ivas1ly/uwu-metrics/internal/server/handlers"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/decompress"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/reqlogger"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/memory"
)

func NewRouter(log *zap.Logger, ms memory.Storage) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Compress(defaultCompressLevel))
	router.Use(decompress.New(log))
	router.Use(reqlogger.New(log))

	handlers.NewRoutes(router, ms, log)

	router.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("route not found :(", zap.String("path", r.URL.Path))
		http.NotFound(w, r)
	}))

	return router
}

func pingDB(ctx context.Context, log *zap.Logger, db *postgres.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db != nil {
			log.Info("check database connection")

			err := db.Pool.Ping(ctx)
			if err != nil {
				log.Info("can't ping database", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			log.Info("database ping OK")
			w.WriteHeader(http.StatusOK)
		} else {
			log.Info("database connection string is empty, nothing to ping")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
