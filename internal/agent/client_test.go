package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/server/entity"
	"github.com/ivas1ly/uwu-metrics/internal/server/handlers"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/memory"
)

const endpoint = "/updates/"

func TestClientSendRequest(t *testing.T) {
	storage := NewTestStorage()
	logger := zap.Must(zap.NewDevelopment())
	router := chi.NewRouter()

	handlers.NewRoutes(router, storage, logger)

	ts := httptest.NewServer(router)
	defer ts.Close()

	metrics := &Metrics{}
	metrics.UpdateMetrics()

	client := Client{
		Metrics: metrics,
		Logger:  zap.Must(zap.NewDevelopment()),
		URL:     ts.URL + endpoint,
	}

	payload := make([]MetricsPayload, 0, defaultPayloadCap)

	for key, value := range client.Metrics.PrepareGaugeReport() {
		val := value

		mp := MetricsPayload{
			ID:    key,
			MType: entity.GaugeType,
			Delta: nil,
			Value: &val,
		}

		payload = append(payload, mp)
	}

	for key, value := range client.Metrics.PrepareCounterReport() {
		val := value
		mp := MetricsPayload{
			ID:    key,
			MType: entity.CounterType,
			Delta: &val,
			Value: nil,
		}

		payload = append(payload, mp)
	}

	b, err := json.Marshal(&payload)
	assert.NoError(t, err)

	err = client.sendRequest(http.MethodPost, b)
	assert.NoError(t, err)
}

func TestClientSendReport(t *testing.T) {
	storage := NewTestStorage()
	logger := zap.Must(zap.NewDevelopment())
	router := chi.NewRouter()

	handlers.NewRoutes(router, storage, logger)

	ts := httptest.NewServer(router)
	defer ts.Close()

	metrics := &Metrics{}
	metrics.UpdateMetrics()

	client := Client{
		Metrics: metrics,
		Logger:  zap.Must(zap.NewDevelopment()),
		URL:     ts.URL + endpoint,
	}

	assert.NoError(t, client.SendReport())
}

type testStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func NewTestStorage() memory.Storage {
	return &testStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

func (ts *testStorage) UpdateGauge(name string, value float64) {
	ts.gauge[name] = value
}

func (ts *testStorage) UpdateCounter(name string, value int64) {
	ts.counter[name] += value
}

func (ts *testStorage) GetMetrics() entity.Metrics {
	return entity.Metrics{Counter: ts.counter, Gauge: ts.gauge}
}

func (ts *testStorage) SetMetrics(metrics entity.Metrics) {
	ts.counter = metrics.Counter
	ts.gauge = metrics.Gauge
}

func (ts *testStorage) GetCounter(name string) (int64, error) {
	counter, ok := ts.counter[name]
	if !ok {
		return 0, fmt.Errorf("counter metric %s doesn't exist", name)
	}
	return counter, nil
}

func (ts *testStorage) GetGauge(name string) (float64, error) {
	gauge, ok := ts.gauge[name]
	if !ok {
		return 0, fmt.Errorf("gauge metric %s doesn't exist", name)
	}
	return gauge, nil
}
