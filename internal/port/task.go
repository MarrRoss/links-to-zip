package port

import (
	"context"
	"workmate_tz/internal/domain/model"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, task *model.Task) error
	GetTaskByID(ctx context.Context, id model.ID) (*model.Task, error)
	PrintAllTasks()
}
