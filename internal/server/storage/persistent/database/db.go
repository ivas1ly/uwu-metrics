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

/*
EXPLAIN ANALYZE
INSERT INTO metrics (id, mtype, mdelta, mvalue)
VALUES ('OtherSys', 'gauge', null, 1010)
ON CONFLICT (id) DO UPDATE SET mvalue = EXCLUDED.mvalue;
                                          QUERY PLAN
-----------------------------------------------------------------------------------------------
 Insert on metrics  (cost=0.00..0.01 rows=0 width=0) (actual time=0.074..0.074 rows=0 loops=1)
   Conflict Resolution: UPDATE
   Conflict Arbiter Indexes: metrics_pkey
   Tuples Inserted: 0
   Conflicting Tuples: 1
   ->  Result  (cost=0.00..0.01 rows=1 width=80) (actual time=0.001..0.002 rows=1 loops=1)
 Planning Time: 0.069 ms
 Execution Time: 0.089 ms
(8 rows)

EXPLAIN ANALYZE
INSERT INTO metrics (id, mtype, mdelta, mvalue)
VALUES ('PollCount', 'counter', 15, null)
ON CONFLICT (id) DO UPDATE SET mdelta = EXCLUDED.mdelta;
                                          QUERY PLAN
-----------------------------------------------------------------------------------------------
 Insert on metrics  (cost=0.00..0.01 rows=0 width=0) (actual time=0.071..0.071 rows=0 loops=1)
   Conflict Resolution: UPDATE
   Conflict Arbiter Indexes: metrics_pkey
   Tuples Inserted: 0
   Conflicting Tuples: 1
   ->  Result  (cost=0.00..0.01 rows=1 width=80) (actual time=0.002..0.002 rows=1 loops=1)
 Planning Time: 0.117 ms
 Execution Time: 0.087 ms
(8 rows)
*/

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

// NewDBStorage creates new persistent storage in the database.
func NewDBStorage(storage memory.Storage, db *postgres.DB, connTimeout time.Duration) persistent.Storage {
	return &dbStorage{
		memoryStorage: storage,
		db:            db,
		timeout:       connTimeout,
	}
}

// Save takes the metrics from memory and saves them to the database.
func (ds *dbStorage) Save(ctx context.Context) error {
	metrics := ds.memoryStorage.GetMetrics()

	tx, err := ds.db.Pool.Begin(ctx)
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

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Restore fetches the last saved metrics from the database and restores them to in-memory storage.
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
