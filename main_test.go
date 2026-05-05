package main

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pq "github.com/lib/pq"
	store "github.com/xhrobj-hex/go-project-278/internal/db"
	"github.com/xhrobj-hex/go-project-278/internal/router"
)

type fakeLinksStore struct {
	links   []store.Link
	listErr error

	created       store.Link
	createErr     error
	lastCreateArg store.CreateLinkParams

	gotByID    store.Link
	getByIDErr error
	lastGetID  int64

	updated       store.Link
	updateErr     error
	lastUpdateArg store.UpdateLinkParams

	deletedRows  int64
	deleteErr    error
	lastDeleteID int64
}

func (f fakeLinksStore) ListLinks(ctx context.Context) ([]store.Link, error) {
	return f.links, f.listErr
}

func (f *fakeLinksStore) CreateLink(ctx context.Context, arg store.CreateLinkParams) (store.Link, error) {
	f.lastCreateArg = arg
	return f.created, f.createErr
}

func (f *fakeLinksStore) GetLinkByID(ctx context.Context, id int64) (store.Link, error) {
	f.lastGetID = id
	return f.gotByID, f.getByIDErr
}

func (f *fakeLinksStore) UpdateLink(ctx context.Context, arg store.UpdateLinkParams) (store.Link, error) {
	f.lastUpdateArg = arg
	return f.updated, f.updateErr
}

func (f *fakeLinksStore) DeleteLink(ctx context.Context, id int64) (int64, error) {
	f.lastDeleteID = id
	return f.deletedRows, f.deleteErr
}

func TestPing(t *testing.T) {
	r := router.New("http://localhost:8080", nil)

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
	r := router.New("http://localhost:8080", &fakeLinksStore{})

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
	r := router.New("http://localhost:8080", storeFake)

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
	r := router.New("http://localhost:8080", storeFake)

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
	r := router.New("http://localhost:8080", storeFake)

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
	r := router.New("http://localhost:8080", storeFake)

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

func TestGetLinkByID(t *testing.T) {
	storeFake := &fakeLinksStore{
		gotByID: store.Link{
			ID:          1,
			OriginalUrl: "https://example.com/long-url",
			ShortName:   "exmpl",
		},
	}
	r := router.New("http://localhost:8080", storeFake)

	rq := httptest.NewRequest(http.MethodGet, "/api/links/1", nil)
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusOK; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	want := `{"id":1,"original_url":"https://example.com/long-url","short_name":"exmpl","short_url":"http://localhost:8080/r/exmpl"}`
	if got := rc.Body.String(); got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}

	if got, want := storeFake.lastGetID, int64(1); got != want {
		t.Fatalf("id: got %d, want %d", got, want)
	}
}

func TestGetLinkByIDNotFound(t *testing.T) {
	storeFake := &fakeLinksStore{
		getByIDErr: sql.ErrNoRows,
	}
	r := router.New("http://localhost:8080", storeFake)

	rq := httptest.NewRequest(http.MethodGet, "/api/links/999", nil)
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusNotFound; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	if got, want := rc.Body.String(), `{"error":"link not found"}`; got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}

	if got, want := storeFake.lastGetID, int64(999); got != want {
		t.Fatalf("id: got %d, want %d", got, want)
	}
}

func TestGetLinkByIDInvalidID(t *testing.T) {
	storeFake := &fakeLinksStore{}
	r := router.New("http://localhost:8080", storeFake)

	rq := httptest.NewRequest(http.MethodGet, "/api/links/abc", nil)
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusBadRequest; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	if got, want := rc.Body.String(), `{"error":"invalid id"}`; got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}

	if got, want := storeFake.lastGetID, int64(0); got != want {
		t.Fatalf("id: got %d, want %d", got, want)
	}
}

