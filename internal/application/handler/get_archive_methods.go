package handler

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
	"workmate_tz/internal/domain/model"
	"workmate_tz/internal/observability"
	"workmate_tz/internal/port"

	"golang.org/x/sync/errgroup"
)

func (h *AppHandler) GetArchiveWithStatusCreated(
	ctx context.Context,
	task *model.Task,
) (
	*GetTaskArchive,
	string,
	error,
) {
	task, err := ChangeTaskStatus(h.taskStorage, ctx, task, model.StatusProcessing)
	if err != nil {
		h.observer.Logger.Error().Err(err).Msgf("failed to change task %v status", task.ID)
		return nil, "", err
	}
	task, files, err := ChooseFilesForDownload(h.observer, h.fileStorage, h.taskStorage, ctx, task)
	if err != nil {
		h.observer.Logger.Error().Err(err).
			Msgf("failed to choose files for download for task %v", task.ID)
		return nil, "", err
	}
	archiveDir, err := MakeArchiveDir(task.ID)
	if err != nil {
		h.observer.Logger.Error().Err(err).Msgf("failed to create archive dir for task %v", task.ID)
		return nil, "", err
	}
	task, downloadedFiles, err := DownloadFilesWithRetry(
		ctx,
		h.observer,
		h.taskStorage,
		h.fileStorage,
		task,
		files,
		archiveDir,
	)
	if err != nil {
		h.observer.Logger.Error().Err(err).Msgf("failed to download files for task %v", task.ID)
		return nil, "", err
	}
	if len(downloadedFiles) == 0 {
		task, err := ChangeTaskStatus(h.taskStorage, ctx, task, model.StatusError)
		if err != nil {
			h.observer.Logger.Error().Err(err).Msgf("failed to change task %v status", task.ID)
			return nil, "", err
		}
		h.observer.Logger.Error().Msgf("failed to download any files for task %v", task.ID)
		return nil, "", fmt.Errorf("failed to download any files for task %v", task.ID)
	}
	zipPath := filepath.Join(archiveDir, "archive.zip")
	err = CreateZipArchive(h.observer, zipPath, downloadedFiles)
	if err != nil {
		_, err := ChangeTaskStatus(h.taskStorage, ctx, task, model.StatusError)
		if err != nil {
			h.observer.Logger.Error().Err(err).Msgf("failed to change task %v status", task.ID)
			return nil, "", err
		}
		h.observer.Logger.Error().Err(err).Msgf("failed to create zip for task %v", task.ID)
		return nil, "", fmt.Errorf("failed to create zip for task %v: %w", task.ID, err)
	}
	_, err = ChangeTaskStatus(h.taskStorage, ctx, task, model.StatusFinished)
	if err != nil {
		h.observer.Logger.Error().Err(err).Msgf("failed to change task %v status", task.ID)
		return nil, "", err
	}
	return &GetTaskArchive{}, zipPath, nil
}

func (h *AppHandler) GetArchiveWithStatusProcessing(
	ctx context.Context,
	task *model.Task,
) (
	*GetTaskArchive,
	string,
	error,
) {
	resp := NewGetTaskArchive(task.ID, model.StatusProcessing, model.MsgTaskProcessing)
	return resp, "", nil
}

func (h *AppHandler) GetArchiveWithStatusFinished(
	ctx context.Context,
	task *model.Task,
) (
	*GetTaskArchive,
	string,
	error,
) {
	archiveDir := filepath.Join(model.BaseTempDir, task.ID.String())
	zipPath := filepath.Join(archiveDir, "archive.zip")
	//zipFile, err := os.Open(zipPath)
	//if err != nil {
	//	h.observer.Logger.Error().Err(err).Msgf("failed to open zip file for task %v", task.ID)
	//	return nil, "", err
	//}
	return &GetTaskArchive{}, zipPath, nil
}

