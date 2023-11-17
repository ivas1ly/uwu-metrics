package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/lib/logger"
	"github.com/ivas1ly/uwu-metrics/internal/server/handlers"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage"
)

func Run(cfg *Config) {
	log := logger.New(defaultLogLevel).
		With(zap.String("app", "server"))

	memStorage := storage.NewMemStorage()
	router := chi.NewRouter()
	handlers.NewRoutes(router, memStorage, log)

	router.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("route not found :(", zap.String("path", r.URL.Path))
		http.NotFound(w, r)
	}))

	server := &http.Server{
		Addr:              cfg.Endpoint,
		Handler:           router,
		ReadTimeout:       defaultReadTimeout,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
		WriteTimeout:      defaultWriteTimeout,
		IdleTimeout:       defaultIdleTimeout,
	}

	log.Info("server started", zap.String("addr", cfg.Endpoint))
	err := server.ListenAndServe()
	if err != nil {
		// net/http recovers panic by default
		panic(err)
	}
}
