package tools

import (
	"context"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/schemas"
	"time"
)

type AdminRepository struct{ store interfaces.Store }

func NewAdminRepository(store interfaces.Store) *AdminRepository {
	return &AdminRepository{store: store}
}
func (r *AdminRepository) Within(ctx context.Context, fn func(interfaces.AdminRepository) error) error {
	return r.store.Transaction(ctx, func(tx interfaces.Store) error { return fn(NewAdminRepository(tx)) })
}
func adminLock(lock bool) string {
	if lock {
		return " FOR UPDATE"
	}
	return ""
}
func (r *AdminRepository) User(ctx context.Context, id uint64, lock bool) (core.AdminUser, error) {
	var v core.AdminUser
	err := r.store.Query(ctx, &v, "SELECT id,name,nickname,email,password_hash,status,birth_date,profile_image_url,created_at,updated_at FROM users WHERE id=?"+adminLock(lock), id)
	if err == nil && v.ID == 0 {
		err = core.NotFound()
	}
	return v, err
}
func (r *AdminRepository) Users(ctx context.Context, status string, p schemas.Page) ([]core.AdminUser, int64, error) {
	where := "1=1"
	args := []any{}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var total int64
	if err := r.store.Query(ctx, &total, "SELECT COUNT(*) FROM users WHERE "+where, args...); err != nil {
		return nil, 0, err
	}
	var rows []core.AdminUser
	args = append(args, p.PageSize, p.Offset())
	err := r.store.Query(ctx, &rows, "SELECT id,name,nickname,status,created_at FROM users WHERE "+where+" ORDER BY id LIMIT ? OFFSET ?", args...)
	return rows, total, err
}
func (r *AdminRepository) CreateAdministrativeUser(ctx context.Context, u core.AdminUser) (uint64, error) {
	id, err := r.store.Insert(ctx, "users", map[string]any{"name": u.Name, "nickname": u.Nickname, "email": u.Email, "password_hash": u.PasswordHash, "status": "active", "birth_date": u.BirthDate, "profile_image_url": "", "created_at": u.CreatedAt, "updated_at": u.UpdatedAt})
	if err != nil {
		return 0, err
	}
	_, err = r.store.Exec(ctx, "INSERT INTO readers(user_id,biography,profile_visibility) VALUES(?,'','public')", id)
	return id, err
}
func (r *AdminRepository) Administrator(ctx context.Context, id uint64, lock bool) (core.AdministratorRecord, error) {
	var row core.AdministratorRecord
	err := r.store.Query(ctx, &row, "SELECT a.user_id,u.name,u.nickname,a.role,a.status,a.created_at,a.updated_at FROM administrators a JOIN users u ON u.id=a.user_id WHERE a.user_id=?"+adminLock(lock), id)
	if err == nil && row.UserID == 0 {
		err = core.NotFound()
	}
	return row, err
}
func (r *AdminRepository) Administrators(ctx context.Context, p schemas.Page) ([]core.AdministratorRecord, int64, error) {
	var n int64
	if err := r.store.Query(ctx, &n, "SELECT COUNT(*) FROM administrators"); err != nil {
		return nil, 0, err
	}
	var rows []core.AdministratorRecord
	err := r.store.Query(ctx, &rows, "SELECT a.user_id,u.name,u.nickname,a.role,a.status,a.created_at,a.updated_at FROM administrators a JOIN users u ON u.id=a.user_id ORDER BY a.user_id LIMIT ? OFFSET ?", p.PageSize, p.Offset())
	return rows, n, err
}
func (r *AdminRepository) LockSystemAdministrators(ctx context.Context) (int64, error) {
	var rows []struct{ UserID uint64 }
	err := r.store.Query(ctx, &rows, "SELECT a.user_id FROM administrators a JOIN users u ON u.id=a.user_id WHERE a.role='system_administrator' AND a.status='active' AND u.status='active' ORDER BY a.user_id FOR UPDATE")
	return int64(len(rows)), err
}
func (r *AdminRepository) SaveAdministrator(ctx context.Context, id uint64, role, status string, create bool) error {
	var err error
	if create {
		_, err = r.store.Exec(ctx, "INSERT INTO administrators(user_id,role,status,created_at,updated_at) VALUES(?,?,?,UTC_TIMESTAMP(6),UTC_TIMESTAMP(6))", id, role, status)
	} else {
		_, err = r.store.Exec(ctx, "UPDATE administrators SET role=?,status=?,updated_at=UTC_TIMESTAMP(6) WHERE user_id=?", role, status, id)
	}
	return err
}
func (r *AdminRepository) RevokeSessions(ctx context.Context, id uint64, now time.Time) error {
	_, err := r.store.Exec(ctx, "UPDATE sessions SET revoked_at=? WHERE user_id=? AND revoked_at IS NULL", now, id)
	return err
}

