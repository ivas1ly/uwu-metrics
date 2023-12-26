package sethash

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/lib/logger"
)

const (
	defaultLogLevel          = "info"
	defaultTestClientTimeout = 3 * time.Second
)

func TestSetHash(t *testing.T) {
	log := logger.New(defaultLogLevel, zap.NewDevelopmentConfig()).
		With(zap.String("app", "test"))

	t.Run("with key", func(t *testing.T) {
		r := chi.NewRouter()
		r.Use(New(log, []byte("some key")))
		r.Get("/", testResponse(t))

		ts := httptest.NewServer(r)
		defer ts.Close()

		testText := "Test Data"

		resp, respBody := testRequest(t, ts, http.MethodGet, "/", bytes.NewBuffer([]byte(testText)))
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, testText, respBody)
		assert.Equal(t, "de33e02b3f3a2b76e1db241bfdf4a71142ff202a3714bf3d6772eae0da74aa19", resp.Header.Get("HashSHA256"))
	})

	t.Run("without key", func(t *testing.T) {
		r := chi.NewRouter()
		r.Use(New(log, nil))
		r.Get("/", testResponse(t))

		ts := httptest.NewServer(r)
		defer ts.Close()

		testText := "Test Data"

		resp, respBody := testRequest(t, ts, http.MethodGet, "/", bytes.NewBuffer([]byte(testText)))
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, testText, respBody)
		assert.Equal(t, "", resp.Header.Get("HashSHA256"))
	})
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestClientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, ts.URL+path, body)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func testResponse(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		b, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		_, _ = w.Write(b)
	}
}
