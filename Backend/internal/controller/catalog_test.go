package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/services"
)

type catalogHTTPRepository struct {
	interfaces.CatalogRepository
	calls int
}

func (r *catalogHTTPRepository) Work(ctx context.Context, id uint64, age int, lock bool) (core.Work, error) {
	r.calls++
	return core.Work{ID: id, Title: "Oculta", Status: "draft"}, nil
}
func TestCatalogHTTPRejectsUnknownFieldsInvalidIDsAndInvisibleWorks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &catalogHTTPRepository{}
	engine := gin.New()
	group := engine.Group("/api/v1")
	auth := func(c *gin.Context) { c.Set("actor", core.Actor{ID: 2, Reader: true}) }
	RegisterCatalog(group, services.NewCatalog(repo, 1), auth)
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{{"GET", "/api/v1/works/0", "", 400}, {"GET", "/api/v1/works/4", "", 404}, {"GET", "/api/v1/works?page_size=101", "", 400}, {"POST", "/api/v1/admin/works", `{"title":"Teste","role":"system_administrator"}`, 400}, {"POST", "/api/v1/admin/works", `{"title":"Teste"}`, 403}} {
		request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, request)
		if response.Code != tc.status {
			t.Errorf("%s %s status %d body=%s", tc.method, tc.path, response.Code, response.Body.String())
		}
		if strings.Contains(response.Body.String(), "Oculta") {
			t.Error("draft title leaked")
		}
	}
	if repo.calls != 1 {
		t.Fatalf("invalid requests reached repository %d times", repo.calls)
	}
}
func TestCatalogHTTPAnonymousPersonalRouteIsRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterCatalog(engine.Group("/api/v1"), services.NewCatalog(&catalogHTTPRepository{}, 1), func(c *gin.Context) {})
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/me/recommendations", nil))
	if response.Code != 401 {
		t.Fatalf("got %d: %s", response.Code, response.Body.String())
	}
}
