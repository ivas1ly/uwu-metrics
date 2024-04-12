package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/ivas1ly/uwu-metrics/internal/server/service"
	"go.uber.org/zap"
)

func Example_updateURL() {
	testStorage := NewTestStorage()
	logger := zap.Must(zap.NewDevelopment())
	router := chi.NewRouter()
	metricsService := service.NewMetricsService(testStorage)

	NewRoutes(router, metricsService, logger)

	ts := httptest.NewServer(router)
	defer ts.Close()

	client := &http.Client{
		Timeout: defaultTestClientTimeout,
	}

	var req *http.Request
	var err error
	var resp *http.Response

	req, err = http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL+"/update/gauge/owo/123.456", nil)
	if err != nil {
		fmt.Println(err.Error())
	}

	resp, err = client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	resp.Body.Close()

	fmt.Println("gauge", resp.Status)

	req, err = http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL+"/update/counter/uwu/123", nil)
	if err != nil {
		fmt.Println(err.Error())
	}

	resp, err = client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	resp.Body.Close()

	fmt.Println("counter", resp.Status)

	// Output:
	// gauge 200 OK
	// counter 200 OK
}

func Example_valueURL() {
	testStorage := NewTestStorage()
	logger := zap.Must(zap.NewDevelopment())
	router := chi.NewRouter()
	metricsService := service.NewMetricsService(testStorage)

	NewRoutes(router, metricsService, logger)

	testStorage.UpdateCounter("uwu", 123)
	testStorage.UpdateGauge("owo", 123.456)

	ts := httptest.NewServer(router)
	defer ts.Close()

	client := &http.Client{
		Timeout: defaultTestClientTimeout,
	}

	var req *http.Request
	var err error
	var resp *http.Response
	var value string
	var respBody []byte

	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/value/gauge/owo", nil)
	if err != nil {
		fmt.Println(err.Error())
	}

	resp, err = client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}

	respBody, err = io.ReadAll(resp.Body)
	value = string(respBody)
	resp.Body.Close()

	fmt.Println("gauge", resp.Status, value)

	req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/value/counter/uwu", nil)
	if err != nil {
		fmt.Println(err.Error())
	}

	resp, err = client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}

	respBody, err = io.ReadAll(resp.Body)
	value = string(respBody)
	resp.Body.Close()

	fmt.Println("counter", resp.Status, value)

	// Output:
	// gauge 200 OK 123.456
	// counter 200 OK 123
}

func Example_webpage() {
	testStorage := NewTestStorage()
	logger := zap.Must(zap.NewDevelopment())
	router := chi.NewRouter()
	metricsService := service.NewMetricsService(testStorage)

	NewRoutes(router, metricsService, logger)

	testStorage.UpdateCounter("uwu", 123)
	testStorage.UpdateGauge("owo", 123.456)

	ts := httptest.NewServer(router)
	defer ts.Close()

	client := &http.Client{
		Timeout: defaultTestClientTimeout,
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/", nil)
	if err != nil {
		fmt.Println(err.Error())
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	resp.Body.Close()

	fmt.Println("webpage", resp.Status, resp.Header.Get("Content-Type"))

	// Output:
	// webpage 200 OK text/html; charset=utf-8
}

func Example_updateJSON() {
	path := "/update"

	testStorage := NewTestStorage()
	logger := zap.Must(zap.NewDevelopment())
	router := chi.NewRouter()
	metricsService := service.NewMetricsService(testStorage)

	NewRoutes(router, metricsService, logger)

	ts := httptest.NewServer(router)
	defer ts.Close()

	client := &http.Client{
		Timeout: defaultTestClientTimeout,
	}

	var req *http.Request
	var err error
	var resp *http.Response
	var value string
	var respBody []byte

	body := `{"value":789.456,"id":"test gauge","type":"gauge"}`
	req, err = http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL+path,
		bytes.NewBuffer([]byte(body)))
	if err != nil {
		fmt.Println(err.Error())
	}

	resp, err = client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}

	respBody, err = io.ReadAll(resp.Body)
	value = strings.TrimSpace(string(respBody))
	resp.Body.Close()

	fmt.Println("gauge", resp.Status, resp.Header.Get("Content-Type"), value)

	body = `{"delta":1,"id":"test counter","type":"counter"}`
	req, err = http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL+path,
		bytes.NewBuffer([]byte(body)))
	if err != nil {
		fmt.Println(err.Error())
	}

	resp, err = client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}

	respBody, err = io.ReadAll(resp.Body)
	value = strings.TrimSpace(string(respBody))
	resp.Body.Close()

	fmt.Println("counter", resp.Status, resp.Header.Get("Content-Type"), value)

	// Output:
	// gauge 200 OK application/json {"value":789.456,"id":"test gauge","type":"gauge"}
	// counter 200 OK application/json {"delta":1,"id":"test counter","type":"counter"}
}

func Example_updatesJSON() {
	path := "/updates"

	testStorage := NewTestStorage()
	logger := zap.Must(zap.NewDevelopment())
	router := chi.NewRouter()
	metricsService := service.NewMetricsService(testStorage)

	NewRoutes(router, metricsService, logger)

	ts := httptest.NewServer(router)
	defer ts.Close()

	client := &http.Client{
		Timeout: defaultTestClientTimeout,
	}

	var req *http.Request
	var err error
	var resp *http.Response

	body := `[{"delta": 1,"id": "my counter","type": "counter"},
{"value": 789.456,"id": "my gauge","type": "gauge"}]`
	req, err = http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL+path,
		bytes.NewBuffer([]byte(body)))
	if err != nil {
		fmt.Println(err.Error())
	}

	resp, err = client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	resp.Body.Close()

	fmt.Println("gauge", resp.Status)

	// Output:
	// gauge 200 OK
}

func Example_valueJSON() {
	path := "/value"

	testStorage := NewTestStorage()
	logger := zap.Must(zap.NewDevelopment())
	router := chi.NewRouter()
	metricsService := service.NewMetricsService(testStorage)

	NewRoutes(router, metricsService, logger)

	ts := httptest.NewServer(router)
	defer ts.Close()

	testStorage.UpdateCounter("uwu", 123)
	testStorage.UpdateGauge("owo", 123.456)

	client := &http.Client{
		Timeout: defaultTestClientTimeout,
	}

	var req *http.Request
	var err error
	var resp *http.Response
	var value string
	var respBody []byte

	body := `{"id":"owo","type":"gauge"}`
	req, err = http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL+path,
		bytes.NewBuffer([]byte(body)))
	if err != nil {
		fmt.Println(err.Error())
	}

	resp, err = client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}

	respBody, err = io.ReadAll(resp.Body)
	value = strings.TrimSpace(string(respBody))
	resp.Body.Close()

	fmt.Println("gauge", resp.Status, resp.Header.Get("Content-Type"), value)

	body = `{"id":"uwu","type":"counter"}`
	req, err = http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL+path,
		bytes.NewBuffer([]byte(body)))
	if err != nil {
		fmt.Println(err.Error())
	}

	resp, err = client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}

	respBody, err = io.ReadAll(resp.Body)
	value = strings.TrimSpace(string(respBody))
	resp.Body.Close()

	fmt.Println("counter", resp.Status, resp.Header.Get("Content-Type"), value)

	// Output:
	// gauge 200 OK application/json {"value":123.456,"id":"owo","type":"gauge"}
	// counter 200 OK application/json {"delta":123,"id":"uwu","type":"counter"}
}
