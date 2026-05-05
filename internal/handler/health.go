package handler

import "github.com/gin-gonic/gin"

func Ping(c *gin.Context) {
	c.String(200, "pong")
}

func TestSentry(c *gin.Context) {
	panic("test sentry panic")
}
