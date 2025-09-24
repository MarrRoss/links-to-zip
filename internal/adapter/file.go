package adapter

import (
	"context"
	"fmt"
	"sync"
	"workmate_tz/internal/domain/model"
	"workmate_tz/internal/observability"
)

type FileRepositoryImpl struct {
	observer *observability.Observability
	files    sync.Map
}

func NewFileRepositoryImpl(
	observer *observability.Observability,
) (*FileRepositoryImpl, error) {
	return &FileRepositoryImpl{
		observer: observer,
	}, nil
}

func (repo *FileRepositoryImpl) CreateFile(
	ctx context.Context,
	file *model.TaskFile,
) error {
	repo.files.Store(file.ID, file)
	return nil
}

func (repo *FileRepositoryImpl) PrintAllFiles() {
	repo.files.Range(func(key, value any) bool {
		file, ok := value.(*model.TaskFile)
		if !ok {
			fmt.Println("invalid type stored in files map")
			return true
		}
		fmt.Printf("File ID: %s, Status: %s\n", file.ID.String(), file.Status)
		return true
	})
}
