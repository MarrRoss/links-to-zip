package adapter

import (
	"context"
	"fmt"
	"sync"
	"workmate_tz/internal/domain/model"
	"workmate_tz/internal/observability"
)

type TaskRepositoryImpl struct {
	observer *observability.Observability
	tasks    sync.Map
}

func NewTaskRepositoryImpl(
	observer *observability.Observability,
) (*TaskRepositoryImpl, error) {
	return &TaskRepositoryImpl{
		observer: observer,
	}, nil
}

func (repo *TaskRepositoryImpl) CreateTask(
	ctx context.Context,
	task *model.Task,
) error {
	repo.tasks.Store(task.ID, task)
	return nil
}

func (repo *TaskRepositoryImpl) GetTaskByID(
	ctx context.Context,
	id model.ID,
) (*model.Task, error) {
	value, ok := repo.tasks.Load(id)
	if !ok {
		return nil, fmt.Errorf("task not found by id: %v", id)
	}
	task, ok := value.(*model.Task)
	if !ok {
		return nil, fmt.Errorf("invalid type stored in tasks map for id: %v", id)
	}
	return task, nil
}

func (repo *TaskRepositoryImpl) PrintAllTasks() {
	repo.tasks.Range(func(key, value any) bool {
		task, ok := value.(*model.Task)
		if !ok {
			fmt.Println("invalid type stored in tasks map")
			return true
		}
		fmt.Printf("Task ID: %s, Status: %s\n", task.ID.String(), task.Status)
		return true
	})
}
