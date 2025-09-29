package model

const (
	StatusCreated    = "created"
	StatusProcessing = "processing"
	StatusError      = "error"
	StatusFinished   = "finished"

	MaxAttemptsForFile = 3

	BaseTempDir = "./tmp"

	MsgTaskProcessing = "task in process"
	MsgTaskError      = "error in task processing"
	MsgUnknownStatus  = "unknown task status"
)
