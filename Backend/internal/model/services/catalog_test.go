package services

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/schemas"
)

func TestNormalizeISBNCanonicalAndChecksums(t *testing.T) {
	for _, raw := range []string{"0-306-40615-2", "9780306406157", " 978-0-306-40615-7 "} {
		got, err := NormalizeISBN(raw)
		if err != nil || got != "9780306406157" {
			t.Fatalf("%q: got %q, %v", raw, got, err)
		}
	}
	if got, err := NormalizeISBN("0-8044-2957-X"); err != nil || got != "9780804429573" {
		t.Fatalf("X conversion: %q %v", got, err)
	}
	for _, raw := range []string{"", "9780306406158", "0306406153", "978030640615X", "9780306406157; DROP TABLE works", "1111111111111", "９７８０３０６４０６１５７"} {
		if _, err := NormalizeISBN(raw); err == nil {
			t.Errorf("accepted invalid ISBN %q", raw)
		}
	}
}

type catalogFake struct {
	interfaces.CatalogRepository
	mu         *sync.Mutex
	work       core.Work
	edition    core.Edition
	suggestion core.CatalogSuggestion
	filter     interfaces.CatalogFilter
	saved      int
	audits     int
	failAudit  bool
}

func (f *catalogFake) Within(ctx context.Context, fn func(interfaces.CatalogRepository) error) error {
	if f.mu != nil {
		f.mu.Lock()
		defer f.mu.Unlock()
	}
	copy := *f
	copy.mu = nil
	if err := fn(&copy); err != nil {
		return err
	}
	mutex := f.mu
	*f = copy
	f.mu = mutex
	return nil
}
func (f *catalogFake) Work(ctx context.Context, id uint64, age int, lock bool) (core.Work, error) {
	if f.work.ID != id {
		return core.Work{}, core.NotFound()
	}
	return f.work, nil
}
func (f *catalogFake) SaveWork(ctx context.Context, w *core.Work, authors, genres []uint64) error {
	if w.ID == 0 {
		w.ID = 1
		w.Version = 1
	} else {
		w.Version++
	}
	w.Authors = []core.Author{}
	w.Genres = []core.Genre{}
	for _, id := range authors {
		w.Authors = append(w.Authors, core.Author{ID: id, Name: "Autor"})
	}
	for _, id := range genres {
		w.Genres = append(w.Genres, core.Genre{ID: id, Name: "Gênero", Active: true})
	}
	f.work = *w
	f.saved++
	return nil
}
func (f *catalogFake) Author(context.Context, uint64) (core.Author, error) {
	return core.Author{ID: 1, Name: "Autor"}, nil
}
func (f *catalogFake) Genre(context.Context, uint64) (core.Genre, error) {
	return core.Genre{ID: 1, Name: "Ficção", Active: true}, nil
}
func (f *catalogFake) Publisher(context.Context, uint64) (core.Publisher, error) {
	return core.Publisher{ID: 1, Name: "Editora"}, nil
}
func (f *catalogFake) Edition(ctx context.Context, id uint64, lock bool) (core.Edition, error) {
	if f.edition.ID != id {
		return core.Edition{}, core.NotFound()
	}
	return f.edition, nil
}
func (f *catalogFake) EditionByISBN(context.Context, string) (core.Edition, error) {
	return core.Edition{}, core.NotFound()
}
func (f *catalogFake) ValidateEditionPages(context.Context, uint64, int) error { return nil }
func (f *catalogFake) Audit(context.Context, core.Actor, string, string, uint64, string) error {
	if f.failAudit {
		return errors.New("audit unavailable")
	}
	f.audits++
	return nil
}
func (f *catalogFake) Suggestion(ctx context.Context, id uint64, lock bool) (core.CatalogSuggestion, error) {
	if id != f.suggestion.ID {
		return core.CatalogSuggestion{}, core.NotFound()
	}
	return f.suggestion, nil
}
func (f *catalogFake) SaveSuggestion(ctx context.Context, s *core.CatalogSuggestion) error {
	if s.ID == 0 {
		s.ID = 1
	}
	f.suggestion = *s
	return nil
}
func (f *catalogFake) ListWorks(ctx context.Context, q interfaces.CatalogFilter) ([]core.Work, int64, error) {
	f.filter = q
	return []core.Work{f.work}, 1, nil
}

