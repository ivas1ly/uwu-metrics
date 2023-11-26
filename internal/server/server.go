package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/lib/logger"
	"github.com/ivas1ly/uwu-metrics/internal/server/handlers"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/decompress"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/reqlogger"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/memory"
)

func Run(cfg *Config) {
	log := logger.New(defaultLogLevel).
		With(zap.String("app", "server"))

	memStorage := memory.NewMemStorage()

	router := chi.NewRouter()

	router.Use(middleware.Compress(defaultCompressLevel))
	router.Use(decompress.New(log))
	router.Use(reqlogger.New(log))

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
