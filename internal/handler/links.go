package handler

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	store "github.com/xhrobj-hex/go-project-278/internal/db"
)

type LinksStore interface {
	ListLinks(ctx context.Context) ([]store.Link, error)
	CreateLink(ctx context.Context, arg store.CreateLinkParams) (store.Link, error)
	GetLinkByID(ctx context.Context, id int64) (store.Link, error)
	UpdateLink(ctx context.Context, arg store.UpdateLinkParams) (store.Link, error)
	DeleteLink(ctx context.Context, id int64) (int64, error)
}

type LinkHandler struct {
	baseURL string
	queries LinksStore
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

func NewLinkHandler(baseURL string, queries LinksStore) *LinkHandler {
	return &LinkHandler{
		baseURL: baseURL,
		queries: queries,
	}
}

func (h *LinkHandler) List(c *gin.Context) {
	links, err := h.queries.ListLinks(c.Request.Context())
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
			ShortURL:    buildShortURL(h.baseURL, link.ShortName),
		})
	}

	c.JSON(http.StatusOK, rs)
}

// NOTE: см. Алекс Сюй - System Design - гл. 8
func (h *LinkHandler) Create(c *gin.Context) {
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

	link, err := h.queries.CreateLink(c.Request.Context(), store.CreateLinkParams{
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
		ShortURL:    buildShortURL(h.baseURL, link.ShortName),
	})
}

func (h *LinkHandler) GetById(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id",
		})
		return
	}

	link, err := h.queries.GetLinkByID(c.Request.Context(), id)
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
		ShortURL:    buildShortURL(h.baseURL, link.ShortName),
	})
}

func (h *LinkHandler) Update(c *gin.Context) {
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

	link, err := h.queries.UpdateLink(c.Request.Context(), store.UpdateLinkParams{
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
		ShortURL:    buildShortURL(h.baseURL, link.ShortName),
	})
}

func (h *LinkHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id",
		})
		return
	}

	rowsAffected, err := h.queries.DeleteLink(c.Request.Context(), id)
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
