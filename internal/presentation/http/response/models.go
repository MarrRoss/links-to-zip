package response

import (
	"workmate_tz/internal/domain/model"

	"github.com/google/uuid"
)

type AddTaskResponse struct {
	ID uuid.UUID `json:"id"`
}

func NewAddTaskResponse(id model.ID) *AddTaskResponse {
	return &AddTaskResponse{
		ID: id.ToRaw(),
	}
}
