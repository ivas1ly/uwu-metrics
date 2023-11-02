package server

import (
	"log"
	"log/slog"
	"net/http"
	"os"

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
	mux := http.NewServeMux()
	handler := handlers.NewMetricsHandler(memStorage, logger)
	handler.NewMetricsRoutes(mux)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	log.Printf("server started on port %s", addr)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		// net/http recovers panic by default
		panic(err)
	}
}
