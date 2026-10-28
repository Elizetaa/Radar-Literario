package tools

import (
	"context"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/schemas"
	"time"
)

type readingRepository struct{ store interfaces.Store }

func NewReadingRepository(store interfaces.Store) interfaces.ReadingRepository {
	return &readingRepository{store: store}
}
func (r *readingRepository) Within(ctx context.Context, fn func(interfaces.ReadingRepository) error) error {
	return r.store.Transaction(ctx, func(s interfaces.Store) error { return fn(NewReadingRepository(s)) })
}
func (r *readingRepository) LockReader(ctx context.Context, id uint64) error {
	var row struct{ UserID uint64 }
	e := r.store.Query(ctx, &row, "SELECT user_id FROM readers WHERE user_id=? FOR UPDATE", id)
	if e == nil && row.UserID == 0 {
		return core.NotFound()
	}
	return e
}
func (r *readingRepository) Edition(ctx context.Context, id uint64) (core.ReadingEdition, error) {
	var row core.ReadingEdition
	e := r.store.Query(ctx, &row, "SELECT e.id,e.work_id,e.page_count,GREATEST(e.minimum_age,w.minimum_age) minimum_age,e.status,w.status work_status FROM editions e JOIN works w ON w.id=e.work_id WHERE e.id=? FOR UPDATE", id)
	if e == nil && row.ID == 0 {
		e = core.NotFound()
	}
	return row, e
}
func (r *readingRepository) Record(ctx context.Context, readerID, workID uint64) (*core.ReadingRecord, error) {
	var row core.ReadingRecord
	e := r.store.Query(ctx, &row, "SELECT * FROM reading_records WHERE reader_id=? AND work_id=? FOR UPDATE", readerID, workID)
	if e != nil {
		return nil, e
	}
	if row.ID == 0 {
		return nil, nil
	}
	return &row, nil
}
func (r *readingRepository) SaveRecord(ctx context.Context, v *core.ReadingRecord) error {
	if v.ID == 0 {
		id, e := r.store.Insert(ctx, "reading_records", map[string]any{"reader_id": v.ReaderID, "work_id": v.WorkID, "edition_id": v.EditionID, "current_page": v.CurrentPage, "credited_peak": v.CreditedPeak, "status": v.Status, "started_at": v.StartedAt, "finished_at": v.FinishedAt, "created_at": v.CreatedAt, "updated_at": v.UpdatedAt})
		v.ID = id
		return e
	}
	_, e := r.store.Exec(ctx, "UPDATE reading_records SET edition_id=?,current_page=?,credited_peak=?,status=?,started_at=?,finished_at=?,updated_at=? WHERE id=? AND reader_id=?", v.EditionID, v.CurrentPage, v.CreditedPeak, v.Status, v.StartedAt, v.FinishedAt, v.UpdatedAt, v.ID, v.ReaderID)
	return e
}
func (r *readingRepository) RemoveRecord(ctx context.Context, readerID, workID uint64) error {
	_, e := r.store.Exec(ctx, "DELETE FROM reading_records WHERE reader_id=? AND work_id=?", readerID, workID)
	return e
}
func (r *readingRepository) AddProgress(ctx context.Context, v core.ProgressEvent) error {
	_, e := r.store.Insert(ctx, "reading_progress_events", map[string]any{"record_id": v.RecordID, "reader_id": v.ReaderID, "previous_page": v.PreviousPage, "current_page": v.CurrentPage, "credited_pages": v.CreditedPages, "change_kind": v.ChangeKind, "created_at": v.CreatedAt})
	return e
}
func (r *readingRepository) Shelf(ctx context.Context, readerID uint64, age int, status string, page schemas.Page) ([]core.ReadingRecord, int64, error) {
	from := " FROM reading_records rr JOIN works w ON w.id=rr.work_id JOIN editions e ON e.id=rr.edition_id WHERE rr.reader_id=? AND w.status='published' AND e.status='published' AND GREATEST(w.minimum_age,e.minimum_age)<=?"
	args := []any{readerID, age}
	switch status {
	case "incomplete":
		from += " AND rr.status IN('in_progress','paused')"
	case "preferred":
		from += " AND EXISTS(SELECT 1 FROM reader_preferred_works rp WHERE rp.reader_id=rr.reader_id AND rp.work_id=rr.work_id)"
	case "":
	default:
		from += " AND rr.status=?"
		args = append(args, status)
	}
	var total int64
	if e := r.store.Query(ctx, &total, "SELECT COUNT(*)"+from, args...); e != nil {
		return nil, 0, e
	}
	rows := []core.ReadingRecord{}
	e := r.store.Query(ctx, &rows, "SELECT rr.*,w.title work_title,e.title edition_title,e.language"+from+" ORDER BY rr.updated_at DESC,rr.id DESC LIMIT ? OFFSET ?", append(args, page.PageSize, page.Offset())...)
	return rows, total, e
}
func (r *readingRepository) ProfileVisible(ctx context.Context, id uint64, actor core.Actor) error {
	var row struct {
		Visibility string
		Status     string
	}
	e := r.store.Query(ctx, &row, "SELECT r.profile_visibility visibility,u.status FROM readers r JOIN users u ON u.id=r.user_id WHERE r.user_id=?", id)
	if e != nil {
		return e
	}
	if row.Status != "active" {
		return core.NotFound()
	}
	if id == actor.ID {
		return nil
	}
	if row.Visibility == "public" || (row.Visibility == "registered_users_only" && actor.ID != 0) {
		return nil
	}
	return core.NotFound()
}

