package services

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/schemas"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

type Admin struct {
	repo interfaces.AdminRepository
	now  func() time.Time
}

func NewAdmin(repo interfaces.AdminRepository) *Admin {
	return &Admin{repo: repo, now: func() time.Time { return time.Now().UTC() }}
}
func adminHas(values []string, v string) bool {
	for _, x := range values {
		if x == v {
			return true
		}
	}
	return false
}

var adminRoles = []string{"user_manager", "catalog_manager", "administrator", "system_administrator"}
var reportStates = []string{"open", "in_review", "awaiting_information", "resolved", "dismissed", "canceled"}
var reportReasons = []string{"spam", "harassment", "hate_speech", "threats", "inappropriate_content", "privacy_violation", "copyright_violation", "impersonation", "misinformation", "other"}
var ticketStates = []string{"open", "in_progress", "resolved", "closed"}

func adminText(s string, min, max int) bool {
	n := utf8.RuneCountInString(strings.TrimSpace(s))
	return n >= min && n <= max
}
func adminReader(a core.Actor) error {
	if a.ID == 0 {
		return core.ErrUnauthenticated
	}
	if !a.Reader {
		return core.Forbidden()
	}
	return nil
}
func adminIsNotFound(err error) bool { var e *core.Error; return errors.As(err, &e) && e.Status == 404 }
func adminDTO(v core.AdministratorRecord) schemas.Administrator {
	return schemas.Administrator{ID: v.UserID, Name: v.Name, Nickname: v.Nickname, Role: v.Role, Status: v.Status, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
func (s *Admin) Administrators(ctx context.Context, a core.Actor, p schemas.Page) (schemas.List, error) {
	if e := a.Require("system"); e != nil {
		return schemas.List{}, e
	}
	rows, n, e := s.repo.Administrators(ctx, p)
	if e != nil {
		return schemas.List{}, e
	}
	out := make([]schemas.Administrator, 0, len(rows))
	for _, v := range rows {
		out = append(out, adminDTO(v))
	}
	return schemas.Paginated(out, p, n), nil
}
func (s *Admin) Administrator(ctx context.Context, a core.Actor, id uint64) (schemas.Administrator, error) {
	if e := a.Require("system"); e != nil {
		return schemas.Administrator{}, e
	}
	v, e := s.repo.Administrator(ctx, id, false)
	return adminDTO(v), e
}
func (s *Admin) CreateAdministrator(ctx context.Context, a core.Actor, in schemas.CreateAdministrator) (schemas.Administrator, error) {
	if e := a.Require("system"); e != nil {
		return schemas.Administrator{}, e
	}
	if !adminHas(adminRoles, in.Role) || !adminText(in.Reason, 3, 2000) {
		return schemas.Administrator{}, core.Invalid("INVALID_ADMINISTRATOR", "Informe papel válido e motivo da criação.")
	}
	var user core.AdminUser
	if in.UserID == nil {
		in.Email = core.NormalizeEmail(in.Email)
		in.Nickname = core.NormalizeNickname(in.Nickname)
		address, e := mail.ParseAddress(in.Email)
		if e != nil || address.Address != in.Email || len(in.Email) > 254 || !adminText(in.Name, 2, 120) || !regexp.MustCompile(`^[a-z0-9_]{3,30}$`).MatchString(in.Nickname) {
			return schemas.Administrator{}, core.Invalid("INVALID_ACCOUNT", "Informe nome, apelido e e-mail válidos.")
		}
		if e := core.ValidatePassword(in.Password); e != nil {
			return schemas.Administrator{}, e
		}
		date, e := core.Date(in.BirthDate)
		if e != nil {
			return schemas.Administrator{}, e
		}
		if date == nil || date.After(s.now()) {
			return schemas.Administrator{}, core.Invalid("INVALID_BIRTH_DATE", "Informe nascimento válido.")
		}
		hash, e := core.HashPassword(in.Password)
		if e != nil {
			return schemas.Administrator{}, e
		}
		user = core.AdminUser{Name: strings.TrimSpace(in.Name), Nickname: in.Nickname, Email: in.Email, PasswordHash: hash, Status: "active", BirthDate: date, CreatedAt: s.now(), UpdatedAt: s.now()}
	} else if *in.UserID == 0 || in.Name != "" || in.Nickname != "" || in.Email != "" || in.Password != "" || in.BirthDate != "" {
		return schemas.Administrator{}, core.Invalid("INVALID_ACCOUNT", "Use user_id existente ou os dados de uma conta nova.")
	}
	var out schemas.Administrator
	err := s.repo.Within(ctx, func(tx interfaces.AdminRepository) error {
		id := uint64(0)
		if in.UserID == nil {
			var e error
			id, e = tx.CreateAdministrativeUser(ctx, user)
			if e != nil {
				return e
			}
		} else {
			id = *in.UserID
			u, e := tx.User(ctx, id, true)
			if e != nil {
				return e
			}
			if u.Status != "active" {
				return core.Conflict("Somente uma conta ativa pode receber papel administrativo.")
			}
			blocked, e := tx.HasBlockingPenalty(ctx, id, s.now())
			if e != nil {
				return e
			}
			if blocked {
				return core.Conflict("Revogue a punição vigente antes de atribuir papel administrativo.")
			}
		}
		if e := tx.SaveAdministrator(ctx, id, in.Role, "active", true); e != nil {
			return e
		}
		if e := tx.RevokeSessions(ctx, id, s.now()); e != nil {
			return e
		}
		if e := tx.Audit(ctx, a.ID, "administrator.create", "administrator", id, in.Reason, s.now()); e != nil {
			return e
		}
		v, e := tx.Administrator(ctx, id, false)
		out = adminDTO(v)
		return e
	})
	return out, err
}
func (s *Admin) UpdateAdministrator(ctx context.Context, a core.Actor, id uint64, in schemas.PatchAdministrator) (schemas.Administrator, error) {
	if e := a.Require("system"); e != nil {
		return schemas.Administrator{}, e
	}
	if !adminText(in.Reason, 3, 2000) || (in.Role == nil && in.Status == nil) {
		return schemas.Administrator{}, core.Invalid("INVALID_ADMINISTRATOR", "Informe alteração e motivo.")
	}
	if in.Role != nil && !adminHas(adminRoles, *in.Role) {
		return schemas.Administrator{}, core.Invalid("INVALID_ROLE", "Papel administrativo inválido.")
	}
	if in.Status != nil && !adminHas([]string{"active", "inactive"}, *in.Status) {
		return schemas.Administrator{}, core.Invalid("INVALID_STATUS", "Estado administrativo inválido.")
	}
	var out schemas.Administrator
	err := s.repo.Within(ctx, func(tx interfaces.AdminRepository) error {
		n, e := tx.LockSystemAdministrators(ctx)
		if e != nil {
			return e
		}
		v, e := tx.Administrator(ctx, id, true)
		if e != nil {
			return e
		}
		role, status := v.Role, v.Status
		if in.Role != nil {
			role = *in.Role
		}
		if in.Status != nil {
			status = *in.Status
		}
		if v.Role == "system_administrator" && v.Status == "active" && (role != "system_administrator" || status != "active") && n <= 1 {
			return core.Conflict("O último administrador de sistema ativo deve ser preservado.")
		}
		if e = tx.SaveAdministrator(ctx, id, role, status, false); e != nil {
			return e
		}
		if e = tx.RevokeSessions(ctx, id, s.now()); e != nil {
			return e
		}
		if e = tx.Audit(ctx, a.ID, "administrator.update", "administrator", id, in.Reason, s.now()); e != nil {
			return e
		}
		v, e = tx.Administrator(ctx, id, false)
		out = adminDTO(v)
		return e
	})
	return out, err
}
func (s *Admin) RevokeAdministrator(ctx context.Context, a core.Actor, id uint64, reason string) error {
	status := "inactive"
	_, e := s.UpdateAdministrator(ctx, a, id, schemas.PatchAdministrator{Status: &status, Reason: reason})
	return e
}
func (s *Admin) Users(ctx context.Context, a core.Actor, status string, p schemas.Page) (schemas.List, error) {
	if e := a.Require("moderation"); e != nil {
		return schemas.List{}, e
	}
	if status != "" && !adminHas([]string{"pending_verification", "active", "inactive", "suspended", "deactivated", "banned"}, status) {
		return schemas.List{}, core.Invalid("INVALID_STATUS", "Estado de conta inválido.")
	}
	rows, n, e := s.repo.Users(ctx, status, p)
	if e != nil {
		return schemas.List{}, e
	}
	out := make([]schemas.AdministrativeUser, 0, len(rows))
	for _, v := range rows {
		out = append(out, schemas.AdministrativeUser{ID: v.ID, Name: v.Name, Nickname: v.Nickname, Status: v.Status, CreatedAt: v.CreatedAt})
	}
	return schemas.Paginated(out, p, n), nil
}
func (s *Admin) User(ctx context.Context, a core.Actor, id uint64) (schemas.AdministrativeUser, error) {
	if e := a.Require("moderation"); e != nil {
		return schemas.AdministrativeUser{}, e
	}
	v, e := s.repo.User(ctx, id, false)
	if e != nil {
		return schemas.AdministrativeUser{}, e
	}
	if e = s.repo.Audit(ctx, a.ID, "user.inspect", "user", id, "Consulta administrativa de identificação mínima.", s.now()); e != nil {
		return schemas.AdministrativeUser{}, e
	}
	return schemas.AdministrativeUser{ID: v.ID, Name: v.Name, Nickname: v.Nickname, Status: v.Status, CreatedAt: v.CreatedAt}, nil
}
func reportDTO(v core.Report, admin bool) schemas.Report {
	out := schemas.Report{ID: v.ID, ReportedUserID: v.ReportedUserID, ReportedPostID: v.ReportedPostID, ReportedCommentID: v.ReportedCommentID, Reason: v.Reason, Description: v.Description, Status: v.Status, ResolutionNote: v.ResolutionNote, ResolvedAt: v.ResolvedAt, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
	if admin {
		out.ReporterID = &v.ReporterID
		out.AssignedTo = v.AssignedTo
		out.ReportDataURL = v.ReportDataURL
	}
	return out
}
func reportTarget(in schemas.CreateReport) (string, uint64, error) {
	n := 0
	kind := ""
	id := uint64(0)
	for k, v := range map[string]*uint64{"user": in.ReportedUserID, "post": in.ReportedPostID, "comment": in.ReportedCommentID} {
		if v != nil {
			n++
			kind = k
			id = *v
		}
	}
	if n != 1 || id == 0 {
		return "", 0, core.Invalid("INVALID_REPORT_TARGET", "Informe exatamente um alvo válido para a denúncia.")
	}
	return kind, id, nil
}
func (s *Admin) CreateReport(ctx context.Context, a core.Actor, in schemas.CreateReport) (schemas.Report, error) {
	if e := adminReader(a); e != nil {
		return schemas.Report{}, e
	}
	kind, id, e := reportTarget(in)
	if e != nil {
		return schemas.Report{}, e
	}
	if !adminHas(reportReasons, in.Reason) || !adminText(in.Description, 3, 5000) {
		return schemas.Report{}, core.Invalid("INVALID_REPORT", "Informe motivo e descrição da denúncia.")
	}
	if e = core.ValidateURL(in.ReportDataURL); e != nil {
		return schemas.Report{}, e
	}
	target, e := s.repo.Target(ctx, kind, id)
	if e != nil {
		return schemas.Report{}, e
	}
	if target.UserStatus == "deactivated" || target.Status == "deleted" {
		return schemas.Report{}, core.NotFound()
	}
	if kind != "user" && (target.Status != "published" || target.UserStatus != "active" || target.WorkStatus != "published" || (kind == "comment" && target.ParentStatus != "published") || target.MinimumAge > a.Age(s.now())) {
		return schemas.Report{}, core.NotFound()
	}
	now := s.now()
	v := core.Report{ReporterID: a.ID, ReportedUserID: in.ReportedUserID, ReportedPostID: in.ReportedPostID, ReportedCommentID: in.ReportedCommentID, Reason: in.Reason, Description: strings.TrimSpace(in.Description), ReportDataURL: in.ReportDataURL, Status: "open", CreatedAt: now, UpdatedAt: now}
	v.ID, e = s.repo.CreateReport(ctx, v)
	return reportDTO(v, false), e
}
func (s *Admin) Reports(ctx context.Context, a core.Actor, admin bool, status string, p schemas.Page) (schemas.List, error) {
	var owner *uint64
	if admin {
		if e := a.Require("moderation"); e != nil {
			return schemas.List{}, e
		}
	} else {
		if e := adminReader(a); e != nil {
			return schemas.List{}, e
		}
		owner = &a.ID
	}
	if status != "" && !adminHas(reportStates, status) {
		return schemas.List{}, core.Invalid("INVALID_REPORT_STATUS", "Estado de denúncia inválido.")
	}
	rows, n, e := s.repo.Reports(ctx, owner, status, p)
	if e != nil {
		return schemas.List{}, e
	}
	if admin {
		if e = s.repo.Audit(ctx, a.ID, "report.list", "report", 0, "Consulta autorizada da fila de denúncias.", s.now()); e != nil {
			return schemas.List{}, e
		}
	}
	out := make([]schemas.Report, 0, len(rows))
	for _, v := range rows {
		out = append(out, reportDTO(v, admin))
	}
	return schemas.Paginated(out, p, n), nil
}
func (s *Admin) getReport(ctx context.Context, r interfaces.AdminRepository, a core.Actor, id uint64, admin, lock bool) (core.Report, error) {
	if admin {
		if e := a.Require("moderation"); e != nil {
			return core.Report{}, e
		}
	} else {
		if e := adminReader(a); e != nil {
			return core.Report{}, e
		}
	}
	v, e := r.Report(ctx, id, lock)
	if e != nil {
		return v, e
	}
	if !admin && v.ReporterID != a.ID {
		return v, core.NotFound()
	}
	return v, nil
}
func (s *Admin) Report(ctx context.Context, a core.Actor, id uint64, admin bool) (schemas.Report, error) {
	v, e := s.getReport(ctx, s.repo, a, id, admin, false)
	if e != nil {
		return schemas.Report{}, e
	}
	if admin {
		e = s.repo.Audit(ctx, a.ID, "report.inspect", "report", id, "Inspeção autorizada de denúncia e evidência.", s.now())
	}
	return reportDTO(v, admin), e
}
func reportTransition(from, to string) bool {
	if from == to {
		return true
	}
	switch from {
	case "open":
		return to == "in_review" || to == "awaiting_information" || to == "resolved" || to == "dismissed"
	case "in_review":
		return to == "awaiting_information" || to == "resolved" || to == "dismissed"
	case "awaiting_information":
		return to == "in_review" || to == "resolved" || to == "dismissed"
	}
	return false
}
func (s *Admin) validateAssignee(ctx context.Context, r interfaces.AdminRepository, id *uint64, permission string) error {
	if id == nil {
		return nil
	}
	if *id == 0 {
		return core.Invalid("INVALID_ASSIGNEE", "Responsável inválido.")
	}
	v, e := r.Administrator(ctx, *id, false)
	if e != nil {
		return e
	}
	u, e := r.User(ctx, *id, false)
	if e != nil {
		return e
	}
	if v.Status != "active" || u.Status != "active" {
		return core.Invalid("INVALID_ASSIGNEE", "Escolha um administrador ativo.")
	}
	return (core.Actor{ID: *id, Role: v.Role}).Require(permission)
}
func (s *Admin) UpdateReport(ctx context.Context, a core.Actor, id uint64, in schemas.PatchReport) (schemas.Report, error) {
	if e := a.Require("moderation"); e != nil {
		return schemas.Report{}, e
	}
	if in.Status == nil && !in.AssignedTo.Set && in.ResolutionNote == nil {
		return schemas.Report{}, core.Invalid("EMPTY_PATCH", "Informe uma alteração.")
	}
	var out schemas.Report
	err := s.repo.Within(ctx, func(tx interfaces.AdminRepository) error {
		v, e := s.getReport(ctx, tx, a, id, true, true)
		if e != nil {
			return e
		}
		if adminHas([]string{"resolved", "dismissed", "canceled"}, v.Status) {
			return core.Conflict("Denúncia encerrada não pode ser alterada.")
		}
		if in.AssignedTo.Set {
			if e = s.validateAssignee(ctx, tx, in.AssignedTo.Value, "moderation"); e != nil {
				return e
			}
			v.AssignedTo = in.AssignedTo.Value
		}
		if in.Status != nil {
			if !reportTransition(v.Status, *in.Status) {
				return core.Conflict("Transição de denúncia não permitida.")
			}
			v.Status = *in.Status
		}
		if in.ResolutionNote != nil {
			if !adminText(*in.ResolutionNote, 3, 5000) {
				return core.Invalid("INVALID_RESOLUTION", "Informe uma resolução válida.")
			}
			v.ResolutionNote = in.ResolutionNote
		}
		if v.Status == "resolved" || v.Status == "dismissed" {
			if v.ResolutionNote == nil {
				return core.Invalid("RESOLUTION_REQUIRED", "A conclusão exige uma resolução.")
			}
			now := s.now()
			v.ResolvedAt = &now
		}
		v.UpdatedAt = s.now()
		if e = tx.SaveReport(ctx, v); e != nil {
			return e
		}
		reason := "Atualização de estado ou atribuição da denúncia."
		if v.ResolutionNote != nil {
			reason = *v.ResolutionNote
		}
		if e = tx.Audit(ctx, a.ID, "report.update", "report", id, reason, s.now()); e != nil {
			return e
		}
		out = reportDTO(v, true)
		return nil
	})
	return out, err
}
func (s *Admin) CancelReport(ctx context.Context, a core.Actor, id uint64) error {
	return s.repo.Within(ctx, func(tx interfaces.AdminRepository) error {
		v, e := s.getReport(ctx, tx, a, id, false, true)
		if e != nil {
			return e
		}
		if v.Status == "canceled" {
			return nil
		}
		if v.Status != "open" || v.AssignedTo != nil {
			return core.Conflict("Somente denúncia ainda não analisada pode ser cancelada.")
		}
		v.Status = "canceled"
		v.UpdatedAt = s.now()
		return tx.SaveReport(ctx, v)
	})
}
func (s *Admin) ReportMessages(ctx context.Context, a core.Actor, id uint64, admin bool, p schemas.Page) (schemas.List, error) {
	if _, e := s.getReport(ctx, s.repo, a, id, admin, false); e != nil {
		return schemas.List{}, e
	}
	rows, n, e := s.repo.ReportMessages(ctx, id, admin, p)
	if e != nil {
		return schemas.List{}, e
	}
	if admin {
		if e = s.repo.Audit(ctx, a.ID, "report.messages.inspect", "report", id, "Consulta autorizada ao histórico da denúncia.", s.now()); e != nil {
			return schemas.List{}, e
		}
	}
	out := make([]schemas.ReportMessage, 0, len(rows))
	for _, v := range rows {
		out = append(out, schemas.ReportMessage{ID: v.ID, AuthorID: v.AuthorID, Text: v.Text, Visibility: v.Visibility, CreatedAt: v.CreatedAt})
	}
	return schemas.Paginated(out, p, n), nil
}
func (s *Admin) AddReportMessage(ctx context.Context, a core.Actor, id uint64, admin bool, in schemas.CreateReportMessage) (schemas.ReportMessage, error) {
	if !adminText(in.Text, 1, 5000) {
		return schemas.ReportMessage{}, core.Invalid("INVALID_MESSAGE", "Informe uma mensagem de até 5000 caracteres.")
	}
	if in.Visibility == "" {
		in.Visibility = "reporter"
	}
	if !adminHas([]string{"reporter", "internal"}, in.Visibility) || (!admin && (in.Visibility != "reporter" || in.RequestInformation)) {
		return schemas.ReportMessage{}, core.Forbidden()
	}
	var out schemas.ReportMessage
	err := s.repo.Within(ctx, func(tx interfaces.AdminRepository) error {
		v, e := s.getReport(ctx, tx, a, id, admin, true)
		if e != nil {
			return e
		}
		if adminHas([]string{"resolved", "dismissed", "canceled"}, v.Status) {
			return core.Conflict("Denúncia encerrada não aceita mensagens.")
		}
		if !admin && v.Status != "awaiting_information" {
			return core.Conflict("Aguarde uma solicitação de informações antes de complementar.")
		}
		if in.RequestInformation {
			if in.Visibility != "reporter" {
				return core.Invalid("INVALID_VISIBILITY", "Solicitação de informações deve ser visível ao denunciante.")
			}
			v.Status = "awaiting_information"
		} else if !admin {
			v.Status = "in_review"
		}
		now := s.now()
		message := core.ReportMessage{ReportID: id, AuthorID: a.ID, Text: strings.TrimSpace(in.Text), Visibility: in.Visibility, CreatedAt: now}
		message.ID, e = tx.CreateReportMessage(ctx, message)
		if e != nil {
			return e
		}
		v.UpdatedAt = now
		if e = tx.SaveReport(ctx, v); e != nil {
			return e
		}
		if admin {
			if e = tx.Audit(ctx, a.ID, "report.message", "report", id, "Mensagem administrativa de acompanhamento.", now); e != nil {
				return e
			}
		}
		out = schemas.ReportMessage{ID: message.ID, AuthorID: a.ID, Text: message.Text, Visibility: message.Visibility, CreatedAt: now}
		return nil
	})
	return out, err
}
func (s *Admin) Moderate(ctx context.Context, a core.Actor, kind string, id uint64, in schemas.Moderation) error {
	if e := a.Require("moderation"); e != nil {
		return e
	}
	if !adminHas([]string{"hidden", "removed", "published"}, in.Status) || !adminText(in.Reason, 3, 2000) {
		return core.Invalid("INVALID_MODERATION", "Informe estado de moderação e motivo válidos.")
	}
	return s.repo.Within(ctx, func(tx interfaces.AdminRepository) error {
		v, e := tx.Target(ctx, kind, id)
		if e != nil {
			return e
		}
		if v.Status == "deleted" || v.Status == "draft" {
			return core.Conflict("Conteúdo excluído ou rascunho não pode ser restaurado por moderação.")
		}
		if in.Status == "published" && !adminHas([]string{"hidden", "removed", "published"}, v.Status) {
			return core.Conflict("Restauração não permitida.")
		}
		now := s.now()
		if e = tx.Moderate(ctx, kind, id, in.Status, now); e != nil {
			return e
		}
		if e = tx.Audit(ctx, a.ID, "content."+in.Status, kind, id, in.Reason, now); e != nil {
			return e
		}
		if v.OwnerID != a.ID {
			return tx.NotifyModeration(ctx, v.OwnerID, a.ID, kind, id, fmt.Sprintf("moderation:%s:%d:%d", kind, id, now.UnixNano()), now)
		}
		return nil
	})
}

func (s *Admin) InspectContent(ctx context.Context, a core.Actor, kind string, id uint64) (schemas.ContentEvidence, error) {
	if e := a.Require("moderation"); e != nil {
		return schemas.ContentEvidence{}, e
	}
	v, e := s.repo.ContentEvidence(ctx, kind, id)
	if e != nil {
		return schemas.ContentEvidence{}, e
	}
	if e = s.repo.Audit(ctx, a.ID, "content.inspect", kind, id, "Inspeção autorizada de evidência para moderação.", s.now()); e != nil {
		return schemas.ContentEvidence{}, e
	}
	return schemas.ContentEvidence{ID: v.ID, ReaderID: v.ReaderID, PostID: v.PostID, WorkID: v.WorkID, Kind: v.Kind, Title: v.Title, Text: v.Text, Status: v.Status, ContainsSpoiler: v.ContainsSpoiler, ContainsSensitiveContent: v.ContainsSensitiveContent, MinimumAge: v.MinimumAge, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}, nil
}
func penaltyDTO(v core.Penalty) schemas.Penalty {
	return schemas.Penalty{ID: v.ID, UserID: v.UserID, AppliedBy: v.AppliedBy, RevokedBy: v.RevokedBy, Type: v.Type, Reason: v.Reason, StartsAt: v.StartsAt, EndsAt: v.EndsAt, RevokedAt: v.RevokedAt, RevocationReason: v.RevocationReason, CreatedAt: v.CreatedAt}
}
func validatePenalty(in schemas.CreatePenalty, now time.Time) (time.Time, error) {
	if in.UserID == 0 || !adminHas([]string{"warning", "restriction", "suspension", "ban"}, in.Type) || !adminText(in.Reason, 3, 2000) {
		return now, core.Invalid("INVALID_PENALTY", "Informe leitor, tipo e motivo válidos.")
	}
	start := now
	if in.StartsAt != nil {
		start = in.StartsAt.UTC()
	}
	if in.Type == "ban" || in.Type == "warning" {
		if in.EndsAt != nil {
			return start, core.Invalid("INVALID_PENALTY_PERIOD", "Advertência e banimento não recebem data final.")
		}
	} else if in.EndsAt == nil || !in.EndsAt.After(start) || !in.EndsAt.After(now) {
		return start, core.Invalid("INVALID_PENALTY_PERIOD", "Suspensão/restrição exige fim futuro posterior ao início.")
	}
	if start.Before(now.Add(-time.Minute)) {
		return start, core.Invalid("INVALID_PENALTY_PERIOD", "O início não pode ser retroativo.")
	}
	seen := map[uint64]bool{}
	for _, id := range in.ReportIDs {
		if id == 0 || seen[id] {
			return start, core.Invalid("INVALID_REPORT_IDS", "Informe denúncias distintas e válidas.")
		}
		seen[id] = true
	}
	return start, nil
}
func (s *Admin) ApplyPenalty(ctx context.Context, a core.Actor, in schemas.CreatePenalty) (schemas.Penalty, error) {
	if e := a.Require("moderation"); e != nil {
		return schemas.Penalty{}, e
	}
	now := s.now()
	start, e := validatePenalty(in, now)
	if e != nil {
		return schemas.Penalty{}, e
	}
	var out schemas.Penalty
	err := s.repo.Within(ctx, func(tx interfaces.AdminRepository) error {
		u, e := tx.User(ctx, in.UserID, true)
		if e != nil {
			return e
		}
		if u.Status == "deactivated" {
			return core.Conflict("Conta desativada não recebe punição.")
		}
		if administrator, ae := tx.Administrator(ctx, in.UserID, false); ae == nil {
			if administrator.Status == "active" {
				return core.Forbidden()
			}
		} else if !adminIsNotFound(ae) {
			return ae
		}
		for _, rid := range in.ReportIDs {
			report, e := tx.Report(ctx, rid, false)
			if e != nil {
				return e
			}
			target := in.UserID
			if report.ReportedUserID != nil {
				target = *report.ReportedUserID
			} else {
				kind, id := "post", uint64(0)
				if report.ReportedPostID != nil {
					id = *report.ReportedPostID
				} else {
					kind = "comment"
					id = *report.ReportedCommentID
				}
				v, e := tx.Target(ctx, kind, id)
				if e != nil {
					return e
				}
				target = v.OwnerID
			}
			if target != in.UserID {
				return core.Invalid("UNRELATED_REPORT", "A denúncia não corresponde ao leitor punido.")
			}
		}
		v := core.Penalty{UserID: in.UserID, AppliedBy: a.ID, Type: in.Type, Reason: strings.TrimSpace(in.Reason), StartsAt: start, EndsAt: in.EndsAt, CreatedAt: now}
		v.ID, e = tx.CreatePenalty(ctx, v)
		if e != nil {
			return e
		}
		if e = tx.LinkPenaltyReports(ctx, v.ID, in.ReportIDs); e != nil {
			return e
		}
		if (in.Type == "suspension" || in.Type == "ban") && !start.After(now) {
			if e = tx.RevokeSessions(ctx, in.UserID, now); e != nil {
				return e
			}
		}
		if e = tx.Audit(ctx, a.ID, "penalty.apply", "penalty", v.ID, in.Reason, now); e != nil {
			return e
		}
		if e = tx.NotifyModeration(ctx, in.UserID, a.ID, "user", in.UserID, fmt.Sprintf("penalty:%d", v.ID), now); e != nil {
			return e
		}
		out = penaltyDTO(v)
		return nil
	})
	return out, err
}
func (s *Admin) Penalties(ctx context.Context, a core.Actor, user *uint64, p schemas.Page) (schemas.List, error) {
	if e := a.Require("moderation"); e != nil {
		return schemas.List{}, e
	}
	rows, n, e := s.repo.Penalties(ctx, user, p)
	if e != nil {
		return schemas.List{}, e
	}
	out := make([]schemas.Penalty, 0, len(rows))
	for _, v := range rows {
		out = append(out, penaltyDTO(v))
	}
	return schemas.Paginated(out, p, n), nil
}
func (s *Admin) RevokePenalty(ctx context.Context, a core.Actor, id uint64, reason string) error {
	if e := a.Require("moderation"); e != nil {
		return e
	}
	if !adminText(reason, 3, 2000) {
		return core.Invalid("INVALID_REASON", "Informe o motivo da revogação.")
	}
	return s.repo.Within(ctx, func(tx interfaces.AdminRepository) error {
		v, e := tx.Penalty(ctx, id, true)
		if e != nil {
			return e
		}
		if v.RevokedAt != nil {
			return nil
		}
		now := s.now()
		if e = tx.RevokePenalty(ctx, id, a.ID, reason, now); e != nil {
			return e
		}
		return tx.Audit(ctx, a.ID, "penalty.revoke", "penalty", id, reason, now)
	})
}
func faqDTO(v core.FAQEntry) schemas.FAQ {
	return schemas.FAQ{ID: v.ID, Question: v.Question, Answer: v.Answer, SortOrder: v.SortOrder, Status: v.Status, UpdatedAt: v.UpdatedAt}
}
func (s *Admin) FAQs(ctx context.Context, a core.Actor, admin bool, p schemas.Page) (schemas.List, error) {
	if admin {
		if e := a.Require("support"); e != nil {
			return schemas.List{}, e
		}
	}
	rows, n, e := s.repo.FAQs(ctx, admin, p)
	if e != nil {
		return schemas.List{}, e
	}
	out := make([]schemas.FAQ, 0, len(rows))
	for _, v := range rows {
		out = append(out, faqDTO(v))
	}
	return schemas.Paginated(out, p, n), nil
}
func (s *Admin) FAQ(ctx context.Context, a core.Actor, id uint64, admin bool) (schemas.FAQ, error) {
	if admin {
		if e := a.Require("support"); e != nil {
			return schemas.FAQ{}, e
		}
	}
	v, e := s.repo.FAQ(ctx, id, admin)
	return faqDTO(v), e
}
func (s *Admin) SaveFAQ(ctx context.Context, a core.Actor, id uint64, in schemas.SaveFAQ) (schemas.FAQ, error) {
	if e := a.Require("support"); e != nil {
		return schemas.FAQ{}, e
	}
	var out schemas.FAQ
	err := s.repo.Within(ctx, func(tx interfaces.AdminRepository) error {
		now := s.now()
		v := core.FAQEntry{Status: "draft", CreatedAt: now, UpdatedAt: now}
		if id != 0 {
			var e error
			v, e = tx.FAQ(ctx, id, true)
			if e != nil {
				return e
			}
		}
		if in.Question != nil {
			v.Question = strings.TrimSpace(*in.Question)
		}
		if in.Answer != nil {
			v.Answer = strings.TrimSpace(*in.Answer)
		}
		if in.SortOrder != nil {
			v.SortOrder = *in.SortOrder
		}
		if in.Status != nil {
			v.Status = *in.Status
		}
		if !adminText(v.Question, 3, 500) || !adminText(v.Answer, 1, 10000) || v.SortOrder < 0 || !adminHas([]string{"draft", "published"}, v.Status) {
			return core.Invalid("INVALID_FAQ", "Informe pergunta, resposta, ordem e estado válidos.")
		}
		v.UpdatedAt = now
		var e error
		v.ID, e = tx.SaveFAQ(ctx, v, id == 0)
		if e != nil {
			return e
		}
		if e = tx.Audit(ctx, a.ID, "faq.save", "faq", v.ID, "Manutenção da FAQ.", now); e != nil {
			return e
		}
		out = faqDTO(v)
		return nil
	})
	return out, err
}
func (s *Admin) DeleteFAQ(ctx context.Context, a core.Actor, id uint64) error {
	if e := a.Require("support"); e != nil {
		return e
	}
	return s.repo.Within(ctx, func(tx interfaces.AdminRepository) error {
		if _, e := tx.FAQ(ctx, id, true); e != nil {
			return e
		}
		if e := tx.DeleteFAQ(ctx, id); e != nil {
			return e
		}
		return tx.Audit(ctx, a.ID, "faq.delete", "faq", id, "Exclusão de pergunta da FAQ.", s.now())
	})
}
func ticketDTO(v core.SupportTicket) schemas.SupportTicket {
	return schemas.SupportTicket{ID: v.ID, RequesterID: v.RequesterID, AssignedTo: v.AssignedTo, Subject: v.Subject, Status: v.Status, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
func (s *Admin) CreateTicket(ctx context.Context, a core.Actor, in schemas.CreateSupportTicket) (schemas.SupportTicket, error) {
	if e := adminReader(a); e != nil {
		return schemas.SupportTicket{}, e
	}
	if !adminText(in.Subject, 3, 200) || !adminText(in.Text, 1, 10000) {
		return schemas.SupportTicket{}, core.Invalid("INVALID_SUPPORT_TICKET", "Informe assunto e mensagem válidos.")
	}
	var out schemas.SupportTicket
	err := s.repo.Within(ctx, func(tx interfaces.AdminRepository) error {
		now := s.now()
		v := core.SupportTicket{RequesterID: a.ID, Subject: strings.TrimSpace(in.Subject), Status: "open", CreatedAt: now, UpdatedAt: now}
		var e error
		v.ID, e = tx.CreateTicket(ctx, v)
		if e != nil {
			return e
		}
		if _, e = tx.CreateTicketMessage(ctx, core.SupportMessage{TicketID: v.ID, AuthorID: a.ID, Text: strings.TrimSpace(in.Text), CreatedAt: now}); e != nil {
			return e
		}
		out = ticketDTO(v)
		return nil
	})
	return out, err
}
func (s *Admin) getTicket(ctx context.Context, r interfaces.AdminRepository, a core.Actor, id uint64, admin, lock bool) (core.SupportTicket, error) {
	if admin {
		if e := a.Require("support"); e != nil {
			return core.SupportTicket{}, e
		}
	} else if a.ID == 0 {
		return core.SupportTicket{}, core.ErrUnauthenticated
	}
	v, e := r.Ticket(ctx, id, lock)
	if e != nil {
		return v, e
	}
	if !admin && v.RequesterID != a.ID {
		if e = a.Require("support"); e != nil {
			return v, core.NotFound()
		}
	}
	return v, nil
}
func (s *Admin) Ticket(ctx context.Context, a core.Actor, id uint64, admin bool) (schemas.SupportTicket, error) {
	v, e := s.getTicket(ctx, s.repo, a, id, admin, false)
	if e == nil && (admin || v.RequesterID != a.ID) {
		e = s.repo.Audit(ctx, a.ID, "support.inspect", "support_ticket", id, "Consulta autorizada de atendimento.", s.now())
	}
	return ticketDTO(v), e
}
func (s *Admin) Tickets(ctx context.Context, a core.Actor, admin bool, status string, p schemas.Page) (schemas.List, error) {
	var owner *uint64
	if admin {
		if e := a.Require("support"); e != nil {
			return schemas.List{}, e
		}
	} else {
		if e := adminReader(a); e != nil {
			return schemas.List{}, e
		}
		owner = &a.ID
	}
	if status != "" && !adminHas(ticketStates, status) {
		return schemas.List{}, core.Invalid("INVALID_TICKET_STATUS", "Estado de chamado inválido.")
	}
	rows, n, e := s.repo.Tickets(ctx, owner, status, p)
	if e != nil {
		return schemas.List{}, e
	}
	out := make([]schemas.SupportTicket, 0, len(rows))
	for _, v := range rows {
		out = append(out, ticketDTO(v))
	}
	return schemas.Paginated(out, p, n), nil
}
func ticketTransition(from, to string) bool {
	if from == to {
		return true
	}
	switch from {
	case "open":
		return to == "in_progress" || to == "resolved" || to == "closed"
	case "in_progress":
		return to == "resolved" || to == "closed"
	case "resolved":
		return to == "in_progress" || to == "closed"
	}
	return false
}
func (s *Admin) UpdateTicket(ctx context.Context, a core.Actor, id uint64, in schemas.PatchSupportTicket) (schemas.SupportTicket, error) {
	if e := a.Require("support"); e != nil {
		return schemas.SupportTicket{}, e
	}
	if in.Status == nil && !in.AssignedTo.Set {
		return schemas.SupportTicket{}, core.Invalid("EMPTY_PATCH", "Informe uma alteração.")
	}
	var out schemas.SupportTicket
	err := s.repo.Within(ctx, func(tx interfaces.AdminRepository) error {
		v, e := s.getTicket(ctx, tx, a, id, true, true)
		if e != nil {
			return e
		}
		if in.Status != nil {
			if !ticketTransition(v.Status, *in.Status) {
				return core.Conflict("Transição de chamado não permitida.")
			}
			v.Status = *in.Status
		}
		if in.AssignedTo.Set {
			if e = s.validateAssignee(ctx, tx, in.AssignedTo.Value, "support"); e != nil {
				return e
			}
			v.AssignedTo = in.AssignedTo.Value
		}
		v.UpdatedAt = s.now()
		if e = tx.SaveTicket(ctx, v); e != nil {
			return e
		}
		if e = tx.Audit(ctx, a.ID, "support.update", "support_ticket", id, "Alteração de estado ou responsável do chamado.", s.now()); e != nil {
			return e
		}
		out = ticketDTO(v)
		return nil
	})
	return out, err
}
func (s *Admin) TicketMessages(ctx context.Context, a core.Actor, id uint64, p schemas.Page) (schemas.List, error) {
	ticket, e := s.getTicket(ctx, s.repo, a, id, false, false)
	if e != nil {
		return schemas.List{}, e
	}
	rows, n, e := s.repo.TicketMessages(ctx, id, p)
	if e != nil {
		return schemas.List{}, e
	}
	if ticket.RequesterID != a.ID {
		if e = s.repo.Audit(ctx, a.ID, "support.messages.inspect", "support_ticket", id, "Consulta autorizada de mensagens do atendimento.", s.now()); e != nil {
			return schemas.List{}, e
		}
	}
	out := make([]schemas.SupportMessage, 0, len(rows))
	for _, v := range rows {
		out = append(out, schemas.SupportMessage{ID: v.ID, AuthorID: v.AuthorID, Text: v.Text, CreatedAt: v.CreatedAt})
	}
	return schemas.Paginated(out, p, n), nil
}

func (s *Admin) AdminTicketMessages(ctx context.Context, a core.Actor, id uint64, p schemas.Page) (schemas.List, error) {
	if e := a.Require("support"); e != nil {
		return schemas.List{}, e
	}
	return s.TicketMessages(ctx, a, id, p)
}
func (s *Admin) AddTicketMessage(ctx context.Context, a core.Actor, id uint64, forceAdmin bool, in schemas.CreateSupportMessage) (schemas.SupportMessage, error) {
	if !adminText(in.Text, 1, 10000) {
		return schemas.SupportMessage{}, core.Invalid("INVALID_MESSAGE", "Informe uma mensagem de até 10000 caracteres.")
	}
	var out schemas.SupportMessage
	err := s.repo.Within(ctx, func(tx interfaces.AdminRepository) error {
		v, e := s.getTicket(ctx, tx, a, id, forceAdmin, true)
		if e != nil {
			return e
		}
		if v.Status == "closed" {
			return core.Conflict("Chamado encerrado não aceita mensagens; abra outro atendimento.")
		}
		admin := a.ID != v.RequesterID || forceAdmin
		if admin {
			if e = a.Require("support"); e != nil {
				return e
			}
		}
		now := s.now()
		message := core.SupportMessage{TicketID: id, AuthorID: a.ID, Text: strings.TrimSpace(in.Text), CreatedAt: now}
		message.ID, e = tx.CreateTicketMessage(ctx, message)
		if e != nil {
			return e
		}
		if v.Status == "resolved" {
			v.Status = "in_progress"
		}
		v.UpdatedAt = now
		if e = tx.SaveTicket(ctx, v); e != nil {
			return e
		}
		if admin {
			if e = tx.Audit(ctx, a.ID, "support.message", "support_ticket", id, "Resposta administrativa ao chamado.", now); e != nil {
				return e
			}
		}
		out = schemas.SupportMessage{ID: message.ID, AuthorID: a.ID, Text: message.Text, CreatedAt: now}
		return nil
	})
	return out, err
}
func (s *Admin) CloseTicket(ctx context.Context, a core.Actor, id uint64) error {
	return s.repo.Within(ctx, func(tx interfaces.AdminRepository) error {
		v, e := s.getTicket(ctx, tx, a, id, false, true)
		if e != nil {
			return e
		}
		if v.Status == "closed" {
			return nil
		}
		v.Status = "closed"
		v.UpdatedAt = s.now()
		if e = tx.SaveTicket(ctx, v); e != nil {
			return e
		}
		if v.RequesterID != a.ID {
			return tx.Audit(ctx, a.ID, "support.close", "support_ticket", id, "Encerramento administrativo do chamado.", s.now())
		}
		return nil
	})
}
func (s *Admin) AuditLogs(ctx context.Context, a core.Actor, kind string, p schemas.Page) (schemas.List, error) {
	if e := a.Require("audit"); e != nil {
		return schemas.List{}, e
	}
	if a.Role == "catalog_manager" && !adminHas([]string{"work", "edition", "author", "publisher", "genre", "catalog_suggestion"}, kind) {
		return schemas.List{}, core.Forbidden()
	}
	if a.Role == "user_manager" && !adminHas([]string{"user", "report", "penalty", "post", "comment", "faq", "support_ticket"}, kind) {
		return schemas.List{}, core.Forbidden()
	}
	rows, n, e := s.repo.AuditLogs(ctx, kind, p)
	if e != nil {
		return schemas.List{}, e
	}
	out := make([]schemas.AuditLog, 0, len(rows))
	for _, v := range rows {
		out = append(out, schemas.AuditLog{ID: v.ID, ActorID: v.ActorID, TargetID: v.TargetID, TargetType: v.TargetType, Action: v.Action, Reason: v.Reason, CreatedAt: v.CreatedAt})
	}
	return schemas.Paginated(out, p, n), nil
}
