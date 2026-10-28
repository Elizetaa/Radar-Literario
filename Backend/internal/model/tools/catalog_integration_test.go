//go:build integration

package tools_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/schemas"
	"radar-literario/backend/internal/model/services"
	"radar-literario/backend/internal/model/tools"
)

type catalogFixture struct {
	db                     *gorm.DB
	store                  interfaces.Store
	svc                    *services.Catalog
	admin, reader, other   core.Actor
	prefix                 string
	authors, genres, works []uint64
	publisher              uint64
}

func catalogPtr[T any](v T) *T { return &v }
func newCatalogFixture(t *testing.T) *catalogFixture {
	t.Helper()
	db := readingTestDatabase(t)
	f := &catalogFixture{db: db, store: tools.NewStore(db), prefix: fmt.Sprintf("c%x", time.Now().UnixNano())}
	f.svc = services.NewCatalog(tools.NewCatalogRepository(f.store), 1)
	born := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	user := func(suffix string) core.Actor {
		id := f.insert(t, "users", map[string]any{"name": "Conta fictícia catálogo", "nickname": f.prefix + suffix, "email": f.prefix + suffix + "@example.test", "password_hash": "fixture-not-authenticatable", "birth_date": born})
		f.insert(t, "readers", map[string]any{"user_id": id})
		return core.Actor{ID: id, Reader: true, BirthDate: &born}
	}
	f.admin = user("a")
	f.admin.Role = "catalog_manager"
	f.reader = user("r")
	f.other = user("o")
	f.insert(t, "administrators", map[string]any{"user_id": f.admin.ID, "role": "catalog_manager", "status": "active"})
	f.publisher = f.insert(t, "publishers", map[string]any{"name": "Editora " + f.prefix})
	for _, suffix := range []string{"a", "b"} {
		f.authors = append(f.authors, f.insert(t, "authors", map[string]any{"name": "Autor " + f.prefix + suffix, "biography": "Biografia fictícia"}))
		f.genres = append(f.genres, f.insert(t, "genres", map[string]any{"name": "Gênero " + f.prefix + suffix, "active": true}))
	}
	t.Cleanup(func() {
		ids := []uint64{f.admin.ID, f.reader.ID, f.other.ID}
		queries := []struct {
			q string
			a []any
		}{
			{"DELETE FROM notifications WHERE recipient_id IN ? OR actor_id IN ?", []any{ids, ids}},
			{"DELETE FROM post_likes WHERE reader_id IN ?", []any{ids}},
			{"DELETE FROM reviews WHERE reader_id IN ?", []any{ids}},
			{"DELETE FROM posts WHERE reader_id IN ?", []any{ids}},
			{"DELETE FROM reading_records WHERE reader_id IN ?", []any{ids}},
			{"DELETE FROM reader_genres WHERE reader_id IN ?", []any{ids}},
			{"DELETE FROM reader_preferred_works WHERE reader_id IN ?", []any{ids}},
			{"DELETE FROM catalog_suggestions WHERE reader_id IN ?", []any{ids}},
			{"DELETE FROM audit_logs WHERE actor_id IN ?", []any{ids}},
			{"DELETE FROM editions WHERE work_id IN ?", []any{f.works}},
			{"DELETE FROM work_authors WHERE work_id IN ?", []any{f.works}},
			{"DELETE FROM work_genres WHERE work_id IN ?", []any{f.works}},
			{"DELETE FROM works WHERE id IN ?", []any{f.works}},
			{"DELETE FROM authors WHERE id IN ?", []any{f.authors}},
			{"DELETE FROM genres WHERE id IN ?", []any{f.genres}},
			{"DELETE FROM publishers WHERE id=?", []any{f.publisher}},
			{"DELETE FROM administrators WHERE user_id=?", []any{f.admin.ID}},
			{"DELETE FROM readers WHERE user_id IN ?", []any{ids}},
			{"DELETE FROM users WHERE id IN ?", []any{ids}},
		}
		for _, q := range queries {
			if e := db.Exec(q.q, q.a...).Error; e != nil {
				t.Errorf("limpeza fixture catálogo: %T", e)
			}
		}
	})
	return f
}
func (f *catalogFixture) insert(t *testing.T, table string, values map[string]any) uint64 {
	t.Helper()
	id, e := f.store.Insert(context.Background(), table, values)
	if e != nil {
		t.Fatalf("fixture %s: %v", table, e)
	}
	return id
}
func (f *catalogFixture) work(t *testing.T, suffix string, genres []uint64, age int, status string) schemas.Work {
	t.Helper()
	w, e := f.svc.SaveWork(context.Background(), f.admin, 0, schemas.WorkInput{Title: catalogPtr(f.prefix + suffix), Synopsis: catalogPtr("Sinopse fictícia"), OriginalLanguage: catalogPtr("pt"), Status: &status, MinimumAge: &age, AuthorIDs: &f.authors, GenreIDs: &genres})
	if e != nil {
		t.Fatal(e)
	}
	f.works = append(f.works, w.ID)
	return w
}
func (f *catalogFixture) edition(t *testing.T, work uint64, isbn json.RawMessage) schemas.Edition {
	t.Helper()
	in := schemas.EditionInput{WorkID: &work, PublisherID: &f.publisher, Title: catalogPtr("Edição fictícia"), ISBN: isbn, ISBNAbsenceReason: catalogPtr("Fixture sem identificador real"), EditionNumber: catalogPtr(1), PageCount: catalogPtr(100), Language: catalogPtr("pt"), MinimumAge: catalogPtr(0), Status: catalogPtr("published")}
	e, err := f.svc.SaveEdition(context.Background(), f.admin, 0, in)
	if err != nil {
		t.Fatal(err)
	}
	return e
}
func catalogConflict(t *testing.T, e error) {
	t.Helper()
	var d *core.Error
	if !errors.As(e, &d) || d.Status != 409 {
		t.Fatalf("esperado conflito, recebido %v", e)
	}
}

