package storage

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/ivas1ly/uwu-metrics/internal/metrics"
)

type Storage interface {
	Update(metric metrics.Metric) error
	Get()
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

func (ms *MemStorage) Update(metric metrics.Metric) error {
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
		if _, ok := ms.counter[metric.Name]; ok {
			ms.counter[metric.Name] += value
		} else {
			ms.counter[metric.Name] = value
		}
	default:
		return errors.New("unknown metric type")
	}

	ms.Get()

	return nil
}

func (ms *MemStorage) Get() {
	log.Printf("collection counter: %+v\n", ms.counter)
	log.Printf("collection gauge: %+v\n", ms.gauge)
}
