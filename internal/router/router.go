package router

import (
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/xhrobj-hex/go-project-278/internal/handler"
)

func New(baseURL string, links handler.LinksStore) *gin.Engine {
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
