package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"workmate_tz/internal/domain/exception"
	"workmate_tz/internal/domain/model"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type GetTaskArchiveQuery struct {
	TaskID uuid.UUID
}

type GetTaskArchive struct {
	Status  string
	Message string
}

func (h *AppHandler) GetTaskArchive(
	ctx context.Context,
	qry GetTaskArchiveQuery,
) (
	*GetTaskArchive,
	//*FileError,
	io.Reader,
	error,
) {
	id := model.UUIDtoID(qry.TaskID)
	task, err := h.taskStorage.GetTaskByID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get task: %w", err)
	}

	switch task.Status {
	case model.StatusCreated:
		task.Status = model.StatusProcessing
		files, errIDs, err := h.fileStorage.GetFiles(ctx, task.Files)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get files for task %v: %w", task.ID, err)
		}
		if len(files) == 0 {
			return nil, nil, fmt.Errorf("no valid files in task")
		}
		filesForDwnld := make([]*model.TaskFile, 0)
		for _, file := range files {
			if file.AttemptCount < file.MaxAttempts {
				filesForDwnld = append(filesForDwnld, file)
			}
		}
		archiveDir, err := MakeArchiveDir(task.ID)
		if err != nil {
			return nil, nil, err
		}

	case model.StatusProcessing:
		resp := NewGetTaskArchive(task.Status, "task in processing")
		return resp, nil, nil
	case model.StatusFinished:
		// обработка для finished
	case model.StatusError:
		resp := NewGetTaskArchive(task.Status, "error in task processing")
		return resp, nil, nil
	default:
		// для неизвестного статуса
	}

}

func NewGetTaskArchive(status, message string) *GetTaskArchive {
	return &GetTaskArchive{
		Status:  status,
		Message: message,
	}
}

func MakeArchiveDir(taskID model.ID) (string, error) {
	baseTempDir := model.BaseTempDir
	tempDir := filepath.Join(baseTempDir, taskID.String())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create archive dir: %w", err)
	}
	return tempDir, nil
}
