package agent

import (
	"testing"

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
