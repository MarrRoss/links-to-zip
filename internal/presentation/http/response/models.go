package response

import (
	"time"
	"workmate_tz/internal/application/exception"
	"workmate_tz/internal/application/handler"
	"workmate_tz/internal/domain/model"

	"github.com/google/uuid"
)

type AddTaskResponse struct {
	ID   uuid.UUID           `json:"id"`
	Errs []FileErrorResponse `json:"errors"`
}

type GetTaskShortStatusResponse struct {
	ID      uuid.UUID  `json:"id"`
	Status  string     `json:"status"`
	EndedAt *time.Time `json:"ended_at,omitempty"`
}

type GetTaskResponse struct {
	ID        uuid.UUID      `json:"id"`
	Name      *string        `json:"name"`
	Status    string         `json:"status"`
	Files     []FileResponse `json:"files"`
	Errors    []string       `json:"errors"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	EndedAt   *time.Time     `json:"ended_at,omitempty"`
}

type FileResponse struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Link         string     `json:"link"`
	Status       string     `json:"status"`
	Error        *string    `json:"error,omitempty"`
	AttemptCount int        `json:"attempt_count"`
	MaxAttempts  int        `json:"max_attempts"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
}

func NewGetTaskResponse(task *model.Task, files []handler.GetFile) *GetTaskResponse {
	filesResp := make([]FileResponse, len(files))
	for key, file := range files {
		filesResp[key] = NewFileResponse(file)
	}
	return &GetTaskResponse{
		ID:        task.ID.ToRaw(),
		Name:      task.Name,
		Status:    task.Status,
		Files:     filesResp,
		Errors:    task.Errors,
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
		EndedAt:   task.EndedAt,
	}
}

func NewFileResponse(file handler.GetFile) FileResponse {
	return FileResponse{
		ID:           file.ID,
		Name:         file.Name,
		Link:         file.Link,
		Status:       file.Status,
		Error:        file.Error,
		AttemptCount: file.AttemptCount,
		MaxAttempts:  file.MaxAttempts,
		CreatedAt:    file.CreatedAt,
		UpdatedAt:    file.UpdatedAt,
		EndedAt:      file.EndedAt,
	}
}

//type GetTaskArchiveResponse struct {
//	Status  *string `json:"status,omitempty"`
//	Message *string `json:"message,omitempty"`
//}

type FileErrorResponse struct {
	Link string `json:"link"`
	Err  error  `json:"error"`
}

func NewAddTaskResponse(id model.ID, linksErrs []exception.FileError) *AddTaskResponse {
	errsResp := make([]FileErrorResponse, len(linksErrs))
	for key, value := range linksErrs {
		errsResp[key] = FileErrorResponse{
			Link: value.Link,
			Err:  value.Err,
		}
	}
	return &AddTaskResponse{
		ID:   id.ToRaw(),
		Errs: errsResp,
	}
}
