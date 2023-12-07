package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func New(ctx context.Context, connString string, connAttempts int, connTimeout time.Duration) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("can't create new connection pool: %w", err)
	}

	withTimeout, cancel := context.WithTimeout(ctx, connTimeout)
	defer cancel()

	err = DoWithAttempts(func() error {
		if err = pool.Ping(withTimeout); err != nil {
			return fmt.Errorf("failed to ping database: %w", err)
		}
		return nil
	}, connAttempts, connTimeout)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return pool, nil
}

func DoWithAttempts(fn func() error, connAttempts int, connTimeout time.Duration) error {
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
