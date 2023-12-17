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
	"github.com/ivas1ly/uwu-metrics/internal/lib/migrate"
	"github.com/ivas1ly/uwu-metrics/internal/lib/postgres"
	"github.com/ivas1ly/uwu-metrics/internal/server/handlers"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/decompress"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/reqlogger"
	"github.com/ivas1ly/uwu-metrics/internal/server/middleware/writesync"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/memory"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/persistent"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/persistent/database"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/persistent/file"
)

func Run(cfg *Config) {
	log := logger.New(defaultLogLevel).
		With(zap.String("app", "server"))

	ctx := context.Background()

	withCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	ctxDB, cancelDB := context.WithTimeout(ctx, defaultDatabaseConnTimeout)
	defer cancelDB()

	memStorage := memory.NewMemStorage()

	var persistentStorage persistent.Storage
	if cfg.FileStoragePath != "" {
		persistentStorage = file.NewFileStorage(cfg.FileStoragePath, defaultFilePerm, memStorage)
	}

	var db *postgres.DB
	var err error
	if cfg.DatabaseDSN != "" {
		log.Info("received connection string", zap.String("connString", cfg.DatabaseDSN))

		err = migrate.RunMigrations(cfg.DatabaseDSN, defaultDatabaseConnAttempts, defaultDatabaseConnTimeout)
		if err != nil {
			log.Panic("can't run migrations", zap.Error(err))
		}
		log.Info("migrations up success", zap.String("status", "OK"))

		db, err = postgres.New(ctxDB, cfg.DatabaseDSN, defaultDatabaseConnAttempts, defaultDatabaseConnTimeout)
		if err != nil {
			log.Panic("can't connect to database", zap.Error(err))
		}
		defer db.Close()

		persistentStorage = database.NewDBStorage(memStorage, db, defaultDatabaseConnTimeout)
	}

	if cfg.Restore && cfg.FileStoragePath != "" && db == nil {
		if err = persistentStorage.Restore(ctxDB); err != nil {
			log.Info("failed to restore metrics from file, new file created", zap.String("error", err.Error()))
		} else {
			log.Info("file restored", zap.String("file", cfg.FileStoragePath))
		}
	}

	if db != nil && cfg.Restore {
		if err = persistentStorage.Restore(ctxDB); err != nil {
			log.Info("failed to restore metrics from database", zap.String("error", err.Error()))
		} else {
			log.Info("metrics restored from database")
		}
	}

	router := chi.NewRouter()

	router.Use(middleware.Compress(defaultCompressLevel))
	router.Use(decompress.New(log))
	router.Use(reqlogger.New(log))

	if cfg.StoreInterval == 0 {
		log.Info("all data will be saved synchronously", zap.Int("store interval", cfg.StoreInterval))
		router.Use(writesync.New(ctxDB, log, persistentStorage))
	}

	handlers.NewRoutes(router, memStorage, log)

	router.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("route not found :(", zap.String("path", r.URL.Path))
		http.NotFound(w, r)
	}))

	if cfg.FileStoragePath != "" && cfg.StoreInterval > 0 {
		go writeMetricsAsync(withCancel, persistentStorage, cfg.StoreInterval, log)
	}

	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		if db != nil {
			log.Info("check database connection")

			err = db.Ping(ctxDB)
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
	})

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	if err = runServer(ctx, cfg.Endpoint, router, log); err != nil {
		log.Info("unexpected server error", zap.Error(err))
	}

	// stop writeMetricsAsync job
	cancel()

	if err = persistentStorage.Save(ctxDB); err != nil {
		log.Info("can't save metrics before shutting down", zap.Error(err))
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

func writeMetricsAsync(ctx context.Context, storage persistent.Storage, interval int, log *zap.Logger) {
	saveTicker := time.NewTicker(time.Duration(interval) * time.Second)

	log.Info("start persist metrics job", zap.Int("interval", interval))

	withTimeout, cancel := context.WithTimeout(ctx, defaultDatabaseConnTimeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			log.Info("received done context")
			return
		case st := <-saveTicker.C:
			if err := storage.Save(withTimeout); err != nil {
				log.Info("[ERROR] ticker can't save metrics with interval",
					zap.Int("interval", interval), zap.Error(err))
			}
			log.Info("[OK] ticker metrics saved", zap.Time("saved at", st))
		}
	}
}
