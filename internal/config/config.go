package config

import (
	"log/slog"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTPPort    string `env:"HTTP_PORT" env-default:"8080"`
	PostgresDSN string `env:"POSTGRES_DSN" env-default:"postgres://postgres:postgres@db:5432/booking?sslmode=disable"`
	JWTSecret   string `env:"JWT_SECRET"   env-default:"secret"`
	LogLevel    string `env:"LOG_LEVEL" env-default:"DEBUG"`
}

func MustLoad() Config {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		slog.Error("failed to read env", "error", err)
		panic(err)
	}

	return cfg
}
