package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ivas1ly/uwu-metrics/internal/server/entity"
)

func TestMemoryStorage(t *testing.T) {
	ms := NewMemStorage()

	t.Run("update counter", func(t *testing.T) {
		var testValueCounter int64 = 64

		ms.UpdateCounter("my counter", testValueCounter)
		value, err := ms.GetCounter("my counter")
		assert.NoError(t, err)
		assert.Equal(t, testValueCounter, value)
	})

	t.Run("update gauge", func(t *testing.T) {
		testValueGauge := 128.32

		ms.UpdateGauge("OwO gauge", testValueGauge)
		value, err := ms.GetGauge("OwO gauge")
		assert.NoError(t, err)
		assert.Equal(t, testValueGauge, value)
	})

	t.Run("check for metrics", func(t *testing.T) {
		counter, err := ms.GetCounter("UwU")
		assert.Error(t, err)
		assert.Equal(t, int64(0), counter)

		gauge, err := ms.GetGauge("teeeeest!!")
		assert.Error(t, err)
		assert.Equal(t, float64(0), gauge)
	})

	t.Run("check get/set metrics", func(t *testing.T) {
		metrics := entity.Metrics{
			Counter: make(map[string]int64),
			Gauge:   make(map[string]float64),
		}

		metrics.Counter["counter 1"] = 678
		metrics.Counter["counter 2"] = 123
		metrics.Gauge["gauge 1"] = 123.456
		metrics.Gauge["gauge 2"] = 789.456
		metrics.Gauge["gauge 2"] = 0

		ms.SetMetrics(metrics)

		result := ms.GetMetrics()
		assert.Equal(t, metrics, result)
	})
}
