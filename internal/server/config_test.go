package server

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Run("check default values", func(t *testing.T) {
		config := NewConfig()

		assert.Equal(t, config.Endpoint, net.JoinHostPort(defaultHost, defaultPort))
		assert.Equal(t, config.DatabaseDSN, "")
		assert.Equal(t, config.StoreInterval, defaultStoreInterval)
		assert.Equal(t, config.FileStoragePath, defaultFileStoragePath)
		assert.Equal(t, config.Restore, defaultFileRestore)
	})
}
