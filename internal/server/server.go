package server

import (
	"log"
	"log/slog"
	"net/http"
	"os"
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

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	log.Printf("server started on port %s", addr)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
