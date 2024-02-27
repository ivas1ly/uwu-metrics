package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/agent/metrics"
	"github.com/ivas1ly/uwu-metrics/internal/utils/hash"
)

const defaultPayloadCap = 40

var (
	retryIntervals = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
)

// Client is a structure for storing HTTP client parameters.
type Client struct {
	Metrics *metrics.Metrics
	Logger  *zap.Logger
	URL     string
	Key     []byte
}

// MetricsPayload structure to convert metrics into JSON format for sending to the server.
type MetricsPayload struct {
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	ID    string   `json:"id"`
	MType string   `json:"type"`
}

// SendReport prepares and sends metrics to the server.
// If the metrics cannot be sent, it will retry to send the metrics
// to the server several more times (3 times in total).
func (c *Client) SendReport() error {
	payload := make([]MetricsPayload, 0, defaultPayloadCap)

	for key, value := range c.Metrics.PrepareGaugeReport() {
		val := value

		mp := MetricsPayload{
			ID:    key,
			MType: metrics.GaugeType,
			Delta: nil,
			Value: &val,
		}

		payload = append(payload, mp)
	}

	for key, value := range c.Metrics.PrepareCounterReport() {
		val := value
		mp := MetricsPayload{
			ID:    key,
			MType: metrics.CounterType,
			Delta: &val,
			Value: nil,
		}

		payload = append(payload, mp)
	}

	body, err := json.Marshal(&payload)
	if err != nil {
		c.Logger.Info("failed to marshal json", zap.Error(err))
		return err
	}

	for _, interval := range retryIntervals {
		err = c.sendRequest(http.MethodPost, body)
		if err != nil {
			c.Logger.Info("can't send request, trying again", zap.Error(err),
				zap.Duration("with interval", interval))
			time.Sleep(interval)
		} else {
			c.Logger.Info("request sent successfully")
			break
		}
	}
	if err != nil {
		return err
	}

	return nil
}

// sendRequest wrapper method for net/http client.
func (c *Client) sendRequest(method string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultClientTimeout)
	defer cancel()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(body); err != nil {
		c.Logger.Info("can't compress body", zap.Error(err))
		return err
	}

	if err := gw.Close(); err != nil {
		c.Logger.Info("can't close gzip writer", zap.Error(err))
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, c.URL, &buf)
	if err != nil {
		c.Logger.Info("can't create new HTTP request", zap.Error(err))
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	if len(c.Key) > 0 {
		c.Logger.Info("hash key found, set the header with a body hash")

		var sign string
		if sign, err = hash.Hash(body, c.Key); err == nil {
			c.Logger.Info("hash", zap.String("val", sign))
			req.Header.Set("HashSHA256", sign)
		} else {
			c.Logger.Info("can't set hash header, skip header")
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.Logger.Info("can't send the HTTP request", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	return nil
}
