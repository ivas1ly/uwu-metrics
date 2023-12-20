package agent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Run("check default values", func(t *testing.T) {
		cfg := NewConfig()

		assert.Equal(t, cfg.ReportInterval, defaultReportInterval*time.Second)
		assert.Equal(t, cfg.PollInterval, defaultPollInterval*time.Second)
		assert.Equal(t, cfg.EndpointHost, defaultEndpointHost)
	})
}
