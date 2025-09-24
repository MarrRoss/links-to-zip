package handler

import (
	"workmate_tz/internal/observability"
	"workmate_tz/internal/port"
)

type AppHandler struct {
	observer    *observability.Observability
	fileStorage port.FileRepository
	taskStorage port.TaskRepository
}

func NewAppHandler(
	observer *observability.Observability,
	fileStorage port.FileRepository,
	taskStorage port.TaskRepository,
) *AppHandler {
	return &AppHandler{
		observer:    observer,
		fileStorage: fileStorage,
		taskStorage: taskStorage,
	}
}
