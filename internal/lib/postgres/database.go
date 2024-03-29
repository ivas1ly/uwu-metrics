package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type DB struct {
	*pgxpool.Pool
}

// New creates a connection to a PostgreSQL database.
func New(ctx context.Context, connString string, connAttempts int, connTimeout time.Duration) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("can't create new connection pool: %w", err)
	}

	err = retryWithAttempts(func() error {
		if err = pool.Ping(ctx); err != nil {
			return fmt.Errorf("failed to ping database: %w", err)
		}
		return nil
	}, connAttempts, connTimeout)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &DB{
		Pool: pool,
	}, nil
}

func retryWithAttempts(fn func() error, connAttempts int, connTimeout time.Duration) error {
	var err error

	for connAttempts > 0 {
		if err = fn(); err != nil {
			zap.L().Info("trying to connect Postgres...", zap.Int("attempts left", connAttempts))
			time.Sleep(connTimeout)
			connAttempts--
			continue
		}
		return nil
	}

	return err
}
