package memory

import (
	"fmt"

	"github.com/ivas1ly/uwu-metrics/internal/server/entity"
)

// A construct to verify the implementation of an interface.
var _ Storage = (*memStorage)(nil)

// Storage is the interface that groups the in-memory storage methods.
type Storage interface {
	UpdateCounter(name string, value int64)
	UpdateGauge(name string, value float64)
	GetCounter(name string) (int64, error)
	GetGauge(name string) (float64, error)
	GetMetrics() entity.Metrics
	SetMetrics(metrics entity.Metrics)
}

type memStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

// NewMemStorage creates a new temporary storage in-memory for gauges and counter metrics.
func NewMemStorage() Storage {
	return &memStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

// UpdateGauge adds a new metric of type gauge to the storage.
func (ms *memStorage) UpdateGauge(name string, value float64) {
	ms.gauge[name] = value
}

// UpdateCounter adds a new metric of type counter to the storage.
func (ms *memStorage) UpdateCounter(name string, value int64) {
	ms.counter[name] += value
}

// GetMetrics gets all metrics from in-memory storage.
func (ms *memStorage) GetMetrics() entity.Metrics {
	return entity.Metrics{Counter: ms.counter, Gauge: ms.gauge}
}

// SetMetrics sets all metrics to in-memory storage.
func (ms *memStorage) SetMetrics(metrics entity.Metrics) {
	ms.counter = metrics.Counter
	ms.gauge = metrics.Gauge
}

// GetCounter gets a metric of type counter by its name.
func (ms *memStorage) GetCounter(name string) (int64, error) {
	counter, ok := ms.counter[name]
	if !ok {
		return 0, fmt.Errorf("counter metric %q doesn't exist", name)
	}
	return counter, nil
}

// GetGauge gets a metric of type gauge by its name.
func (ms *memStorage) GetGauge(name string) (float64, error) {
	gauge, ok := ms.gauge[name]
	if !ok {
		return 0, fmt.Errorf("gauge metric %q doesn't exist", name)
	}
	return gauge, nil
}
