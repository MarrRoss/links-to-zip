package config

import (
	"github.com/caarlos0/env/v8"
	"sync"
)

var (
	FactoryInst Factory
	onlyOnce    sync.Once
)

func Load() *Factory {
	onlyOnce.Do(func() {
		if err := env.Parse(&FactoryInst); err != nil {
			panic(err)
		}
	})
	return &FactoryInst
}
