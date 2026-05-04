package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	store "github.com/xhrobj-hex/go-project-278/internal/db"
)

type fakeLinksStore struct {
	links     []store.Link
	listErr   error
	created   store.Link
	createErr error
}

func (f fakeLinksStore) ListLinks(ctx context.Context) ([]store.Link, error) {
	return f.links, f.listErr
}

func (f fakeLinksStore) CreateLink(ctx context.Context, arg store.CreateLinkParams) (store.Link, error) {
	return f.created, f.createErr
}

func TestPing(t *testing.T) {
	r := setupRouter("http://localhost:8080", nil)

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

func TestListLinksEmpty(t *testing.T) {
	r := setupRouter("http://localhost:8080", fakeLinksStore{})

	rq := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusOK; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	if got, want := rc.Body.String(), "[]"; got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}
}

func TestCreateLink(t *testing.T) {
	r := setupRouter("http://localhost:8080", fakeLinksStore{
		created: store.Link{
			ID:          1,
			OriginalUrl: "https://example.com/long-url",
			ShortName:   "exmpl",
		},
	})

	body := `{"original_url":"https://example.com/long-url","short_name":"exmpl"}`
	rq := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(body))
	rq.Header.Set("Content-Type", "application/json")
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusCreated; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	want := `{"id":1,"original_url":"https://example.com/long-url","short_name":"exmpl","short_url":"http://localhost:8080/r/exmpl"}`
	if got := rc.Body.String(); got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}
}

func TestCreateLinkRequiresOriginalURL(t *testing.T) {
	r := setupRouter("http://localhost:8080", fakeLinksStore{})

	body := `{"short_name":"exmpl"}`
	rq := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(body))
	rq.Header.Set("Content-Type", "application/json")
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusBadRequest; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	if got, want := rc.Body.String(), `{"error":"original_url is required"}`; got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}
}
