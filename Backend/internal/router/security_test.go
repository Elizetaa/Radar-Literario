package router_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"radar-literario/backend/internal/config"
	"radar-literario/backend/internal/model/services"
	"radar-literario/backend/internal/router"
	"strings"
	"testing"
	"time"
)

func TestHTTPReadinessCORSPreauthCSRFAndRateLimit(t *testing.T) {
	cfg := config.Config{Env: "test", CookieName: "radar_session", SameSite: "lax", EncryptionKey: make([]byte, 32), SessionIdle: 30 * time.Minute, SessionAbsolute: time.Hour, RequestTimeout: time.Second, BodyLimit: 1024, RateLimit: 2, CORSOrigins: []string{"http://localhost:5173"}, CSRFOrigins: []string{"http://localhost:5173"}}
	s := services.NewAccounts(nil, cfg, nil)
	h := router.New(cfg, router.Dependencies{Accounts: s, Ready: func(context.Context) error { return errors.New("database disconnected") }})
	req := func(method, path, origin, csrf string, cookies []*http.Cookie) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(`{}`))
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		if csrf != "" {
			r.Header.Set("X-CSRF-Token", csrf)
		}
		for _, cookie := range cookies {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	if w := req("GET", "/health", "", "", nil); w.Code != 200 {
		t.Fatalf("liveness %d", w.Code)
	}
	if w := req("GET", "/ready", "", "", nil); w.Code != 503 || strings.Contains(w.Body.String(), "disconnected") {
		t.Fatalf("readiness leaked dependency: %s", w.Body.String())
	}
	if w := req("GET", "/api/v1/auth/csrf", "https://untrusted.test", "", nil); w.Code != 403 {
		t.Fatal("CORS allowed forbidden origin")
	}
	w := req("OPTIONS", "/api/v1/readers", "http://localhost:5173", "", nil)
	if w.Code != 204 || w.Header().Get("Access-Control-Allow-Credentials") != "true" || w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatal("CORS preflight invalid")
	}
	w = req("GET", "/api/v1/auth/csrf", "http://localhost:5173", "", nil)
	var response struct {
		Data struct {
			Token string `json:"csrf_token"`
		}
	}
	if e := json.Unmarshal(w.Body.Bytes(), &response); e != nil {
		t.Fatal(e)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || cookies[0].Path != "/" || cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatal("preauth cookie missing protections")
	}
	if len(response.Data.Token) != 43 {
		t.Fatal("preauth CSRF invalid")
	}
	for _, token := range []string{"", strings.Repeat("a", 43)} {
		if w = req("POST", "/api/v1/readers", "http://localhost:5173", token, cookies); w.Code != 403 {
			t.Fatal("invalid CSRF accepted")
		}
	}
	if w = req("POST", "/api/v1/readers", "https://untrusted.test", response.Data.Token, cookies); w.Code != 403 {
		t.Fatal("CSRF origin accepted")
	}
	for i := 0; i < 2; i++ {
		if w = req("POST", "/api/v1/readers", "http://localhost:5173", response.Data.Token, cookies); w.Code != 400 {
			t.Fatalf("validation expected400 got%d", w.Code)
		}
	}
	if w = req("POST", "/api/v1/readers", "http://localhost:5173", response.Data.Token, cookies); w.Code != 429 || w.Header().Get("Retry-After") == "" {
		t.Fatal("auth rate limit absent")
	}
	if w = req("GET", "/api/v1/me", "http://localhost:5173", "", nil); w.Code != 401 {
		t.Fatal("unauthenticated private route allowed")
	}
}
