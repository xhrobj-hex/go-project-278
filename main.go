package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	router := setupRouter()

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func setupRouter() *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	return r
}
