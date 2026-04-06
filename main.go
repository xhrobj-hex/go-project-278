package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.New()

	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
