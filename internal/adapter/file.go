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

func (repo *FileRepositoryImpl) GetFiles(
	ctx context.Context,
	ids []model.ID,
) (
	[]*model.TaskFile,
	[]model.ID,
	error,
) {
	var foundFiles []*model.TaskFile
	var notFoundIDs []model.ID

	for _, id := range ids {
		if val, ok := repo.files.Load(id); ok {
			if file, ok := val.(*model.TaskFile); ok {
				foundFiles = append(foundFiles, file)
			} else {
				repo.observer.Logger.Error().
					Msgf("invalid type stored in files map for file id: %v", id)
				notFoundIDs = append(notFoundIDs, id)
			}
		} else {
			notFoundIDs = append(notFoundIDs, id)
		}
	}

	return foundFiles, notFoundIDs, nil
}

func (repo *FileRepositoryImpl) UpdateFile(
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
