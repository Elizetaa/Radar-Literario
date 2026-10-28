package tools_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/schemas"
	"radar-literario/backend/internal/model/services"
	"radar-literario/backend/internal/model/tools"
	"strings"
	"testing"
	"time"
)

func adminIntegration(t *testing.T) (*gorm.DB, interfaces.Store, *services.Admin, core.Actor, core.Actor, core.Actor) {
	t.Helper()
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN não configurado; requer MariaDB isolado com migrações aplicadas")
	}
	cfg, e := mysqlDriver.ParseDSN(dsn)
	if e != nil || !strings.Contains(cfg.DBName, "test") {
		t.Fatal("TEST_DB_DSN deve selecionar banco de testes")
	}
	db, e := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if e != nil {
		t.Fatal(e)
	}
	pool, e := db.DB()
	if e != nil {
		t.Fatal(e)
	}
	pool.SetMaxOpenConns(2)
	t.Cleanup(func() { _ = pool.Close() })
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	t.Cleanup(func() { tx.Rollback() })
	store := tools.NewStore(tx)
	ctx := context.Background()
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	birth := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	create := func(n int) core.Actor {
		nick := fmt.Sprintf("a%d%s", n, suffix)
		id, e := store.Insert(ctx, "users", map[string]any{"name": "Conta fictícia administrativa", "nickname": nick, "email": nick + "@example.test", "password_hash": "fixture-not-for-login", "status": "active", "birth_date": birth})
		if e != nil {
			t.Fatal(e)
		}
		if _, e = store.Exec(ctx, "INSERT INTO readers(user_id,biography,profile_visibility) VALUES(?,'','public')", id); e != nil {
			t.Fatal(e)
		}
		return core.Actor{ID: id, Reader: true, BirthDate: &birth}
	}
	admin, reader, other := create(1), create(2), create(3)
	admin.Role = "system_administrator"
	if _, e = store.Exec(ctx, "INSERT INTO administrators(user_id,role,status) VALUES(?,'system_administrator','active')", admin.ID); e != nil {
		t.Fatal(e)
	}
	return tx, store, services.NewAdmin(tools.NewAdminRepository(store)), admin, reader, other
}
func adminIntegrationStatus(t *testing.T, e error, status int) {
	t.Helper()
	var de *core.Error
	if !errors.As(e, &de) || de.Status != status {
		t.Fatalf("wanted %d, got %v", status, e)
	}
}
func TestAdminMariaDBReportsAndSupport(t *testing.T) {
	_, store, s, admin, reader, other := adminIntegration(t)
	ctx := context.Background()
	id := other.ID
	report, e := s.CreateReport(ctx, reader, schemas.CreateReport{ReportedUserID: &id, Reason: "spam", Description: "Teste de persistência de denúncia."})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Report(ctx, other, report.ID, false); e == nil {
		t.Fatal("other reader accessed report")
	} else {
		adminIntegrationStatus(t, e, 404)
	}
	_, e = s.AddReportMessage(ctx, admin, report.ID, true, schemas.CreateReportMessage{Text: "Nota restrita aos moderadores.", Visibility: "internal"})
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.AddReportMessage(ctx, admin, report.ID, true, schemas.CreateReportMessage{Text: "Informe mais detalhes.", RequestInformation: true})
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.AddReportMessage(ctx, reader, report.ID, false, schemas.CreateReportMessage{Text: "A mensagem foi repetida várias vezes."})
	if e != nil {
		t.Fatal(e)
	}
	list, e := s.ReportMessages(ctx, reader, report.ID, false, schemas.Page{Page: 1, PageSize: 20})
	if e != nil {
		t.Fatal(e)
	}
	messages := list.Data.([]schemas.ReportMessage)
	if len(messages) != 2 {
		t.Fatalf("wanted 2 public messages, got %d", len(messages))
	}
	for _, v := range messages {
		if v.Visibility != "reporter" {
			t.Fatal("internal message leaked")
		}
	}
	status := "resolved"
	reason := "Denúncia analisada sem exposição de evidência."
	if _, e = s.UpdateReport(ctx, admin, report.ID, schemas.PatchReport{Status: &status, ResolutionNote: &reason}); e != nil {
		t.Fatal(e)
	}
	ticket, e := s.CreateTicket(ctx, reader, schemas.CreateSupportTicket{Subject: "Dúvida persistente", Text: "Como registrar minha leitura?"})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.TicketMessages(ctx, other, ticket.ID, schemas.Page{Page: 1, PageSize: 20}); e == nil {
		t.Fatal("other reader accessed support messages")
	} else {
		adminIntegrationStatus(t, e, 404)
	}
	if _, e = s.AddTicketMessage(ctx, admin, ticket.ID, true, schemas.CreateSupportMessage{Text: "Use a estante para iniciar sua leitura."}); e != nil {
		t.Fatal(e)
	}
	list, e = s.TicketMessages(ctx, reader, ticket.ID, schemas.Page{Page: 1, PageSize: 20})
	if e != nil {
		t.Fatal(e)
	}
	if list.Pagination.Total != 2 {
		t.Fatalf("wanted persisted conversation, got %d", list.Pagination.Total)
	}
	if e = s.CloseTicket(ctx, reader, ticket.ID); e != nil {
		t.Fatal(e)
	}
	if _, e = s.AddTicketMessage(ctx, reader, ticket.ID, false, schemas.CreateSupportMessage{Text: "Nova mensagem"}); e == nil {
		t.Fatal("closed support ticket accepted message")
	}
	var count int64
	if e = store.Query(ctx, &count, "SELECT COUNT(*) FROM audit_logs WHERE actor_id=?", admin.ID); e != nil || count < 3 {
		t.Fatalf("audit records missing: %d, %v", count, e)
	}
}
func TestAdminMariaDBConstraintsAndPenaltyEffects(t *testing.T) {
	_, store, s, admin, reader, other := adminIntegration(t)
	ctx := context.Background()
	now := time.Now().UTC()
	for _, q := range []string{
		"INSERT INTO reports(reporter_id,reason,description,status,created_at,updated_at) VALUES(?,'spam','sem alvo','open',UTC_TIMESTAMP(6),UTC_TIMESTAMP(6))",
		"INSERT INTO reports(reporter_id,reported_user_id,reported_post_id,reason,description,status,created_at,updated_at) VALUES(?,?,1,'spam','dois alvos','open',UTC_TIMESTAMP(6),UTC_TIMESTAMP(6))",
	} {
		args := []any{reader.ID}
		if strings.Contains(q, "dois alvos") {
			args = append(args, other.ID)
		}
		if _, e := store.Exec(ctx, q, args...); e == nil {
			t.Fatal("MariaDB accepted invalid report target")
		}
	}
	if _, e := store.Exec(ctx, "INSERT INTO penalties(user_id,applied_by,type,reason,starts_at,ends_at,created_at) VALUES(?,?,'suspension','fim ausente',?,NULL,?)", reader.ID, admin.ID, now, now); e == nil {
		t.Fatal("MariaDB accepted suspension without end")
	}
	if _, e := store.Exec(ctx, "INSERT INTO reports(reporter_id,reported_user_id,reason,description,status,created_at,updated_at) VALUES(?,18446744073709551614,'spam','FK inválida','open',?,?)", reader.ID, now, now); e == nil {
		t.Fatal("MariaDB accepted invalid user FK")
	}
	token := fmt.Sprintf("%064x", reader.ID)
	if _, e := store.Exec(ctx, "INSERT INTO sessions(user_id,token_hash,csrf_hash,created_at,last_seen_at,expires_at) VALUES(?,?,?,?,?,?)", reader.ID, token, token, now, now, now.Add(time.Hour)); e != nil {
		t.Fatal(e)
	}
	end := now.Add(time.Hour)
	penalty, e := s.ApplyPenalty(ctx, admin, schemas.CreatePenalty{UserID: reader.ID, Type: "suspension", Reason: "Suspensão de teste", EndsAt: &end})
	if e != nil {
		t.Fatal(e)
	}
	var n int64
	if e = store.Query(ctx, &n, "SELECT COUNT(*) FROM sessions WHERE user_id=? AND revoked_at IS NOT NULL", reader.ID); e != nil || n != 1 {
		t.Fatalf("sessions not revoked: %d %v", n, e)
	}
	second, e := s.ApplyPenalty(ctx, admin, schemas.CreatePenalty{UserID: reader.ID, Type: "ban", Reason: "Punição sobreposta de teste"})
	if e != nil {
		t.Fatal(e)
	}
	if e = s.RevokePenalty(ctx, admin, penalty.ID, "Revisão de suspensão"); e != nil {
		t.Fatal(e)
	}
	if e = store.Query(ctx, &n, "SELECT COUNT(*) FROM penalties WHERE user_id=? AND revoked_at IS NULL AND starts_at<=UTC_TIMESTAMP(6) AND (ends_at IS NULL OR ends_at>UTC_TIMESTAMP(6)) AND type IN ('suspension','ban')", reader.ID); e != nil || n != 1 {
		t.Fatalf("overlapping penalty was lost: %d %v", n, e)
	}
	if e = s.RevokePenalty(ctx, admin, second.ID, "Revisão do banimento"); e != nil {
		t.Fatal(e)
	}
	_, e = s.ApplyPenalty(ctx, admin, schemas.CreatePenalty{UserID: admin.ID, Type: "ban", Reason: "Operação proibida"})
	adminIntegrationStatus(t, e, 403)
}
func TestAdminMariaDBFAQAndAdministrator(t *testing.T) {
	_, store, s, admin, reader, _ := adminIntegration(t)
	ctx := context.Background()
	question, answer, status := "Como usar a estante?", "Selecione uma edição da obra.", "draft"
	faq, e := s.SaveFAQ(ctx, admin, 0, schemas.SaveFAQ{Question: &question, Answer: &answer, Status: &status})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.FAQ(ctx, core.Actor{}, faq.ID, false); e == nil {
		t.Fatal("unpublished FAQ exposed")
	} else {
		adminIntegrationStatus(t, e, 404)
	}
	status = "published"
	if _, e = s.SaveFAQ(ctx, admin, faq.ID, schemas.SaveFAQ{Status: &status}); e != nil {
		t.Fatal(e)
	}
	if _, e = s.FAQ(ctx, core.Actor{}, faq.ID, false); e != nil {
		t.Fatal(e)
	}
	created, e := s.CreateAdministrator(ctx, admin, schemas.CreateAdministrator{UserID: &reader.ID, Role: "catalog_manager", Reason: "Atribuição de teste"})
	if e != nil {
		t.Fatal(e)
	}
	if created.Role != "catalog_manager" {
		t.Fatal("administrative role not persisted")
	}
	if e = s.RevokeAdministrator(ctx, admin, reader.ID, "Revogação de teste"); e != nil {
		t.Fatal(e)
	}
	record, e := s.Administrator(ctx, admin, reader.ID)
	if e != nil || record.Status != "inactive" {
		t.Fatalf("administrative revocation failed: %v", e)
	}
	body, e := json.Marshal(record)
	if e != nil {
		t.Fatal(e)
	}
	if strings.Contains(string(body), "password") || strings.Contains(string(body), "email") {
		t.Fatal("administrator JSON leaks credentials")
	}
	var n int64
	if e = store.Query(ctx, &n, "SELECT COUNT(*) FROM audit_logs WHERE actor_id=? AND target_type='administrator'", admin.ID); e != nil || n != 2 {
		t.Fatalf("administrative audit missing: %d %v", n, e)
	}
}