const readingPostSelect = `SELECT p.*,w.minimum_age work_minimum_age,w.status work_status,u.status author_status,u.name author_name,u.nickname author_nickname,d.edition_id,d.current_page,d.occurred_at,v.rating,
 (SELECT COUNT(*) FROM post_likes pl JOIN users lu ON lu.id=pl.reader_id WHERE pl.post_id=p.id AND lu.status='active') like_count,
 (SELECT COUNT(*) FROM comments c JOIN users cu ON cu.id=c.reader_id WHERE c.post_id=p.id AND c.status='published' AND cu.status='active') comment_count,
 EXISTS(SELECT 1 FROM post_likes pl WHERE pl.post_id=p.id AND pl.reader_id=?) liked
 FROM posts p JOIN works w ON w.id=p.work_id JOIN users u ON u.id=p.reader_id LEFT JOIN reading_diaries d ON d.post_id=p.id LEFT JOIN reviews v ON v.post_id=p.id `

func (r *readingRepository) Post(ctx context.Context, id, viewer uint64) (core.ReadingPost, error) {
	var row core.ReadingPost
	e := r.store.Query(ctx, &row, readingPostSelect+" WHERE p.id=?", viewer, id)
	if e == nil && row.ID == 0 {
		e = core.NotFound()
	}
	return row, e
}
func (r *readingRepository) Posts(ctx context.Context, actor core.Actor, f schemas.PostFilter) ([]core.ReadingPost, int64, error) {
	where := " WHERE u.status='active' AND w.status='published' AND GREATEST(w.minimum_age,p.minimum_age)<=?"
	args := []any{actor.Age(time.Now())}
	if f.Own {
		where += " AND p.reader_id=? AND p.status IN('draft','published')"
		args = append(args, actor.ID)
	} else {
		where += " AND p.status='published'"
	}
	if f.Kind != "" {
		where += " AND p.kind=?"
		args = append(args, f.Kind)
	}
	if f.ReaderID != 0 {
		where += " AND p.reader_id=?"
		args = append(args, f.ReaderID)
	}
	if f.WorkID != 0 {
		where += " AND p.work_id=?"
		args = append(args, f.WorkID)
	}
	var total int64
	if e := r.store.Query(ctx, &total, "SELECT COUNT(*) FROM posts p JOIN works w ON w.id=p.work_id JOIN users u ON u.id=p.reader_id"+where, args...); e != nil {
		return nil, 0, e
	}
	rows := []core.ReadingPost{}
	queryArgs := append([]any{actor.ID}, args...)
	queryArgs = append(queryArgs, f.Page.PageSize, f.Page.Offset())
	e := r.store.Query(ctx, &rows, readingPostSelect+where+" ORDER BY p.created_at DESC,p.id DESC LIMIT ? OFFSET ?", queryArgs...)
	return rows, total, e
}
func (r *readingRepository) CreatePost(ctx context.Context, p *core.ReadingPost) error {
	id, e := r.store.Insert(ctx, "posts", map[string]any{"reader_id": p.ReaderID, "work_id": p.WorkID, "kind": p.Kind, "title": p.Title, "text": p.Text, "data_url": nilIfEmpty(p.DataURL), "status": p.Status, "contains_spoiler": p.ContainsSpoiler, "contains_sensitive_content": p.ContainsSensitiveContent, "minimum_age": p.MinimumAge, "published_at": p.PublishedAt, "effects_applied_at": p.EffectsAppliedAt, "created_at": p.CreatedAt, "updated_at": p.UpdatedAt})
	p.ID = id
	return e
}
func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
func (r *readingRepository) SavePost(ctx context.Context, p core.ReadingPost) error {
	n, e := r.store.Exec(ctx, "UPDATE posts SET title=?,text=?,data_url=?,status=?,contains_spoiler=?,contains_sensitive_content=?,minimum_age=?,published_at=?,effects_applied_at=?,updated_at=? WHERE id=? AND reader_id=? AND status IN('draft','published')", p.Title, p.Text, nilIfEmpty(p.DataURL), p.Status, p.ContainsSpoiler, p.ContainsSensitiveContent, p.MinimumAge, p.PublishedAt, p.EffectsAppliedAt, p.UpdatedAt, p.ID, p.ReaderID)
	if e != nil || n > 0 {
		return e
	}
	// A second identical save can affect zero rows. Use a current locking read,
	// rather than the transaction snapshot, to distinguish it from moderation.
	return r.ensureWritableState(ctx, "posts", p.ID, p.ReaderID)
}

