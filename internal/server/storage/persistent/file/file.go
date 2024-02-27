package file

import (
	"context"
	"encoding/json"
	"os"

	"github.com/ivas1ly/uwu-metrics/internal/server/entity"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/memory"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/persistent"
)

type fileStorage struct {
	memoryStorage memory.Storage
	fileName      string
	perm          os.FileMode
}

// NewFileStorage creates new persistent storage in the file.
func NewFileStorage(fileName string, perm os.FileMode, storage memory.Storage) persistent.Storage {
	return &fileStorage{
		fileName:      fileName,
		perm:          perm,
		memoryStorage: storage,
	}
}

// Save takes the metrics from memory and saves them to the file.
func (fs *fileStorage) Save(_ context.Context) error {
	file, err := os.OpenFile(fs.fileName, os.O_WRONLY|os.O_CREATE, fs.perm)
	if err != nil {
		return err
	}

	defer file.Close()

	metrics := fs.memoryStorage.GetMetrics()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(&metrics)
	if err != nil {
		return err
	}

	return nil
}

// Restore fetches the last saved metrics from the file and restores them to in-memory storage.
func (fs *fileStorage) Restore(_ context.Context) error {
	file, err := os.OpenFile(fs.fileName, os.O_RDONLY|os.O_CREATE, fs.perm)
	if err != nil {
		return err
	}

	var metrics entity.Metrics
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&metrics); err != nil {
		return err
	}

	fs.memoryStorage.SetMetrics(metrics)

	return nil
}
