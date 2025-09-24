package logging

import (
	"github.com/rs/zerolog"
	"os"
	"time"
)

func NewZeroLogger(level zerolog.Level) zerolog.Logger {
	logger := zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
	).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()
	return logger
}