func TestMariaDBCatalogSuggestionConcurrentDecisionVersionAndRollback(t *testing.T) {
	f := newCatalogFixture(t)
	ctx := context.Background()
	w := f.work(t, " original", f.genres, 0, "published")
	proposal, _ := json.Marshal(map[string]any{"title": f.prefix + " aprovado"})
	suggestion, e := f.svc.CreateSuggestion(ctx, f.reader, schemas.SuggestionInput{Operation: "update", Resource: "work", TargetID: &w.ID, Proposal: proposal})
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	results := make(chan error, 8)
	start := make(chan struct{})
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, e := f.svc.DecideSuggestion(ctx, f.admin, suggestion.ID, true, "Aprovação fictícia concorrente")
			results <- e
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	success := 0
	for e := range results {
		if e == nil {
			success++
		} else {
			catalogConflict(t, e)
		}
	}
	if success != 1 {
		t.Fatalf("aprovações=%d, esperado1", success)
	}
	w, e = f.svc.Work(ctx, f.admin, w.ID, true)
	if e != nil {
		t.Fatal(e)
	}
	if w.Title != f.prefix+" aprovado" || w.Version != 2 {
		t.Fatalf("aplicação duplicada ou perdida: %#v", w)
	}
	var decisions int64
	if e = f.db.Raw("SELECT COUNT(*) FROM audit_logs WHERE target_type='catalog_suggestion' AND target_id=? AND action='catalog_suggestion.approved'", suggestion.ID).Scan(&decisions).Error; e != nil || decisions != 1 {
		t.Fatalf("auditoria duplicada: %d, %v", decisions, e)
	}
	stale, e := f.svc.CreateSuggestion(ctx, f.reader, schemas.SuggestionInput{Operation: "update", Resource: "work", TargetID: &w.ID, Proposal: json.RawMessage(`{"synopsis":"Proposta antiga"}`)})
	if e != nil {
		t.Fatal(e)
	}
	_, e = f.svc.SaveWork(ctx, f.admin, w.ID, schemas.WorkInput{Synopsis: catalogPtr("Alteração administrativa recente")})
	if e != nil {
		t.Fatal(e)
	}
	_, e = f.svc.DecideSuggestion(ctx, f.admin, stale.ID, true, "Não deve sobrescrever")
	catalogConflict(t, e)
	state, e := f.svc.Suggestion(ctx, f.reader, stale.ID, false)
	if e != nil || state.Status != "pending" {
		t.Fatalf("conflito consumiu sugestão: %v %#v", e, state)
	}
	w, e = f.svc.Work(ctx, f.admin, w.ID, true)
	if e != nil || w.Synopsis != "Alteração administrativa recente" {
		t.Fatal("sobrescreveu alteração concorrente")
	}
	failing, e := f.svc.CreateSuggestion(ctx, f.reader, schemas.SuggestionInput{Operation: "update", Resource: "work", TargetID: &w.ID, Proposal: json.RawMessage(`{"synopsis":"Não deve persistir após falha de auditoria"}`)})
	if e != nil {
		t.Fatal(e)
	}
	failSvc := services.NewCatalog(catalogAuditFailure{tools.NewCatalogRepository(f.store)}, 1)
	if _, e = failSvc.DecideSuggestion(ctx, f.admin, failing.ID, true, "Falha controlada de auditoria"); e == nil {
		t.Fatal("falha controlada não ocorreu")
	}
	after, e := f.svc.Work(ctx, f.admin, w.ID, true)
	if e != nil || after.Version != w.Version || after.Synopsis != w.Synopsis {
		t.Fatal("transação não reverteu catálogo")
	}
	state, e = f.svc.Suggestion(ctx, f.reader, failing.ID, false)
	if e != nil || state.Status != "pending" {
		t.Fatal("transação não reverteu decisão")
	}
}

