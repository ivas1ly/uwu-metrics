package memory

import (
	"fmt"

	"github.com/ivas1ly/uwu-metrics/internal/server/entity"
)

var _ Storage = (*memStorage)(nil)

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

func NewMemStorage() Storage {
	return &memStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

func (ms *memStorage) UpdateGauge(name string, value float64) {
	ms.gauge[name] = value
}

func (ms *memStorage) UpdateCounter(name string, value int64) {
	ms.counter[name] += value
}

func (ms *memStorage) GetMetrics() entity.Metrics {
	return entity.Metrics{Counter: ms.counter, Gauge: ms.gauge}
}

func (ms *memStorage) SetMetrics(metrics entity.Metrics) {
	ms.counter = metrics.Counter
	ms.gauge = metrics.Gauge
}

func (ms *memStorage) GetCounter(name string) (int64, error) {
	counter, ok := ms.counter[name]
	if !ok {
		return 0, fmt.Errorf("counter metric %q doesn't exist", name)
	}
	return counter, nil
}

func (ms *memStorage) GetGauge(name string) (float64, error) {
	gauge, ok := ms.gauge[name]
	if !ok {
		return 0, fmt.Errorf("gauge metric %q doesn't exist", name)
	}
	return gauge, nil
}
