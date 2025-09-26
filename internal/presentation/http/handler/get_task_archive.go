package handler

import (
	"fmt"
	"workmate_tz/internal/application/handler"
	"workmate_tz/internal/presentation/http/request"

	"github.com/gofiber/fiber/v2"
)

func (h *PresentHandler) GetTaskArchive(ctx *fiber.Ctx) error {
	var pathReq request.GetTaskRequest
	if err := ctx.ParamsParser(&pathReq); err != nil {
		h.observer.Logger.Trace().Err(err).Msg("failed to parse id request")
		return fmt.Errorf("failed to parse id request")
	}

	qry := handler.GetTaskArchiveQuery{
		TaskID: pathReq.ID,
	}
	archive, err := h.appHandler.GetTaskArchive(ctx.UserContext(), qry)
	if err != nil {
		h.observer.Logger.Error().Err(err).Msgf("failed to get task archive")
		return fmt.Errorf("failed to get task archive")
	}
	return ctx.Status(fiber.StatusOK).JSON(archive)
}
