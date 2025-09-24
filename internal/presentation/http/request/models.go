package request

import "workmate_tz/internal/application/handler"

type AddTaskRequest struct {
	TaskName *string        `json:"task_name,omitempty"`
	Files    []FilesRequest `json:"files"`
}

type FilesRequest struct {
	FileName *string `json:"file_name,omitempty"`
	Link     string  `json:"link"`
}

func AddTaskRequestToCommand(req AddTaskRequest) handler.AddTaskCommand {
	files := make([]handler.AddFileCommand, len(req.Files))
	for key, file := range req.Files {
		files[key] = handler.AddFileCommand{
			FileName: file.FileName,
			Link:     file.Link,
		}
	}
	return handler.AddTaskCommand{
		TaskName: req.TaskName,
		Files:    files,
	}
}
