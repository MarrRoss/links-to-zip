package handler

import (
	"time"
	"workmate_tz/internal/domain/model"

	"github.com/google/uuid"
)

type GetTaskStatusQuery struct {
	TaskID uuid.UUID
}

type GetFile struct {
	ID           uuid.UUID
	Name         string
	Link         string
	Status       string
	Error        *string
	AttemptCount int
	MaxAttempts  int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	EndedAt      *time.Time
}

func SliceToGetFiles(files []*model.TaskFile) []GetFile {
	getFiles := make([]GetFile, 0, len(files))
	for _, file := range files {
		var fileError *string
		if file.Error != "" {
			fileError = &file.Error
		}
		getFile := GetFile{
			ID:           file.ID.ToRaw(),
			Name:         file.Name,
			Link:         file.Link,
			Status:       file.Status,
			Error:        fileError,
			AttemptCount: file.AttemptCount,
			MaxAttempts:  file.MaxAttempts,
			CreatedAt:    file.CreatedAt,
			UpdatedAt:    file.UpdatedAt,
			EndedAt:      file.EndedAt,
		}
		getFiles = append(getFiles, getFile)
	}
	return getFiles
}
