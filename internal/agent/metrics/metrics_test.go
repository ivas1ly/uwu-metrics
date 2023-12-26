package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpdateMetrics(t *testing.T) {
	tests := []struct {
		name      string
		updates   int
		pollCount int64
	}{
		{
			name:      "no updates",
			updates:   0,
			pollCount: 0,
		},
		{
			name:      "1 update",
			updates:   1,
			pollCount: 1,
		},
		{
			name:      "10 updates",
			updates:   10,
			pollCount: 10,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ms := Metrics{}
			prevRandomValue := ms.RandomValue

			for upd := 0; upd < test.updates; upd++ {
				ms.UpdateMetrics()
				assert.NotEqual(t, prevRandomValue, ms.RandomValue)
				prevRandomValue = ms.RandomValue
			}
			assert.Equal(t, test.pollCount, ms.PollCount)
		})
	}
}

func TestUpdatePsutilMetrics(t *testing.T) {
	ms := Metrics{}

	err := ms.UpdatePsutilMetrics()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, ms.UtilizationPerCPU[0], 0.0)
}

func TestPrepareGaugeReport(t *testing.T) {
	ms := Metrics{}
	ms.UpdateMetrics()
	err := ms.UpdatePsutilMetrics()
	assert.NoError(t, err)

	report := ms.PrepareGaugeReport()

	assert.Greater(t, report["TotalMemory"], 0.0)
	assert.GreaterOrEqual(t, report["CPUutilization0"], 0.0)
	assert.Greater(t, report["RandomValue"], 0.0)
	assert.Greater(t, report["Alloc"], 0.0)
	assert.Greater(t, report["NextGC"], 0.0)
}

func TestPrepareCounterReport(t *testing.T) {
	ms := Metrics{}
	ms.UpdateMetrics()
	time.Sleep(1 * time.Second)
	ms.UpdateMetrics()

	report := ms.PrepareCounterReport()

	assert.Equal(t, report["PollCount"], int64(2))
}
