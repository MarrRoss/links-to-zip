package api

type Config struct {
	Host             string `env:"API_HOST" envDefault:"0.0.0.0"`
	Port             string `env:"API_PORT" envDefault:"8080"`
	OriginCORS       string `env:"API_ORIGIN_CORS" envDefault:"*"`
	AllowCredentials bool   `env:"API_ALLOW_CREDENTIALS" envDefault:"false"`
}
