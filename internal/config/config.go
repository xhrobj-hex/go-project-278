package config

import (
	"errors"
	"os"
)

type Config struct {
	DatabaseURL string
	Port        string
	BaseURL     string
	SentryDSN   string
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"), // NOTE: на Render платформа будет подсовывать порт в env PORT
		BaseURL:     os.Getenv("BASE_URL"),
		SentryDSN:   os.Getenv("SENTRY_DSN"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:" + cfg.Port
	}

	return cfg, nil
}
