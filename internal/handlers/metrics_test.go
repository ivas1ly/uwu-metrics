package handlers

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ivas1ly/uwu-metrics/internal/metrics"
	"github.com/ivas1ly/uwu-metrics/internal/storage"
)

func TestMetricsHandler(t *testing.T) {
	testStorage := NewTestStorage()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	router := chi.NewRouter()

	handlers := MetricsHandler{
		storage: testStorage,
		logger:  logger,
	}

	handlers.NewMetricsRoutes(router)

	ts := httptest.NewServer(router)
	defer ts.Close()

	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name   string
		path   string
		method string
		want   want
	}{
		{
			name:   "with correct type, name and value / gauge",
			path:   "/update/gauge/owo/123.456",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
		{
			name:   "with correct type, name and value / counter",
			path:   "/update/counter/uwu/123",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
		{
			name:   "without metrics name / gauge",
			path:   "/update/gauge/",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name:   "without metrics name / counter",
			path:   "/update/counter/",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name:   "without metrics value / gauge",
			path:   "/update/gauge/uwu",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name:   "without metrics value / counter",
			path:   "/update/counter/owo",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name:   "with incorrect value / gauge",
			path:   "/update/gauge/uwu/wow",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name:   "with incorrect value / counter",
			path:   "/update/counter/owo/wow",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name:   "with incorrect value / counter",
			path:   "/update/counter/owo/123.456",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name:   "with incorrect type",
			path:   "/update/hello/owo/123.456",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name:   "method not allowed",
			path:   "/update/counter/owo/123",
			method: http.MethodPatch,
			want: want{
				contentType: "",
				statusCode:  http.StatusMethodNotAllowed,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := testRequest(t, ts, test.method, test.path)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, test.want.statusCode, res.StatusCode)
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string) *http.Response {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	return resp
}

type TestStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func NewTestStorage() storage.Storage {
	return &TestStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

func (ts *TestStorage) Update(metric metrics.Metric) error {
	switch metric.Type {
	case "gauge":
		value, err := strconv.ParseFloat(metric.Value, 64)
		if err != nil {
			return fmt.Errorf("incorrect metric value: %w", err)
		}
		ts.gauge[metric.Name] = value
	case "counter":
		value, err := strconv.ParseInt(metric.Value, 10, 64)
		if err != nil {
			return fmt.Errorf("incorrect metric value: %w", err)
		}
		if _, ok := ts.counter[metric.Name]; ok {
			ts.counter[metric.Name] += value
		} else {
			ts.counter[metric.Name] = value
		}
	default:
		return errors.New("unknown metric type")
	}

	ts.Get()

	return nil
}

func (ts *TestStorage) Get() {
	log.Printf("collection counter: %+v\n", ts.counter)
	log.Printf("collection gauge: %+v\n", ts.gauge)
}
