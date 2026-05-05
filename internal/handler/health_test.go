package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/ping", Ping)

	rq := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusOK; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	if got, want := rc.Body.String(), "pong"; got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}
}