type catalogAuditFailure struct{ interfaces.CatalogRepository }

func (r catalogAuditFailure) Within(ctx context.Context, fn func(interfaces.CatalogRepository) error) error {
	return r.CatalogRepository.Within(ctx, func(tx interfaces.CatalogRepository) error { return fn(catalogAuditFailure{tx}) })
}
func (r catalogAuditFailure) Audit(context.Context, core.Actor, string, string, uint64, string) error {
	return errors.New("falha controlada de auditoria")
}

func TestMariaDBCatalogISBNEquivalenceAndConcurrentUnique(t *testing.T) {
	f := newCatalogFixture(t)
	ctx := context.Background()
	w := f.work(t, " ISBN", f.genres, 0, "published")
	base := fmt.Sprintf("%09d", time.Now().UnixNano()%1000000000)
	sum := 0
	for i, c := range base {
		sum += (10 - i) * int(c-'0')
	}
	digit := (11 - sum%11) % 11
	last := string(rune('0' + digit))
	if digit == 10 {
		last = "X"
	}
	isbn10 := base + last
	isbn13, e := services.NormalizeISBN(isbn10)
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	results := make(chan error, 8)
	start := make(chan struct{})
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			isbn := isbn10
			if i%2 == 0 {
				isbn = isbn13
			}
			raw, _ := json.Marshal(isbn)
			_, e := f.svc.SaveEdition(ctx, f.admin, 0, schemas.EditionInput{WorkID: &w.ID, PublisherID: &f.publisher, Title: catalogPtr("Edição concorrente"), ISBN: raw, EditionNumber: catalogPtr(1), PageCount: catalogPtr(100), Language: catalogPtr("pt"), MinimumAge: catalogPtr(0), Status: catalogPtr("published")})
			results <- e
		}(i)
	}
	close(start)
	wg.Wait()
	close(results)
	success := 0
	for e := range results {
		if e == nil {
			success++
		} else {
			catalogConflict(t, e)
		}
	}
	if success != 1 {
		t.Fatalf("edições criadas=%d", success)
	}
	by10, e := f.svc.ISBN(ctx, f.reader, isbn10)
	if e != nil {
		t.Fatal(e)
	}
	by13, e := f.svc.ISBN(ctx, f.reader, isbn13)
	if e != nil || by10.ID != by13.ID || by13.ISBN == nil || *by13.ISBN != isbn13 {
		t.Fatal("ISBN equivalente não aponta para edição única")
	}
	if len(w.Authors) != 2 {
		t.Fatal("relação N:M perdeu autor")
	}
	list, e := f.svc.Works(ctx, f.reader, interfaces.CatalogFilter{Page: schemas.Page{Page: 1, PageSize: 20}, Title: f.prefix, Author: w.Authors[1].Name})
	if e != nil || list.Pagination.Total != 1 {
		t.Fatalf("busca segundo autor: %v, %#v", e, list)
	}
}

