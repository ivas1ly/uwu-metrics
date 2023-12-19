package server

import (
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

func NewRouter(ms memory.Storage, db *postgres.DB, log *zap.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Compress(defaultCompressLevel))
	// an error occurs here, can't use "middleware" package name for my own middlewares
	router.Use(decompress.New(log))
	router.Use(reqlogger.New(log))

	handlers.NewRoutes(router, ms, log)

	router.Get("/ping", pingDB(log, db))

	router.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("route not found :(", zap.String("path", r.URL.Path))
		http.NotFound(w, r)
	}))

	return router
}

func pingDB(log *zap.Logger, db *postgres.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db != nil {
			log.Info("check database connection")

			err := db.Pool.Ping(r.Context())
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
