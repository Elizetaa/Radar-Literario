package tools

import (
	"context"
	"fmt"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"sort"
	"strings"
	"time"
)

type accountRepository struct{ s interfaces.Store }

func NewAccountRepository(s interfaces.Store) interfaces.AccountRepository {
	return &accountRepository{s: s}
}
func (r *accountRepository) Within(ctx context.Context, fn func(interfaces.AccountRepository) error) error {
	return r.s.Transaction(ctx, func(s interfaces.Store) error { return fn(NewAccountRepository(s)) })
}

const accountSelect = `SELECT u.*,r.biography,r.profile_visibility,r.sex,r.country_of_origin,r.favorite_work_id,COALESCE(a.role,'') AS role,(r.user_id IS NOT NULL) AS is_reader FROM users u LEFT JOIN readers r ON r.user_id=u.id LEFT JOIN administrators a ON a.user_id=u.id AND a.status='active' `

func (r *accountRepository) Account(ctx context.Context, id uint64, lock bool) (core.Account, error) {
	var a core.Account
	q := accountSelect + "WHERE u.id=?"
	if lock {
		q += " FOR UPDATE"
	}
	e := r.s.Query(ctx, &a, q, id)
	if e == nil && a.ID == 0 {
		e = core.NotFound()
	}
	return a, e
}
func (r *accountRepository) ByEmail(ctx context.Context, email string) (core.Account, error) {
	var a core.Account
	e := r.s.Query(ctx, &a, accountSelect+"WHERE u.email=?", email)
	if e == nil && a.ID == 0 {
		e = core.NotFound()
	}
	return a, e
}
func (r *accountRepository) Create(ctx context.Context, u core.User) (uint64, error) {
	id, e := r.s.Insert(ctx, "users", map[string]any{"name": u.Name, "nickname": u.Nickname, "email": u.Email, "password_hash": u.PasswordHash, "birth_date": u.BirthDate, "status": "active", "created_at": u.CreatedAt, "updated_at": u.UpdatedAt})
	if e != nil {
		return 0, e
	}
	_, e = r.s.Exec(ctx, "INSERT INTO readers(user_id) VALUES (?)", id)
	return id, e
}
func updateColumns(ctx context.Context, s interfaces.Store, table, key string, id uint64, values map[string]any) error {
	if len(values) == 0 {
		return nil
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		if !identifier.MatchString(k) {
			return fmt.Errorf("invalid update column")
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sets []string
	var args []any
	for _, k := range keys {
		sets = append(sets, "`"+k+"`=?")
		args = append(args, values[k])
	}
	args = append(args, id)
	_, e := s.Exec(ctx, "UPDATE `"+table+"` SET "+strings.Join(sets, ",")+" WHERE `"+key+"`=?", args...)
	return e
}
func (r *accountRepository) UpdateAccount(ctx context.Context, id uint64, user, reader map[string]any) error {
	if _, changed := user["email"]; changed {
		var previousEmail string
		if e := r.s.Query(ctx, &previousEmail, "SELECT email FROM users WHERE id=? FOR UPDATE", id); e != nil {
			return e
		}
		if _, e := r.s.Exec(ctx, "DELETE FROM mail_outbox WHERE recipient=?", previousEmail); e != nil {
			return e
		}
		if _, e := r.s.Exec(ctx, "DELETE FROM password_resets WHERE user_id=?", id); e != nil {
			return e
		}
	}
	if len(user) > 0 {
		user["updated_at"] = time.Now().UTC()
	}
	if e := updateColumns(ctx, r.s, "users", "id", id, user); e != nil {
		return e
	}
	return updateColumns(ctx, r.s, "readers", "user_id", id, reader)
}
func (r *accountRepository) WorkEligible(ctx context.Context, id uint64, age int) error {
	var n int64
	e := r.s.Query(ctx, &n, "SELECT COUNT(*) FROM works WHERE id=? AND status='published' AND minimum_age<=?", id, age)
	if e == nil && n == 0 {
		return core.NotFound()
	}
	return e
}
func (r *accountRepository) Session(ctx context.Context, hash string) (core.Session, error) {
	var s core.Session
	e := r.s.Query(ctx, &s, "SELECT * FROM sessions WHERE token_hash=?", hash)
	if e == nil && s.ID == 0 {
		e = core.ErrUnauthenticated
	}
	return s, e
}
func (r *accountRepository) AddSession(ctx context.Context, s core.Session) error {
	_, e := r.s.Insert(ctx, "sessions", map[string]any{"user_id": s.UserID, "token_hash": s.TokenHash, "csrf_hash": s.CSRFHash, "created_at": s.CreatedAt, "last_seen_at": s.LastSeenAt, "expires_at": s.ExpiresAt})
	return e
}
func (r *accountRepository) TouchSession(ctx context.Context, id uint64, now, idleCutoff time.Time) (bool, error) {
	n, e := r.s.Exec(ctx, "UPDATE sessions SET last_seen_at=GREATEST(last_seen_at,?) WHERE id=? AND revoked_at IS NULL AND expires_at>? AND last_seen_at>?", now, id, now, idleCutoff)
	if e != nil {
		return false, e
	}
	if n > 0 {
		return true, nil
	}
	var count int64
	e = r.s.Query(ctx, &count, "SELECT COUNT(*) FROM sessions WHERE id=? AND revoked_at IS NULL AND expires_at>? AND last_seen_at>?", id, now, idleCutoff)
	return count == 1, e
}
func (r *accountRepository) Revoke(ctx context.Context, user uint64, token string, now time.Time) error {
	q := "UPDATE sessions SET revoked_at=? WHERE revoked_at IS NULL AND user_id=?"
	args := []any{now, user}
	if token != "" {
		q += " AND token_hash=?"
		args = append(args, token)
	}
	_, e := r.s.Exec(ctx, q, args...)
	return e
}
func (r *accountRepository) RevokeToken(ctx context.Context, token string, now time.Time) error {
	_, e := r.s.Exec(ctx, "UPDATE sessions SET revoked_at=? WHERE token_hash=? AND revoked_at IS NULL", now, token)
	return e
}
func (r *accountRepository) Penalties(ctx context.Context, id uint64, now time.Time) (bool, bool, error) {
	var p struct{ Blocked, Restricted bool }
	e := r.s.Query(ctx, &p, `SELECT COALESCE(MAX(type IN ('suspension','ban')),0) AS blocked,COALESCE(MAX(type='restriction'),0) AS restricted FROM penalties WHERE user_id=? AND revoked_at IS NULL AND starts_at<=? AND (ends_at IS NULL OR ends_at>?)`, id, now, now)
	return p.Blocked, p.Restricted, e
}
func (r *accountRepository) SaveReset(ctx context.Context, t core.Reset, recipient, payload string) error {
	if _, e := r.s.Exec(ctx, "UPDATE password_resets SET consumed_at=? WHERE user_id=? AND consumed_at IS NULL", t.CreatedAt, t.UserID); e != nil {
		return e
	}
	if _, e := r.s.Insert(ctx, "password_resets", map[string]any{"user_id": t.UserID, "token_hash": t.TokenHash, "expires_at": t.ExpiresAt, "created_at": t.CreatedAt}); e != nil {
		return e
	}
	_, e := r.s.Insert(ctx, "mail_outbox", map[string]any{"recipient": recipient, "payload": payload, "available_at": t.CreatedAt, "created_at": t.CreatedAt})
	return e
}
func (r *accountRepository) ConsumeReset(ctx context.Context, hash string, now time.Time) (uint64, error) {
	var t core.Reset
	if e := r.s.Query(ctx, &t, "SELECT * FROM password_resets WHERE token_hash=? FOR UPDATE", hash); e != nil {
		return 0, e
	}
	if t.ID == 0 || t.ConsumedAt != nil || !now.Before(t.ExpiresAt) {
		return 0, core.Invalid("INVALID_RESET_TOKEN", "O link de recuperação é inválido ou expirou.")
	}
	var status string
	if e := r.s.Query(ctx, &status, "SELECT status FROM users WHERE id=? FOR UPDATE", t.UserID); e != nil {
		return 0, e
	}
	if status != "active" {
		return 0, core.Invalid("INVALID_RESET_TOKEN", "O link de recuperação é inválido ou expirou.")
	}
	_, e := r.s.Exec(ctx, "UPDATE password_resets SET consumed_at=? WHERE user_id=? AND consumed_at IS NULL", now, t.UserID)
	return t.UserID, e
}
func (r *accountRepository) ChangeHash(ctx context.Context, id uint64, hash string) error {
	_, e := r.s.Exec(ctx, "UPDATE users SET password_hash=?,updated_at=? WHERE id=?", hash, time.Now().UTC(), id)
	return e
}
func (r *accountRepository) Anonymize(ctx context.Context, id uint64, now time.Time) error {
	var email string
	if e := r.s.Query(ctx, &email, "SELECT email FROM users WHERE id=? FOR UPDATE", id); e != nil {
		return e
	}
	var n int64
	if e := r.s.Query(ctx, &n, "SELECT COUNT(*) FROM administrators WHERE user_id=? AND status='active'", id); e != nil {
		return e
	}
	if n > 0 {
		return core.Conflict("Revogue o perfil administrativo antes de excluir a conta.")
	}
	statements := []string{"UPDATE posts SET status='deleted',updated_at=? WHERE reader_id=?", "UPDATE comments SET status='deleted',updated_at=? WHERE reader_id=?", "UPDATE users SET name='Conta excluída',nickname=CONCAT('~deleted_',id),email=CONCAT('~deleted_',id,'@invalid.test'),password_hash='',birth_date=NULL,profile_image_url='',email_verified_at=NULL,status='deactivated',deleted_at=?,updated_at=? WHERE id=?"}
	for i, q := range statements {
		var e error
		if i == 2 {
			_, e = r.s.Exec(ctx, q, now, now, id)
		} else {
			_, e = r.s.Exec(ctx, q, now, id)
		}
		if e != nil {
			return e
		}
	}
	for _, q := range []string{"DELETE FROM post_likes WHERE reader_id=?", "DELETE FROM comment_likes WHERE reader_id=?", "DELETE FROM notifications WHERE recipient_id=?", "DELETE FROM reading_records WHERE reader_id=?", "DELETE FROM reader_genres WHERE reader_id=?", "DELETE FROM reader_preferred_works WHERE reader_id=?", "UPDATE readers SET biography='',sex=NULL,country_of_origin=NULL,favorite_work_id=NULL,profile_visibility='private' WHERE user_id=?", "DELETE FROM password_resets WHERE user_id=?"} {
		if _, e := r.s.Exec(ctx, q, id); e != nil {
			return e
		}
	}
	if _, e := r.s.Exec(ctx, "DELETE FROM mail_outbox WHERE recipient=?", email); e != nil {
		return e
	}
	if _, e := r.s.Exec(ctx, "UPDATE support_messages m JOIN support_tickets t ON t.id=m.ticket_id SET m.text='' WHERE t.requester_id=?", id); e != nil {
		return e
	}
	if _, e := r.s.Exec(ctx, "UPDATE support_tickets SET subject='Conta excluída',status='closed',updated_at=? WHERE requester_id=?", now, id); e != nil {
		return e
	}
	return r.Revoke(ctx, id, "", now)
}
func (r *accountRepository) Bootstrap(ctx context.Context, u core.User, allowWeak bool) (uint64, error) {
	var rows []struct{ UserID uint64 }
	if e := r.s.Query(ctx, &rows, "SELECT user_id FROM administrators WHERE role='system_administrator' AND status='active' FOR UPDATE"); e != nil {
		return 0, e
	}
	if len(rows) > 0 {
		a, e := r.ByEmail(ctx, u.Email)
		if e == nil && a.Role == "system_administrator" && a.Nickname == u.Nickname {
			return a.ID, nil
		}
		return 0, core.Conflict("O primeiro administrador já foi criado; use a gestão autenticada.")
	}
	id, e := r.Create(ctx, u)
	if e != nil {
		return 0, e
	}
	if _, e = r.s.Exec(ctx, "INSERT INTO administrators(user_id,role,status) VALUES (?,'system_administrator','active')", id); e != nil {
		return 0, e
	}
	_, e = r.s.Exec(ctx, "INSERT INTO audit_logs(actor_id,target_type,target_id,action,reason) VALUES (?,'administrator',?,'bootstrap',?)", id, id, fmt.Sprintf("Criação inicial controlada; exceção de senha local: %t", allowWeak))
	return id, e
}
func (r *accountRepository) PendingMail(ctx context.Context, now time.Time) ([]core.Outbox, error) {
	var rows []core.Outbox
	e := r.s.Query(ctx, &rows, "SELECT id,recipient,payload,attempts,available_at FROM mail_outbox WHERE delivered_at IS NULL AND available_at<=? AND (claimed_at IS NULL OR claimed_at<?) AND attempts<10 ORDER BY id LIMIT 10", now, now.Add(-time.Minute))
	return rows, e
}
func (r *accountRepository) ClaimMail(ctx context.Context, id uint64, now time.Time) (bool, error) {
	n, e := r.s.Exec(ctx, "UPDATE mail_outbox SET claimed_at=?,attempts=attempts+1 WHERE id=? AND delivered_at IS NULL AND (claimed_at IS NULL OR claimed_at<?)", now, id, now.Add(-time.Minute))
	return n == 1, e
}
func (r *accountRepository) FinishMail(ctx context.Context, id uint64, ok bool, now time.Time) error {
	if ok {
		_, e := r.s.Exec(ctx, "UPDATE mail_outbox SET delivered_at=?,payload='',recipient='',claimed_at=NULL WHERE id=?", now, id)
		return e
	}
	_, e := r.s.Exec(ctx, "UPDATE mail_outbox SET available_at=?,claimed_at=NULL WHERE id=?", now.Add(5*time.Minute), id)
	return e
}
func (r *accountRepository) Purge(ctx context.Context, before time.Time) error {
	for _, q := range []string{"UPDATE posts SET text='Conteúdo removido',title='',data_url=NULL WHERE status IN ('deleted','removed') AND updated_at<?", "UPDATE comments SET text='' WHERE status IN ('deleted','removed') AND updated_at<?", "UPDATE report_messages m JOIN reports r ON r.id=m.report_id SET m.text='' WHERE r.status IN ('resolved','dismissed','canceled') AND r.updated_at<?", "UPDATE reports SET description='',report_data_url='',resolution_note='' WHERE status IN ('resolved','dismissed','canceled') AND updated_at<?", "UPDATE audit_logs SET reason='' WHERE created_at<?", "UPDATE penalties SET reason='',revocation_reason=NULL WHERE created_at<?", "UPDATE support_messages m JOIN support_tickets t ON t.id=m.ticket_id SET m.text='' WHERE t.status='closed' AND t.updated_at<?", "UPDATE support_tickets SET subject='Atendimento encerrado' WHERE status='closed' AND updated_at<?", "UPDATE catalog_suggestions SET proposal=JSON_OBJECT(),decision_reason='' WHERE status IN ('approved','rejected') AND updated_at<?", "DELETE FROM sessions WHERE expires_at<?", "DELETE FROM password_resets WHERE expires_at<?", "DELETE FROM mail_outbox WHERE created_at<?"} {
		if _, e := r.s.Exec(ctx, q, before); e != nil {
			return e
		}
	}
	return nil
}
