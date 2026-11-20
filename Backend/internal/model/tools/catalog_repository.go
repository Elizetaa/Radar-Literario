package tools

import (
	"context"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/schemas"
	"strings"
	"time"
)

type catalogRepository struct{ store interfaces.Store }

func NewCatalogRepository(store interfaces.Store) interfaces.CatalogRepository {
	return &catalogRepository{store: store}
}
func (r *catalogRepository) Within(ctx context.Context, fn func(interfaces.CatalogRepository) error) error {
	return r.store.Transaction(ctx, func(tx interfaces.Store) error { return fn(&catalogRepository{store: tx}) })
}
func (r *catalogRepository) count(ctx context.Context, sql string, args ...any) (int64, error) {
	var result []struct{ Total int64 }
	err := r.store.Query(ctx, &result, sql, args...)
	if err != nil {
		return 0, err
	}
	if len(result) == 0 {
		return 0, nil
	}
	return result[0].Total, nil
}
func catalogPlaceholders(count int) string {
	return strings.TrimSuffix(strings.Repeat("?,", count), ",")
}

func (r *catalogRepository) ListWorks(ctx context.Context, f interfaces.CatalogFilter) ([]core.Work, int64, error) {
	where := []string{"1=1"}
	args := []any{}
	if !f.Admin {
		where = append(where, "w.status='published' AND w.minimum_age<=?")
		args = append(args, f.MaxAge)
	} else if f.Status != "" {
		where = append(where, "w.status=?")
		args = append(args, f.Status)
	}
	if f.Title != "" {
		where = append(where, "w.title LIKE ?")
		args = append(args, "%"+f.Title+"%")
	}
	if f.Author != "" {
		where = append(where, "EXISTS(SELECT 1 FROM work_authors wa JOIN authors a ON a.id=wa.author_id WHERE wa.work_id=w.id AND a.name LIKE ?)")
		args = append(args, "%"+f.Author+"%")
	}
	if f.GenreID > 0 {
		where = append(where, "EXISTS(SELECT 1 FROM work_genres wg JOIN genres g ON g.id=wg.genre_id WHERE wg.work_id=w.id AND wg.genre_id=? AND g.active=1)")
		args = append(args, f.GenreID)
	}
	if f.ISBN != "" {
		where = append(where, "EXISTS(SELECT 1 FROM editions e WHERE e.work_id=w.id AND e.isbn=? AND e.status='published' AND e.minimum_age<=?)")
		args = append(args, f.ISBN, f.MaxAge)
	}
	if f.Language != "" {
		where = append(where, "EXISTS(SELECT 1 FROM editions e WHERE e.work_id=w.id AND e.language=? AND e.status='published' AND e.minimum_age<=?)")
		args = append(args, f.Language, f.MaxAge)
	}
	if f.Preferred {
		where = append(where, "EXISTS(SELECT 1 FROM reader_preferred_works pw WHERE pw.work_id=w.id AND pw.reader_id=?)")
		args = append(args, f.ReaderID)
	}
	if f.Order == "recommended" {
		where = append(where, "NOT EXISTS(SELECT 1 FROM reading_records rr WHERE rr.work_id=w.id AND rr.reader_id=?)")
		args = append(args, f.ReaderID)
	}
	condition := strings.Join(where, " AND ")
	ratingJoin := ` LEFT JOIN (SELECT rv.work_id,AVG(rv.rating) average_rating,COUNT(*) review_count FROM reviews rv JOIN posts p ON p.id=rv.post_id JOIN users u ON u.id=p.reader_id WHERE p.status='published' AND u.status='active' AND p.minimum_age<=? GROUP BY rv.work_id) rates ON rates.work_id=w.id `
	base := ` FROM works w ` + ratingJoin + ` WHERE ` + condition
	baseArgs := append([]any{f.MaxAge}, args...)
	if f.Order == "top-rated" {
		base += ` AND COALESCE(rates.review_count,0)>=?`
		baseArgs = append(baseArgs, f.MinimumReviews)
	}
	total, err := r.count(ctx, "SELECT COUNT(*) total"+base, baseArgs...)
	if err != nil {
		return nil, 0, err
	}
	selectSQL := `SELECT w.*,rates.average_rating,COALESCE(rates.review_count,0) review_count,(SELECT COUNT(DISTINCT rr.reader_id) FROM reading_records rr JOIN users ru ON ru.id=rr.reader_id WHERE rr.work_id=w.id AND ru.status='active') reader_count`
	selectArgs := []any{}
	if f.Order == "recommended" {
		selectSQL += `,(SELECT COUNT(*) FROM work_genres wg JOIN genres ag ON ag.id=wg.genre_id AND ag.active=1 WHERE wg.work_id=w.id AND wg.genre_id IN (SELECT rg.genre_id FROM reader_genres rg WHERE rg.reader_id=? UNION SELECT hwg.genre_id FROM reviews hr JOIN posts hp ON hp.id=hr.post_id JOIN works hw ON hw.id=hr.work_id JOIN work_genres hwg ON hwg.work_id=hr.work_id WHERE hp.reader_id=? AND hp.status='published' AND hr.rating>=4 AND hw.status='published' AND GREATEST(hp.minimum_age,hw.minimum_age)<=?)) affinity`
		selectArgs = append(selectArgs, f.ReaderID, f.ReaderID, f.MaxAge)
	}
	order := "w.title ASC,w.id ASC"
	switch f.Order {
	case "popular":
		order = "reader_count DESC,COALESCE(rates.average_rating,0) DESC,w.id ASC"
	case "top-rated":
		order = "rates.average_rating DESC,rates.review_count DESC,w.id ASC"
	case "recommended":
		order = "affinity DESC,COALESCE(rates.average_rating,0) DESC,COALESCE(rates.review_count,0) DESC,w.id ASC"
	case "newest":
		order = "w.created_at DESC,w.id DESC"
	}
	allArgs := append(selectArgs, baseArgs...)
	allArgs = append(allArgs, f.Page.PageSize, f.Page.Offset())
	works := []core.Work{}
	if err = r.store.Query(ctx, &works, selectSQL+base+" ORDER BY "+order+" LIMIT ? OFFSET ?", allArgs...); err != nil {
		return nil, 0, err
	}
	if err = r.populateRelations(ctx, works); err != nil {
		return nil, 0, err
	}
	return works, total, nil
}
func (r *catalogRepository) populateRelations(ctx context.Context, works []core.Work) error {
	if len(works) == 0 {
		return nil
	}
	ids := make([]any, len(works))
	index := map[uint64]int{}
	for i := range works {
		ids[i] = works[i].ID
		index[works[i].ID] = i
		works[i].Authors = []core.Author{}
		works[i].Genres = []core.Genre{}
	}
	var authors []struct {
		WorkID uint64
		core.Author
	}
	var genres []struct {
		WorkID uint64
		core.Genre
	}
	if err := r.store.Query(ctx, &authors, "SELECT wa.work_id,a.* FROM authors a JOIN work_authors wa ON wa.author_id=a.id WHERE wa.work_id IN ("+catalogPlaceholders(len(ids))+") ORDER BY a.name,a.id", ids...); err != nil {
		return err
	}
	if err := r.store.Query(ctx, &genres, "SELECT wg.work_id,g.* FROM genres g JOIN work_genres wg ON wg.genre_id=g.id WHERE wg.work_id IN ("+catalogPlaceholders(len(ids))+") ORDER BY g.name,g.id", ids...); err != nil {
		return err
	}
	for _, v := range authors {
		at := index[v.WorkID]
		works[at].Authors = append(works[at].Authors, v.Author)
	}
	for _, v := range genres {
		at := index[v.WorkID]
		works[at].Genres = append(works[at].Genres, v.Genre)
	}
	return nil
}
func (r *catalogRepository) Work(ctx context.Context, id uint64, age int, lock bool) (core.Work, error) {
	rows := []core.Work{}
	query := "SELECT * FROM works WHERE id=?"
	if lock {
		query += " FOR UPDATE"
	}
	if err := r.store.Query(ctx, &rows, query, id); err != nil {
		return core.Work{}, err
	}
	if len(rows) == 0 {
		return core.Work{}, core.NotFound()
	}
	if err := r.populateRelations(ctx, rows); err != nil {
		return core.Work{}, err
	}
	var rating []struct {
		AverageRating *float64
		ReviewCount   int64
	}
	if err := r.store.Query(ctx, &rating, `SELECT AVG(rv.rating) average_rating,COUNT(*) review_count FROM reviews rv JOIN posts p ON p.id=rv.post_id JOIN users u ON u.id=p.reader_id WHERE rv.work_id=? AND p.status='published' AND u.status='active' AND p.minimum_age<=?`, id, age); err != nil {
		return core.Work{}, err
	}
	if len(rating) > 0 {
		rows[0].AverageRating = rating[0].AverageRating
		rows[0].ReviewCount = rating[0].ReviewCount
	}
	return rows[0], nil
}
func (r *catalogRepository) SaveWork(ctx context.Context, w *core.Work, authors, genres []uint64) error {
	if w.ID == 0 {
		id, err := r.store.Insert(ctx, "works", map[string]any{"title": w.Title, "synopsis": w.Synopsis, "original_language": w.OriginalLanguage, "first_publication_date": w.FirstPublicationDate, "status": w.Status, "minimum_age": w.MinimumAge, "version": 1, "created_at": time.Now().UTC(), "updated_at": time.Now().UTC()})
		if err != nil {
			return err
		}
		w.ID = id
		w.Version = 1
	} else {
		_, err := r.store.Exec(ctx, `UPDATE works SET title=?,synopsis=?,original_language=?,first_publication_date=?,status=?,minimum_age=?,version=version+1,updated_at=UTC_TIMESTAMP(6) WHERE id=?`, w.Title, w.Synopsis, w.OriginalLanguage, w.FirstPublicationDate, w.Status, w.MinimumAge, w.ID)
		if err != nil {
			return err
		}
		w.Version++
	}
	if _, err := r.store.Exec(ctx, "DELETE FROM work_authors WHERE work_id=?", w.ID); err != nil {
		return err
	}
	for _, id := range authors {
		if _, err := r.store.Exec(ctx, "INSERT INTO work_authors(work_id,author_id) VALUES(?,?)", w.ID, id); err != nil {
			return err
		}
	}
	if _, err := r.store.Exec(ctx, "DELETE FROM work_genres WHERE work_id=?", w.ID); err != nil {
		return err
	}
	for _, id := range genres {
		if _, err := r.store.Exec(ctx, "INSERT INTO work_genres(work_id,genre_id) VALUES(?,?)", w.ID, id); err != nil {
			return err
		}
	}
	return nil
}
func (r *catalogRepository) ListEditions(ctx context.Context, f interfaces.CatalogFilter) ([]core.Edition, int64, error) {
	where := "1=1"
	args := []any{}
	if f.WorkID > 0 {
		where += " AND e.work_id=?"
		args = append(args, f.WorkID)
	}
	if !f.Admin {
		where += " AND e.status='published' AND w.status='published' AND GREATEST(e.minimum_age,w.minimum_age)<=?"
		args = append(args, f.MaxAge)
	} else if f.Status != "" {
		where += " AND e.status=?"
		args = append(args, f.Status)
	}
	base := " FROM editions e JOIN works w ON w.id=e.work_id WHERE " + where
	total, err := r.count(ctx, "SELECT COUNT(*) total"+base, args...)
	if err != nil {
		return nil, 0, err
	}
	rows := []core.Edition{}
	args = append(args, f.Page.PageSize, f.Page.Offset())
	err = r.store.Query(ctx, &rows, "SELECT e.*"+base+" ORDER BY e.id LIMIT ? OFFSET ?", args...)
	return rows, total, err
}
func (r *catalogRepository) Edition(ctx context.Context, id uint64, lock bool) (core.Edition, error) {
	rows := []core.Edition{}
	q := "SELECT * FROM editions WHERE id=?"
	if lock {
		q += " FOR UPDATE"
	}
	err := r.store.Query(ctx, &rows, q, id)
	if err != nil {
		return core.Edition{}, err
	}
	if len(rows) == 0 {
		return core.Edition{}, core.NotFound()
	}
	return rows[0], nil
}
func (r *catalogRepository) EditionByISBN(ctx context.Context, isbn string) (core.Edition, error) {
	rows := []core.Edition{}
	err := r.store.Query(ctx, &rows, "SELECT * FROM editions WHERE isbn=?", isbn)
	if err != nil {
		return core.Edition{}, err
	}
	if len(rows) == 0 {
		return core.Edition{}, core.NotFound()
	}
	return rows[0], nil
}
func (r *catalogRepository) SaveEdition(ctx context.Context, e *core.Edition) error {
	if e.ID == 0 {
		id, err := r.store.Insert(ctx, "editions", map[string]any{"work_id": e.WorkID, "publisher_id": e.PublisherID, "title": e.Title, "isbn": e.ISBN, "isbn_absence_reason": e.ISBNAbsenceReason, "edition_number": e.EditionNumber, "page_count": e.PageCount, "publication_date": e.PublicationDate, "language": e.Language, "image_url": e.ImageURL, "status": e.Status, "minimum_age": e.MinimumAge, "version": 1, "created_at": time.Now().UTC(), "updated_at": time.Now().UTC()})
		if err != nil {
			return err
		}
		e.ID = id
		e.Version = 1
		return nil
	}
	_, err := r.store.Exec(ctx, `UPDATE editions SET publisher_id=?,title=?,isbn=?,isbn_absence_reason=?,edition_number=?,page_count=?,publication_date=?,language=?,image_url=?,status=?,minimum_age=?,version=version+1,updated_at=UTC_TIMESTAMP(6) WHERE id=?`, e.PublisherID, e.Title, e.ISBN, e.ISBNAbsenceReason, e.EditionNumber, e.PageCount, e.PublicationDate, e.Language, e.ImageURL, e.Status, e.MinimumAge, e.ID)
	if err == nil {
		e.Version++
	}
	return err
}
func (r *catalogRepository) ValidateEditionPages(ctx context.Context, id uint64, pages int) error {
	count, err := r.count(ctx, `SELECT (SELECT COUNT(*) FROM reading_records WHERE edition_id=? AND current_page>?) + (SELECT COUNT(*) FROM reading_diaries WHERE edition_id=? AND current_page>?) total`, id, pages, id, pages)
	if err != nil {
		return err
	}
	if count > 0 {
		return core.Conflict("O novo total de páginas invalidaria registros de leitura ou diários históricos.")
	}
	return nil
}
func (r *catalogRepository) ListAuthors(ctx context.Context, p schemas.Page) ([]core.Author, int64, error) {
	rows := []core.Author{}
	n, err := r.count(ctx, "SELECT COUNT(*) total FROM authors")
	if err != nil {
		return nil, 0, err
	}
	err = r.store.Query(ctx, &rows, "SELECT * FROM authors ORDER BY name,id LIMIT ? OFFSET ?", p.PageSize, p.Offset())
	return rows, n, err
}
func (r *catalogRepository) Author(ctx context.Context, id uint64) (core.Author, error) {
	rows := []core.Author{}
	err := r.store.Query(ctx, &rows, "SELECT * FROM authors WHERE id=?", id)
	if err != nil {
		return core.Author{}, err
	}
	if len(rows) == 0 {
		return core.Author{}, core.NotFound()
	}
	return rows[0], nil
}
func (r *catalogRepository) SaveAuthor(ctx context.Context, a *core.Author) error {
	if a.ID == 0 {
		id, err := r.store.Insert(ctx, "authors", map[string]any{"name": a.Name, "biography": a.Biography, "birth_date": a.BirthDate, "death_date": a.DeathDate, "country_of_origin": a.CountryOfOrigin, "sex": a.Sex})
		a.ID = id
		return err
	}
	_, err := r.store.Exec(ctx, "UPDATE authors SET name=?,biography=?,birth_date=?,death_date=?,country_of_origin=?,sex=?,updated_at=UTC_TIMESTAMP(6) WHERE id=?", a.Name, a.Biography, a.BirthDate, a.DeathDate, a.CountryOfOrigin, a.Sex, a.ID)
	return err
}
func (r *catalogRepository) DeleteAuthor(ctx context.Context, id uint64) error {
	_, err := r.store.Exec(ctx, "DELETE FROM authors WHERE id=?", id)
	return err
}
func (r *catalogRepository) ListPublishers(ctx context.Context, p schemas.Page) ([]core.Publisher, int64, error) {
	rows := []core.Publisher{}
	n, err := r.count(ctx, "SELECT COUNT(*) total FROM publishers")
	if err != nil {
		return nil, 0, err
	}
	err = r.store.Query(ctx, &rows, "SELECT * FROM publishers ORDER BY name,id LIMIT ? OFFSET ?", p.PageSize, p.Offset())
	return rows, n, err
}
func (r *catalogRepository) Publisher(ctx context.Context, id uint64) (core.Publisher, error) {
	rows := []core.Publisher{}
	err := r.store.Query(ctx, &rows, "SELECT * FROM publishers WHERE id=?", id)
	if err != nil {
		return core.Publisher{}, err
	}
	if len(rows) == 0 {
		return core.Publisher{}, core.NotFound()
	}
	return rows[0], nil
}
func (r *catalogRepository) SavePublisher(ctx context.Context, p *core.Publisher) error {
	if p.ID == 0 {
		id, err := r.store.Insert(ctx, "publishers", map[string]any{"name": p.Name, "website_url": p.WebsiteURL, "country_of_origin": p.CountryOfOrigin})
		p.ID = id
		return err
	}
	_, err := r.store.Exec(ctx, "UPDATE publishers SET name=?,website_url=?,country_of_origin=?,updated_at=UTC_TIMESTAMP(6) WHERE id=?", p.Name, p.WebsiteURL, p.CountryOfOrigin, p.ID)
	return err
}
func (r *catalogRepository) DeletePublisher(ctx context.Context, id uint64) error {
	_, err := r.store.Exec(ctx, "DELETE FROM publishers WHERE id=?", id)
	return err
}
func (r *catalogRepository) ListGenres(ctx context.Context, p schemas.Page, all bool) ([]core.Genre, int64, error) {
	where := " WHERE active=1"
	if all {
		where = ""
	}
	rows := []core.Genre{}
	n, err := r.count(ctx, "SELECT COUNT(*) total FROM genres"+where)
	if err != nil {
		return nil, 0, err
	}
	err = r.store.Query(ctx, &rows, "SELECT * FROM genres"+where+" ORDER BY name,id LIMIT ? OFFSET ?", p.PageSize, p.Offset())
	return rows, n, err
}
func (r *catalogRepository) Genre(ctx context.Context, id uint64) (core.Genre, error) {
	rows := []core.Genre{}
	err := r.store.Query(ctx, &rows, "SELECT * FROM genres WHERE id=?", id)
	if err != nil {
		return core.Genre{}, err
	}
	if len(rows) == 0 {
		return core.Genre{}, core.NotFound()
	}
	return rows[0], nil
}
func (r *catalogRepository) SaveGenre(ctx context.Context, g *core.Genre) error {
	if g.ID == 0 {
		id, err := r.store.Insert(ctx, "genres", map[string]any{"name": g.Name, "active": g.Active})
		g.ID = id
		return err
	}
	_, err := r.store.Exec(ctx, "UPDATE genres SET name=?,active=? WHERE id=?", g.Name, g.Active, g.ID)
	return err
}
func (r *catalogRepository) Interests(ctx context.Context, id uint64) ([]core.Genre, error) {
	rows := []core.Genre{}
	err := r.store.Query(ctx, &rows, "SELECT g.* FROM genres g JOIN reader_genres rg ON rg.genre_id=g.id WHERE rg.reader_id=? ORDER BY g.name,g.id", id)
	return rows, err
}
func (r *catalogRepository) SetInterests(ctx context.Context, id uint64, genres []uint64) error {
	if _, err := r.store.Exec(ctx, "DELETE FROM reader_genres WHERE reader_id=?", id); err != nil {
		return err
	}
	for _, g := range genres {
		if _, err := r.store.Exec(ctx, "INSERT INTO reader_genres(reader_id,genre_id) VALUES(?,?)", id, g); err != nil {
			return err
		}
	}
	return nil
}
func (r *catalogRepository) SetPreferred(ctx context.Context, id, workID uint64, enabled bool) error {
	if enabled {
		_, err := r.store.Exec(ctx, "INSERT INTO reader_preferred_works(reader_id,work_id) VALUES(?,?) ON DUPLICATE KEY UPDATE reader_id=VALUES(reader_id)", id, workID)
		return err
	}
	_, err := r.store.Exec(ctx, "DELETE FROM reader_preferred_works WHERE reader_id=? AND work_id=?", id, workID)
	return err
}
func (r *catalogRepository) ListSuggestions(ctx context.Context, reader uint64, status string, p schemas.Page) ([]core.CatalogSuggestion, int64, error) {
	where := "1=1"
	args := []any{}
	if reader > 0 {
		where += " AND reader_id=?"
		args = append(args, reader)
	}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	n, err := r.count(ctx, "SELECT COUNT(*) total FROM catalog_suggestions WHERE "+where, args...)
	if err != nil {
		return nil, 0, err
	}
	args = append(args, p.PageSize, p.Offset())
	rows := []core.CatalogSuggestion{}
	err = r.store.Query(ctx, &rows, "SELECT * FROM catalog_suggestions WHERE "+where+" ORDER BY created_at DESC,id DESC LIMIT ? OFFSET ?", args...)
	return rows, n, err
}
func (r *catalogRepository) Suggestion(ctx context.Context, id uint64, lock bool) (core.CatalogSuggestion, error) {
	rows := []core.CatalogSuggestion{}
	q := "SELECT * FROM catalog_suggestions WHERE id=?"
	if lock {
		q += " FOR UPDATE"
	}
	err := r.store.Query(ctx, &rows, q, id)
	if err != nil {
		return core.CatalogSuggestion{}, err
	}
	if len(rows) == 0 {
		return core.CatalogSuggestion{}, core.NotFound()
	}
	return rows[0], nil
}
func (r *catalogRepository) SaveSuggestion(ctx context.Context, s *core.CatalogSuggestion) error {
	if s.ID == 0 {
		id, err := r.store.Insert(ctx, "catalog_suggestions", map[string]any{"reader_id": s.ReaderID, "operation": s.Operation, "resource": s.Resource, "target_id": s.TargetID, "expected_version": s.ExpectedVersion, "proposal": s.Proposal, "status": s.Status})
		s.ID = id
		return err
	}
	_, err := r.store.Exec(ctx, "UPDATE catalog_suggestions SET status=?,decision_by=?,decision_reason=?,decided_at=?,updated_at=UTC_TIMESTAMP(6) WHERE id=?", s.Status, s.DecisionBy, s.DecisionReason, s.DecidedAt, s.ID)
	return err
}
func (r *catalogRepository) Audit(ctx context.Context, a core.Actor, action, kind string, id uint64, reason string) error {
	_, err := r.store.Insert(ctx, "audit_logs", map[string]any{"actor_id": a.ID, "action": action, "target_type": kind, "target_id": id, "reason": reason, "created_at": time.Now().UTC()})
	return err
}

var _ interfaces.CatalogRepository = (*catalogRepository)(nil)