func (h *AppHandler) GetArchiveWithStatusError(
	ctx context.Context,
	task *model.Task,
) (
	*GetTaskArchive,
	string,
	error,
) {
	resp := NewGetTaskArchive(task.ID, model.StatusError, model.MsgTaskError)
	return resp, "", nil
}

func (h *AppHandler) GetArchiveWithStatusUnknown(
	ctx context.Context,
	task *model.Task,
) (
	*GetTaskArchive,
	string,
	error,
) {
	resp := NewGetTaskArchive(task.ID, task.Status, model.MsgUnknownStatus)
	return resp, "", nil
}

func ChangeTaskStatus(
	taskStorage port.TaskRepository,
	ctx context.Context,
	task *model.Task,
	status string,
) (
	*model.Task,
	error,
) {
	err := task.SetStatus(status)
	if err != nil {
		return nil, fmt.Errorf("failed to set domain task status: %w", err)
	}
	err = taskStorage.UpdateTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to update task status in storage: %w", err)
	}
	return task, nil
}

func ChangeFileStatus(
	fileStorage port.FileRepository,
	ctx context.Context,
	file *model.TaskFile,
	status string,
) (
	*model.TaskFile,
	error,
) {
	err := file.SetStatus(status)
	if err != nil {
		return nil, fmt.Errorf("failed to set domain file status: %w", err)
	}
	err = fileStorage.UpdateFile(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("failed to update file status in storage: %w", err)
	}
	return file, nil
}

func ChangeTaskError(
	taskStorage port.TaskRepository,
	ctx context.Context,
	task *model.Task,
	errorMsg string,
) (
	*model.Task,
	error,
) {
	task.SetError(errorMsg)
	err := taskStorage.UpdateTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to update task in storage: %w", err)
	}
	return task, nil
}

func ChooseFilesForDownload(
	observer *observability.Observability,
	fileStorage port.FileRepository,
	taskStorage port.TaskRepository,
	ctx context.Context,
	task *model.Task,
) (
	*model.Task,
	[]*model.TaskFile,
	error,
) {
	files, err := GetTaskFiles(observer, fileStorage, taskStorage, task, ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get files for task %v: %w", task.ID, err)
	}
	filesForDwnld := make([]*model.TaskFile, 0)
	for _, file := range files {
		if file.AttemptCount < file.MaxAttempts {
			filesForDwnld = append(filesForDwnld, file)
		}
	}
	return task, filesForDwnld, nil
}

func GetTaskFiles(
	observer *observability.Observability,
	fileStorage port.FileRepository,
	taskStorage port.TaskRepository,
	task *model.Task,
	ctx context.Context,
) ([]*model.TaskFile, error) {
	files, errIDs, err := fileStorage.GetFiles(ctx, task.Files)
	if err != nil {
		return nil, fmt.Errorf("failed to get files for task %v: %w", task.ID, err)
	}
	if len(errIDs) > 0 {
		task, err = ChangeTaskError(taskStorage, ctx, task,
			fmt.Sprintf("Failed to get from db files with ids: %v; ", errIDs))
		if err != nil {
			observer.Logger.Error().Err(err).Msgf("failed to change task %v", task.ID)
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no valid files in task")
	}
	return files, nil
}

func MakeArchiveDir(taskID model.ID) (string, error) {
	tempDir := filepath.Join(model.BaseTempDir, taskID.String())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create archive dir: %w", err)
	}
	return tempDir, nil
}

func NewGetTaskArchive(id model.ID, status, message string) *GetTaskArchive {
	return &GetTaskArchive{
		ID:      id.ToRaw(),
		Status:  status,
		Message: message,
	}
}

