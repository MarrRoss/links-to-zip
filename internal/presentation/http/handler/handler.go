package handler

import (
	"workmate_tz/internal/application/handler"
	"workmate_tz/internal/observability"
)

type PresentHandler struct {
	observer   *observability.Observability
	appHandler *handler.AppHandler
}

func NewPresentHandler(
	observer *observability.Observability,
	appHandler *handler.AppHandler,
) *PresentHandler {
	return &PresentHandler{
		observer:   observer,
		appHandler: appHandler,
	}
}
