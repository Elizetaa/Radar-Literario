//go:build integration

package router_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"radar-literario/backend/internal/config"
	"radar-literario/backend/internal/database"
	"radar-literario/backend/internal/model/services"
	"radar-literario/backend/internal/model/tools"
	"radar-literario/backend/internal/router"
	"strings"
	"testing"
	"time"
)

type client struct {
	t       *testing.T
	h       http.Handler
	cookies map[string]*http.Cookie
	csrf    string
}

func (c *client) request(method, path string, body any, want int) map[string]any {
	c.t.Helper()
	var input io.Reader
	if body != nil {
		b, e := json.Marshal(body)
		if e != nil {
			c.t.Fatal(e)
		}
		input = bytes.NewReader(b)
	}
	r := httptest.NewRequest(method, path, input)
	r.Header.Set("Origin", "http://localhost:5173")
	r.Header.Set("Content-Type", "application/json")
	if c.csrf != "" {
		r.Header.Set("X-CSRF-Token", c.csrf)
	}
	for _, v := range c.cookies {
		r.AddCookie(v)
	}
	w := httptest.NewRecorder()
	c.h.ServeHTTP(w, r)
	if w.Code != want {
		c.t.Fatalf("%s %s: want %d got %d: %s", method, path, want, w.Code, w.Body.String())
	}
	for _, v := range w.Result().Cookies() {
		if v.MaxAge < 0 {
			delete(c.cookies, v.Name)
		} else {
			c.cookies[v.Name] = v
		}
	}
	var out map[string]any
	if w.Body.Len() > 0 {
		if e := json.Unmarshal(w.Body.Bytes(), &out); e != nil {
			c.t.Fatal(e)
		}
	}
	return out
}
func (c *client) preauth() {
	c.t.Helper()
	out := c.request("GET", "/api/v1/auth/csrf", nil, 200)
	c.csrf = out["data"].(map[string]any)["csrf_token"].(string)
}
func (c *client) login(email string) {
	c.t.Helper()
	c.preauth()
	out := c.request("POST", "/api/v1/auth/login", map[string]any{"email": email, "password": "testing-password"}, 200)
	c.csrf = out["data"].(map[string]any)["csrf_token"].(string)
}
func data(v map[string]any) map[string]any { return v["data"].(map[string]any) }
func id(v map[string]any) uint64           { return uint64(data(v)["id"].(float64)) }
func integration(t *testing.T) (*gorm.DB, config.Config) {
	t.Helper()
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("configure TEST_DB_DSN e aplique migrações no banco isolado")
	}
	d, e := mysqlDriver.ParseDSN(dsn)
	if e != nil || !strings.HasSuffix(d.DBName, "_test") {
		t.Fatal("TEST_DB_DSN deve usar banco com sufixo _test")
	}
	db, e := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if e != nil {
		t.Fatal(e)
	}
	pool, e := db.DB()
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { pool.Close() })
	if e = database.SchemaReady(context.Background(), pool); e != nil {
		t.Fatal(e)
	}
	cfg := config.Config{Env: "test", CookieName: "test_session", SameSite: "lax", EncryptionKey: make([]byte, 32), SessionIdle: 30 * time.Minute, SessionAbsolute: 7 * 24 * time.Hour, ResetTTL: 30 * time.Minute, RequestTimeout: 15 * time.Second, BodyLimit: 1 << 20, RateLimit: 1000, CORSOrigins: []string{"http://localhost:5173"}, CSRFOrigins: []string{"http://localhost:5173"}, FrontendURL: "http://localhost:5173", RetentionDays: 180}
	return db, cfg
}
func TestMariaDBHTTPFullReaderAndAdministrativeFlow(t *testing.T) {
	db, cfg := integration(t)
	tx := db.Begin()
	defer tx.Rollback()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	store := tools.NewStore(tx)
	accounts := services.NewAccounts(tools.NewAccountRepository(store), cfg, nil)
	catalog := services.NewCatalog(tools.NewCatalogRepository(store), 1)
	reading := services.NewReading(tools.NewReadingRepository(store), services.ReadingOptions{MaxCommentDepth: 5, Recommendations: catalog.Recommendations, PreferredWorks: catalog.Preferred})
	adminService := services.NewAdmin(tools.NewAdminRepository(store))
	h := router.New(cfg, router.Dependencies{Accounts: accounts, Catalog: catalog, Reading: reading, Admin: adminService, Ready: store.Ready})
	newClient := func() *client { return &client{t: t, h: h, cookies: map[string]*http.Cookie{}} }
	guest := newClient()
	first := newClient()
	second := newClient()
	admin := newClient()
	prefix := fmt.Sprintf("h%x", time.Now().UnixNano())
	firstEmail := prefix + "1@example.test"
	secondEmail := prefix + "2@example.test"
	adminEmail := prefix + "3@example.test"
	guest.request("GET", "/health", nil, 200)
	guest.request("GET", "/ready", nil, 200)
	guest.request("GET", "/api/usuarios/1", nil, 404)
	guest.request("POST", "/api/v1/readers", map[string]any{}, 403)
	guest.preauth()
	guest.request("POST", "/api/v1/readers", map[string]any{"name": "Invalid", "role": "system_administrator"}, 400)
	register := func(c *client, n, email, birth string) uint64 {
		c.preauth()
		v := c.request("POST", "/api/v1/readers", map[string]any{"name": "Leitor de Teste", "nickname": n, "email": email, "password": "testing-password", "birth_date": birth}, 201)
		if strings.Contains(fmt.Sprint(v), "password_hash") {
			t.Fatal("credential leaked")
		}
		return id(v)
	}
	firstID := register(first, prefix+"1", firstEmail, "1990-01-01")
	secondID := register(second, prefix+"2", secondEmail, "1992-06-20")
	adminID := register(admin, prefix+"3", adminEmail, "1990-01-01")
	if e := tx.Exec("INSERT INTO administrators(user_id,role,status) VALUES (?,'system_administrator','active')", adminID).Error; e != nil {
		t.Fatal(e)
	}
	guest.request("POST", "/api/v1/readers", map[string]any{"name": "Leitor", "nickname": prefix + "4", "email": strings.ToUpper(firstEmail), "password": "testing-password", "birth_date": "1990-01-01"}, 409)
	first.login(firstEmail)
	second.login(secondEmail)
	admin.login(adminEmail)
	if cookie := first.cookies[cfg.CookieName]; cookie == nil || !cookie.HttpOnly || cookie.Path != "/" || cookie.SameSite != http.SameSiteLaxMode || cookie.MaxAge <= 0 {
		t.Fatal("session cookie missing protections")
	}
	switcher := newClient()
	for name, cookie := range first.cookies {
		switcher.cookies[name] = cookie
	}
	switcher.login(secondEmail)
	first.request("GET", "/api/v1/me", nil, 401)
	first.login(firstEmail)
	second.request("PATCH", "/api/v1/me", map[string]any{"role": "administrator"}, 400)
	first.request("POST", "/api/v1/admin/genres", map[string]any{"name": "Forbidden"}, 403)
	genreID := id(admin.request("POST", "/api/v1/admin/genres", map[string]any{"name": prefix + " Ficção", "active": true}, 201))
	authorID := id(admin.request("POST", "/api/v1/admin/authors", map[string]any{"name": "Autor " + prefix, "biography": "Pessoa fictícia"}, 201))
	publisherID := id(admin.request("POST", "/api/v1/admin/publishers", map[string]any{"name": "Editora " + prefix}, 201))
	workID := id(admin.request("POST", "/api/v1/admin/works", map[string]any{"title": "Obra " + prefix, "synopsis": "Sinopse fictícia.", "original_language": "pt-BR", "minimum_age": 0, "status": "published", "author_ids": []uint64{authorID}, "genre_ids": []uint64{genreID}}, 201))
	editionID := id(admin.request("POST", "/api/v1/admin/editions", map[string]any{"work_id": workID, "publisher_id": publisherID, "title": "Edição de Teste", "isbn": "9780306406157", "edition_number": 1, "page_count": 200, "language": "pt-BR", "minimum_age": 0, "status": "published"}, 201))
	path := fmt.Sprintf("/api/v1/works/%d", workID)
	guest.request("GET", path, nil, 200)
	guest.request("GET", "/api/v1/catalog/isbn/0306406152", nil, 200)
	shelfPath := fmt.Sprintf("/api/v1/me/shelf/%d", workID)
	shelf := map[string]any{"edition_id": editionID, "status": "planned", "current_page": 0}
	first.request("PUT", shelfPath, shelf, 200)
	first.request("PUT", shelfPath, shelf, 200)
	var count int64
	tx.Raw("SELECT COUNT(*) FROM reading_records WHERE reader_id=? AND work_id=?", firstID, workID).Scan(&count)
	if count != 1 {
		t.Fatal("shelf duplicate")
	}
	first.request("PUT", shelfPath, map[string]any{"edition_id": editionID, "status": "in_progress", "current_page": 201}, 400)
	diaryID := id(first.request("POST", "/api/v1/reading-entries", map[string]any{"work_id": workID, "edition_id": editionID, "title": "Diário", "text": "Registro na página vinte.", "current_page": 20, "status": "draft", "contains_spoiler": false, "contains_sensitive_content": false, "minimum_age": 0}, 201))
	var page int
	tx.Raw("SELECT current_page FROM reading_records WHERE reader_id=? AND work_id=?", firstID, workID).Scan(&page)
	if page != 0 {
		t.Fatal("draft modified progress")
	}
	diaryPath := fmt.Sprintf("/api/v1/reading-entries/%d", diaryID)
	first.request("PATCH", diaryPath, map[string]any{"status": "published"}, 200)
	first.request("PATCH", diaryPath, map[string]any{"text": "Texto editado sem novo progresso."}, 200)
	tx.Raw("SELECT current_page FROM reading_records WHERE reader_id=? AND work_id=?", firstID, workID).Scan(&page)
	if page != 20 {
		t.Fatal("published diary missing shelf effect")
	}
	second.request("PATCH", diaryPath, map[string]any{"text": "Ataque"}, 403)
	review := func(c *client, rating int) uint64 {
		return id(c.request("POST", "/api/v1/reviews", map[string]any{"work_id": workID, "title": "Avaliação", "text": "Comentário de avaliação.", "rating": rating, "status": "published", "contains_spoiler": false, "contains_sensitive_content": false, "minimum_age": 0}, 201))
	}
	firstReview := review(first, 5)
	second.request("PUT", shelfPath, map[string]any{"edition_id": editionID, "status": "in_progress", "current_page": 40}, 200)
	review(second, 3)
	work := data(guest.request("GET", path, nil, 200))
	if work["average_rating"] != float64(4) {
		t.Fatalf("wrong average: %#v", work)
	}
	commentID := id(second.request("POST", fmt.Sprintf("/api/v1/posts/%d/comments", diaryID), map[string]any{"text": "Comentário da segunda conta."}, 201))
	first.request("PATCH", fmt.Sprintf("/api/v1/comments/%d", commentID), map[string]any{"text": "Alteração indevida"}, 403)
	second.request("PUT", fmt.Sprintf("/api/v1/posts/%d/like", diaryID), nil, 204)
	second.request("PUT", fmt.Sprintf("/api/v1/posts/%d/like", diaryID), nil, 204)
	tx.Raw("SELECT COUNT(*) FROM post_likes WHERE reader_id=? AND post_id=?", secondID, diaryID).Scan(&count)
	if count != 1 {
		t.Fatal("duplicated like")
	}
	notifications := first.request("GET", "/api/v1/me/notifications", nil, 200)
	if len(notifications["data"].([]any)) < 2 {
		t.Fatal("comment/like notifications missing")
	}
	first.request("PATCH", diaryPath, map[string]any{"contains_spoiler": true}, 200)
	hidden := data(guest.request("GET", diaryPath, nil, 200))
	if hidden["text"] != nil || hidden["title"] != nil {
		t.Fatal("spoiler leaked")
	}
	revealed := data(guest.request("GET", diaryPath+"?reveal_spoilers=true", nil, 200))
	if revealed["text"] == nil {
		t.Fatal("explicit reveal failed")
	}
	admin.request("PATCH", fmt.Sprintf("/api/v1/admin/works/%d", workID), map[string]any{"minimum_age": 18}, 200)
	guest.request("GET", diaryPath+"?reveal_spoilers=true", nil, 404)
	guest.request("GET", fmt.Sprintf("/api/v1/posts/%d/comments?reveal_spoilers=true", diaryID), nil, 404)
	admin.request("PATCH", fmt.Sprintf("/api/v1/admin/works/%d", workID), map[string]any{"minimum_age": 0}, 200)
	reportID := id(second.request("POST", "/api/v1/reports", map[string]any{"reported_post_id": firstReview, "reason": "other", "description": "Solicitação fictícia de revisão."}, 201))
	first.request("GET", fmt.Sprintf("/api/v1/me/reports/%d", reportID), nil, 404)
	admin.request("POST", fmt.Sprintf("/api/v1/admin/posts/%d/moderation", firstReview), map[string]any{"status": "hidden", "reason": "Análise fictícia"}, 204)
	work = data(guest.request("GET", path, nil, 200))
	if work["average_rating"] != float64(3) {
		t.Fatal("moderation did not update average")
	}
	first.request("DELETE", fmt.Sprintf("/api/v1/reviews/%d", firstReview), nil, 409)
	first.request("PUT", "/api/v1/me/interests", map[string]any{"genre_ids": []uint64{genreID}}, 200)
	first.request("PUT", fmt.Sprintf("/api/v1/me/preferred-works/%d", workID), nil, 204)
	first.request("PATCH", "/api/v1/me", map[string]any{"favorite_work_id": workID}, 200)
	first.request("GET", "/api/v1/me/statistics?year=2026", nil, 200)
	first.request("GET", "/api/v1/me/reading-report?group=preferred", nil, 200)
	first.request("GET", "/api/v1/me/recommendations", nil, 200)
	ticketID := id(first.request("POST", "/api/v1/support-tickets", map[string]any{"subject": "Ajuda de teste", "text": "Mensagem fictícia de suporte."}, 201))
	second.request("GET", fmt.Sprintf("/api/v1/support-tickets/%d", ticketID), nil, 404)
	admin.request("POST", fmt.Sprintf("/api/v1/admin/support-tickets/%d/messages", ticketID), map[string]any{"text": "Resposta administrativa."}, 201)
	var hash string
	tx.Raw("SELECT password_hash FROM users WHERE id=?", firstID).Scan(&hash)
	if hash == "testing-password" || !strings.HasPrefix(hash, "$2") {
		t.Fatal("password not protected")
	}
	public := data(guest.request("GET", fmt.Sprintf("/api/v1/readers/%d", firstID), nil, 200))
	if _, ok := public["email"]; ok {
		t.Fatal("public email leaked")
	}
	if _, ok := public["birth_date"]; ok {
		t.Fatal("public birth date leaked")
	}
	second.request("POST", "/api/v1/auth/logout", nil, 204)
	second.request("GET", "/api/v1/me", nil, 401)
	first.request("DELETE", "/api/v1/me", map[string]any{"password": "testing-password"}, 204)
	guest.request("GET", fmt.Sprintf("/api/v1/readers/%d", firstID), nil, 404)
	guest.request("GET", diaryPath, nil, 404)
	tx.Raw("SELECT COUNT(*) FROM reading_records WHERE reader_id=?", firstID).Scan(&count)
	if count != 0 {
		t.Fatal("deleted account kept shelf")
	}
}
