package http

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/ivas1ly/uwu-metrics/internal/agent/metrics"
	"github.com/ivas1ly/uwu-metrics/internal/utils/hash"
)

const (
	defaultPayloadCap    = 40
	defaultClientTimeout = 3 * time.Second
)

var (
	retryIntervals = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
)

type Client interface {
	SendReport() error
}

// Client is a structure for storing HTTP client parameters.
type httpClient struct {
	Metrics      *metrics.Metrics
	Logger       *zap.Logger
	RSAPublicKey *rsa.PublicKey
	LocalIP      *net.IP
	URL          string
	HashKey      []byte
}

func NewClient(metrics *metrics.Metrics, localIP *net.IP, publicKey *rsa.PublicKey,
	url string, hashKey []byte, logger *zap.Logger) Client {
	return &httpClient{
		Metrics:      metrics,
		Logger:       logger,
		RSAPublicKey: publicKey,
		LocalIP:      localIP,
		URL:          url,
		HashKey:      hashKey,
	}
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
func (c *httpClient) SendReport() error {
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
func (c *httpClient) sendRequest(method string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultClientTimeout)
	defer cancel()

	var sign string
	var err error
	if len(c.HashKey) > 0 {
		c.Logger.Info("hash key found, set the header with a body hash")

		if sign, err = hash.Hash(body, c.HashKey); err == nil {
			c.Logger.Info("hash", zap.String("val", sign))
		} else {
			c.Logger.Info("can't set hash header, skip header")
		}
	}

	var encrypted []byte
	if c.RSAPublicKey != nil {
		encrypted, err = rsa.EncryptPKCS1v15(rand.Reader, c.RSAPublicKey, body)
		if err != nil {
			c.Logger.Info("can't encrypt body with RSA public key", zap.Error(err))
			return err
		}

		body = encrypted
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err = gw.Write(body); err != nil {
		c.Logger.Info("can't compress body", zap.Error(err))
		return err
	}

	if err = gw.Close(); err != nil {
		c.Logger.Info("can't close gzip writer", zap.Error(err))
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, c.URL, &buf)
	if err != nil {
		c.Logger.Info("can't create new HTTP request", zap.Error(err))
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.LocalIP != nil {
		req.Header.Set("X-Real-IP", c.LocalIP.String())
	}
	req.Header.Set("Content-Encoding", "gzip")

	if sign != "" {
		req.Header.Set("HashSHA256", sign)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.Logger.Info("can't send the HTTP request", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	return nil
}
