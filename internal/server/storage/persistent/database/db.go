package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ivas1ly/uwu-metrics/internal/lib/postgres"
	"github.com/ivas1ly/uwu-metrics/internal/server/entity"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/memory"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/persistent"
)

// upsert https://www.postgresql.org/docs/current/sql-insert.html#SQL-ON-CONFLICT
const (
	//nolint:lll // const query
	saveGauge = "INSERT INTO metrics (id, mtype, mdelta, mvalue) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO UPDATE SET mvalue = EXCLUDED.mvalue;"
	//nolint:lll // const query
	saveCounter = "INSERT INTO metrics (id, mtype, mdelta, mvalue) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO UPDATE SET mdelta = EXCLUDED.mdelta;"
	getMetrics  = "SELECT id, mtype, mdelta, mvalue FROM metrics;"
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

func (ds *dbStorage) Save() error {
	metrics := ds.memoryStorage.GetMetrics()

	ctx := context.Background()

	withTimeout, cancel := context.WithTimeout(ctx, ds.timeout)
	defer cancel()

	batch := &pgx.Batch{}

	for id, metric := range metrics.Gauge {
		batch.Queue(saveGauge, id, "gauge", nil, metric)
	}

	for id, metric := range metrics.Counter {
		batch.Queue(saveCounter, id, "counter", metric, nil)
	}

	results := ds.db.Pool.SendBatch(withTimeout, batch)

	err := results.Close()
	if err != nil {
		return err
	}

	return nil
}

func (ds *dbStorage) Restore() error {
	var metrics entity.Metrics
	metrics.Counter = make(map[string]int64)
	metrics.Gauge = make(map[string]float64)

	ctx := context.Background()

	withTimeout, cancel := context.WithTimeout(ctx, ds.timeout)
	defer cancel()

	rows, err := ds.db.Pool.Query(withTimeout, getMetrics)
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

		err := rows.Scan(
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