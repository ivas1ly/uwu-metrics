package storage

import (
	"log"

	"github.com/ivas1ly/uwu-metrics/internal/metrics"
)

type Storage interface {
	Update(name string, metric metrics.Metric)
	Get()
}

type MemStorage struct {
	collection map[string]metrics.Metric
}

func NewMemStorage() Storage {
	return &MemStorage{
		collection: make(map[string]metrics.Metric),
	}
}

func (ms *MemStorage) Update(name string, metrics metrics.Metric) {
	switch metrics.Type {
	case "gauge":
		ms.collection[name] = metrics
	case "counter":
		if value, ok := ms.collection[name]; ok {
			newCounter := metrics.Counter + value.Counter
			metrics.Counter = newCounter
			ms.collection[name] = metrics
		} else {
			ms.collection[name] = metrics
		}
	}
	ms.Get()
}

func (ms *MemStorage) Get() {
	log.Printf("collection: %+v\n", ms.collection)
}
