package main

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
	goose "github.com/pressly/goose/v3"
	"github.com/xhrobj-hex/go-project-278/internal/config"
	"github.com/xhrobj-hex/go-project-278/internal/database"
	store "github.com/xhrobj-hex/go-project-278/internal/db"
	"github.com/xhrobj-hex/go-project-278/internal/router"
)

//go:embed db/migrations/*.sql
var migrationsFS embed.FS

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if cfg.SentryDSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn: cfg.SentryDSN,
		})
		if err != nil {
			log.Printf("Sentry initialization failed: %v", err)
		}
		defer sentry.Flush(2 * time.Second)
	}

	dbConn, err := database.Connect(cfg.DatabaseURL)
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

	httpRouter := router.New(cfg.BaseURL, cfg.FrontendOrigin, queries)

	log.Printf("server started on port %s", cfg.Port)

	echoBanner()

	return httpRouter.Run(":" + cfg.Port)
}

func runMigrations(db *sql.DB) error {
	goose.SetBaseFS(migrationsFS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "db/migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func echoBanner() {
	fmt.Println(`
  _________.__                   __                              
 /   _____/|  |__   ____________/  |_  ____   ____   ___________ 
 \_____  \ |  |  \ /  _ \_  __ \   __\/ __ \ /    \_/ __ \_  __ \
 /        \|   Y  (  <_> )  | \/|  | \  ___/|   |  \  ___/|  | \/
/_______  /|___|  /\____/|__|   |__|  \___  >___|  /\___  >__|   
        \/      \/                        \/     \/     \/
	`)
}
