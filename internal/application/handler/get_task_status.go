package handler

import (
	"context"
	"fmt"
	"workmate_tz/internal/domain/model"
)

func (h *AppHandler) GetTaskStatus(
	ctx context.Context,
	qry GetTaskStatusQuery,
) (
	*model.Task,
	[]GetFile,
	error,
) {
	id := model.UUIDtoID(qry.TaskID)
	task, err := h.taskStorage.GetTaskByID(ctx, id)
	if err != nil {
		h.observer.Logger.Error().Err(err).Msgf("failed to get task %v from storage", id)
		return nil, nil, fmt.Errorf("failed to get task %v from storage: %w", id, err)
	}
	files, err := GetTaskFiles(h.observer, h.fileStorage, h.taskStorage, task, ctx)
	if err != nil {
		h.observer.Logger.Error().Err(err).Msgf("failed to get files for task %v", task.ID)
		return nil, nil, fmt.Errorf("failed to get files for task %v: %w", task.ID, err)
	}
	filesStruct := SliceToGetFiles(files)
	return task, filesStruct, nil
}
