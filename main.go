package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/xhrobj-hex/go-project-278/internal/db"
)

type linkResponse struct {
	ID          int64  `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name"`
	ShortURL    string `json:"short_url"`
}

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

	queries := db.New(dbConn)

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

func setupRouter(baseURL string, queries *db.Queries) *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	r.GET("/test-sentry", func(c *gin.Context) {
		panic("test sentry panic")
	})

	r.GET("/api/links", func(c *gin.Context) {
		links, err := queries.ListLinks(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to list links",
			})
			return
		}

		rs := make([]linkResponse, 0, len(links))
		for _, link := range links {
			rs = append(rs, linkResponse{
				ID:          link.ID,
				OriginalURL: link.OriginalUrl,
				ShortName:   link.ShortName,
				ShortURL:    buildShortURL(baseURL, link.ShortName),
			})
		}

		c.JSON(http.StatusOK, rs)
	})

	return r
}

func buildShortURL(baseURL string, shortName string) string {
	return baseURL + "/r/" + shortName
}