func catalogAdmin() core.Actor {
	birth := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	return core.Actor{ID: 9, Role: "catalog_manager", Reader: true, BirthDate: &birth}
}
func catalogFixture() *catalogFake {
	return &catalogFake{mu: &sync.Mutex{}, work: core.Work{ID: 1, Title: "Obra", Synopsis: "Sinopse", OriginalLanguage: "pt-br", Status: "published", MinimumAge: 0, Version: 3, Authors: []core.Author{{ID: 1}}, Genres: []core.Genre{{ID: 1, Active: true}}}}
}
func testPtr[T any](v T) *T { return &v }

func TestCatalogPublicDetailProtectsDraftAndAge(t *testing.T) {
	ctx := context.Background()
	repo := catalogFixture()
	s := NewCatalog(repo, 2)
	repo.work.Status = "draft"
	if _, err := s.Work(ctx, core.Actor{}, 1, false); err == nil {
		t.Fatal("visitor read draft")
	}
	repo.work.Status = "published"
	repo.work.MinimumAge = 18
	if _, err := s.Work(ctx, core.Actor{}, 1, false); err == nil {
		t.Fatal("visitor read adult work")
	}
	if _, err := s.Work(ctx, catalogAdmin(), 1, true); err != nil {
		t.Fatal(err)
	}
	reader := catalogAdmin()
	reader.Role = ""
	if _, err := s.Work(ctx, reader, 1, true); err == nil {
		t.Fatal("reader accessed admin work")
	}
}

func TestCatalogReaderCannotMutateOrMassAssignCatalog(t *testing.T) {
	repo := catalogFixture()
	s := NewCatalog(repo, 1)
	a := catalogAdmin()
	a.Role = ""
	if _, err := s.SaveWork(context.Background(), a, 1, schemas.WorkInput{Title: testPtr("alterada")}); err == nil {
		t.Fatal("reader changed catalog")
	}
	if repo.saved != 0 {
		t.Fatal("unauthorized persistence")
	}
	var in schemas.WorkInput
	if err := catalogStrict(json.RawMessage(`{"title":"x","reader_id":123}`), &in); err == nil {
		t.Fatal("unknown field accepted")
	}
}

func TestCatalogWorkChangeAndAuditAreAtomic(t *testing.T) {
	repo := catalogFixture()
	repo.failAudit = true
	s := NewCatalog(repo, 1)
	if _, err := s.SaveWork(context.Background(), catalogAdmin(), 1, schemas.WorkInput{Title: testPtr("alterada")}); err == nil {
		t.Fatal("audit error ignored")
	}
	if repo.work.Title != "Obra" || repo.work.Version != 3 || repo.saved != 0 {
		t.Fatalf("rollback failed: %+v", repo.work)
	}
}

func TestCatalogPublishedWorkRequiresExplicitClassificationRelations(t *testing.T) {
	repo := catalogFixture()
	s := NewCatalog(repo, 1)
	base := schemas.WorkInput{Title: testPtr("Nova"), Synopsis: testPtr("Sinopse"), OriginalLanguage: testPtr("pt"), Status: testPtr("published"), AuthorIDs: testPtr([]uint64{1}), GenreIDs: testPtr([]uint64{1})}
	if _, err := s.SaveWork(context.Background(), catalogAdmin(), 0, base); err == nil {
		t.Fatal("missing age accepted")
	}
	base.MinimumAge = testPtr(0)
	base.AuthorIDs = testPtr([]uint64{1, 1})
	if _, err := s.SaveWork(context.Background(), catalogAdmin(), 0, base); err == nil {
		t.Fatal("duplicate relation accepted")
	}
}

func TestCatalogEditionValidationRejectsPagesDatesAndISBN(t *testing.T) {
	repo := catalogFixture()
	s := NewCatalog(repo, 1)
	ctx := context.Background()
	base := schemas.EditionInput{WorkID: testPtr(uint64(1)), PublisherID: testPtr(uint64(1)), Title: testPtr("Edição"), PageCount: testPtr(0), Language: testPtr("pt"), MinimumAge: testPtr(0), ISBN: json.RawMessage(`"0306406152"`)}
	if _, err := s.prepareEdition(ctx, repo, 0, base); err == nil {
		t.Fatal("zero pages accepted")
	}
	base.PageCount = testPtr(100)
	base.ISBN = json.RawMessage(`"0306406153"`)
	if _, err := s.prepareEdition(ctx, repo, 0, base); err == nil {
		t.Fatal("invalid ISBN accepted")
	}
	base.ISBN = json.RawMessage(`null`)
	if _, err := s.prepareEdition(ctx, repo, 0, base); err == nil {
		t.Fatal("missing ISBN reason accepted")
	}
	base.ISBNAbsenceReason = testPtr("Publicação independente sem ISBN")
	got, err := s.prepareEdition(ctx, repo, 0, base)
	if err != nil || got.ISBN != nil {
		t.Fatalf("manual edition: %+v %v", got, err)
	}
}

