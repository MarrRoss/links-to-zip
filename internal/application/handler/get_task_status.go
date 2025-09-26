package handler

import (
	"context"
	"fmt"
	"workmate_tz/internal/domain/model"

	"github.com/google/uuid"
)

type GetTaskStatusQuery struct {
	TaskID uuid.UUID
}

func (h *AppHandler) GetTaskStatus(
	ctx context.Context,
	qry GetTaskStatusQuery,
) (
	string,
	error,
) {
	id := model.UUIDtoID(qry.TaskID)
	task, err := h.taskStorage.GetTaskByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to get task: %w", err)
	}
	return task.Status, nil
}
