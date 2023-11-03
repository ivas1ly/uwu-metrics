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

const (
	addr = ":8080"
)

func Run() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	memStorage := storage.NewMemStorage()
	router := chi.NewRouter()

	handler := handlers.NewMetricsHandler(memStorage, logger)
	handler.NewMetricsRoutes(router)

	router.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("route not found :(", slog.String("path", r.URL.Path))
		http.NotFound(w, r)
	}))

	log.Printf("server started on port %s", addr)
	err := http.ListenAndServe(addr, router)
	if err != nil {
		// net/http recovers panic by default
		panic(err)
	}
}
