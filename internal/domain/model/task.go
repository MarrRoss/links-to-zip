package model

import (
	"time"
	"workmate_tz/internal/domain/exception"
)

type Task struct {
	ID          ID
	Name        *string
	Status      string
	Files       []ID
	Errors      []string
	ArchivePath *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	EndedAt     *time.Time
}

func NewTask(name *string, fileIDs []ID) (*Task, error) {
	id := NewID()
	now := time.Now()
	newTask := Task{
		ID:          id,
		Name:        name,
		Status:      StatusCreated,
		Files:       fileIDs,
		Errors:      []string{},
		ArchivePath: nil,
		CreatedAt:   now,
		UpdatedAt:   now,
		EndedAt:     nil,
	}
	return &newTask, nil
}

func (task *Task) SetStatus(status string) error {
	//if task.EndedAt != nil {
	//	return domain.ErrTaskIsDeleted
	//}
	if status != StatusProcessing && status != StatusFinished && status != StatusError {
		return exception.ErrInvalidStatus
	}
	task.Status = status

	now := time.Now()
	task.UpdatedAt = now
	if status == StatusFinished || status == StatusError {
		task.EndedAt = &now
	}
	return nil
}

func (task *Task) SetError(err string) {
	task.Errors = append(task.Errors, err)
	task.UpdatedAt = time.Now()
}

func (task *Task) SetArchivePath(path string) error {
	//if task.EndedAt != nil {
	//	return domain.ErrTaskIsDeleted
	//}
	if path == "" {
		return exception.ErrInvalidArchivePath
	}
	task.ArchivePath = &path
	task.Status = StatusFinished
	task.UpdatedAt = time.Now()
	timeNow := time.Now()
	task.EndedAt = &timeNow
	return nil
}
