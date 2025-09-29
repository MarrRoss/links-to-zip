package handler

import (
	"io"
	"workmate_tz/internal/application/handler"
	"workmate_tz/internal/presentation/http/request"

	"github.com/gofiber/fiber/v2"
)

func (h *PresentHandler) GetTaskArchive(ctx *fiber.Ctx) error {
	var pathReq request.GetTaskRequest
	if err := ctx.ParamsParser(&pathReq); err != nil {
		h.observer.Logger.Trace().Err(err).Msg("failed to parse id request")
		return ctx.Status(fiber.StatusBadRequest).SendString("failed to parse id request")
	}

	qry := handler.GetTaskArchiveQuery{
		TaskID: pathReq.ID,
	}
	answer, archive, err := h.appHandler.GetTaskArchive(ctx.UserContext(), qry)
	if err != nil {
		h.observer.Logger.Error().Err(err).Msgf("failed to get task archive")
		return ctx.JSON("failed to get task archive")
	}

	if archive != nil {
		defer func(archive io.ReadCloser) {
			err := archive.Close()
			if err != nil {
				h.observer.Logger.Error().Err(err).Msg("failed to close archive")
				return
			}
		}(archive)
		ctx.Set("Content-Disposition", "attachment; filename=\"archive.zip\"")
		ctx.Set("Content-Type", "application/zip")
		return ctx.SendStream(archive)
	}

	return ctx.JSON(answer)
}
