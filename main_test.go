package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPing(t *testing.T) {
	r := setupRouter("http://localhost:8080", nil)

	rq := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if rc.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rc.Code)
	}

	if rc.Body.String() != "pong" {
		t.Fatalf("expected body %q, got %q", "pong", rc.Body.String())
	}
}
