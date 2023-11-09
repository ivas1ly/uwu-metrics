package server

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/ivas1ly/uwu-metrics/internal/handlers"
	"github.com/ivas1ly/uwu-metrics/internal/storage"
)

func Run(cfg *Config) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts)).
		With(slog.String("app", "server"))
	slog.SetDefault(logger)

	memStorage := storage.NewMemStorage()
	router := chi.NewRouter()

	handler := handlers.NewMetricsHandler(memStorage, logger)
	handler.NewRoutes(router)

	router.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("route not found :(", slog.String("path", r.URL.Path))
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

	log.Printf("server started on %s", cfg.Endpoint)
	err := server.ListenAndServe()
	if err != nil {
		// net/http recovers panic by default
		panic(err)
	}
}
