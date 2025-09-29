package model

import (
	"fmt"
	"net/url"
	"path"
	"time"
	"workmate_tz/internal/domain/exception"
)

type TaskFile struct {
	ID           ID
	Name         string
	Link         string
	Status       string
	Error        string
	AttemptCount int
	MaxAttempts  int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	EndedAt      *time.Time
}

func NewFile(link string, parsedLink url.URL) (*TaskFile, error) {
	if link == "" {
		return nil, exception.ErrInvalidFileLink
	}
	id := NewID()
	name := CreateFileName(link, parsedLink, id)
	now := time.Now()
	newFileLink := TaskFile{
		ID:           id,
		Name:         name,
		Link:         link,
		Status:       StatusCreated,
		Error:        "",
		AttemptCount: 0,
		MaxAttempts:  MaxAttemptsForFile,
		CreatedAt:    now,
		UpdatedAt:    now,
		EndedAt:      nil,
	}
	return &newFileLink, nil
}

func CreateFileName(link string, parsedLink url.URL, id ID) string {
	defaultName := fmt.Sprintf("file-%s%s", id, path.Ext(link))
	base := path.Base(parsedLink.Path)
	if base == "/" || base == "." || base == "" {
		return defaultName
	}
	return base
}

func (file *TaskFile) SetStatus(status string) error {
	//if file.EndedAt != nil {
	//	return domain.ErrFileIsDeleted
	//}
	if status != StatusProcessing && status != StatusFinished && status != StatusError {
		return exception.ErrInvalidStatus
	}
	file.Status = status

	now := time.Now()
	file.UpdatedAt = now
	if status == StatusFinished || status == StatusError {
		file.EndedAt = &now
	}
	return nil
}

func (file *TaskFile) IncrementAttemptCount() {
	//if file.EndedAt != nil {
	//	return domain.ErrFileIsDeleted
	//}
	file.AttemptCount++
	file.UpdatedAt = time.Now()
}

func (file *TaskFile) SetError(err string) error {
	//if file.EndedAt != nil {
	//	return domain.ErrFileIsDeleted
	//}
	//if err == "" {
	//	return exception.ErrInvalidFileError
	//}
	file.Error = err
	//file.Status = StatusError
	file.UpdatedAt = time.Now()
	return nil
}
