package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

//nolint:funlen // test func
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
			name:   "update with empty name",
			path:   "/update/counter//123",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				body:        "",
			},
		},
		{
			name:   "update with empty type",
			path:   "/update//test/123",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				body:        "",
			},
		},
		{
			name:   "update with empty value",
			path:   "/update/counter/",
			method: http.MethodPost,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				body:        "",
			},
		},
		{
			name:   "value with empty name",
			path:   "/value/counter//",
			method: http.MethodGet,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				body:        "",
			},
		},
		{
			name:   "value with empty type",
			path:   "/value//test",
			method: http.MethodGet,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
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
		{
			name:   "get metrics webpage",
			path:   "/",
			method: http.MethodGet,
			want: want{
				contentType: "text/html; charset=utf-8",
				statusCode:  http.StatusOK,
				body:        "",
			},
		},
		{
			name:   "value unknown metric type",
			path:   "/value/unknown/test",
			method: http.MethodGet,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				body:        "",
			},
		},
		{
			name:   "route not found",
			path:   "//",
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
			res := testRequest(t, ts, test.method, test.path, nil, "text/plain")
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

//nolint:funlen // test func
func TestMetricsHandlerJSON(t *testing.T) {
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
		body   string
		want   want
	}{
		{
			name:   "update with correct type, id and value / gauge",
			path:   "/update",
			method: http.MethodPost,
			body:   `{"value":789.456,"id":"test gauge","type":"gauge"}`,
			want: want{
				contentType: "application/json",
				body:        `{"value":789.456,"id":"test gauge","type":"gauge"}`,
				statusCode:  200,
			},
		},
		{
			name:   "update with incorrect type, id and value / gauge",
			path:   "/update",
			method: http.MethodPost,
			body:   `{"value":"incorrect!","id":"test gauge","type":"gauge"}`,
			want: want{
				contentType: "application/json",
				body:        `{"message":"can't parse request body"}`,
				statusCode:  400,
			},
		},
		{
			name:   "update with incorrect type, id and value / counter",
			path:   "/update",
			method: http.MethodPost,
			body:   `{"delta":123.456,"id":"test counter","type":"counter"}`,
			want: want{
				contentType: "application/json",
				body:        `{"message":"can't parse request body"}`,
				statusCode:  400,
			},
		},
		{
			name:   "update with correct type, id and delta / counter",
			path:   "/update",
			method: http.MethodPost,
			body:   `{"delta":1,"id":"test counter","type":"counter"}`,
			want: want{
				contentType: "application/json",
				body:        `{"delta":1,"id":"test counter","type":"counter"}`,
				statusCode:  200,
			},
		},
		{
			name:   "update send counter one more / counter",
			path:   "/update",
			method: http.MethodPost,
			body:   `{"delta":1,"id":"test counter","type":"counter"}`,
			want: want{
				contentType: "application/json",
				body:        `{"delta":2,"id":"test counter","type":"counter"}`,
				statusCode:  200,
			},
		},
		{
			name:   "update send counter with empty delta / counter",
			path:   "/update",
			method: http.MethodPost,
			body:   `{"id":"cntr","type":"counter"}`,
			want: want{
				contentType: "application/json",
				body:        "{\"message\":\"empty metric value \\\"counter\\\"\"}",
				statusCode:  400,
			},
		},
		{
			name:   "update send gauge with empty value / gauge",
			path:   "/update",
			method: http.MethodPost,
			body:   `{"id":"gg","type":"gauge"}`,
			want: want{
				contentType: "application/json",
				body:        `{"message":"empty metric value \"gauge\""}`,
				statusCode:  400,
			},
		},
		{
			name:   "update parse unknown request body",
			path:   "/update",
			method: http.MethodPost,
			body:   `{ "OwO" }`,
			want: want{
				contentType: "application/json",
				body:        `{"message":"can't parse request body"}`,
				statusCode:  400,
			},
		},
		{
			name:   "check update request fields, one field is empty",
			path:   "/update",
			method: http.MethodPost,
			body:   `{"delta":1,"type":"test"}`,
			want: want{
				contentType: "application/json",
				body:        `{"message":"field \"id\" is required"}`,
				statusCode:  400,
			},
		},
		{
			name:   "check update request fields, field type is empty",
			path:   "/update",
			method: http.MethodPost,
			body:   `{"delta":1,"id":"test"}`,
			want: want{
				contentType: "application/json",
				body:        `{"message":"field \"type\" is required"}`,
				statusCode:  400,
			},
		},
		{
			name:   "check update request fields, two required fields are empty",
			path:   "/update",
			method: http.MethodPost,
			body:   `{"delta":1}`,
			want: want{
				contentType: "application/json",
				body:        "{\"message\":\"field \\\"type\\\" is required, field \\\"id\\\" is required\"}",
				statusCode:  400,
			},
		},
		{
			name:   "check update empty body",
			path:   "/update",
			method: http.MethodPost,
			body:   ``,
			want: want{
				contentType: "application/json",
				body:        `{"message":"empty request body"}`,
				statusCode:  400,
			},
		},
		{
			name:   "update unknown metric type",
			path:   "/update",
			method: http.MethodPost,
			body:   `{ "id":"unknown","type":"unknown" }`,
			want: want{
				contentType: "application/json",
				body:        `{"message":"unknown metric type \"unknown\""}`,
				statusCode:  400,
			},
		},
		{
			name:   "check value empty body",
			path:   "/value",
			method: http.MethodPost,
			body:   ``,
			want: want{
				contentType: "application/json",
				body:        `{"message":"empty request body"}`,
				statusCode:  400,
			},
		},
		{
			name:   "value parse unknown request body",
			path:   "/value",
			method: http.MethodPost,
			body:   `{ "UwU" }`,
			want: want{
				contentType: "application/json",
				body:        `{"message":"can't parse request body"}`,
				statusCode:  400,
			},
		},
		{
			name:   "value with correct type and id / gauge",
			path:   "/value",
			method: http.MethodPost,
			body:   `{"id":"test gauge","type":"gauge"}`,
			want: want{
				contentType: "application/json",
				body:        `{"value":789.456,"id":"test gauge","type":"gauge"}`,
				statusCode:  200,
			},
		},
		{
			name:   "value with correct type and id / counter",
			path:   "/value",
			method: http.MethodPost,
			body:   `{"id":"test counter","type":"counter"}`,
			want: want{
				contentType: "application/json",
				body:        `{"delta":2,"id":"test counter","type":"counter"}`,
				statusCode:  200,
			},
		},
		{
			name:   "value get unknown metric / counter",
			path:   "/value",
			method: http.MethodPost,
			body:   `{ "id":"unknown","type":"counter" }`,
			want: want{
				contentType: "application/json",
				body:        `{"message":"counter metric unknown doesn't exist"}`,
				statusCode:  404,
			},
		},
		{
			name:   "value get unknown metric / gauge",
			path:   "/value",
			method: http.MethodPost,
			body:   `{ "id":"unknown","type":"gauge" }`,
			want: want{
				contentType: "application/json",
				body:        `{"message":"gauge metric unknown doesn't exist"}`,
				statusCode:  404,
			},
		},
		{
			name:   "value unknown metric type",
			path:   "/value",
			method: http.MethodPost,
			body:   `{ "id":"unknown","type":"unknown" }`,
			want: want{
				contentType: "application/json",
				body:        `{"message":"unknown metric type \"unknown\""}`,
				statusCode:  400,
			},
		},
		{
			name:   "check value request fields, id field is empty",
			path:   "/value",
			method: http.MethodPost,
			body:   `{"type":"test"}`,
			want: want{
				contentType: "application/json",
				body:        `{"message":"field \"id\" is required"}`,
				statusCode:  400,
			},
		},
		{
			name:   "check value request fields, field type is empty",
			path:   "/value",
			method: http.MethodPost,
			body:   `{"id":"test"}`,
			want: want{
				contentType: "application/json",
				body:        `{"message":"field \"type\" is required"}`,
				statusCode:  400,
			},
		},
		{
			name:   "check value request fields, two required fields are empty",
			path:   "/value",
			method: http.MethodPost,
			body:   `{}`,
			want: want{
				contentType: "application/json",
				body:        "{\"message\":\"field \\\"type\\\" is required, field \\\"id\\\" is required\"}",
				statusCode:  400,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := testRequest(t, ts, test.method, test.path, bytes.NewBuffer([]byte(test.body)), "application/json")
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, test.want.statusCode, res.StatusCode)
			defer res.Body.Close()

			if test.want.body != "" {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				// remove "\n" character at the response end of the body
				assert.Equal(t, test.want.body, strings.TrimSpace(string(resBody)))
			}
		})
	}
}

