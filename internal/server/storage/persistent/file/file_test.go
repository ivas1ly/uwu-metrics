package file

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ivas1ly/uwu-metrics/internal/server/entity"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/memory"
)

func TestFileStorage(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-metrics")
	if err != nil {
		t.Fatal(err)
	}
	fileName := tmpFile.Name()

	defer func(name string) {
		err = os.Remove(name)
		if err != nil {
			t.Fatal(err)
		}
	}(fileName)

	ms := memory.NewMemStorage()
	fileStorage := NewFileStorage(fileName, 0666, ms)

	metrics := entity.Metrics{
		Counter: make(map[string]int64),
		Gauge:   make(map[string]float64),
	}

	metrics.Counter["counter 1"] = 678
	metrics.Counter["counter 2"] = 123
	metrics.Gauge["gauge 1"] = 123.456
	metrics.Gauge["gauge 2"] = 789.456
	metrics.Gauge["gauge 2"] = 0

	t.Run("write to file", func(t *testing.T) {
		ms.SetMetrics(metrics)

		err = fileStorage.Save(context.Background())
		assert.NoError(t, err)
		assert.FileExists(t, fileName)

		// with a short declaration govet shows a warning message
		var data []byte
		data, err = os.ReadFile(fileName)
		assert.NoError(t, err)

		//nolint:lll // test json
		assert.Equal(t, string(data), "{\"Counter\":{\"counter 1\":678,\"counter 2\":123},\"Gauge\":{\"gauge 1\":123.456,\"gauge 2\":0}}\n")
	})

	t.Run("read from existed file", func(t *testing.T) {
		assert.FileExists(t, fileName)
		assert.NoError(t, err)

		ms.SetMetrics(entity.Metrics{
			Counter: make(map[string]int64),
			Gauge:   make(map[string]float64),
		})

		err = fileStorage.Restore(context.Background())
		assert.NoError(t, err)

		counter, err := ms.GetCounter("counter 2")
		assert.NoError(t, err)
		assert.Equal(t, int64(123), counter)
	})
}
