package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	router.GET("/panic", func(c *gin.Context) {
		panic("test paic")
	})

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