func TestMariaDBCatalogRecommendationsRankingsAndEligibility(t *testing.T) {
	f := newCatalogFixture(t)
	ctx := context.Background()
	page := schemas.Page{Page: 1, PageSize: 100}
	history := f.work(t, " history", []uint64{f.genres[1]}, 0, "published")
	both := f.work(t, " both", f.genres, 0, "published")
	one := f.work(t, " one", []uint64{f.genres[0]}, 0, "published")
	none := f.work(t, " none", nil, 0, "draft")
	draft := f.work(t, " draft", f.genres, 0, "draft")
	adult := f.work(t, " adult", f.genres, 18, "published")
	_ = none
	_ = draft
	_ = adult
	if _, e := f.svc.SetInterests(ctx, f.reader, []uint64{f.genres[0]}); e != nil {
		t.Fatal(e)
	}
	review := func(reader, work uint64, rating, age int, status string) {
		now := time.Now().UTC()
		post := f.insert(t, "posts", map[string]any{"reader_id": reader, "work_id": work, "kind": "review", "title": "Avaliação fixture", "text": "Texto fictício", "status": status, "contains_spoiler": false, "contains_sensitive_content": false, "minimum_age": age, "created_at": now, "updated_at": now})
		f.insert(t, "reviews", map[string]any{"post_id": post, "reader_id": reader, "work_id": work, "rating": rating})
	}
	review(f.reader.ID, history.ID, 5, 0, "published")
	review(f.other.ID, both.ID, 4, 0, "published")
	review(f.admin.ID, both.ID, 4, 0, "published")
	review(f.other.ID, one.ID, 5, 18, "published")
	review(f.admin.ID, one.ID, 4, 0, "published")
	edition := f.edition(t, history.ID, nil)
	now := time.Now().UTC()
	f.insert(t, "reading_records", map[string]any{"reader_id": f.reader.ID, "work_id": history.ID, "edition_id": edition.ID, "status": "in_progress", "current_page": 5, "created_at": now, "updated_at": now})
	unknown := f.reader
	unknown.BirthDate = nil
	filter := interfaces.CatalogFilter{Page: page, Title: f.prefix, Order: "recommended"}
	list, e := f.svc.Works(ctx, unknown, filter)
	if e != nil {
		t.Fatal(e)
	}
	works := list.Data.([]schemas.Work)
	if len(works) != 2 || works[0].ID != both.ID || works[0].Affinity != 2 || works[1].ID != one.ID || works[1].Affinity != 1 {
		t.Fatalf("recomendações incorretas: %#v", works)
	}
	if works[1].ReviewCount != 1 || works[1].AverageRating == nil || *works[1].AverageRating != 4 {
		t.Fatal("avaliação restrita vazou na média")
	}
	filter.Order = "top-rated"
	ranked := services.NewCatalog(tools.NewCatalogRepository(f.store), 2)
	list, e = ranked.Works(ctx, unknown, filter)
	if e != nil {
		t.Fatal(e)
	}
	works = list.Data.([]schemas.Work)
	if len(works) != 1 || works[0].ID != both.ID || works[0].ReviewCount != 2 {
		t.Fatalf("amostra mínima/idade ranking: %#v", works)
	}
	otherEdition := f.edition(t, both.ID, nil)
	f.insert(t, "reading_records", map[string]any{"reader_id": f.other.ID, "work_id": both.ID, "edition_id": otherEdition.ID, "status": "planned", "current_page": 0, "created_at": now, "updated_at": now})
	f.insert(t, "reading_records", map[string]any{"reader_id": f.admin.ID, "work_id": both.ID, "edition_id": otherEdition.ID, "status": "planned", "current_page": 0, "created_at": now, "updated_at": now})
	filter.Order = "popular"
	list, e = f.svc.Works(ctx, unknown, filter)
	if e != nil {
		t.Fatal(e)
	}
	works = list.Data.([]schemas.Work)
	if len(works) == 0 || works[0].ID != both.ID || works[0].ReaderCount != 2 {
		t.Fatalf("popularidade não usa leitores distintos: %#v", works)
	}
	for i := 0; i < 3; i++ {
		again, e := f.svc.Works(ctx, unknown, filter)
		if e != nil {
			t.Fatal(e)
		}
		if again.Data.([]schemas.Work)[0].ID != both.ID {
			t.Fatal("ranking não determinístico")
		}
	}
}
