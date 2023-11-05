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
	handler.NewMetricsRoutes(router)

	router.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("route not found :(", slog.String("path", r.URL.Path))
		http.NotFound(w, r)
	}))

	log.Printf("server started on %s", cfg.Endpoint)
	err := http.ListenAndServe(cfg.Endpoint, router)
	if err != nil {
		// net/http recovers panic by default
		panic(err)
	}
}