func TestUpdateLink(t *testing.T) {
	storeFake := &fakeLinksStore{
		updated: store.Link{
			ID:          1,
			OriginalUrl: "https://example.com/updated-url",
			ShortName:   "exmpl-upd",
		},
	}
	r := router.New("http://localhost:8080", storeFake)

	body := `{"original_url":"https://example.com/updated-url","short_name":"exmpl-upd"}`
	rq := httptest.NewRequest(http.MethodPut, "/api/links/1", strings.NewReader(body))
	rq.Header.Set("Content-Type", "application/json")
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusOK; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	want := `{"id":1,"original_url":"https://example.com/updated-url","short_name":"exmpl-upd","short_url":"http://localhost:8080/r/exmpl-upd"}`
	if got := rc.Body.String(); got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}

	if got, want := storeFake.lastUpdateArg.ID, int64(1); got != want {
		t.Fatalf("id: got %d, want %d", got, want)
	}

	if got, want := storeFake.lastUpdateArg.OriginalUrl, "https://example.com/updated-url"; got != want {
		t.Fatalf("original_url: got %q, want %q", got, want)
	}

	if got, want := storeFake.lastUpdateArg.ShortName, "exmpl-upd"; got != want {
		t.Fatalf("short_name: got %q, want %q", got, want)
	}
}

func TestUpdateLinkNotFound(t *testing.T) {
	storeFake := &fakeLinksStore{
		updateErr: sql.ErrNoRows,
	}
	r := router.New("http://localhost:8080", storeFake)

	body := `{"original_url":"https://example.com/updated-url","short_name":"exmpl-upd"}`
	rq := httptest.NewRequest(http.MethodPut, "/api/links/999", strings.NewReader(body))
	rq.Header.Set("Content-Type", "application/json")
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusNotFound; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	if got, want := rc.Body.String(), `{"error":"link not found"}`; got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}

	if got, want := storeFake.lastUpdateArg.ID, int64(999); got != want {
		t.Fatalf("id: got %d, want %d", got, want)
	}
}

func TestUpdateLinkInvalidID(t *testing.T) {
	storeFake := &fakeLinksStore{}
	r := router.New("http://localhost:8080", storeFake)

	body := `{"original_url":"https://example.com/updated-url","short_name":"exmpl-upd"}`
	rq := httptest.NewRequest(http.MethodPut, "/api/links/abc", strings.NewReader(body))
	rq.Header.Set("Content-Type", "application/json")
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusBadRequest; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	if got, want := rc.Body.String(), `{"error":"invalid id"}`; got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}

	if got, want := storeFake.lastUpdateArg, (store.UpdateLinkParams{}); got != want {
		t.Fatalf("update args: got %+v, want %+v", got, want)
	}
}

func TestUpdateLinkRequiresOriginalURL(t *testing.T) {
	storeFake := &fakeLinksStore{}
	r := router.New("http://localhost:8080", storeFake)

	body := `{"short_name":"exmpl-upd"}`
	rq := httptest.NewRequest(http.MethodPut, "/api/links/1", strings.NewReader(body))
	rq.Header.Set("Content-Type", "application/json")
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusBadRequest; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	if got, want := rc.Body.String(), `{"error":"original_url is required"}`; got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}

	if got, want := storeFake.lastUpdateArg, (store.UpdateLinkParams{}); got != want {
		t.Fatalf("update args: got %+v, want %+v", got, want)
	}
}

func TestDeleteLink(t *testing.T) {
	storeFake := &fakeLinksStore{
		deletedRows: 1,
	}
	r := router.New("http://localhost:8080", storeFake)

	rq := httptest.NewRequest(http.MethodDelete, "/api/links/1", nil)
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusNoContent; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	if got, want := rc.Body.String(), ""; got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}

	if got, want := storeFake.lastDeleteID, int64(1); got != want {
		t.Fatalf("id: got %d, want %d", got, want)
	}
}

func TestDeleteLinkNotFound(t *testing.T) {
	storeFake := &fakeLinksStore{
		deletedRows: 0,
	}
	r := router.New("http://localhost:8080", storeFake)

	rq := httptest.NewRequest(http.MethodDelete, "/api/links/999", nil)
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusNotFound; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	if got, want := rc.Body.String(), `{"error":"link not found"}`; got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}

	if got, want := storeFake.lastDeleteID, int64(999); got != want {
		t.Fatalf("id: got %d, want %d", got, want)
	}
}

func TestDeleteLinkInvalidID(t *testing.T) {
	storeFake := &fakeLinksStore{}
	r := router.New("http://localhost:8080", storeFake)

	rq := httptest.NewRequest(http.MethodDelete, "/api/links/abc", nil)
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusBadRequest; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	if got, want := rc.Body.String(), `{"error":"invalid id"}`; got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}

	if got, want := storeFake.lastDeleteID, int64(0); got != want {
		t.Fatalf("id: got %d, want %d", got, want)
	}
}
