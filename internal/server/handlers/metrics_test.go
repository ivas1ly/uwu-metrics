package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/server/entity"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/memory"
)

const (
	defaultTestClientTimeout = 3 * time.Second
)

func TestMetricsHandler(t *testing.T) {
	testStorage := NewTestStorage()
	logger := zap.Must(zap.NewDevelopment())
	router := chi.NewRouter()

	NewRoutes(router, testStorage, logger)

	ts := httptest.NewServer(router)
	defer ts.Close()

	type want struct {
		contentType string
		body        string
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
				body:        "",
			},
		},
		{
			name:   "with correct type, name and value / counter",
			path:   "/update/counter/uwu/123",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
				body:        "",
			},
		},
		{
			name:   "without metrics name / gauge",
			path:   "/update/gauge/",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				body:        "",
			},
		},
		{
			name:   "without metrics name / counter",
			path:   "/update/counter/",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				body:        "",
			},
		},
		{
			name:   "without metrics value / gauge",
			path:   "/update/gauge/uwu",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				body:        "",
			},
		},
		{
			name:   "without metrics value / counter",
			path:   "/update/counter/owo",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				body:        "",
			},
		},
		{
			name:   "with incorrect value / gauge",
			path:   "/update/gauge/uwu/wow",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				body:        "",
			},
		},
		{
			name:   "with incorrect value / counter",
			path:   "/update/counter/owo/wow",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				body:        "",
			},
		},
		{
			name:   "with incorrect value / counter",
			path:   "/update/counter/owo/123.456",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				body:        "",
			},
		},
		{
			name:   "with incorrect type",
			path:   "/update/hello/owo/123.456",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				body:        "",
			},
		},
		{
			name:   "method not allowed",
			path:   "/update/counter/owo/123",
			method: http.MethodPatch,
			want: want{
				contentType: "",
				statusCode:  http.StatusMethodNotAllowed,
				body:        "",
			},
		},
		{
			name:   "get metric value with correct name / counter",
			path:   "/value/counter/uwu",
			method: http.MethodGet,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
				body:        "123",
			},
		},
		{
			name:   "get metric value with correct name / gauge",
			path:   "/value/gauge/owo",
			method: http.MethodGet,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusOK,
				body:        "123.456",
			},
		},
		{
			name:   "get metric value with incorrect name",
			path:   "/value/counter/teeeeeeeeeeeeeest",
			method: http.MethodGet,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				body:        "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := testRequest(t, ts, test.method, test.path)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, test.want.statusCode, res.StatusCode)
			defer res.Body.Close()

			if test.want.body != "" {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				assert.Equal(t, test.want.body, string(resBody))
			}
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string) *http.Response {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestClientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, ts.URL+path, nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	return resp
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