func DownloadFilesWithRetry(
	ctx context.Context,
	observer *observability.Observability,
	taskStorage port.TaskRepository,
	fileStorage port.FileRepository,
	task *model.Task,
	files []*model.TaskFile,
	archiveDir string,
) (
	*model.Task,
	[]string,
	error,
) {
	var mu sync.Mutex
	var errLinks []string
	downloadedFiles := make([]string, 0, len(files))

	g, ctx := errgroup.WithContext(ctx)
	for _, file := range files {
		file := file
		g.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				err := DownloadFile(observer, file.Link, archiveDir, file.Name)
				if err == nil {
					mu.Lock()
					downloadedFiles = append(downloadedFiles, filepath.Join(archiveDir, file.Name))
					mu.Unlock()
					_, err := ChangeFileStatus(fileStorage, ctx, file, model.StatusFinished)
					if err != nil {
						return err
					}
					return nil
				}

				file.IncrementAttemptCount()
				err = file.SetError(err.Error())
				if err != nil {
					observer.Logger.Error().Err(err).Msg("failed to set file error")
				}

				if file.AttemptCount >= file.MaxAttempts {
					file, err := ChangeFileStatus(fileStorage, ctx, file, model.StatusError)
					if err != nil {
						return err
					}
					mu.Lock()
					errLinks = append(errLinks, file.Link)
					mu.Unlock()
					return nil
				}

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(3 * time.Second):
					continue
				}
			}
		})
	}

	if err := g.Wait(); err != nil {
		observer.Logger.Error().Err(err).Msg("unexpected error in download group")
	}
	var err error
	if len(errLinks) > 0 {
		task, err = ChangeTaskError(taskStorage, ctx, task,
			fmt.Sprintf("Failed to download files: %v", errLinks))
		if err != nil {
			observer.Logger.Error().Err(err).Msg("failed to update task errors")
		}
	}

	return task, downloadedFiles, nil
}

func DownloadFile(
	observer *observability.Observability,
	url string,
	archiveDir, fileName string,
) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to GET %s: %w", url, err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			observer.Logger.Error().Err(err).Msg("failed to close response body")
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status %d for %s", resp.StatusCode, url)
	}

	filePath := filepath.Join(archiveDir, fileName)
	finalFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed create file %s: %w", filePath, err)
	}
	defer func(finalFile *os.File) {
		err := finalFile.Close()
		if err != nil {
			observer.Logger.Error().Err(err).Msgf("failed to close file for %s", url)
			return
		}
	}(finalFile)

	_, err = io.Copy(finalFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed write to file %s: %w", filePath, err)
	}

	return nil
}

func CreateZipArchive(
	observer *observability.Observability,
	zipPath string,
	filesPaths []string,
) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer func(zipFile *os.File) {
		err := zipFile.Close()
		if err != nil {
			observer.Logger.Error().Err(err).Msgf("failed to close zip file")
			return
		}
	}(zipFile)

	zipWriter := zip.NewWriter(zipFile)
	defer func(zipWriter *zip.Writer) {
		err := zipWriter.Close()
		if err != nil {
			observer.Logger.Error().Err(err).Msgf("failed to close zip writer")
			return
		}
	}(zipWriter)

	for _, filePath := range filesPaths {
		fileToZip, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file %s for archive: %w", filePath, err)
		}

		info, err := fileToZip.Stat()
		if err != nil {
			err := fileToZip.Close()
			if err != nil {
				return fmt.Errorf("failed to close file %s after stat error: %w", filePath, err)
			}
			return fmt.Errorf("failed to stat file %s for archive: %w", filePath, err)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			err := fileToZip.Close()
			if err != nil {
				return fmt.Errorf("failed to close file %s after header creation error: %w", filePath, err)
			}
			return fmt.Errorf("failed to create header for file %s for archive: %w", filePath, err)
		}
		header.Name = filepath.Base(filePath)
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			err := fileToZip.Close()
			if err != nil {
				return fmt.Errorf("failed to close file %s after header writing error: %w", filePath, err)
			}
			return fmt.Errorf("failed to write header for file %s: %w", filePath, err)
		}

		_, copyErr := io.Copy(writer, fileToZip)
		closeErr := fileToZip.Close()
		if copyErr != nil {
			return fmt.Errorf("failed to copy file %s to archive: %w", filePath, copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("failed to close file %s after copying to archive: %w", filePath, closeErr)
		}

	}
	return nil
}
