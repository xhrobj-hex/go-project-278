package handler

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	store "github.com/xhrobj-hex/go-project-278/internal/db"
)

// LinksStore описывает операции хранилища, которые нужны LinkHandler.
type LinksStore interface {
	ListLinks(ctx context.Context, arg store.ListLinksParams) ([]store.Link, error)
	CountLinks(ctx context.Context) (int64, error)
	CreateLink(ctx context.Context, arg store.CreateLinkParams) (store.Link, error)
	GetLinkByID(ctx context.Context, id int64) (store.Link, error)
	GetLinkByShortName(ctx context.Context, shortName string) (store.Link, error)
	UpdateLink(ctx context.Context, arg store.UpdateLinkParams) (store.Link, error)
	DeleteLink(ctx context.Context, id int64) (int64, error)

	CreateLinkVisit(ctx context.Context, arg store.CreateLinkVisitParams) (store.LinkVisit, error)
	ListLinkVisits(ctx context.Context, arg store.ListLinkVisitsParams) ([]store.LinkVisit, error)
	CountLinkVisits(ctx context.Context) (int64, error)
}

// LinkHandler обрабатывает HTTP-запросы для ссылок и аналитики посещений.
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

type linkVisitResponse struct {
	ID        int64     `json:"id"`
	LinkID    int64     `json:"link_id"`
	CreatedAt time.Time `json:"created_at"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Referer   string    `json:"referer"`
	Status    int32     `json:"status"`
}

type createLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
	ShortName   string `json:"short_name" binding:"omitempty,min=3,max=32"`
}

type linksRange struct {
	start int32
	end   int32
}

// NewLinkHandler создаёт обработчик ссылок с указанным базовым URL и хранилищем.
func NewLinkHandler(baseURL string, queries LinksStore) *LinkHandler {
	return &LinkHandler{
		baseURL: baseURL,
		queries: queries,
	}
}

// List возвращает постраничный список сокращённых ссылок.
func (h *LinkHandler) List(c *gin.Context) {
	linksRange, err := parseLinksRange(getRangeValue(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid range",
		})
		return
	}

	total, err := h.queries.CountLinks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to count links",
		})
		return
	}

	links, err := h.queries.ListLinks(c.Request.Context(), store.ListLinksParams{
		Limit:  linksRange.limit(),
		Offset: linksRange.start,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list links",
		})
		return
	}

	if len(links) == 0 {
		c.Header("Content-Range", fmt.Sprintf("links */%d", total))
	} else {
		end := linksRange.start + int32(len(links)) - 1
		c.Header("Content-Range", fmt.Sprintf("links %d-%d/%d", linksRange.start, end, total))
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

// Create создаёт новую сокращённую ссылку.
//
// NOTE: см. Алекс Сюй - System Design - гл. 8
func (h *LinkHandler) Create(c *gin.Context) {
	var req createLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindError(c, err)
		return
	}

	shortName := req.ShortName
	if shortName == "" {
		shortName = generateShortName()
	}

	link, err := h.queries.CreateLink(c.Request.Context(), store.CreateLinkParams{
		OriginalUrl: req.OriginalURL,
		ShortName:   shortName,
	})
	if err != nil {
		if isUniqueViolation(err) {
			c.JSON(http.StatusUnprocessableEntity, validationErrorResponse(
				"short_name",
				"short name already in use",
			))
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

// GetById возвращает сокращённую ссылку по числовому идентификатору.
func (h *LinkHandler) GetByID(c *gin.Context) {
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

// Update обновляет сокращённую ссылку по числовому идентификатору.
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
		handleBindError(c, err)
		return
	}

	shortName := req.ShortName
	if shortName == "" {
		shortName = generateShortName()
	}

	link, err := h.queries.UpdateLink(c.Request.Context(), store.UpdateLinkParams{
		ID:          id,
		OriginalUrl: req.OriginalURL,
		ShortName:   shortName,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "link not found",
			})
			return
		}

		if isUniqueViolation(err) {
			c.JSON(http.StatusUnprocessableEntity, validationErrorResponse(
				"short_name",
				"short name already in use",
			))
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

// Delete удаляет сокращённую ссылку по числовому идентификатору.
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

// Redirect перенаправляет короткий код на исходный URL и сохраняет посещение.
func (h *LinkHandler) Redirect(c *gin.Context) {
	code := c.Param("code")

	link, err := h.queries.GetLinkByShortName(c.Request.Context(), code)
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

	status := http.StatusFound

	_, err = h.queries.CreateLinkVisit(c.Request.Context(), store.CreateLinkVisitParams{
		LinkID:    link.ID,
		Ip:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Referer:   c.GetHeader("Referer"),
		Status:    int32(status),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create link visit",
		})
		return
	}

	c.Redirect(status, link.OriginalUrl)
}

// ListVisits возвращает постраничный список посещений сокращённых ссылок.
func (h *LinkHandler) ListVisits(c *gin.Context) {
	linksRange, err := parseLinksRange(getRangeValue(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid range",
		})
		return
	}

	total, err := h.queries.CountLinkVisits(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to count link visits",
		})
		return
	}

	visits, err := h.queries.ListLinkVisits(c.Request.Context(), store.ListLinkVisitsParams{
		Limit:  linksRange.limit(),
		Offset: linksRange.start,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to list link visits",
		})
		return
	}

	if len(visits) == 0 {
		c.Header("Content-Range", fmt.Sprintf("link_visits */%d", total))
	} else {
		end := linksRange.start + int32(len(visits)) - 1
		c.Header("Content-Range", fmt.Sprintf("link_visits %d-%d/%d", linksRange.start, end, total))
	}

	rs := make([]linkVisitResponse, 0, len(visits))
	for _, visit := range visits {
		rs = append(rs, linkVisitResponse{
			ID:        visit.ID,
			LinkID:    visit.LinkID,
			CreatedAt: visit.CreatedAt,
			IP:        visit.Ip,
			UserAgent: visit.UserAgent,
			Referer:   visit.Referer,
			Status:    visit.Status,
		})
	}

	c.JSON(http.StatusOK, rs)
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

func parseLinksRange(value string) (linksRange, error) {
	if value == "" {
		return linksRange{start: 0, end: 9}, nil
	}

	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "[") || !strings.HasSuffix(value, "]") {
		return linksRange{}, errors.New("invalid range")
	}

	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")

	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return linksRange{}, errors.New("invalid range")
	}

	start, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 32)
	if err != nil {
		return linksRange{}, errors.New("invalid range")
	}

	end, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 32)
	if err != nil {
		return linksRange{}, errors.New("invalid range")
	}

	if start < 0 || end < start {
		return linksRange{}, errors.New("invalid range")
	}

	return linksRange{
		start: int32(start),
		end:   int32(end),
	}, nil
}

func (r linksRange) limit() int32 {
	return r.end - r.start + 1
}

func getRangeValue(c *gin.Context) string {
	if value := c.Query("range"); value != "" {
		return value
	}

	return c.GetHeader("Range")
}
