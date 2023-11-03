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

	"github.com/stretchr/testify/assert"

	"github.com/ivas1ly/uwu-metrics/internal/metrics"
	"github.com/ivas1ly/uwu-metrics/internal/storage"
)

func TestMetricsHandler(t *testing.T) {
	testStorage := NewTestStorage()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handlers := MetricsHandler{
		storage: testStorage,
		logger:  logger,
	}

	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name    string
		request string
		method  string
		want    want
	}{
		{
			name:    "with correct type, name and value / gauge",
			request: "/update/gauge/owo/123.456",
			method:  http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
		{
			name:    "with correct type, name and value / counter",
			request: "/update/counter/uwu/123",
			method:  http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
			},
		},
		{
			name:    "without metrics name / gauge",
			request: "/update/gauge/",
			method:  http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name:    "without metrics name / counter",
			request: "/update/counter/",
			method:  http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name:    "without metrics value / gauge",
			request: "/update/gauge/uwu",
			method:  http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name:    "without metrics value / counter",
			request: "/update/counter/owo",
			method:  http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name:    "with incorrect value / gauge",
			request: "/update/gauge/uwu/wow",
			method:  http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name:    "with incorrect value / counter",
			request: "/update/counter/owo/wow",
			method:  http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name:    "with incorrect value / counter",
			request: "/update/counter/owo/123.456",
			method:  http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name:    "with incorrect type",
			request: "/update/hello/owo/123.456",
			method:  http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name:    "method not allowed",
			request: "/update/counter/owo/123",
			method:  http.MethodPatch,
			want: want{
				contentType: "",
				statusCode:  http.StatusMethodNotAllowed,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.request, nil)
			request.Header.Set("Content-Type", "text/plain")
			h := http.StripPrefix("/update/", http.HandlerFunc(handlers.update))

			nr := httptest.NewRecorder()
			h.ServeHTTP(nr, request)

			res := nr.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, test.want.statusCode, res.StatusCode)
		})
	}
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
