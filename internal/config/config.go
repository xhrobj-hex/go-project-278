package config

import (
	"errors"
	"os"
)

// Config содержит настройки приложения, загружаемые из переменных окружения.
type Config struct {
	// DatabaseURL содержит строку подключения к PostgreSQL.
	DatabaseURL string

	// Port содержит порт, на котором запускается HTTP-сервер.
	Port string

	// BaseURL содержит публичный базовый адрес приложения для генерации коротких ссылок.
	BaseURL string

	// FrontendOrigin содержит origin фронтенда, которому разрешены CORS-запросы.
	FrontendOrigin string

	// SentryDSN содержит DSN для подключения Sentry.
	SentryDSN string
}

// Load загружает конфигурацию приложения из переменных окружения.
func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		Port:           os.Getenv("PORT"), // NOTE: на Render платформа будет автоматически подсовывать порт в env PORT
		BaseURL:        os.Getenv("BASE_URL"),
		FrontendOrigin: os.Getenv("FRONTEND_ORIGIN"),
		SentryDSN:      os.Getenv("SENTRY_DSN"),
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

	if cfg.FrontendOrigin == "" {
		cfg.FrontendOrigin = "http://localhost:5173"
	}

	return cfg, nil
}
