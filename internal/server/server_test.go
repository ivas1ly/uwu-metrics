package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	cfg := NewConfig()

	freePort, err := getFreePort()
	if freePort != 0 {
		t.Logf("%d", freePort)
	}
	assert.NoError(t, err)
	cfg.Endpoint = fmt.Sprintf("%s:%d", "localhost", freePort)
	t.Log(cfg.Endpoint)

	go Run(cfg)

	time.Sleep(5 * time.Second)

	url := "http://" + cfg.Endpoint

	t.Run("test /ping", func(t *testing.T) {
		resp := testServerRequest(t, url, http.MethodGet, "/ping")
		defer resp.Body.Close()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("test webpage", func(t *testing.T) {
		resp := testServerRequest(t, url, http.MethodGet, "/")
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func testServerRequest(t *testing.T, endpoint string, method, path string) *http.Response {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestClientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, endpoint+path, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
