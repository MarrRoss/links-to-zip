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
	exceptionAppl "workmate_tz/internal/application/exception"
	"workmate_tz/internal/domain/model"
	"workmate_tz/internal/observability"
	"workmate_tz/internal/port"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type GetTaskArchiveQuery struct {
	TaskID uuid.UUID
}

type GetTaskArchive struct {
	ID      uuid.UUID
	Status  string
	Message string
}

func (h *AppHandler) GetTaskArchive(
	ctx context.Context,
	qry GetTaskArchiveQuery,
) (
	*GetTaskArchive,
	[]exceptionAppl.FileError,
	io.Reader,
	error,
) {
	id := model.UUIDtoID(qry.TaskID)
	task, err := h.taskStorage.GetTaskByID(ctx, id)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get task: %w", err)
	}

	switch task.Status {
	case model.StatusCreated:
		task, err := ChangeTaskStatus(h.taskStorage, ctx, task, model.StatusProcessing)
		if err != nil {
			return nil, nil, nil, err
		}
		files, errsFromBD, err := ChooseFilesForDownload(h.fileStorage, ctx, task)
		if err != nil {
			return nil, errsFromBD, nil, err
		}
		archiveDir, err := MakeArchiveDir(task.ID)
		if err != nil {
			return nil, errsFromBD, nil, err
		}
		downloadedFiles, fileErrors, err := DownloadFilesWithRetry(ctx, h.observer, files, archiveDir)
		allErrs := append(errsFromBD, fileErrors...)
		if err != nil {
			return nil, allErrs, nil, err
		}
		if len(downloadedFiles) == 0 {
			task, err := ChangeTaskStatus(h.taskStorage, ctx, task, model.StatusError)
			if err != nil {
				return nil, allErrs, nil, err
			}
			return nil, allErrs, nil, fmt.Errorf("failed to download files for task %v, errors: %v", task.ID, allErrs)
		}
		zipPath := filepath.Join(archiveDir, "archive.zip")
		err = CreateZipArchive(h.observer, zipPath, downloadedFiles)
		if err != nil {
			_, err := ChangeTaskStatus(h.taskStorage, ctx, task, model.StatusError)
			if err != nil {
				return nil, nil, err
			}
			return nil, nil, fmt.Errorf("failed to create zip: %w", err)
		}
		zipFile, err := os.Open(zipPath)
		if err != nil {
			_, err := ChangeTaskStatus(h.taskStorage, ctx, task, model.StatusError)
			if err != nil {
				return nil, nil, err
			}
			return nil, nil, fmt.Errorf("failed to open zip file: %w", err)
		}
		_, err = ChangeTaskStatus(h.taskStorage, ctx, task, model.StatusFinished)
		if err != nil {
			return nil, nil, err
		}
		return &GetTaskArchive{
			Status:  model.StatusFinished,
			Message: "Archive created successfully",
		}, zipFile, nil

	case model.StatusProcessing:
		resp := NewGetTaskArchive(task.Status, "task in processing")
		return resp, nil, nil
	case model.StatusFinished:
		// обработка для finished
	case model.StatusError:
		resp := NewGetTaskArchive(task.Status, "error in task processing")
		return resp, nil, nil
	default:
		// для неизвестного статуса
	}

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

func ChooseFilesForDownload(
	fileStorage port.FileRepository,
	ctx context.Context,
	task *model.Task,
) (
	[]*model.TaskFile,
	[]exceptionAppl.FileError,
	error,
) {
	files, errIDs, err := fileStorage.GetFiles(ctx, task.Files)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get files for task %v: %w", task.ID, err)
	}
	errsFromBD := exceptionAppl.ErrIDsToErrStructs(errIDs, exceptionAppl.MsgFileNotFound)
	if len(files) == 0 {
		return nil, errsFromBD, fmt.Errorf("no valid files in task")
	}
	filesForDwnld := make([]*model.TaskFile, 0)
	for _, file := range files {
		if file.AttemptCount < file.MaxAttempts {
			filesForDwnld = append(filesForDwnld, file)
		}
	}
	return filesForDwnld, errsFromBD, nil
}

func MakeArchiveDir(taskID model.ID) (string, error) {
	baseTempDir := model.BaseTempDir
	tempDir := filepath.Join(baseTempDir, taskID.String())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create archive dir: %w", err)
	}
	return tempDir, nil
}

func NewGetTaskArchive(status, message string) *GetTaskArchive {
	return &GetTaskArchive{
		Status:  status,
		Message: message,
	}
}

func DownloadFilesWithRetry(
	ctx context.Context,
	observer *observability.Observability,
	files []*model.TaskFile,
	archiveDir string,
) (
	[]string,
	[]exceptionAppl.FileError,
	error,
) {
	var mu sync.Mutex
	var fileErrors []exceptionAppl.FileError
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
					file.Status = model.StatusFinished
					return nil
				}

				file.AttemptCount++
				file.Error = err.Error()
				file.Status = model.StatusError

				if file.AttemptCount >= file.MaxAttempts {
					mu.Lock()
					fileErrors = append(fileErrors, exceptionAppl.FileError{Link: file.Link, Err: err})
					mu.Unlock()
					return nil
				}
				// повторить попытку после ошибки
			}
		})
	}

	if err := g.Wait(); err != nil {
		return nil, nil, fmt.Errorf("some downloads failed: %w", err)
	}

	return downloadedFiles, fileErrors, nil
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
