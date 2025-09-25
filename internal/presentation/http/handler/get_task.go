package handler

import (
	"fmt"
	"workmate_tz/internal/presentation/http/request"
	"workmate_tz/internal/presentation/http/response"

	"github.com/gofiber/fiber/v2"
)

func (h *PresentHandler) GetTask(ctx *fiber.Ctx) error {
	var bodyReq request.AddTaskRequest
	if err := ctx.BodyParser(&bodyReq); err != nil {
		h.observer.Logger.Trace().Err(err).Msg("failed to parse body request")
		return fmt.Errorf("failed to parse body request")
	}

	cmd := request.AddTaskRequestToCommand(bodyReq)
	id, linksErrs, err := h.appHandler.CreateTask(ctx.UserContext(), cmd)
	if err != nil {
		h.observer.Logger.Error().Err(err).Msgf("failed to create task")
		return fmt.Errorf("failed to create task")
	}
	return ctx.Status(fiber.StatusOK).JSON(response.NewAddTaskResponse(id, linksErrs))
}
