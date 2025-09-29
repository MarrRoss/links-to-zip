package handler

import (
	"context"
	"fmt"
	"time"
	"workmate_tz/internal/domain/model"
)

func (h *AppHandler) GetTaskShortStatus(
	ctx context.Context,
	qry GetTaskStatusQuery,
) (
	model.ID,
	string,
	*time.Time,
	error,
) {
	// todo отдавать структуру
	id := model.UUIDtoID(qry.TaskID)
	task, err := h.taskStorage.GetTaskByID(ctx, id)
	if err != nil {
		h.observer.Logger.Error().Err(err).Msgf("failed to get task %v from storage", id)
		return model.ID{}, "", nil, fmt.Errorf("failed to get task %v from storage: %w", id, err)
	}
	return task.ID, task.Status, task.EndedAt, nil
}
