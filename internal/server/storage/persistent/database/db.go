package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ivas1ly/uwu-metrics/internal/lib/postgres"
	"github.com/ivas1ly/uwu-metrics/internal/server/entity"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/memory"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/persistent"
)

const (
	saveGauge = `INSERT INTO metrics (id, mtype, mdelta, mvalue)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE SET mvalue = EXCLUDED.mvalue;`
	saveCounter = `INSERT INTO metrics (id, mtype, mdelta, mvalue)
VALUES ($1, $2, $3, $4) ON CONFLICT (id)
DO UPDATE SET mdelta = EXCLUDED.mdelta;`
	getMetrics = "SELECT id, mtype, mdelta, mvalue FROM metrics;"
)

type dbStorage struct {
	memoryStorage memory.Storage
	db            *postgres.DB
	timeout       time.Duration
}

func NewDBStorage(storage memory.Storage, db *postgres.DB, connTimeout time.Duration) persistent.Storage {
	return &dbStorage{
		memoryStorage: storage,
		db:            db,
		timeout:       connTimeout,
	}
}

func (ds *dbStorage) Save(ctx context.Context) error {
	metrics := ds.memoryStorage.GetMetrics()

	withTimeout, cancel := context.WithTimeout(ctx, ds.timeout)
	defer cancel()

	tx, err := ds.db.Pool.Begin(withTimeout)
	if err != nil {
		return err
	}
	defer func(tx pgx.Tx) {
		err = tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrTxClosed) {
			return
		}
	}(tx)

	batch := &pgx.Batch{}

	for id, metric := range metrics.Gauge {
		batch.Queue(saveGauge, id, "gauge", nil, metric)
	}

	for id, metric := range metrics.Counter {
		batch.Queue(saveCounter, id, "counter", metric, nil)
	}

	results := tx.SendBatch(ctx, batch)

	// check one affected row
	ct, err := results.Exec()
	if err != nil || ct.RowsAffected() != 1 {
		return err
	}

	err = results.Close()
	if err != nil {
		return err
	}

	err = tx.Commit(withTimeout)
	if err != nil {
		return err
	}

	return nil
}

func (ds *dbStorage) Restore(ctx context.Context) error {
	var metrics entity.Metrics
	metrics.Counter = make(map[string]int64)
	metrics.Gauge = make(map[string]float64)

	rows, err := ds.db.Pool.Query(ctx, getMetrics)
	if err != nil {
		return err
	}
	defer rows.Close()

	type Metric struct {
		mdelta *int64
		mvalue *float64
		id     string
		mtype  string
	}

	for rows.Next() {
		metric := Metric{}

		err = rows.Scan(
			&metric.id,
			&metric.mtype,
			&metric.mdelta,
			&metric.mvalue,
		)
		if err != nil {
			return err
		}

		if metric.mtype == entity.GaugeType {
			metrics.Gauge[metric.id] = *metric.mvalue
		}
		if metric.mtype == entity.CounterType {
			metrics.Counter[metric.id] = *metric.mdelta
		}
	}

	ds.memoryStorage.SetMetrics(metrics)

	return nil
}
