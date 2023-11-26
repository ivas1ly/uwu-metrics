package file

import (
	"encoding/json"
	"os"

	"github.com/ivas1ly/uwu-metrics/internal/server/entity"
	"github.com/ivas1ly/uwu-metrics/internal/server/storage/memory"
)

type PersistentStorage interface {
	Save() error
	Restore() error
}

type fileStorage struct {
	memoryStorage memory.Storage
	fileName      string
	perm          os.FileMode
}

func NewFileStorage(fileName string, perm os.FileMode, storage memory.Storage) PersistentStorage {
	return &fileStorage{
		fileName:      fileName,
		perm:          perm,
		memoryStorage: storage,
	}
}

func (fs *fileStorage) Save() error {
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

func (fs *fileStorage) Restore() error {
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
