package services

import (
	"context"
	"encoding/json"
	"errors"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/schemas"
	"strings"
	"testing"
	"time"
)

type adminFake struct {
	interfaces.AdminRepository
	report                      core.Report
	administrator               core.AdministratorRecord
	user                        core.AdminUser
	target                      core.ReportTarget
	ticket                      core.SupportTicket
	penalty                     core.Penalty
	systemCount                 int64
	writes, audits, revocations int
	failMessage, failAudit      bool
	lastVisibility              string
}

func (f *adminFake) Within(ctx context.Context, fn func(interfaces.AdminRepository) error) error {
	before := *f
	e := fn(f)
	if e != nil {
		*f = before
	}
	return e
}
func (f *adminFake) User(context.Context, uint64, bool) (core.AdminUser, error) {
	if f.user.ID == 0 {
		return core.AdminUser{}, core.NotFound()
	}
	return f.user, nil
}
func (f *adminFake) Administrator(context.Context, uint64, bool) (core.AdministratorRecord, error) {
	if f.administrator.UserID == 0 {
		return core.AdministratorRecord{}, core.NotFound()
	}
	return f.administrator, nil
}
func (f *adminFake) LockSystemAdministrators(context.Context) (int64, error) {
	return f.systemCount, nil
}
func (f *adminFake) SaveAdministrator(_ context.Context, _ uint64, role, status string, _ bool) error {
	f.administrator.Role = role
	f.administrator.Status = status
	f.writes++
	return nil
}
func (f *adminFake) RevokeSessions(context.Context, uint64, time.Time) error {
	f.revocations++
	return nil
}
func (f *adminFake) HasBlockingPenalty(context.Context, uint64, time.Time) (bool, error) {
	return false, nil
}
func (f *adminFake) ContentEvidence(context.Context, string, uint64) (core.ContentEvidence, error) {
	return core.ContentEvidence{ID: 2, Text: "Evidência protegida", Status: "hidden", MinimumAge: 18}, nil
}
func (f *adminFake) Target(context.Context, string, uint64) (core.ReportTarget, error) {
	if f.target.ID == 0 {
		return core.ReportTarget{}, core.NotFound()
	}
	return f.target, nil
}
func (f *adminFake) CreateReport(_ context.Context, v core.Report) (uint64, error) {
	f.writes++
	v.ID = 10
	f.report = v
	return v.ID, nil
}
func (f *adminFake) Report(context.Context, uint64, bool) (core.Report, error) {
	if f.report.ID == 0 {
		return core.Report{}, core.NotFound()
	}
	return f.report, nil
}
func (f *adminFake) SaveReport(_ context.Context, v core.Report) error {
	f.report = v
	f.writes++
	return nil
}
func (f *adminFake) ReportMessages(_ context.Context, _ uint64, admin bool, _ schemas.Page) ([]core.ReportMessage, int64, error) {
	rows := []core.ReportMessage{{ID: 1, Text: "visível", Visibility: "reporter"}}
	if admin {
		rows = append(rows, core.ReportMessage{ID: 2, Text: "interno", Visibility: "internal"})
	}
	return rows, int64(len(rows)), nil
}
func (f *adminFake) CreateReportMessage(_ context.Context, v core.ReportMessage) (uint64, error) {
	f.writes++
	f.lastVisibility = v.Visibility
	return 5, nil
}
func (f *adminFake) Moderate(context.Context, string, uint64, string, time.Time) error {
	f.writes++
	return nil
}
func (f *adminFake) NotifyModeration(context.Context, uint64, uint64, string, uint64, string, time.Time) error {
	f.writes++
	return nil
}
func (f *adminFake) CreatePenalty(_ context.Context, v core.Penalty) (uint64, error) {
	f.writes++
	v.ID = 3
	f.penalty = v
	return 3, nil
}
func (f *adminFake) LinkPenaltyReports(context.Context, uint64, []uint64) error { return nil }
func (f *adminFake) Penalty(context.Context, uint64, bool) (core.Penalty, error) {
	return f.penalty, nil
}
func (f *adminFake) RevokePenalty(_ context.Context, _ uint64, actor uint64, reason string, now time.Time) error {
	f.writes++
	f.penalty.RevokedAt = &now
	f.penalty.RevokedBy = &actor
	f.penalty.RevocationReason = &reason
	return nil
}
func (f *adminFake) Ticket(context.Context, uint64, bool) (core.SupportTicket, error) {
	if f.ticket.ID == 0 {
		return core.SupportTicket{}, core.NotFound()
	}
	return f.ticket, nil
}
func (f *adminFake) CreateTicket(_ context.Context, v core.SupportTicket) (uint64, error) {
	v.ID = 9
	f.ticket = v
	f.writes++
	return 9, nil
}
func (f *adminFake) SaveTicket(_ context.Context, v core.SupportTicket) error {
	f.ticket = v
	f.writes++
	return nil
}
func (f *adminFake) CreateTicketMessage(context.Context, core.SupportMessage) (uint64, error) {
	if f.failMessage {
		return 0, errors.New("message storage failure")
	}
	f.writes++
	return 7, nil
}
func (f *adminFake) Audit(context.Context, uint64, string, string, uint64, string, time.Time) error {
	if f.failAudit {
		return errors.New("audit storage failure")
	}
	f.audits++
	return nil
}
func adminTestService(f *adminFake) *Admin {
	s := NewAdmin(f)
	s.now = func() time.Time { return time.Date(2026, 10, 27, 15, 0, 0, 0, time.UTC) }
	return s
}
func adminTestActor(role string) core.Actor {
	birth := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	return core.Actor{ID: 1, Reader: true, Role: role, BirthDate: &birth}
}
func adminTestStatus(t *testing.T, e error, status int) {
	t.Helper()
	var de *core.Error
	if !errors.As(e, &de) || de.Status != status {
		t.Fatalf("wanted status %d, got %v", status, e)
	}
}

