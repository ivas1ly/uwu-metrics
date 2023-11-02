package agent

import (
	"fmt"
	"log/slog"
	"net/http"
)

type Client struct {
	URL     string
	Metrics *Metrics
	Logger  *slog.Logger
}

func (c *Client) SendReport() {
	for key, value := range c.Metrics.PrepareGaugeReport() {
		path := fmt.Sprintf("%s/%s/%f", "gauge", key, value)

		if err := c.sendRequest(http.MethodPost, path); err != nil {
			continue
		}
	}

	for key, value := range c.Metrics.PrepareCounterReport() {
		path := fmt.Sprintf("%s/%s/%d", "counter", key, int(value))

		if err := c.sendRequest(http.MethodPost, path); err != nil {
			continue
		}
	}
}

func (c *Client) sendRequest(method, path string) error {
	c.Logger.Debug(c.URL + path)
	req, err := http.NewRequest(method, c.URL+path, nil)
	if err != nil {
		c.Logger.Error("can't create new request")
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.Logger.Error("can't make the request")
		return err
	}
	defer resp.Body.Close()

	return nil
}
