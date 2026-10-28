package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/services"
	"strings"
	"testing"
	"time"
)

type readingHTTPRepository struct {
	interfaces.ReadingRepository
	minimumAge int
}

func (r readingHTTPRepository) Post(context.Context, uint64, uint64) (core.ReadingPost, error) {
	return core.ReadingPost{ID: 1, ReaderID: 2, WorkID: 7, Kind: "diary", Title: "Título com spoiler", Text: "Final protegido", ContainsSpoiler: true, Status: "published", AuthorStatus: "active", WorkStatus: "published", WorkMinimumAge: r.minimumAge}, nil
}
func readingHTTPRouter(ageRating int) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	born := time.Now().UTC().AddDate(-14, 0, 0)
	r.Use(func(c *gin.Context) { c.Set("actor", core.Actor{ID: 1, Reader: true, BirthDate: &born}); c.Next() })
	RegisterReading(r.Group("/api/v1"), services.NewReading(readingHTTPRepository{minimumAge: ageRating}, services.ReadingOptions{}), func(c *gin.Context) { c.Next() })
	return r
}
func TestReadingHTTPRejectsUnknownAuthorAndMissingSpoiler(t *testing.T) {
	r := readingHTTPRouter(0)
	for _, body := range []string{`{"reader_id":99,"edition_id":3,"text":"Diário","current_page":0,"contains_spoiler":false}`, `{"edition_id":3,"text":"Diário","current_page":0}`} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/reading-entries", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 400 {
			t.Fatalf("entrada inválida aceitou autoria/ausência de flag: %d %s", w.Code, w.Body.String())
		}
	}
}
func TestReadingHTTPMasksProtectedTextAndRechecksWorkAge(t *testing.T) {
	r := readingHTTPRouter(0)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/reading-entries/1", nil))
	if w.Code != 200 || strings.Contains(w.Body.String(), "Final protegido") || strings.Contains(w.Body.String(), "Título com spoiler") {
		t.Fatalf("spoiler vazou: %d %s", w.Code, w.Body.String())
	}
	r = readingHTTPRouter(18)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/reading-entries/1?reveal_spoilers=true", nil))
	if w.Code != 404 {
		t.Fatalf("confirmação contornou faixa etária: %d", w.Code)
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/posts/0", nil))
	if w.Code != 400 {
		t.Fatalf("ID zero não validado: %d", w.Code)
	}
}
