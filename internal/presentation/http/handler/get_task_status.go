package handler

import (
	"workmate_tz/internal/application/handler"
	"workmate_tz/internal/presentation/http/request"
	"workmate_tz/internal/presentation/http/response"

	"github.com/gofiber/fiber/v2"
)

func (h *PresentHandler) GetTaskStatus(ctx *fiber.Ctx) error {
	var pathReq request.GetTaskRequest
	if err := ctx.ParamsParser(&pathReq); err != nil {
		h.observer.Logger.Trace().Err(err).Msg("failed to parse id request")
		return ctx.Status(fiber.StatusBadRequest).SendString("failed to parse id request")
	}

	qry := handler.GetTaskStatusQuery{
		TaskID: pathReq.ID,
	}
	task, files, err := h.appHandler.GetTaskStatus(ctx.UserContext(), qry)
	if err != nil {
		h.observer.Logger.Error().Err(err).Msgf("failed to get task status")
		return ctx.Status(fiber.StatusInternalServerError).SendString("failed to get task status")
	}
	return ctx.Status(fiber.StatusOK).JSON(response.NewGetTaskResponse(task, files))
}