func TestMetricsUpdatesHandlerJSON(t *testing.T) {
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
		body   string
		want   want
	}{
		{
			name:   "updates with several correct metrics at a time",
			path:   "/updates",
			method: http.MethodPost,
			body: `[{"delta": 1,"id": "my counter","type": "counter"},
{"value": 789.456,"id": "my gauge","type": "gauge"}]`,
			want: want{
				contentType: "application/json",
				body:        "",
				statusCode:  200,
			},
		},
		{
			name:   "updates with several correct metrics / incorrect req fields",
			path:   "/updates",
			method: http.MethodPost,
			body: `[{"delta": 1,"id": "my counter","type": "abc"},
{"value": 789.456,"id": "my gauge","type": "gauge"}]`,
			want: want{
				contentType: "application/json",
				body:        `{"message":"unknown metric type \"abc\""}`,
				statusCode:  400,
			},
		},
		{
			name:   "updates with several metrics / with missing fields",
			path:   "/updates",
			method: http.MethodPost,
			body: `[{"delta": 1,"type": "abc"},
{"value": 789.456,"id": "my gauge","type": "gauge"}]`,
			want: want{
				contentType: "application/json",
				body:        `{"message":"field \"id\" is required"}`,
				statusCode:  400,
			},
		},
		{
			name:   "updates with several metrics / with several missing fields",
			path:   "/updates",
			method: http.MethodPost,
			body: `[{"delta": 1},
{"value": 789.456,"id": "my gauge","type": "gauge"}]`,
			want: want{
				contentType: "application/json",
				body:        `{"message":"field \"type\" is required, field \"id\" is required"}`,
				statusCode:  400,
			},
		},
		{
			name:   "updates with empty body",
			path:   "/updates",
			method: http.MethodPost,
			body:   ``,
			want: want{
				contentType: "application/json",
				body:        `{"message":"empty request body"}`,
				statusCode:  400,
			},
		},
		{
			name:   "updates with incorrect request body",
			path:   "/updates",
			method: http.MethodPost,
			body:   `{ "UwU" }`,
			want: want{
				contentType: "application/json",
				body:        `{"message":"can't parse request body"}`,
				statusCode:  400,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := testRequest(t, ts, test.method, test.path, bytes.NewBuffer([]byte(test.body)), "application/json")
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, test.want.statusCode, res.StatusCode)
			defer res.Body.Close()

			if test.want.body != "" {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				// remove "\n" character at the response end of the body
				assert.Equal(t, test.want.body, strings.TrimSpace(string(resBody)))
			}
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader, header string) *http.Response {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestClientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, ts.URL+path, body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", header)

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
