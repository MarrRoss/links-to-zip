package handler

import (
	"context"
	"fmt"
	"workmate_tz/internal/domain/model"

	"github.com/google/uuid"
)

type GetTaskArchiveQuery struct {
	TaskID uuid.UUID
}

type GetTaskArchive struct {
	ID      uuid.UUID
	Status  string
	Message string
}

func (h *AppHandler) GetTaskArchive(
	ctx context.Context,
	qry GetTaskArchiveQuery,
) (
	*GetTaskArchive,
	string,
	error,
) {
	id := model.UUIDtoID(qry.TaskID)
	task, err := h.taskStorage.GetTaskByID(ctx, id)
	if err != nil {
		h.observer.Logger.Error().Err(err).Msgf("failed to get task %v from storage", id)
		return nil, "", fmt.Errorf("failed to get task %v from storage: %w", id, err)
	}

	switch task.Status {
	case model.StatusCreated:
		return h.GetArchiveWithStatusCreated(ctx, task)
	case model.StatusProcessing:
		return h.GetArchiveWithStatusProcessing(ctx, task)
	case model.StatusFinished:
		return h.GetArchiveWithStatusFinished(ctx, task)
	case model.StatusError:
		return h.GetArchiveWithStatusError(ctx, task)
	default:
		return h.GetArchiveWithStatusUnknown(ctx, task)
	}
}
