package checkhash

import (
	"bytes"
	"context"
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

	"github.com/ivas1ly/uwu-metrics/internal/lib/logger"
	"github.com/ivas1ly/uwu-metrics/internal/utils/hash"
)

const (
	defaultLogLevel          = "info"
	defaultTestClientTimeout = 3 * time.Second
	testBody                 = "Somebody once told me the world is gonna roll me"
)

func TestCheckHash(t *testing.T) {
	log := logger.New(defaultLogLevel, zap.NewDevelopmentConfig())

	key := []byte("some key")
	h, err := hash.Hash([]byte(testBody), key)
	assert.NoError(t, err)

	type want struct {
		body       string
		statusCode int
	}

	tests := []struct {
		name   string
		header string
		body   string
		key    []byte
		want   want
	}{
		{
			name:   "with key",
			header: h,
			body:   testBody,
			key:    key,
			want: want{
				body:       "",
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "without key",
			header: "",
			body:   testBody,
			key:    nil,
			want: want{
				body:       "",
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "with a different key",
			header: h,
			body:   testBody,
			key:    []byte("1"),
			want: want{
				body:       `{"message":"can't check hash"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "with a key but without hash header",
			header: "",
			body:   testBody,
			key:    key,
			want: want{
				body:       "",
				statusCode: http.StatusOK,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Use(New(log, test.key))
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, respBody := testRequest(t, ts, http.MethodPost, "/", test.header, bytes.NewBuffer([]byte(test.body)))
			defer resp.Body.Close()
			assert.Equal(t, test.want.statusCode, resp.StatusCode)
			assert.Equal(t, test.want.body, strings.TrimSpace(respBody))
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, header string,
	body io.Reader) (*http.Response, string) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestClientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, ts.URL+path, body)
	require.NoError(t, err)

	if header != "" {
		req.Header.Set("HashSHA256", header)
	}

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}
