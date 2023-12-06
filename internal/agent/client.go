package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/agent/entity"
)

type Client struct {
	Metrics *Metrics
	Logger  *zap.Logger
	URL     string
}

type MetricsPayload struct {
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	ID    string   `json:"id"`
	MType string   `json:"type"`
}

func (c *Client) SendReport() {
	for key, value := range c.Metrics.PrepareGaugeReport() {
		val := value
		body, err := json.Marshal(&MetricsPayload{
			ID:    key,
			MType: entity.GaugeType,
			Delta: nil,
			Value: &val,
		})
		if err != nil {
			c.Logger.Info("failed to marshal json", zap.Error(err))
			continue
		}

		if err := c.sendRequest(http.MethodPost, body); err != nil {
			continue
		}
	}

	for key, value := range c.Metrics.PrepareCounterReport() {
		val := value
		body, err := json.Marshal(&MetricsPayload{
			ID:    key,
			MType: entity.CounterType,
			Delta: &val,
			Value: nil,
		})
		if err != nil {
			c.Logger.Info("failed to marshal json", zap.Error(err))
			continue
		}

		if err := c.sendRequest(http.MethodPost, body); err != nil {
			continue
		}
	}
}

func (c *Client) sendRequest(method string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultClientTimeout)
	defer cancel()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(body); err != nil {
		c.Logger.Error("can't compress body", zap.Error(err))
		return err
	}

	if err := gw.Close(); err != nil {
		c.Logger.Error("can't close gzip writer", zap.Error(err))
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, c.URL, &buf)
	if err != nil {
		c.Logger.Error("can't create new HTTP request", zap.Error(err))
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.Logger.Error("can't send the HTTP request", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	return nil
}
