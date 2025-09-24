package observability

import "github.com/rs/zerolog"

type Observability struct {
	Logger zerolog.Logger
}

func New(logger zerolog.Logger) *Observability {
	return &Observability{
		Logger: logger,
	}
}

func (o Observability) GetLogger() *zerolog.Logger {
	return &o.Logger
}
