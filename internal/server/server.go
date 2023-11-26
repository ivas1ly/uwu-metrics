package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/lib/logger"
	"github.com/ivas1ly/uwu-metrics/internal/server/handlers"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/decompress"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/reqlogger"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/writesync"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/memory"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/persistent/file"
)

func Run(cfg *Config) {
	log := logger.New(defaultLogLevel).
		With(zap.String("app", "server"))

	ctx := context.Background()
	withCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	memStorage := memory.NewMemStorage()

	var persistentStorage file.PersistentStorage
	if cfg.FileStoragePath != "" {
		persistentStorage = file.NewFileStorage(cfg.FileStoragePath, defaultFilePerm, memStorage)
	}

	if cfg.FileRestore && cfg.FileStoragePath != "" {
		if err := persistentStorage.Restore(); err != nil {
			log.Info("failed to restore metrics from file, new file created", zap.String("error", err.Error()))
		} else {
			log.Info("file restored", zap.String("file", cfg.FileStoragePath))
		}
	}

	router := chi.NewRouter()

	router.Use(middleware.Compress(defaultCompressLevel))
	router.Use(decompress.New(log))
	router.Use(reqlogger.New(log))

	if cfg.StoreInterval == 0 {
		log.Info("all data will be saved synchronously", zap.Int("store interval", cfg.StoreInterval))
		router.Use(writesync.New(log, persistentStorage))
	}

	handlers.NewRoutes(router, memStorage, log)

	router.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("route not found :(", zap.String("path", r.URL.Path))
		http.NotFound(w, r)
	}))

	if cfg.FileStoragePath != "" && cfg.StoreInterval > 0 {
		go writeMetricsAsync(withCancel, persistentStorage, cfg.StoreInterval, log)
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	if err := runServer(ctx, cfg.Endpoint, router, log); err != nil {
		log.Info("unexpected server error", zap.Error(err))
	}

	// stop writeMetricsAsync job
	cancel()

	if err := persistentStorage.Save(); err != nil {
		log.Error("can't save metrics before shutting down", zap.Error(err))
	} else {
		log.Info("all metrics saved successfully")
	}
}

func runServer(ctx context.Context, endpoint string, router *chi.Mux, log *zap.Logger) error {
	server := &http.Server{
		Addr:              endpoint,
		Handler:           router,
		ReadTimeout:       defaultReadTimeout,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
		WriteTimeout:      defaultWriteTimeout,
		IdleTimeout:       defaultIdleTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("unexpected server error", zap.Error(err))
		}
	}()

	log.Info("server started", zap.String("addr", endpoint))
	<-ctx.Done()

	log.Info("gracefully shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	go func() {
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Info("unexpected server shutdown error", zap.Error(err))
		}
	}()

	<-shutdownCtx.Done()
	if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
		log.Info("timeout exceeded, forcing shutdown")
		return shutdownCtx.Err()
	}

	return nil
}

func writeMetricsAsync(ctx context.Context, storage file.PersistentStorage, interval int, log *zap.Logger) {
	saveTicker := time.NewTicker(time.Duration(interval) * time.Second)

	log.Info("start persist metrics job", zap.Int("interval", interval))

	for {
		select {
		case <-ctx.Done():
			log.Info("received done context")
			return
		case st := <-saveTicker.C:
			if err := storage.Save(); err != nil {
				log.Error("can't save metrics with interval",
					zap.Int("interval", interval), zap.Error(err))
			}
			log.Info("metrics saved", zap.Time("saved at", st))
		}
	}
}
