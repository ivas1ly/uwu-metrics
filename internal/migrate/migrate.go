package migrate

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	// register postgres driver.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/migrations"
)

func RunMigrations(connString string, connAttempts int, connTimeout time.Duration) error {
	var m *migrate.Migrate
	var err error

	dir, err := iofs.New(migrations.Migrations, ".")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	for connAttempts > 0 {
		if m, err = migrate.NewWithSourceInstance("iofs", dir, connString); err != nil {
			zap.L().Info("trying to connect, attempts left", zap.Int("attempts", connAttempts))
			time.Sleep(connTimeout)
			connAttempts--
			continue
		}
		break
	}
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("can't up migrations: %w", err)
	}
	defer m.Close()

	if errors.Is(err, migrate.ErrNoChange) {
		zap.L().Info("OK, no change to DB schema")
		return nil
	}

	return err
}
