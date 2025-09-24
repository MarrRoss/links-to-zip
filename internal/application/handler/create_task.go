package handler

import (
	"context"
	"workmate_tz/internal/domain/exception"
	"workmate_tz/internal/domain/model"
)

type AddTaskCommand struct {
	TaskName *string
	Files    []AddFileCommand
}

type AddFileCommand struct {
	FileName *string
	Link     string
}

func (h *AppHandler) CreateTask(ctx context.Context, cmd AddTaskCommand) (model.ID, error) {
	filesIDs := make([]model.ID, len(cmd.Files))
	for key, file := range cmd.Files {
		valid, parsedLink := IsValidURL(file.Link)
		if !valid {
			h.observer.Logger.Trace().Msg("invalid url format")
			return model.ID{}, exception.ErrInvalidURL
		}
		domainFile, err := model.NewFile(file.Link, parsedLink)
		if err != nil {
			h.observer.Logger.Trace().Err(err).
				Msgf("failed to create domain file from link %v", file.Link)
			return model.ID{}, err
		}
		err = h.fileStorage.CreateFile(ctx, domainFile)
		if err != nil {
			h.observer.Logger.Trace().Err(err).Msg("failed to add file to storage")
			return model.ID{}, err
		}
		filesIDs[key] = domainFile.ID
	}
	h.fileStorage.PrintAllFiles()
	newTask, err := model.NewTask(cmd.TaskName, filesIDs)
	if err != nil {
		h.observer.Logger.Trace().Err(err).Msg("failed to create domain task")
		return model.ID{}, err
	}
	err = h.taskStorage.CreateTask(ctx, newTask)
	if err != nil {
		h.observer.Logger.Trace().Err(err).Msg("failed to add task to storage")
		return model.ID{}, err
	}
	h.taskStorage.PrintAllTasks()
	return newTask.ID, nil
}
