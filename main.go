package main

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	goose "github.com/pressly/goose/v3"
	store "github.com/xhrobj-hex/go-project-278/internal/db"
	"github.com/xhrobj-hex/go-project-278/internal/handler"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type config struct {
	databaseURL string
	port        string
	baseURL     string
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if err := initSentry(); err != nil {
		log.Printf("Sentry initialization failed: %v", err)
	}
	defer sentry.Flush(2 * time.Second)

	dbConn, err := connectDB(cfg.databaseURL)
	if err != nil {
		return err
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()

	if err := runMigrations(dbConn); err != nil {
		return err
	}

	queries := store.New(dbConn)

	router := setupRouter(cfg.baseURL, queries)

	log.Printf("server started on port %s", cfg.port)

	return router.Run(":" + cfg.port)
}

func loadConfig() (config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return config{}, errors.New("DATABASE_URL is required")
	}

	port := os.Getenv("PORT") // NOTE: на Render платформа будет подсовывать порт в env PORT
	if port == "" {
		port = "8080"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:" + port
	}

	return config{
		databaseURL: databaseURL,
		port:        port,
		baseURL:     baseURL,
	}, nil
}

func initSentry() error {
	return sentry.Init(sentry.ClientOptions{}) // NOTE: Dsn берется из env SENTRY_DSN
}

func connectDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	goose.SetBaseFS(migrationsFS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func setupRouter(baseURL string, links handler.LinksStore) *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))

	r.GET("/ping", handler.Ping)
	r.GET("/test-sentry", handler.TestSentry)

	linkHandler := handler.NewLinkHandler(baseURL, links)

	r.GET("/api/links", linkHandler.List)
	r.POST("/api/links", linkHandler.Create)
	r.GET("/api/links/:id", linkHandler.GetById)
	r.PUT("/api/links/:id", linkHandler.Update)
	r.DELETE("/api/links/:id", linkHandler.Delete)

	return r
}
