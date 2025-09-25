package response

import (
	"time"
	"workmate_tz/internal/application/handler"
	"workmate_tz/internal/domain/model"

	"github.com/google/uuid"
)

type AddTaskResponse struct {
	ID   uuid.UUID           `json:"id"`
	Errs []FileErrorResponse `json:"errors"`
}

type GetTaskResponse struct {
	ID        uuid.UUID      `json:"id"`
	Name      *string        `json:"name"`
	Status    string         `json:"status"`
	Files     []FileResponse `json:"files"`
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

type FileErrorResponse struct {
	Link string `json:"link"`
	Err  error  `json:"error"`
}

func NewAddTaskResponse(id model.ID, linksErrs []handler.FileError) *AddTaskResponse {
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
