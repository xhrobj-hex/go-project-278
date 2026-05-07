package router

import (
	"net/http"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/xhrobj-hex/go-project-278/internal/handler"
)

// New создает HTTP-роутер приложения и подключает middleware.
//
// baseURL используется при формировании коротких ссылок,
// frontendOrigin задает разрешенный origin фронтенда для CORS-запросов,
// links предоставляет доступ к хранилищу ссылок.
func New(baseURL, frontendOrigin string, links handler.LinksStore) *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			frontendOrigin,
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
		},
		ExposeHeaders: []string{
			"Content-Range",
		},
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
