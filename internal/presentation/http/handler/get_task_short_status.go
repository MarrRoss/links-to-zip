package handler

import (
	"workmate_tz/internal/application/handler"
	"workmate_tz/internal/presentation/http/request"

	"github.com/gofiber/fiber/v2"
)

func (h *PresentHandler) GetTaskShortStatus(ctx *fiber.Ctx) error {
	var pathReq request.GetTaskRequest
	if err := ctx.ParamsParser(&pathReq); err != nil {
		h.observer.Logger.Trace().Err(err).Msg("failed to parse id request")
		return ctx.Status(fiber.StatusBadRequest).SendString("failed to parse id request")
	}

	qry := handler.GetTaskStatusQuery{
		TaskID: pathReq.ID,
	}
	_, status, _, err := h.appHandler.GetTaskShortStatus(ctx.UserContext(), qry)
	if err != nil {
		h.observer.Logger.Error().Err(err).Msgf("failed to get task short status")
		return ctx.Status(fiber.StatusInternalServerError).SendString("failed to get task short status")
	}
	return ctx.Status(fiber.StatusOK).JSON(status)
}
