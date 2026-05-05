package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	pq "github.com/lib/pq"
	goose "github.com/pressly/goose/v3"
	store "github.com/xhrobj-hex/go-project-278/internal/db"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type linksStore interface {
	ListLinks(ctx context.Context) ([]store.Link, error)
	CreateLink(ctx context.Context, arg store.CreateLinkParams) (store.Link, error)
	GetLinkByID(ctx context.Context, id int64) (store.Link, error)
	UpdateLink(ctx context.Context, arg store.UpdateLinkParams) (store.Link, error)
	DeleteLink(ctx context.Context, id int64) (int64, error)
}

type linkResponse struct {
	ID          int64  `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name"`
	ShortURL    string `json:"short_url"`
}

type createLinkRequest struct {
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name"`
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

func setupRouter(baseURL string, queries linksStore) *gin.Engine {
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

	r.POST("/api/links", func(c *gin.Context) {
		var req createLinkRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request body",
			})
			return
		}

		if req.OriginalURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "original_url is required",
			})
			return
		}

		shortName := req.ShortName
		if shortName == "" {
			shortName = generateShortName()
			// ???: стоит ли проверять сгенеренное имя на уникальность?
			// даже при n = 6 это будет невероятное везение ...
		}

		link, err := queries.CreateLink(c.Request.Context(), store.CreateLinkParams{
			OriginalUrl: req.OriginalURL,
			ShortName:   shortName,
		})
		if err != nil {
			if isUniqueViolation(err) {
				c.JSON(http.StatusConflict, gin.H{
					"error": "short_name already exists",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to create link",
			})
			return
		}

		c.JSON(http.StatusCreated, linkResponse{
			ID:          link.ID,
			OriginalURL: link.OriginalUrl,
			ShortName:   link.ShortName,
			ShortURL:    buildShortURL(baseURL, link.ShortName),
		})
	})

	r.GET("/api/links/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid id",
			})
			return
		}

		link, err := queries.GetLinkByID(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "link not found",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get link",
			})
			return
		}

		c.JSON(http.StatusOK, linkResponse{
			ID:          link.ID,
			OriginalURL: link.OriginalUrl,
			ShortName:   link.ShortName,
			ShortURL:    buildShortURL(baseURL, link.ShortName),
		})
	})

	r.PUT("/api/links/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid id",
			})
			return
		}

		var req createLinkRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request body",
			})
			return
		}

		if req.OriginalURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "original_url is required",
			})
			return
		}

		link, err := queries.UpdateLink(c.Request.Context(), store.UpdateLinkParams{
			ID:          id,
			OriginalUrl: req.OriginalURL,
			ShortName:   req.ShortName,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "link not found",
				})
				return
			}

			if isUniqueViolation(err) {
				c.JSON(http.StatusConflict, gin.H{
					"error": "short_name already exists",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to update link",
			})
			return
		}

		c.JSON(http.StatusOK, linkResponse{
			ID:          link.ID,
			OriginalURL: link.OriginalUrl,
			ShortName:   link.ShortName,
			ShortURL:    buildShortURL(baseURL, link.ShortName),
		})
	})

	r.DELETE("/api/links/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid id",
			})
			return
		}

		rowsAffected, err := queries.DeleteLink(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to delete link",
			})
			return
		}

		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "link not found",
			})
			return
		}

		c.Status(http.StatusNoContent)
	})

	return r
}

func buildShortURL(baseURL, shortName string) string {
	return strings.TrimRight(baseURL, "/") + "/r/" + shortName
}

func generateShortName() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const n = 6 // NOTE: 62 ^ 6 = 56_800_235_584

	b := make([]byte, n)
	for i := range b {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "generated"
		}
		b[i] = alphabet[randomIndex.Int64()]
	}

	return string(b)
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	// NOTE: https://www.postgresql.org/docs/current/plpgsql-errors-and-messages.html
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
