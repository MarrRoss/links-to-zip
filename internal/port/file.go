package port

import (
	"context"
	"workmate_tz/internal/domain/model"
)

type FileRepository interface {
	CreateFile(ctx context.Context, file *model.TaskFile) error
	PrintAllFiles()
}