func TestCatalogSuggestionOwnershipConflictAndOnceOnly(t *testing.T) {
	ctx := context.Background()
	repo := catalogFixture()
	target := uint64(1)
	version := uint64(2)
	repo.suggestion = core.CatalogSuggestion{ID: 4, ReaderID: 2, Operation: "update", Resource: "work", TargetID: &target, ExpectedVersion: &version, Proposal: `{"title":"Sugerida"}`, Status: "pending"}
	s := NewCatalog(repo, 1)
	reader := catalogAdmin()
	reader.ID = 3
	reader.Role = ""
	if _, err := s.Suggestion(ctx, reader, 4, false); err == nil {
		t.Fatal("other reader saw suggestion")
	}
	if _, err := s.DecideSuggestion(ctx, catalogAdmin(), 4, true, "Revisão válida"); err == nil {
		t.Fatal("stale target accepted")
	}
	if repo.saved != 0 || repo.suggestion.Status != "pending" {
		t.Fatal("stale proposal changed state")
	}
	version = 3
	repo.suggestion.ExpectedVersion = &version
	if _, err := s.DecideSuggestion(ctx, catalogAdmin(), 4, true, "Revisão válida"); err != nil {
		t.Fatal(err)
	}
	if repo.saved != 1 || repo.work.Title != "Sugerida" {
		t.Fatal("approval not applied")
	}
	if _, err := s.DecideSuggestion(ctx, catalogAdmin(), 4, true, "Outra revisão"); err == nil {
		t.Fatal("second approval accepted")
	}
	if repo.saved != 1 {
		t.Fatal("approval reapplied")
	}
}

func TestCatalogConcurrentSuggestionApprovalDoesNotReapply(t *testing.T) {
	ctx := context.Background()
	repo := catalogFixture()
	target, version := uint64(1), uint64(3)
	repo.suggestion = core.CatalogSuggestion{ID: 1, ReaderID: 2, Operation: "update", Resource: "work", TargetID: &target, ExpectedVersion: &version, Proposal: `{"title":"Concorrente"}`, Status: "pending"}
	s := NewCatalog(repo, 1)
	var group sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			_, err := s.DecideSuggestion(ctx, catalogAdmin(), 1, true, "Revisão concorrente")
			results <- err
		}()
	}
	group.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		}
	}
	if success != 1 || repo.saved != 1 {
		t.Fatalf("successes=%d saves=%d", success, repo.saved)
	}
}

func TestCatalogRecommendationsUseActorAndMinimumSample(t *testing.T) {
	repo := catalogFixture()
	repo.work.Affinity = 2
	s := NewCatalog(repo, 3)
	a := catalogAdmin()
	a.ID = 21
	p := schemas.Page{Page: 2, PageSize: 10}
	result, err := s.Recommendations(context.Background(), a, p)
	if err != nil {
		t.Fatal(err)
	}
	if repo.filter.ReaderID != 21 || repo.filter.Order != "recommended" || repo.filter.MinimumReviews != 3 || repo.filter.Page.Offset() != 10 {
		t.Fatalf("wrong filter %+v", repo.filter)
	}
	rows := result.Data.([]schemas.Work)
	if len(rows) != 1 || rows[0].RecommendationReason == "" {
		t.Fatal("missing recommendation explanation")
	}
	if _, err = s.Recommendations(context.Background(), core.Actor{}, p); err == nil {
		t.Fatal("anonymous personal recommendations")
	}
}

func TestCatalogDatePatchSupportsNullAndRejectsBadFormat(t *testing.T) {
	old := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	if value, err := catalogDate(nil, &old); err != nil || value != &old {
		t.Fatal("omission changed date")
	}
	if value, err := catalogDate(json.RawMessage(`null`), &old); err != nil || value != nil {
		t.Fatal("null did not clear date")
	}
	if _, err := catalogDate(json.RawMessage(`"01/01/2000"`), nil); err == nil {
		t.Fatal("bad civil date accepted")
	}
}
