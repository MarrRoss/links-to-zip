package handler

import (
	"context"
	"fmt"
	"sync"
	"workmate_tz/internal/domain/exception"
	"workmate_tz/internal/domain/model"

	"golang.org/x/sync/errgroup"
)

type AddTaskCommand struct {
	TaskName *string
	Files    []AddFileCommand
}

type AddFileCommand struct {
	FileName *string
	Link     string
}

type FileError struct {
	Link string
	Err  error
}

func (h *AppHandler) CreateTask(
	ctx context.Context,
	cmd AddTaskCommand,
) (
	model.ID,
	[]FileError,
	error,
) {
	var g errgroup.Group
	var mu sync.Mutex

	filesIDs := make([]model.ID, len(cmd.Files))
	var fileErrors []FileError

	for _, file := range cmd.Files {
		file := file

		g.Go(func() error {
			valid, parsedLink := IsValidURL(file.Link)
			if !valid {
				mu.Lock()
				fileErrors = append(fileErrors, FileError{
					Link: file.Link,
					Err:  exception.ErrInvalidURL,
				})
				mu.Unlock()
				h.observer.Logger.Trace().Msgf("invalid url format, link %v", file.Link)
				return nil
			}
			domainFile, err := model.NewFile(file.Link, parsedLink)
			if err != nil {
				mu.Lock()
				fileErrors = append(fileErrors, FileError{Link: file.Link, Err: err})
				mu.Unlock()
				h.observer.Logger.Trace().Err(err).
					Msgf("failed to create domain file from link %v", file.Link)
				return nil
			}
			err = h.fileStorage.CreateFile(ctx, domainFile)
			if err != nil {
				mu.Lock()
				fileErrors = append(fileErrors, FileError{Link: file.Link, Err: err})
				mu.Unlock()
				h.observer.Logger.Trace().Err(err).
					Msgf("failed to add file to storage, link %v", file.Link)
				return nil
			}
			mu.Lock()
			filesIDs = append(filesIDs, domainFile.ID)
			mu.Unlock()
			return nil
		})
	}
	if len(filesIDs) == 0 {
		return model.ID{}, fileErrors, fmt.Errorf("no valid links to create task")
	}

	h.fileStorage.PrintAllFiles()
	newTask, err := model.NewTask(cmd.TaskName, filesIDs)
	if err != nil {
		h.observer.Logger.Trace().Err(err).Msg("failed to create domain task")
		return model.ID{}, fileErrors, err
	}
	err = h.taskStorage.CreateTask(ctx, newTask)
	if err != nil {
		h.observer.Logger.Trace().Err(err).Msg("failed to add task to storage")
		return model.ID{}, fileErrors, err
	}
	h.taskStorage.PrintAllTasks()
	return newTask.ID, fileErrors, nil
}
