package config

import (
	"fmt"
	"workmate_tz/config/api"
)

type Factory struct {
	API api.Config
}

func New() *Factory {
	cfg := Load()
	fmt.Printf("Host: %v, Port: %v", cfg.API.Host, cfg.API.Port)
	return cfg
}