func TestReportRequiresExactlyOneTarget(t *testing.T) {
	id := uint64(2)
	for _, in := range []schemas.CreateReport{{}, {ReportedUserID: &id, ReportedPostID: &id}, {ReportedUserID: &id, ReportedPostID: &id, ReportedCommentID: &id}} {
		f := &adminFake{}
		s := adminTestService(f)
		_, e := s.CreateReport(context.Background(), adminTestActor(""), in)
		adminTestStatus(t, e, 400)
		if f.writes != 0 {
			t.Fatal("invalid report wrote data")
		}
	}
}
func TestReportsEnforceAgeAndDoNotExposeReporterIdentity(t *testing.T) {
	f := &adminFake{target: core.ReportTarget{ID: 2, OwnerID: 5, Status: "published", UserStatus: "active", WorkStatus: "published", MinimumAge: 18}}
	s := adminTestService(f)
	id := uint64(2)
	in := schemas.CreateReport{ReportedPostID: &id, Reason: "spam", Description: "mensagem repetida"}
	minor := adminTestActor("")
	birth := time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC)
	minor.BirthDate = &birth
	_, e := s.CreateReport(context.Background(), minor, in)
	adminTestStatus(t, e, 404)
	out, e := s.CreateReport(context.Background(), adminTestActor(""), in)
	if e != nil {
		t.Fatal(e)
	}
	if out.ReporterID != nil || out.AssignedTo != nil {
		t.Fatal("reporter response exposes administrative fields")
	}
	f.report.ReportDataURL = "https://example.test/private"
	out, e = s.Report(context.Background(), adminTestActor(""), 10, false)
	if e != nil {
		t.Fatal(e)
	}
	body, _ := json.Marshal(out)
	if strings.Contains(string(body), "private") || strings.Contains(string(body), "reporter_id") {
		t.Fatalf("unexpected private fields: %s", body)
	}
}
func TestReportOwnershipAndInformationWorkflow(t *testing.T) {
	f := &adminFake{report: core.Report{ID: 8, ReporterID: 2, Status: "open"}}
	s := adminTestService(f)
	_, e := s.Report(context.Background(), adminTestActor(""), 8, false)
	adminTestStatus(t, e, 404)
	owner := adminTestActor("")
	owner.ID = 2
	_, e = s.AddReportMessage(context.Background(), owner, 8, false, schemas.CreateReportMessage{Text: "complemento"})
	adminTestStatus(t, e, 409)
	_, e = s.AddReportMessage(context.Background(), adminTestActor("user_manager"), 8, true, schemas.CreateReportMessage{Text: "Informe quando ocorreu.", RequestInformation: true})
	if e != nil {
		t.Fatal(e)
	}
	if f.report.Status != "awaiting_information" {
		t.Fatal("did not request information")
	}
	_, e = s.AddReportMessage(context.Background(), owner, 8, false, schemas.CreateReportMessage{Text: "Hoje."})
	if e != nil {
		t.Fatal(e)
	}
	if f.report.Status != "in_review" || f.lastVisibility != "reporter" {
		t.Fatal("reply did not return report to review")
	}
	list, e := s.ReportMessages(context.Background(), owner, 8, false, schemas.Page{Page: 1, PageSize: 20})
	if e != nil {
		t.Fatal(e)
	}
	rows := list.Data.([]schemas.ReportMessage)
	if len(rows) != 1 || rows[0].Visibility != "reporter" {
		t.Fatal("internal message leaked")
	}
}
func TestReportConclusionRequiresResolutionAndCannotReopen(t *testing.T) {
	f := &adminFake{report: core.Report{ID: 8, ReporterID: 2, Status: "in_review"}}
	s := adminTestService(f)
	status := "resolved"
	_, e := s.UpdateReport(context.Background(), adminTestActor("user_manager"), 8, schemas.PatchReport{Status: &status})
	adminTestStatus(t, e, 400)
	reason := "Conteúdo removido após análise."
	out, e := s.UpdateReport(context.Background(), adminTestActor("user_manager"), 8, schemas.PatchReport{Status: &status, ResolutionNote: &reason})
	if e != nil {
		t.Fatal(e)
	}
	if out.ResolvedAt == nil {
		t.Fatal("resolution time missing")
	}
	status = "in_review"
	_, e = s.UpdateReport(context.Background(), adminTestActor("user_manager"), 8, schemas.PatchReport{Status: &status})
	adminTestStatus(t, e, 409)
}
func TestLastSystemAdministratorAndAuditAreAtomic(t *testing.T) {
	f := &adminFake{systemCount: 1, administrator: core.AdministratorRecord{UserID: 1, Role: "system_administrator", Status: "active"}}
	s := adminTestService(f)
	status := "inactive"
	_, e := s.UpdateAdministrator(context.Background(), adminTestActor("system_administrator"), 1, schemas.PatchAdministrator{Status: &status, Reason: "Revogação solicitada"})
	adminTestStatus(t, e, 409)
	if f.writes != 0 || f.revocations != 0 {
		t.Fatal("last administrator changed")
	}
	f.systemCount = 2
	f.failAudit = true
	_, e = s.UpdateAdministrator(context.Background(), adminTestActor("system_administrator"), 1, schemas.PatchAdministrator{Status: &status, Reason: "Revogação solicitada"})
	if e == nil {
		t.Fatal("expected audit failure")
	}
	if f.writes != 0 || f.administrator.Status != "active" {
		t.Fatal("failed audit left administrative change committed")
	}
	f.failAudit = false
	_, e = s.UpdateAdministrator(context.Background(), adminTestActor("system_administrator"), 1, schemas.PatchAdministrator{Status: &status, Reason: "Revogação solicitada"})
	if e != nil {
		t.Fatal(e)
	}
	if f.revocations != 1 {
		t.Fatal("role/status change did not revoke sessions")
	}
}
func TestAdministrativeRoleMatrix(t *testing.T) {
	s := adminTestService(&adminFake{})
	for _, role := range []string{"", "catalog_manager", "user_manager", "administrator"} {
		_, e := s.CreateAdministrator(context.Background(), adminTestActor(role), schemas.CreateAdministrator{Role: "user_manager", Reason: "Cadastro"})
		adminTestStatus(t, e, 403)
	}
	for _, role := range []string{"", "catalog_manager"} {
		_, e := s.ApplyPenalty(context.Background(), adminTestActor(role), schemas.CreatePenalty{})
		adminTestStatus(t, e, 403)
	}
	_, e := s.AuditLogs(context.Background(), adminTestActor("catalog_manager"), "penalty", schemas.Page{Page: 1, PageSize: 20})
	adminTestStatus(t, e, 403)
}
func TestPenaltyRejectsAdministratorAndInvalidPeriods(t *testing.T) {
	f := &adminFake{user: core.AdminUser{ID: 2, Status: "active"}, administrator: core.AdministratorRecord{UserID: 2, Role: "catalog_manager", Status: "active"}}
	s := adminTestService(f)
	_, e := s.ApplyPenalty(context.Background(), adminTestActor("system_administrator"), schemas.CreatePenalty{UserID: 2, Type: "ban", Reason: "Violação confirmada"})
	adminTestStatus(t, e, 403)
	if f.writes != 0 {
		t.Fatal("administrator penalized")
	}
	f.administrator = core.AdministratorRecord{}
	end := s.now().Add(-time.Hour)
	for _, in := range []schemas.CreatePenalty{{UserID: 2, Type: "suspension", Reason: "Violação"}, {UserID: 2, Type: "restriction", Reason: "Violação", EndsAt: &end}, {UserID: 2, Type: "ban", Reason: "Violação", EndsAt: &end}} {
		_, e = s.ApplyPenalty(context.Background(), adminTestActor("user_manager"), in)
		adminTestStatus(t, e, 400)
	}
	end = s.now().Add(time.Hour)
	_, e = s.ApplyPenalty(context.Background(), adminTestActor("user_manager"), schemas.CreatePenalty{UserID: 2, Type: "suspension", Reason: "Violação confirmada", EndsAt: &end})
	if e != nil {
		t.Fatal(e)
	}
	if f.revocations != 1 {
		t.Fatal("active suspension must revoke sessions")
	}
}
func TestPenaltyRevocationIsIdempotent(t *testing.T) {
	f := &adminFake{penalty: core.Penalty{ID: 3, UserID: 2, Type: "ban"}}
	s := adminTestService(f)
	a := adminTestActor("user_manager")
	if e := s.RevokePenalty(context.Background(), a, 3, "Decisão revisada"); e != nil {
		t.Fatal(e)
	}
	writes := f.writes
	if e := s.RevokePenalty(context.Background(), a, 3, "Decisão revisada"); e != nil {
		t.Fatal(e)
	}
	if f.writes != writes || f.audits != 1 {
		t.Fatal("repeated revocation duplicated work")
	}
}
func TestSupportOwnershipAndTransactionalCreation(t *testing.T) {
	f := &adminFake{ticket: core.SupportTicket{ID: 9, RequesterID: 2, Status: "open"}}
	s := adminTestService(f)
	_, e := s.Ticket(context.Background(), adminTestActor(""), 9, false)
	adminTestStatus(t, e, 404)
	_, e = s.AddTicketMessage(context.Background(), adminTestActor("catalog_manager"), 9, false, schemas.CreateSupportMessage{Text: "não autorizado"})
	adminTestStatus(t, e, 404)
	f.ticket = core.SupportTicket{}
	f.failMessage = true
	_, e = s.CreateTicket(context.Background(), adminTestActor(""), schemas.CreateSupportTicket{Subject: "Dúvida de leitura", Text: "Como alterar minha edição?"})
	if e == nil {
		t.Fatal("expected message storage failure")
	}
	if f.ticket.ID != 0 || f.writes != 0 {
		t.Fatal("partial support ticket persisted")
	}
	f.failMessage = false
	out, e := s.CreateTicket(context.Background(), adminTestActor(""), schemas.CreateSupportTicket{Subject: "Dúvida de leitura", Text: "Como alterar minha edição?"})
	if e != nil || out.Status != "open" {
		t.Fatalf("create failed: %v", e)
	}
}
func TestClosedSupportIsTerminalAndResolvedCanReopen(t *testing.T) {
	f := &adminFake{ticket: core.SupportTicket{ID: 9, RequesterID: 1, Status: "closed"}}
	s := adminTestService(f)
	_, e := s.AddTicketMessage(context.Background(), adminTestActor(""), 9, false, schemas.CreateSupportMessage{Text: "nova pergunta"})
	adminTestStatus(t, e, 409)
	status := "in_progress"
	_, e = s.UpdateTicket(context.Background(), adminTestActor("user_manager"), 9, schemas.PatchSupportTicket{Status: &status})
	adminTestStatus(t, e, 409)
	f.ticket.Status = "resolved"
	_, e = s.AddTicketMessage(context.Background(), adminTestActor(""), 9, false, schemas.CreateSupportMessage{Text: "Ainda preciso de ajuda."})
	if e != nil {
		t.Fatal(e)
	}
	if f.ticket.Status != "in_progress" {
		t.Fatal("resolved ticket not reopened")
	}
}
func TestDeletedContentCannotBeRestored(t *testing.T) {
	f := &adminFake{target: core.ReportTarget{ID: 2, OwnerID: 2, Status: "deleted"}}
	s := adminTestService(f)
	e := s.Moderate(context.Background(), adminTestActor("user_manager"), "post", 2, schemas.Moderation{Status: "published", Reason: "Revisão da decisão"})
	adminTestStatus(t, e, 409)
	if f.writes != 0 {
		t.Fatal("deleted post restored")
	}
}
func TestAssignmentRequiresActiveModerator(t *testing.T) {
	f := &adminFake{report: core.Report{ID: 2, ReporterID: 3, Status: "open"}, administrator: core.AdministratorRecord{UserID: 5, Role: "catalog_manager", Status: "active"}, user: core.AdminUser{ID: 5, Status: "active"}}
	s := adminTestService(f)
	id := uint64(5)
	_, e := s.UpdateReport(context.Background(), adminTestActor("user_manager"), 2, schemas.PatchReport{AssignedTo: schemas.AdminOptionalID{Set: true, Value: &id}})
	adminTestStatus(t, e, 403)
	if f.writes != 0 {
		t.Fatal("invalid assignment stored")
	}
}

func TestEvidenceRequiresModerationAndAudit(t *testing.T) {
	f := &adminFake{}
	s := adminTestService(f)
	_, e := s.InspectContent(context.Background(), adminTestActor("catalog_manager"), "post", 2)
	adminTestStatus(t, e, 403)
	f.failAudit = true
	out, e := s.InspectContent(context.Background(), adminTestActor("user_manager"), "post", 2)
	if e == nil || out.Text != "" {
		t.Fatal("evidence returned without audit")
	}
	f.failAudit = false
	out, e = s.InspectContent(context.Background(), adminTestActor("user_manager"), "post", 2)
	if e != nil || out.Text != "Evidência protegida" || f.audits != 1 {
		t.Fatalf("audited evidence read failed: %v", e)
	}
}
