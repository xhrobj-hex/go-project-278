package handler

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	pq "github.com/lib/pq"
	store "github.com/xhrobj-hex/go-project-278/internal/db"
)

type fakeLinksStore struct {
	links       []store.Link
	linkVisits  []store.LinkVisit
	listErr     error
	lastListArg store.ListLinksParams

	total    int64
	countErr error

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

func (f *fakeLinksStore) ListLinks(ctx context.Context, arg store.ListLinksParams) ([]store.Link, error) {
	f.lastListArg = arg
	return f.links, f.listErr
}

func (f *fakeLinksStore) CountLinks(ctx context.Context) (int64, error) {
	return f.total, f.countErr
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

func (f *fakeLinksStore) GetLinkByShortName(ctx context.Context, shortName string) (store.Link, error) {
	for _, link := range f.links {
		if link.ShortName == shortName {
			return link, nil
		}
	}

	return store.Link{}, sql.ErrNoRows
}

func (f *fakeLinksStore) CreateLinkVisit(ctx context.Context, arg store.CreateLinkVisitParams) (store.LinkVisit, error) {
	visit := store.LinkVisit{
		ID:        int64(len(f.linkVisits) + 1),
		LinkID:    arg.LinkID,
		Ip:        arg.Ip,
		UserAgent: arg.UserAgent,
		Referer:   arg.Referer,
		Status:    arg.Status,
		CreatedAt: time.Now(),
	}

	f.linkVisits = append(f.linkVisits, visit)

	return visit, nil
}

func (f *fakeLinksStore) ListLinkVisits(ctx context.Context, arg store.ListLinkVisitsParams) ([]store.LinkVisit, error) {
	return f.linkVisits, nil
}

func (f *fakeLinksStore) CountLinkVisits(ctx context.Context) (int64, error) {
	return int64(len(f.linkVisits)), nil
}

func setupTestRouter(baseURL string, links LinksStore) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()

	linkHandler := NewLinkHandler(baseURL, links)

	r.GET("/api/links", linkHandler.List)
	r.POST("/api/links", linkHandler.Create)
	r.GET("/api/links/:id", linkHandler.GetByID)
	r.PUT("/api/links/:id", linkHandler.Update)
	r.DELETE("/api/links/:id", linkHandler.Delete)

	return r
}

func TestListLinksEmpty(t *testing.T) {
	r := setupTestRouter("http://localhost:8080", &fakeLinksStore{})

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
	r := setupTestRouter("http://localhost:8080", storeFake)

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
	r := setupTestRouter("http://localhost:8080", storeFake)

	body := `{"short_name":"exmpl"}`
	rq := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(body))
	rq.Header.Set("Content-Type", "application/json")
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusUnprocessableEntity; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	if got, want := rc.Body.String(), `{"errors":{"original_url":"is required"}}`; got != want {
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
	r := setupTestRouter("http://localhost:8080", storeFake)

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
	r := setupTestRouter("http://localhost:8080", storeFake)

	body := `{"original_url":"https://example.com/long-url","short_name":"exmpl"}`
	rq := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(body))
	rq.Header.Set("Content-Type", "application/json")
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusUnprocessableEntity; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	if got, want := rc.Body.String(), `{"errors":{"short_name":"short name already in use"}}`; got != want {
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
	r := setupTestRouter("http://localhost:8080", storeFake)

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
	r := setupTestRouter("http://localhost:8080", storeFake)

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
	r := setupTestRouter("http://localhost:8080", storeFake)

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
	r := setupTestRouter("http://localhost:8080", storeFake)

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
	r := setupTestRouter("http://localhost:8080", storeFake)

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
	r := setupTestRouter("http://localhost:8080", storeFake)

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
	r := setupTestRouter("http://localhost:8080", storeFake)

	body := `{"short_name":"exmpl-upd"}`
	rq := httptest.NewRequest(http.MethodPut, "/api/links/1", strings.NewReader(body))
	rq.Header.Set("Content-Type", "application/json")
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusUnprocessableEntity; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	if got, want := rc.Body.String(), `{"errors":{"original_url":"is required"}}`; got != want {
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
	r := setupTestRouter("http://localhost:8080", storeFake)

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
	r := setupTestRouter("http://localhost:8080", storeFake)

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
	r := setupTestRouter("http://localhost:8080", storeFake)

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

func TestListLinksWithRange(t *testing.T) {
	storeFake := &fakeLinksStore{
		total: 42,
		links: []store.Link{
			{ID: 1, OriginalUrl: "https://example.com/one", ShortName: "one"},
			{ID: 2, OriginalUrl: "https://example.com/two", ShortName: "two"},
		},
	}

	r := setupTestRouter("http://localhost:8080", storeFake)

	rq := httptest.NewRequest(http.MethodGet, "/api/links?range=[0,1]", nil)
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusOK; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	if got, want := rc.Header().Get("Content-Range"), "links 0-1/42"; got != want {
		t.Fatalf("Content-Range: got %q, want %q", got, want)
	}

	if got, want := storeFake.lastListArg.Limit, int32(2); got != want {
		t.Fatalf("limit: got %d, want %d", got, want)
	}

	if got, want := storeFake.lastListArg.Offset, int32(0); got != want {
		t.Fatalf("offset: got %d, want %d", got, want)
	}
}

func TestListLinksWithRangeAndSpaces(t *testing.T) {
	storeFake := &fakeLinksStore{
		total: 11,
		links: []store.Link{
			{ID: 6, OriginalUrl: "https://example.com/six", ShortName: "six"},
			{ID: 7, OriginalUrl: "https://example.com/seven", ShortName: "7even"},
			{ID: 8, OriginalUrl: "https://example.com/eight", ShortName: "eight"},
			{ID: 9, OriginalUrl: "https://example.com/nine", ShortName: "nine"},
			{ID: 10, OriginalUrl: "https://example.com/ten", ShortName: "X"},
		},
	}

	r := setupTestRouter("http://localhost:8080", storeFake)

	rq := httptest.NewRequest(http.MethodGet, "/api/links?range=[5,%209]", nil)
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusOK; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	if got, want := rc.Header().Get("Content-Range"), "links 5-9/11"; got != want {
		t.Fatalf("Content-Range: got %q, want %q", got, want)
	}

	if got, want := storeFake.lastListArg.Limit, int32(5); got != want {
		t.Fatalf("limit: got %d, want %d", got, want)
	}

	if got, want := storeFake.lastListArg.Offset, int32(5); got != want {
		t.Fatalf("offset: got %d, want %d", got, want)
	}
}

func TestCreateLinkRejectsInvalidOriginalURL(t *testing.T) {
	storeFake := &fakeLinksStore{}
	r := setupTestRouter("http://localhost:8080", storeFake)

	body := `{"original_url":"not-url","short_name":"exmpl"}`
	rq := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(body))
	rq.Header.Set("Content-Type", "application/json")
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusUnprocessableEntity; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	if got, want := rc.Body.String(), `{"errors":{"original_url":"must be a valid URL"}}`; got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}

	if got, want := storeFake.lastCreateArg, (store.CreateLinkParams{}); got != want {
		t.Fatalf("create args: got %+v, want %+v", got, want)
	}
}

func TestCreateLinkRejectsTooShortShortName(t *testing.T) {
	storeFake := &fakeLinksStore{}
	r := setupTestRouter("http://localhost:8080", storeFake)

	body := `{"original_url":"https://example.com","short_name":"ab"}`
	rq := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(body))
	rq.Header.Set("Content-Type", "application/json")
	rc := httptest.NewRecorder()

	r.ServeHTTP(rc, rq)

	if got, want := rc.Code, http.StatusUnprocessableEntity; got != want {
		t.Fatalf("status: got %d, want %d", got, want)
	}

	if got, want := rc.Body.String(), `{"errors":{"short_name":"must be at least 3 characters"}}`; got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}

	if got, want := storeFake.lastCreateArg, (store.CreateLinkParams{}); got != want {
		t.Fatalf("create args: got %+v, want %+v", got, want)
	}
}
