package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/schemas"
	"radar-literario/backend/internal/model/services"
	"strings"
	"testing"
	"time"
)

type administrativeHTTPFake struct{ interfaces.AdminRepository }

func (f administrativeHTTPFake) Ticket(context.Context, uint64, bool) (core.SupportTicket, error) {
	return core.SupportTicket{ID: 1, RequesterID: 2, Subject: "Atendimento privado", Status: "open"}, nil
}
func (f administrativeHTTPFake) TicketMessages(context.Context, uint64, schemas.Page) ([]core.SupportMessage, int64, error) {
	return []core.SupportMessage{{ID: 1, AuthorID: 2, Text: "Mensagem privada"}}, 1, nil
}
func (f administrativeHTTPFake) Audit(context.Context, uint64, string, string, uint64, string, time.Time) error {
	return nil
}
func adminHTTPRouter(actor core.Actor) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("actor", actor); c.Next() })
	RegisterAdmin(r.Group("/api/v1"), services.NewAdmin(administrativeHTTPFake{}), func(c *gin.Context) {
		if Actor(c).ID == 0 {
			Fail(c, core.ErrUnauthenticated)
			return
		}
		c.Next()
	})
	return r
}
func TestAdministrativeHTTPValidationAndRoles(t *testing.T) {
	r := adminHTTPRouter(core.Actor{ID: 1, Reader: true})
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{
		{"POST", "/api/v1/reports", `{"reason":"spam","description":"teste"}`, 400},
		{"POST", "/api/v1/reports", `{"reported_user_id":2,"reason":"spam","description":"teste","reader_id":9}`, 400},
		{"GET", "/api/v1/admin/administrators", "", 403},
		{"POST", "/api/v1/admin/penalties", `{"user_id":2,"type":"ban","reason":"violação"}`, 403},
		{"GET", "/api/v1/admin/posts/1", "", 403},
		{"GET", "/api/v1/me/reports?page_size=101", "", 400},
		{"GET", "/api/v1/faq/invalid", "", 400},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("wanted %d got %d: %s", tc.status, w.Code, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), `"error"`) {
				t.Fatal("missing structured error")
			}
		})
	}
}
func TestAdministrativeHTTPSupportIsolation(t *testing.T) {
	for _, tc := range []struct {
		actor  core.Actor
		path   string
		status int
	}{
		{core.Actor{ID: 1, Reader: true}, "/api/v1/support-tickets/1/messages", 404},
		{core.Actor{ID: 2, Reader: true}, "/api/v1/support-tickets/1/messages", 200},
		{core.Actor{ID: 1, Role: "catalog_manager"}, "/api/v1/support-tickets/1/messages", 404},
		{core.Actor{ID: 1, Role: "user_manager"}, "/api/v1/support-tickets/1/messages", 200},
		{core.Actor{ID: 2, Reader: true}, "/api/v1/admin/support-tickets/1/messages", 403},
		{core.Actor{}, "/api/v1/support-tickets/1", 401},
	} {
		r := adminHTTPRouter(tc.actor)
		req := httptest.NewRequest("GET", tc.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != tc.status {
			t.Fatalf("actor %+v: wanted %d got %d: %s", tc.actor, tc.status, w.Code, w.Body.String())
		}
		if tc.status != 200 && strings.Contains(w.Body.String(), "Mensagem privada") {
			t.Fatal("private conversation leaked")
		}
	}
}
