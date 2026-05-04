package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pq "github.com/lib/pq"
	store "github.com/xhrobj-hex/go-project-278/internal/db"
)

type fakeLinksStore struct {
	links         []store.Link
	listErr       error
	created       store.Link
	createErr     error
	lastCreateArg store.CreateLinkParams
}

func (f fakeLinksStore) ListLinks(ctx context.Context) ([]store.Link, error) {
	return f.links, f.listErr
}

func (f *fakeLinksStore) CreateLink(ctx context.Context, arg store.CreateLinkParams) (store.Link, error) {
	f.lastCreateArg = arg
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
	r := setupRouter("http://localhost:8080", &fakeLinksStore{})

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
	storeFake := &fakeLinksStore{
		created: store.Link{
			ID:          1,
			OriginalUrl: "https://example.com/long-url",
			ShortName:   "exmpl",
		},
	}
	r := setupRouter("http://localhost:8080", storeFake)

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

	if got, want := storeFake.lastCreateArg.OriginalUrl, "https://example.com/long-url"; got != want {
		t.Fatalf("original_url: got %q, want %q", got, want)
	}

	if got, want := storeFake.lastCreateArg.ShortName, "exmpl"; got != want {
		t.Fatalf("short_name: got %q, want %q", got, want)
	}
}

func TestCreateLinkRequiresOriginalURL(t *testing.T) {
	storeFake := &fakeLinksStore{}
	r := setupRouter("http://localhost:8080", storeFake)

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

	if got, want := storeFake.lastCreateArg, (store.CreateLinkParams{}); got != want {
		t.Fatalf("create args: got %+v, want %+v", got, want)
	}
}

func TestCreateLinkGeneratesShortName(t *testing.T) {
	storeFake := &fakeLinksStore{
		created: store.Link{
			ID:          2,
			OriginalUrl: "https://example.com/long-url",
			ShortName:   "generated",
		},
	}
	r := setupRouter("http://localhost:8080", storeFake)

	body := `{"original_url":"https://example.com/long-url"}`
	rq := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(body))
	rq.Header.Set("Content-Type", "application/json")
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusCreated; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	want := `{"id":2,"original_url":"https://example.com/long-url","short_name":"generated","short_url":"http://localhost:8080/r/generated"}`
	if got := rc.Body.String(); got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}

	if got, want := storeFake.lastCreateArg.OriginalUrl, "https://example.com/long-url"; got != want {
		t.Fatalf("original_url: got %q, want %q", got, want)
	}

	if got := storeFake.lastCreateArg.ShortName; got == "" {
		t.Fatal("short_name: got empty value, want generated value")
	}
}

func TestCreateLinkConflictingShortName(t *testing.T) {
	storeFake := &fakeLinksStore{
		createErr: &pq.Error{Code: "23505"},
	}
	r := setupRouter("http://localhost:8080", storeFake)

	body := `{"original_url":"https://example.com/long-url","short_name":"exmpl"}`
	rq := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(body))
	rq.Header.Set("Content-Type", "application/json")
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusConflict; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	if got, want := rc.Body.String(), `{"error":"short_name already exists"}`; got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}

	if got, want := storeFake.lastCreateArg.ShortName, "exmpl"; got != want {
		t.Fatalf("short_name: got %q, want %q", got, want)
	}
}