func (r *AdminRepository) HasBlockingPenalty(ctx context.Context, id uint64, now time.Time) (bool, error) {
	var n int64
	e := r.store.Query(ctx, &n, "SELECT COUNT(*) FROM penalties WHERE user_id=? AND revoked_at IS NULL AND starts_at<=? AND (ends_at IS NULL OR ends_at>?) AND type IN ('ban','suspension','restriction')", id, now, now)
	return n > 0, e
}
func (r *AdminRepository) Target(ctx context.Context, kind string, id uint64) (core.ReportTarget, error) {
	var row core.ReportTarget
	var err error
	switch kind {
	case "user":
		err = r.store.Query(ctx, &row, "SELECT id,id AS owner_id,status,status AS user_status,'' AS work_status,0 AS minimum_age FROM users WHERE id=?", id)
	case "post":
		err = r.store.Query(ctx, &row, "SELECT p.id,p.reader_id AS owner_id,p.status,u.status AS user_status,w.status AS work_status,GREATEST(w.minimum_age,p.minimum_age) AS minimum_age FROM posts p JOIN users u ON u.id=p.reader_id JOIN works w ON w.id=p.work_id WHERE p.id=?", id)
	case "comment":
		err = r.store.Query(ctx, &row, "SELECT c.id,c.reader_id AS owner_id,c.status,p.status AS parent_status,u.status AS user_status,w.status AS work_status,GREATEST(w.minimum_age,p.minimum_age) AS minimum_age FROM comments c JOIN posts p ON p.id=c.post_id JOIN users u ON u.id=c.reader_id JOIN works w ON w.id=p.work_id JOIN users pu ON pu.id=p.reader_id WHERE c.id=? AND pu.status='active'", id)
	default:
		return row, core.NotFound()
	}
	if err == nil && row.ID == 0 {
		err = core.NotFound()
	}
	return row, err
}
func (r *AdminRepository) CreateReport(ctx context.Context, v core.Report) (uint64, error) {
	return r.store.Insert(ctx, "reports", map[string]any{"reporter_id": v.ReporterID, "reported_user_id": v.ReportedUserID, "reported_post_id": v.ReportedPostID, "reported_comment_id": v.ReportedCommentID, "reason": v.Reason, "description": v.Description, "report_data_url": v.ReportDataURL, "status": v.Status, "created_at": v.CreatedAt, "updated_at": v.UpdatedAt})
}

