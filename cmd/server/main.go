package main

import (
	"fmt"
	"log"
	"workmate_tz/config"
	"workmate_tz/internal/adapter"
	handler2 "workmate_tz/internal/application/handler"
	"workmate_tz/internal/observability"
	"workmate_tz/internal/presentation/http/handler"
	"workmate_tz/pkg/logging"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog"
)

func main() {
	cfg := config.New()
	logger := logging.NewZeroLogger(zerolog.TraceLevel)
	observer := observability.New(logger)
	dbTask, err := adapter.NewTaskRepositoryImpl(observer)
	if err != nil {
		log.Fatalf("failed to connect to storage: %v", err)
	}
	dbFile, err := adapter.NewFileRepositoryImpl(observer)
	if err != nil {
		log.Fatalf("failed to connect to storage: %v", err)
	}
	appHandler := handler2.NewAppHandler(observer, dbFile, dbTask)
	presentHandler := handler.NewPresentHandler(observer, appHandler)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(recover.New(recover.Config{EnableStackTrace: true}))
	app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.API.OriginCORS,
		AllowCredentials: cfg.API.AllowCredentials,
	}))

	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: observer.GetLogger(),
		Fields: []string{
			fiberzerolog.FieldIP,
			fiberzerolog.FieldUserAgent,
			fiberzerolog.FieldLatency,
			fiberzerolog.FieldStatus,
			fiberzerolog.FieldMethod,
			fiberzerolog.FieldURL,
			fiberzerolog.FieldError,
			fiberzerolog.FieldBytesReceived,
			fiberzerolog.FieldBytesSent,
		},
	}))

	app.Post("/tasks", presentHandler.CreateTask)
	app.Get("/tasks/:id", presentHandler.GetTask)

	observer.GetLogger().Info().Msg("Starting app")
	err = app.Listen(fmt.Sprintf(":%s", cfg.API.Port))
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
