package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // exposed on a separate port that should be unavailable
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

func Run(cfg Config) {
	log := logger.New(defaultLogLevel, logger.NewDefaultLoggerConfig()).
		With(zap.String("app", "server"))

	ctx := context.Background()

	withCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	ctxDB, cancelDB := context.WithTimeout(ctx, defaultDatabaseConnTimeout)
	defer cancelDB()

	memStorage := memory.NewMemStorage()

	persistentStorage, db, err := setupPersistentStorage(ctxDB, cfg, memStorage, log)
	if err != nil {
		log.Info("can't setup persistent storage", zap.Error(err))
	}
	if db != nil {
		defer db.Close()
	}

	err = restoreMetrics(ctx, cfg, persistentStorage, db, log)
	if err != nil {
		log.Info("can't restore metrics from persistent storage", zap.Error(err))
	}

	router := NewRouter(memStorage, db, cfg.Key, log)

	if cfg.StoreInterval == 0 {
		log.Info("all data will be saved synchronously", zap.Int("store interval", cfg.StoreInterval))
		router.Use(writesync.New(persistentStorage, log))
	}

	if cfg.FileStoragePath != "" && cfg.StoreInterval > 0 {
		log.Info("all data will be saved asynchronously", zap.Int("store interval", cfg.StoreInterval))
		go writeMetricsAsync(withCancel, log, persistentStorage, cfg.StoreInterval)
	}

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
		log.Info("start pprof server")
		//nolint:gosec // use the default configuration for pprof
		if err := http.ListenAndServe(defaultPprofAddr, nil); err != nil {
			log.Fatal("pprof server", zap.Error(err))
		}
	}()

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

//nolint:whitespace //necessary leading newline, line is too long, otherwise warning from lll linter
func setupPersistentStorage(ctx context.Context, cfg Config,
	ms memory.Storage, log *zap.Logger) (ps persistent.Storage, db *postgres.DB, err error) {

	if cfg.DatabaseDSN != "" {
		return newDBStorage(ctx, cfg.DatabaseDSN, ms, log)
	}

	if cfg.FileStoragePath != "" {
		return newFileStorage(cfg.FileStoragePath, ms, log), nil, nil
	}

	return nil, nil, fmt.Errorf("can't setup persistent storage")
}

func newFileStorage(fileStoragePath string, ms memory.Storage, log *zap.Logger) persistent.Storage {
	log.Info("all data will be saved to file")
	persistentStorage := file.NewFileStorage(fileStoragePath, defaultFilePerm, ms)
	return persistentStorage
}

func newDBStorage(ctx context.Context, databaseDSN string, ms memory.Storage,
	log *zap.Logger) (persistent.Storage, *postgres.DB, error) {
	var db *postgres.DB
	var err error

	log.Info("received connection string", zap.String("connString", databaseDSN))

	err = migrate.RunMigrations(databaseDSN, defaultDatabaseConnAttempts, defaultDatabaseConnTimeout)
	if err != nil {
		log.Info("can't run migrations", zap.Error(err))
		return nil, nil, err
	}
	log.Info("migrations up success", zap.String("status", "OK"))

	db, err = postgres.New(ctx, databaseDSN, defaultDatabaseConnAttempts, defaultDatabaseConnTimeout)
	if err != nil {
		log.Info("can't connect to database", zap.Error(err))
		return nil, nil, err
	}

	err = db.Pool.Ping(ctx)
	if err != nil {
		log.Info("can't ping database", zap.Error(err))
		return nil, nil, err
	}

	log.Info("all data will be saved to database")
	persistentStorage := database.NewDBStorage(ms, db, defaultDatabaseConnTimeout)

	return persistentStorage, db, nil
}

func restoreMetrics(ctx context.Context, cfg Config, ps persistent.Storage, db *postgres.DB, log *zap.Logger) error {
	if cfg.Restore && cfg.FileStoragePath != "" && db == nil {
		if err := ps.Restore(ctx); err != nil {
			log.Info("failed to restore metrics from file, new file created", zap.String("error", err.Error()))
			return err
		}
		log.Info("file restored", zap.String("file", cfg.FileStoragePath))
	}

	if db != nil && cfg.Restore {
		if err := ps.Restore(ctx); err != nil {
			log.Info("failed to restore metrics from database", zap.String("error", err.Error()))
			return err
		}
		log.Info("metrics restored from database")
	}
	return nil
}
