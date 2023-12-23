package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/server/storage/memory"
)

const (
	defaultTestClientTimeout = 3 * time.Second
)

func TestRoute(t *testing.T) {
	log := zap.Must(zap.NewDevelopment())
	ms := memory.NewMemStorage()

	router := NewRouter(ms, nil, log)

	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("check for a non-existent route", func(t *testing.T) {
		resp := testRequest(t, ts, http.MethodGet, "/uwu")
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("check the ping route", func(t *testing.T) {
		resp := testRequest(t, ts, http.MethodGet, "/ping")
		defer resp.Body.Close()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string) *http.Response {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestClientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	return resp
}
