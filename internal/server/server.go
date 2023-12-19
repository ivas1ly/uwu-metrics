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
	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/lib/logger"
	"github.com/ivas1ly/uwu-metrics/internal/lib/postgres"
	"github.com/ivas1ly/uwu-metrics/internal/migrate"
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

	persistentStorage, db := SetupPersistentStorage(ctxDB, log, cfg, memStorage)
	if db != nil {
		defer db.Close()
	}

	RestoreMetrics(ctx, log, cfg, persistentStorage, db)

	router := NewRouter(log, memStorage)

	if cfg.StoreInterval == 0 {
		log.Info("all data will be saved synchronously", zap.Int("store interval", cfg.StoreInterval))
		router.Use(writesync.New(ctx, log, persistentStorage))
	}

	if cfg.FileStoragePath != "" && cfg.StoreInterval > 0 {
		log.Info("all data will be saved asynchronously", zap.Int("store interval", cfg.StoreInterval))
		go writeMetricsAsync(withCancel, log, persistentStorage, cfg.StoreInterval)
	}

	router.Get("/ping", pingDB(ctx, log, db))

	notifyCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	if err := runServer(notifyCtx, cfg.Endpoint, router, log); err != nil {
		log.Info("unexpected server error", zap.Error(err))
	}

	// stop writeMetricsAsync job
	cancel()

	if err := persistentStorage.Save(ctx); err != nil {
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

func writeMetricsAsync(ctx context.Context, log *zap.Logger, storage persistent.Storage, interval int) {
	saveTicker := time.NewTicker(time.Duration(interval) * time.Second)

	log.Info("start persist metrics job", zap.Int("interval", interval))

	for {
		select {
		case <-ctx.Done():
			log.Info("received done context")
			return
		case st := <-saveTicker.C:
			err := storage.Save(ctx)
			if err != nil {
				log.Info("[ERROR] ticker can't save metrics with interval",
					zap.Int("interval", interval), zap.Error(err))
				continue
			}
			log.Info("[OK] ticker metrics saved", zap.Time("saved at", st))
		}
	}
}

func SetupPersistentStorage(ctx context.Context, log *zap.Logger, cfg *Config,
	ms memory.Storage) (persistent.Storage, *postgres.DB) {
	var persistentStorage persistent.Storage

	if cfg.FileStoragePath != "" {
		log.Info("all data will be saved to file")
		persistentStorage = file.NewFileStorage(cfg.FileStoragePath, defaultFilePerm, ms)
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

		db, err = postgres.New(ctx, cfg.DatabaseDSN, defaultDatabaseConnAttempts, defaultDatabaseConnTimeout)
		if err != nil {
			log.Panic("can't connect to database", zap.Error(err))
		}

		err = db.Ping(ctx)
		if err != nil {
			log.Panic("can't ping database", zap.Error(err))
		}

		log.Info("all data will be saved to database")
		persistentStorage = database.NewDBStorage(ms, db, defaultDatabaseConnTimeout)
	}

	return persistentStorage, db
}

func RestoreMetrics(ctx context.Context, log *zap.Logger, cfg *Config, ps persistent.Storage, db *postgres.DB) {
	if cfg.Restore && cfg.FileStoragePath != "" && db == nil {
		if err := ps.Restore(ctx); err != nil {
			log.Info("failed to restore metrics from file, new file created", zap.String("error", err.Error()))
		} else {
			log.Info("file restored", zap.String("file", cfg.FileStoragePath))
		}
	}

	if db != nil && cfg.Restore {
		if err := ps.Restore(ctx); err != nil {
			log.Info("failed to restore metrics from database", zap.String("error", err.Error()))
		} else {
			log.Info("metrics restored from database")
		}
	}
}
