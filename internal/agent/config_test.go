package agent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Run("check default values", func(t *testing.T) {
		config := NewConfig()

		assert.Equal(t, config.ReportInterval, defaultReportInterval*time.Second)
		assert.Equal(t, config.PollInterval, defaultPollInterval*time.Second)
		assert.Equal(t, config.EndpointHost, defaultEndpointHost)
	})
}