func (r *readingRepository) ensureWritableState(ctx context.Context, table string, id, readerID uint64) error {
	var row struct{ Status string }
	if e := r.store.Query(ctx, &row, "SELECT status FROM "+table+" WHERE id=? AND reader_id=? FOR UPDATE", id, readerID); e != nil {
		return e
	}
	if row.Status == "published" || (table == "posts" && row.Status == "draft") {
		return nil
	}
	if row.Status == "" {
		return core.NotFound()
	}
	return core.Conflict("O conteúdo mudou de estado durante a operação. Atualize antes de tentar novamente.")
}
func (r *readingRepository) CreateDiary(ctx context.Context, p core.ReadingPost) error {
	_, e := r.store.Insert(ctx, "reading_diaries", map[string]any{"post_id": p.ID, "kind": "diary", "edition_id": p.EditionID, "current_page": p.CurrentPage, "occurred_at": p.OccurredAt})
	return e
}
func (r *readingRepository) SaveDiary(ctx context.Context, p core.ReadingPost) error {
	_, e := r.store.Exec(ctx, "UPDATE reading_diaries SET current_page=? WHERE post_id=?", p.CurrentPage, p.ID)
	return e
}
func (r *readingRepository) CreateReview(ctx context.Context, p core.ReadingPost) error {
	_, e := r.store.Insert(ctx, "reviews", map[string]any{"post_id": p.ID, "kind": "review", "reader_id": p.ReaderID, "work_id": p.WorkID, "rating": p.Rating, "active_slot": 1})
	return e
}
func (r *readingRepository) SaveReview(ctx context.Context, p core.ReadingPost) error {
	_, e := r.store.Exec(ctx, "UPDATE reviews SET rating=? WHERE post_id=?", p.Rating, p.ID)
	return e
}
func (r *readingRepository) ReleaseReview(ctx context.Context, id uint64) error {
	_, e := r.store.Exec(ctx, "UPDATE reviews SET active_slot=NULL WHERE post_id=?", id)
	return e
}

const readingCommentSelect = `SELECT c.*,u.name author_name,u.nickname author_nickname,u.status author_status,
 (SELECT COUNT(*) FROM comment_likes cl JOIN users lu ON lu.id=cl.reader_id WHERE cl.comment_id=c.id AND lu.status='active') like_count,
 EXISTS(SELECT 1 FROM comment_likes cl WHERE cl.comment_id=c.id AND cl.reader_id=?) liked
 FROM comments c JOIN users u ON u.id=c.reader_id `

