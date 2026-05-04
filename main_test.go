package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	store "github.com/xhrobj-hex/go-project-278/internal/db"
)

type fakeLinksLister struct {
	links []store.Link
	err   error
}

func (f fakeLinksLister) ListLinks(ctx context.Context) ([]store.Link, error) {
	return f.links, f.err
}

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

func TestListLinksEmpty(t *testing.T) {
	r := setupRouter("http://localhost:8080", fakeLinksLister{})

	rq := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if rc.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rc.Code)
	}

	if rc.Body.String() != "[]" {
		t.Fatalf("expected body %q, got %q", "[]", rc.Body.String())
	}
}