func (r *AdminRepository) ContentEvidence(ctx context.Context, kind string, id uint64) (core.ContentEvidence, error) {
	var v core.ContentEvidence
	var err error
	switch kind {
	case "post":
		err = r.store.Query(ctx, &v, "SELECT p.id,p.reader_id,p.id AS post_id,p.work_id,p.kind,p.title,p.text,p.status,p.contains_spoiler,p.contains_sensitive_content,GREATEST(w.minimum_age,p.minimum_age) AS minimum_age,p.created_at,p.updated_at FROM posts p JOIN works w ON w.id=p.work_id WHERE p.id=?", id)
	case "comment":
		err = r.store.Query(ctx, &v, "SELECT c.id,c.reader_id,c.post_id,p.work_id,'comment' AS kind,'' AS title,c.text,c.status,p.contains_spoiler,p.contains_sensitive_content,GREATEST(w.minimum_age,p.minimum_age) AS minimum_age,c.created_at,c.updated_at FROM comments c JOIN posts p ON p.id=c.post_id JOIN works w ON w.id=p.work_id WHERE c.id=?", id)
	default:
		return v, core.NotFound()
	}
	if err == nil && v.ID == 0 {
		err = core.NotFound()
	}
	return v, err
}
func (r *AdminRepository) Report(ctx context.Context, id uint64, lock bool) (core.Report, error) {
	var v core.Report
	err := r.store.Query(ctx, &v, "SELECT * FROM reports WHERE id=?"+adminLock(lock), id)
	if err == nil && v.ID == 0 {
		err = core.NotFound()
	}
	return v, err
}
func (r *AdminRepository) Reports(ctx context.Context, owner *uint64, status string, p schemas.Page) ([]core.Report, int64, error) {
	where := "1=1"
	args := []any{}
	if owner != nil {
		where += " AND reporter_id=?"
		args = append(args, *owner)
	}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var n int64
	if err := r.store.Query(ctx, &n, "SELECT COUNT(*) FROM reports WHERE "+where, args...); err != nil {
		return nil, 0, err
	}
	var rows []core.Report
	args = append(args, p.PageSize, p.Offset())
	err := r.store.Query(ctx, &rows, "SELECT * FROM reports WHERE "+where+" ORDER BY created_at DESC,id DESC LIMIT ? OFFSET ?", args...)
	return rows, n, err
}
func (r *AdminRepository) SaveReport(ctx context.Context, v core.Report) error {
	_, err := r.store.Exec(ctx, "UPDATE reports SET assigned_to=?,status=?,resolution_note=?,resolved_at=?,updated_at=? WHERE id=?", v.AssignedTo, v.Status, v.ResolutionNote, v.ResolvedAt, v.UpdatedAt, v.ID)
	return err
}
func (r *AdminRepository) ReportMessages(ctx context.Context, id uint64, admin bool, p schemas.Page) ([]core.ReportMessage, int64, error) {
	where := "report_id=?"
	if !admin {
		where += " AND visibility='reporter'"
	}
	var n int64
	if err := r.store.Query(ctx, &n, "SELECT COUNT(*) FROM report_messages WHERE "+where, id); err != nil {
		return nil, 0, err
	}
	var rows []core.ReportMessage
	err := r.store.Query(ctx, &rows, "SELECT * FROM report_messages WHERE "+where+" ORDER BY id LIMIT ? OFFSET ?", id, p.PageSize, p.Offset())
	return rows, n, err
}
func (r *AdminRepository) CreateReportMessage(ctx context.Context, v core.ReportMessage) (uint64, error) {
	return r.store.Insert(ctx, "report_messages", map[string]any{"report_id": v.ReportID, "author_id": v.AuthorID, "text": v.Text, "visibility": v.Visibility, "created_at": v.CreatedAt})
}
func (r *AdminRepository) Moderate(ctx context.Context, kind string, id uint64, status string, now time.Time) error {
	table := "posts"
	if kind == "comment" {
		table = "comments"
	} else if kind != "post" {
		return core.NotFound()
	}
	n, err := r.store.Exec(ctx, "UPDATE "+table+" SET status=?,updated_at=? WHERE id=? AND status IN('published','hidden','removed')", status, now, id)
	if err != nil || n > 0 {
		return err
	}
	var row struct{ Status string }
	if err = r.store.Query(ctx, &row, "SELECT status FROM "+table+" WHERE id=? FOR UPDATE", id); err != nil {
		return err
	}
	if row.Status == "published" || row.Status == "hidden" || row.Status == "removed" {
		return nil
	}
	if row.Status == "" {
		return core.NotFound()
	}
	return core.Conflict("Conteúdo excluído ou rascunho não pode ser alterado por moderação.")
}
func (r *AdminRepository) NotifyModeration(ctx context.Context, recipient, actor uint64, kind string, target uint64, key string, now time.Time) error {
	var exists int64
	if err := r.store.Query(ctx, &exists, "SELECT COUNT(*) FROM readers WHERE user_id=?", recipient); err != nil {
		return err
	}
	if exists == 0 {
		return nil
	}
	var postID, commentID *uint64
	if kind == "post" {
		postID = &target
	}
	if kind == "comment" {
		commentID = &target
		var parent uint64
		if e := r.store.Query(ctx, &parent, "SELECT post_id FROM comments WHERE id=?", target); e != nil {
			return e
		}
		if parent != 0 {
			postID = &parent
		}
	}
	_, err := r.store.Exec(ctx, "INSERT INTO notifications(recipient_id,actor_id,event,post_id,comment_id,event_key,created_at) VALUES(?,?,'moderation',?,?,?,?) ON DUPLICATE KEY UPDATE event_key=VALUES(event_key)", recipient, actor, postID, commentID, key, now)
	return err
}
func (r *AdminRepository) CreatePenalty(ctx context.Context, v core.Penalty) (uint64, error) {
	return r.store.Insert(ctx, "penalties", map[string]any{"user_id": v.UserID, "applied_by": v.AppliedBy, "type": v.Type, "reason": v.Reason, "starts_at": v.StartsAt, "ends_at": v.EndsAt, "created_at": v.CreatedAt})
}
func (r *AdminRepository) Penalty(ctx context.Context, id uint64, lock bool) (core.Penalty, error) {
	var v core.Penalty
	err := r.store.Query(ctx, &v, "SELECT * FROM penalties WHERE id=?"+adminLock(lock), id)
	if err == nil && v.ID == 0 {
		err = core.NotFound()
	}
	return v, err
}
func (r *AdminRepository) Penalties(ctx context.Context, user *uint64, p schemas.Page) ([]core.Penalty, int64, error) {
	where := "1=1"
	args := []any{}
	if user != nil {
		where = "user_id=?"
		args = append(args, *user)
	}
	var n int64
	if err := r.store.Query(ctx, &n, "SELECT COUNT(*) FROM penalties WHERE "+where, args...); err != nil {
		return nil, 0, err
	}
	var rows []core.Penalty
	args = append(args, p.PageSize, p.Offset())
	err := r.store.Query(ctx, &rows, "SELECT * FROM penalties WHERE "+where+" ORDER BY created_at DESC,id DESC LIMIT ? OFFSET ?", args...)
	return rows, n, err
}
func (r *AdminRepository) LinkPenaltyReports(ctx context.Context, id uint64, reports []uint64) error {
	for _, rid := range reports {
		if _, err := r.store.Exec(ctx, "INSERT INTO report_penalties(report_id,penalty_id) VALUES(?,?)", rid, id); err != nil {
			return err
		}
	}
	return nil
}
func (r *AdminRepository) RevokePenalty(ctx context.Context, id, actor uint64, reason string, now time.Time) error {
	_, err := r.store.Exec(ctx, "UPDATE penalties SET revoked_by=?,revoked_at=?,revocation_reason=? WHERE id=? AND revoked_at IS NULL", actor, now, reason, id)
	return err
}
func (r *AdminRepository) FAQ(ctx context.Context, id uint64, admin bool) (core.FAQEntry, error) {
	where := "id=?"
	if !admin {
		where += " AND status='published'"
	}
	var v core.FAQEntry
	err := r.store.Query(ctx, &v, "SELECT * FROM faq_entries WHERE "+where, id)
	if err == nil && v.ID == 0 {
		err = core.NotFound()
	}
	return v, err
}
func (r *AdminRepository) FAQs(ctx context.Context, admin bool, p schemas.Page) ([]core.FAQEntry, int64, error) {
	where := "1=1"
	if !admin {
		where = "status='published'"
	}
	var n int64
	if err := r.store.Query(ctx, &n, "SELECT COUNT(*) FROM faq_entries WHERE "+where); err != nil {
		return nil, 0, err
	}
	var rows []core.FAQEntry
	err := r.store.Query(ctx, &rows, "SELECT * FROM faq_entries WHERE "+where+" ORDER BY sort_order,id LIMIT ? OFFSET ?", p.PageSize, p.Offset())
	return rows, n, err
}
func (r *AdminRepository) SaveFAQ(ctx context.Context, v core.FAQEntry, create bool) (uint64, error) {
	if create {
		return r.store.Insert(ctx, "faq_entries", map[string]any{"question": v.Question, "answer": v.Answer, "sort_order": v.SortOrder, "status": v.Status, "created_at": v.CreatedAt, "updated_at": v.UpdatedAt})
	}
	_, err := r.store.Exec(ctx, "UPDATE faq_entries SET question=?,answer=?,sort_order=?,status=?,updated_at=? WHERE id=?", v.Question, v.Answer, v.SortOrder, v.Status, v.UpdatedAt, v.ID)
	return v.ID, err
}
func (r *AdminRepository) DeleteFAQ(ctx context.Context, id uint64) error {
	_, err := r.store.Exec(ctx, "DELETE FROM faq_entries WHERE id=?", id)
	return err
}
func (r *AdminRepository) CreateTicket(ctx context.Context, v core.SupportTicket) (uint64, error) {
	return r.store.Insert(ctx, "support_tickets", map[string]any{"requester_id": v.RequesterID, "subject": v.Subject, "status": v.Status, "created_at": v.CreatedAt, "updated_at": v.UpdatedAt})
}
func (r *AdminRepository) Ticket(ctx context.Context, id uint64, lock bool) (core.SupportTicket, error) {
	var v core.SupportTicket
	err := r.store.Query(ctx, &v, "SELECT * FROM support_tickets WHERE id=?"+adminLock(lock), id)
	if err == nil && v.ID == 0 {
		err = core.NotFound()
	}
	return v, err
}
func (r *AdminRepository) Tickets(ctx context.Context, owner *uint64, status string, p schemas.Page) ([]core.SupportTicket, int64, error) {
	where := "1=1"
	args := []any{}
	if owner != nil {
		where += " AND requester_id=?"
		args = append(args, *owner)
	}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var n int64
	if err := r.store.Query(ctx, &n, "SELECT COUNT(*) FROM support_tickets WHERE "+where, args...); err != nil {
		return nil, 0, err
	}
	var rows []core.SupportTicket
	args = append(args, p.PageSize, p.Offset())
	err := r.store.Query(ctx, &rows, "SELECT * FROM support_tickets WHERE "+where+" ORDER BY created_at DESC,id DESC LIMIT ? OFFSET ?", args...)
	return rows, n, err
}
func (r *AdminRepository) SaveTicket(ctx context.Context, v core.SupportTicket) error {
	_, err := r.store.Exec(ctx, "UPDATE support_tickets SET status=?,assigned_to=?,updated_at=? WHERE id=?", v.Status, v.AssignedTo, v.UpdatedAt, v.ID)
	return err
}
func (r *AdminRepository) TicketMessages(ctx context.Context, id uint64, p schemas.Page) ([]core.SupportMessage, int64, error) {
	var n int64
	if err := r.store.Query(ctx, &n, "SELECT COUNT(*) FROM support_messages WHERE ticket_id=?", id); err != nil {
		return nil, 0, err
	}
	var rows []core.SupportMessage
	err := r.store.Query(ctx, &rows, "SELECT * FROM support_messages WHERE ticket_id=? ORDER BY id LIMIT ? OFFSET ?", id, p.PageSize, p.Offset())
	return rows, n, err
}
func (r *AdminRepository) CreateTicketMessage(ctx context.Context, v core.SupportMessage) (uint64, error) {
	return r.store.Insert(ctx, "support_messages", map[string]any{"ticket_id": v.TicketID, "author_id": v.AuthorID, "text": v.Text, "created_at": v.CreatedAt})
}
func (r *AdminRepository) Audit(ctx context.Context, actor uint64, action, kind string, target uint64, reason string, now time.Time) error {
	_, err := r.store.Insert(ctx, "audit_logs", map[string]any{"actor_id": actor, "action": action, "target_type": kind, "target_id": target, "reason": reason, "created_at": now})
	return err
}
func (r *AdminRepository) AuditLogs(ctx context.Context, kind string, p schemas.Page) ([]core.AuditLog, int64, error) {
	where := "1=1"
	args := []any{}
	if kind != "" {
		where = "target_type=?"
		args = append(args, kind)
	}
	var n int64
	if err := r.store.Query(ctx, &n, "SELECT COUNT(*) FROM audit_logs WHERE "+where, args...); err != nil {
		return nil, 0, err
	}
	var rows []core.AuditLog
	args = append(args, p.PageSize, p.Offset())
	err := r.store.Query(ctx, &rows, "SELECT * FROM audit_logs WHERE "+where+" ORDER BY created_at DESC,id DESC LIMIT ? OFFSET ?", args...)
	return rows, n, err
}
