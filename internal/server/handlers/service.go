package handlers

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/lib/postgres"
)

func PingDB(db *postgres.DB, log *zap.Logger) http.HandlerFunc {
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

func NotFound(log *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("route not found :(", zap.String("path", r.URL.Path))
		http.NotFound(w, r)
	}
}