func (r *readingRepository) Comment(ctx context.Context, id, viewer uint64) (core.ReadingComment, error) {
	var row core.ReadingComment
	e := r.store.Query(ctx, &row, readingCommentSelect+" WHERE c.id=?", viewer, id)
	if e == nil && row.ID == 0 {
		e = core.NotFound()
	}
	return row, e
}
func (r *readingRepository) Comments(ctx context.Context, postID, viewer uint64, parent *uint64, page schemas.Page) ([]core.ReadingComment, int64, error) {
	where := " WHERE c.post_id=? AND (c.status IN('published','deleted') OR EXISTS(SELECT 1 FROM comments child WHERE child.parent_id=c.id))"
	args := []any{postID}
	if parent == nil {
		where += " AND c.parent_id IS NULL"
	} else {
		where += " AND c.parent_id=?"
		args = append(args, *parent)
	}
	var total int64
	if e := r.store.Query(ctx, &total, "SELECT COUNT(*) FROM comments c"+where, args...); e != nil {
		return nil, 0, e
	}
	rows := []core.ReadingComment{}
	queryArgs := append([]any{viewer}, args...)
	queryArgs = append(queryArgs, page.PageSize, page.Offset())
	e := r.store.Query(ctx, &rows, readingCommentSelect+where+" ORDER BY c.created_at ASC,c.id ASC LIMIT ? OFFSET ?", queryArgs...)
	return rows, total, e
}
func (r *readingRepository) CreateComment(ctx context.Context, c *core.ReadingComment) error {
	id, e := r.store.Insert(ctx, "comments", map[string]any{"post_id": c.PostID, "reader_id": c.ReaderID, "parent_id": c.ParentID, "depth": c.Depth, "text": c.Text, "status": c.Status, "created_at": c.CreatedAt, "updated_at": c.UpdatedAt})
	c.ID = id
	return e
}
func (r *readingRepository) SaveComment(ctx context.Context, c core.ReadingComment) error {
	n, e := r.store.Exec(ctx, "UPDATE comments SET text=?,status=?,updated_at=? WHERE id=? AND reader_id=? AND status='published'", c.Text, c.Status, c.UpdatedAt, c.ID, c.ReaderID)
	if e != nil || n > 0 {
		return e
	}
	return r.ensureWritableState(ctx, "comments", c.ID, c.ReaderID)
}
func (r *readingRepository) Like(ctx context.Context, target string, targetID, readerID uint64, add bool) (bool, error) {
	table, column := "post_likes", "post_id"
	if target == "comment" {
		table, column = "comment_likes", "comment_id"
	} else if target != "post" {
		return false, core.Invalid("INVALID_TARGET", "Alvo de curtida inválido.")
	}
	if !add {
		n, e := r.store.Exec(ctx, "DELETE FROM "+table+" WHERE reader_id=? AND "+column+"=?", readerID, targetID)
		return n > 0, e
	}
	var existing uint64
	e := r.store.Query(ctx, &existing, "SELECT id FROM "+table+" WHERE reader_id=? AND "+column+"=?", readerID, targetID)
	if e != nil || existing != 0 {
		return false, e
	}
	_, e = r.store.Insert(ctx, table, map[string]any{"reader_id": readerID, column: targetID, "created_at": time.Now().UTC()})
	return e == nil, e
}
func (r *readingRepository) Notify(ctx context.Context, n core.ReadingNotification) error {
	_, e := r.store.Exec(ctx, "INSERT INTO notifications(recipient_id,actor_id,event,post_id,comment_id,event_key,created_at) VALUES(?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE event_key=VALUES(event_key)", n.RecipientID, n.ActorID, n.Event, n.PostID, n.CommentID, n.EventKey, n.CreatedAt)
	return e
}
func (r *readingRepository) Notifications(ctx context.Context, a core.Actor, unread *bool, page schemas.Page) ([]core.ReadingNotification, int64, error) {
	where := " WHERE n.recipient_id=?"
	args := []any{a.ID}
	if unread != nil {
		if *unread {
			where += " AND n.read_at IS NULL"
		} else {
			where += " AND n.read_at IS NOT NULL"
		}
	}
	var total int64
	if e := r.store.Query(ctx, &total, "SELECT COUNT(*) FROM notifications n"+where, args...); e != nil {
		return nil, 0, e
	}
	rows := []core.ReadingNotification{}
	queryArgs := append([]any{a.Age(time.Now())}, args...)
	queryArgs = append(queryArgs, page.PageSize, page.Offset())
	e := r.store.Query(ctx, &rows, `SELECT n.*, (p.status='published' AND u.status='active' AND w.status='published' AND GREATEST(w.minimum_age,p.minimum_age)<=? AND (n.comment_id IS NULL OR (c.status='published' AND cu.status='active'))) visible FROM notifications n LEFT JOIN posts p ON p.id=n.post_id LEFT JOIN works w ON w.id=p.work_id LEFT JOIN users u ON u.id=p.reader_id LEFT JOIN comments c ON c.id=n.comment_id LEFT JOIN users cu ON cu.id=c.reader_id`+where+" ORDER BY n.created_at DESC,n.id DESC LIMIT ? OFFSET ?", queryArgs...)
	return rows, total, e
}
func (r *readingRepository) ReadNotification(ctx context.Context, id, reader uint64, now time.Time) error {
	var own uint64
	e := r.store.Query(ctx, &own, "SELECT id FROM notifications WHERE id=? AND recipient_id=?", id, reader)
	if e != nil {
		return e
	}
	if own == 0 {
		return core.NotFound()
	}
	_, e = r.store.Exec(ctx, "UPDATE notifications SET read_at=COALESCE(read_at,?) WHERE id=? AND recipient_id=?", now, id, reader)
	return e
}
func (r *readingRepository) Statistics(ctx context.Context, id uint64, year int) (core.ReadingStatistics, error) {
	var out core.ReadingStatistics
	if e := r.store.Query(ctx, &out.CurrentPages, "SELECT COALESCE(SUM(current_page),0) FROM reading_records WHERE reader_id=?", id); e != nil {
		return out, e
	}
	out.Statuses = []core.StatusCount{}
	if e := r.store.Query(ctx, &out.Statuses, "SELECT status,COUNT(*) count FROM reading_records WHERE reader_id=? GROUP BY status", id); e != nil {
		return out, e
	}
	var avg struct{ Value *float64 }
	if e := r.store.Query(ctx, &avg, "SELECT AVG(v.rating) value FROM reviews v JOIN posts p ON p.id=v.post_id WHERE p.reader_id=? AND p.status='published'", id); e != nil {
		return out, e
	}
	out.AverageRating = avg.Value
	if e := r.store.Query(ctx, &out.DiaryCount, "SELECT COUNT(*) FROM posts WHERE reader_id=? AND kind='diary' AND status='published'", id); e != nil {
		return out, e
	}
	var lang struct{ Language *string }
	if e := r.store.Query(ctx, &lang, "SELECT e.language FROM reading_records rr JOIN editions e ON e.id=rr.edition_id WHERE rr.reader_id=? AND rr.status='completed' GROUP BY e.language ORDER BY COUNT(*) DESC,e.language ASC LIMIT 1", id); e != nil {
		return out, e
	}
	out.MostReadLanguage = lang.Language
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)
	out.Months = []core.ReadingPeriod{}
	if e := r.store.Query(ctx, &out.Months, `SELECT month,SUM(works) works,SUM(pages) pages FROM (
 SELECT MONTH(finished_at) month,COUNT(*) works,0 pages FROM reading_records WHERE reader_id=? AND status='completed' AND finished_at>=? AND finished_at<? GROUP BY MONTH(finished_at)
 UNION ALL SELECT MONTH(created_at),0,SUM(credited_pages) FROM reading_progress_events WHERE reader_id=? AND created_at>=? AND created_at<? GROUP BY MONTH(created_at)) periods GROUP BY month`, id, start, end, id, start, end); e != nil {
		return out, e
	}
	out.Genres = []core.GenreCount{}
	e := r.store.Query(ctx, &out.Genres, "SELECT g.id,g.name,COUNT(*) count FROM reading_records rr JOIN work_genres wg ON wg.work_id=rr.work_id JOIN genres g ON g.id=wg.genre_id WHERE rr.reader_id=? AND rr.status='completed' GROUP BY g.id,g.name ORDER BY count DESC,g.id ASC LIMIT 10", id)
	return out, e
}
