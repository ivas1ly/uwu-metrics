package storage

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/ivas1ly/uwu-metrics/internal/entity"
)

type Storage interface {
	Update(metric entity.Metric) error
	GetCounter(name string) (int64, error)
	GetGauge(name string) (float64, error)
	GetMetrics() entity.Metrics
}

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func NewMemStorage() Storage {
	return &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

func (ms *MemStorage) Update(metric entity.Metric) error {
	switch metric.Type {
	case "gauge":
		value, err := strconv.ParseFloat(metric.Value, 64)
		if err != nil {
			return fmt.Errorf("incorrect metric value: %w", err)
		}
		ms.gauge[metric.Name] = value
	case "counter":
		value, err := strconv.ParseInt(metric.Value, 10, 64)
		if err != nil {
			return fmt.Errorf("incorrect metric value: %w", err)
		}
		ms.counter[metric.Name] += value
	default:
		return errors.New("unknown metric type")
	}

	return nil
}

func (ms *MemStorage) GetMetrics() entity.Metrics {
	return entity.Metrics{Counter: ms.counter, Gauge: ms.gauge}
}

func (ms *MemStorage) GetCounter(name string) (int64, error) {
	counter, ok := ms.counter[name]
	if !ok {
		return 0, fmt.Errorf("counter metric %s doesn't exist", name)
	}
	return counter, nil
}

func (ms *MemStorage) GetGauge(name string) (float64, error) {
	gauge, ok := ms.gauge[name]
	if !ok {
		return 0, fmt.Errorf("gauge metric %s doesn't exist", name)
	}
	return gauge, nil
}
