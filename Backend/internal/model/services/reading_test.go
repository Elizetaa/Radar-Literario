package services

import (
	"context"
	"errors"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/schemas"
	"testing"
	"time"
)

type readingFake struct {
	interfaces.ReadingRepository
	record        *core.ReadingRecord
	posts         map[uint64]core.ReadingPost
	comments      map[uint64]core.ReadingComment
	events        []core.ProgressEvent
	likes         map[string]bool
	notifications map[string]core.ReadingNotification
	edition       core.ReadingEdition
	nextID        uint64
	failDiary     bool
	statReader    uint64
}

func newReadingFake() *readingFake {
	return &readingFake{posts: map[uint64]core.ReadingPost{}, comments: map[uint64]core.ReadingComment{}, likes: map[string]bool{}, notifications: map[string]core.ReadingNotification{}, edition: core.ReadingEdition{ID: 3, WorkID: 7, PageCount: 100, Status: "published", WorkStatus: "published"}}
}
func copyReadingRecord(v *core.ReadingRecord) *core.ReadingRecord {
	if v == nil {
		return nil
	}
	c := *v
	return &c
}
func (f *readingFake) Within(ctx context.Context, fn func(interfaces.ReadingRepository) error) error {
	before := copyReadingRecord(f.record)
	events := append([]core.ProgressEvent{}, f.events...)
	posts := map[uint64]core.ReadingPost{}
	for k, v := range f.posts {
		posts[k] = v
	}
	if e := fn(f); e != nil {
		f.record = before
		f.events = events
		f.posts = posts
		return e
	}
	return nil
}
func (f *readingFake) LockReader(context.Context, uint64) error { return nil }
func (f *readingFake) Edition(context.Context, uint64) (core.ReadingEdition, error) {
	return f.edition, nil
}
func (f *readingFake) Record(_ context.Context, reader, work uint64) (*core.ReadingRecord, error) {
	if f.record == nil || f.record.ReaderID != reader || f.record.WorkID != work {
		return nil, nil
	}
	return copyReadingRecord(f.record), nil
}
func (f *readingFake) SaveRecord(_ context.Context, r *core.ReadingRecord) error {
	if r.ID == 0 {
		r.ID = 9
	}
	f.record = copyReadingRecord(r)
	return nil
}
func (f *readingFake) RemoveRecord(context.Context, uint64, uint64) error {
	f.record = nil
	f.events = nil
	return nil
}
func (f *readingFake) AddProgress(_ context.Context, e core.ProgressEvent) error {
	f.events = append(f.events, e)
	return nil
}
func (f *readingFake) Post(_ context.Context, id, viewer uint64) (core.ReadingPost, error) {
	p, ok := f.posts[id]
	if !ok {
		return p, core.NotFound()
	}
	p.AuthorStatus = "active"
	p.WorkStatus = "published"
	p.WorkMinimumAge = f.edition.MinimumAge
	return p, nil
}
func (f *readingFake) CreatePost(_ context.Context, p *core.ReadingPost) error {
	f.nextID++
	p.ID = f.nextID
	f.posts[p.ID] = *p
	return nil
}
func (f *readingFake) SavePost(_ context.Context, p core.ReadingPost) error {
	f.posts[p.ID] = p
	return nil
}
func (f *readingFake) CreateDiary(context.Context, core.ReadingPost) error {
	if f.failDiary {
		return errors.New("falha simulada no subtipo")
	}
	return nil
}
func (f *readingFake) SaveDiary(_ context.Context, p core.ReadingPost) error {
	v := f.posts[p.ID]
	v.CurrentPage = p.CurrentPage
	f.posts[p.ID] = v
	return nil
}
func (f *readingFake) CreateReview(_ context.Context, p core.ReadingPost) error {
	for id, v := range f.posts {
		if id != p.ID && v.Kind == "review" && v.ReaderID == p.ReaderID && v.WorkID == p.WorkID && v.Status != "deleted" {
			return core.Conflict("Já existe uma avaliação.")
		}
	}
	return nil
}
func (f *readingFake) SaveReview(_ context.Context, p core.ReadingPost) error {
	f.posts[p.ID] = p
	return nil
}
func (f *readingFake) ReleaseReview(context.Context, uint64) error { return nil }
func (f *readingFake) Comment(_ context.Context, id, viewer uint64) (core.ReadingComment, error) {
	v, ok := f.comments[id]
	if !ok {
		return v, core.NotFound()
	}
	return v, nil
}
func (f *readingFake) CreateComment(_ context.Context, c *core.ReadingComment) error {
	f.nextID++
	c.ID = f.nextID
	f.comments[c.ID] = *c
	return nil
}
func (f *readingFake) SaveComment(_ context.Context, c core.ReadingComment) error {
	f.comments[c.ID] = c
	return nil
}
func (f *readingFake) Like(_ context.Context, target string, id, reader uint64, add bool) (bool, error) {
	key := target + time.Unix(int64(id), int64(reader)).String()
	before := f.likes[key]
	f.likes[key] = add
	return before != add, nil
}
func (f *readingFake) Notify(_ context.Context, n core.ReadingNotification) error {
	f.notifications[n.EventKey] = n
	return nil
}
func (f *readingFake) Statistics(_ context.Context, id uint64, year int) (core.ReadingStatistics, error) {
	f.statReader = id
	return core.ReadingStatistics{CurrentPages: 35, Statuses: []core.StatusCount{{Status: "in_progress", Count: 1}}, Months: []core.ReadingPeriod{{Month: 1, Pages: 20, Works: 0}, {Month: 12, Pages: 15, Works: 1}}}, nil
}
func testReader() core.Actor {
	birth := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	return core.Actor{ID: 1, Reader: true, BirthDate: &birth}
}
func testReading(f *readingFake) *Reading {
	s := NewReading(f, ReadingOptions{})
	s.now = func() time.Time { return time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC) }
	return s
}
func intptr(v int) *int    { return &v }
func boolptr(v bool) *bool { return &v }
func assertReadingCode(t *testing.T, e error, code string) {
	t.Helper()
	var domain *core.Error
	if !errors.As(e, &domain) || domain.Code != code {
		t.Fatalf("esperado %s, recebido %v", code, e)
	}
}

func TestReadingProgressCreditsOnlyNewAdvanceAndRequiresCorrection(t *testing.T) {
	f := newReadingFake()
	s := testReading(f)
	a := testReader()
	ctx := context.Background()
	for _, step := range []struct {
		page int
		kind string
	}{{20, "advance"}, {35, "advance"}, {10, "correction"}, {35, "advance"}, {40, "advance"}} {
		_, e := s.PutShelf(ctx, a, 7, schemas.ShelfInput{EditionID: 3, CurrentPage: intptr(step.page), Status: "in_progress", ChangeKind: step.kind})
		if e != nil {
			t.Fatal(e)
		}
	}
	credit := 0
	for _, e := range f.events {
		credit += e.CreditedPages
	}
	if credit != 40 {
		t.Fatalf("crédito duplicado após correção: %d", credit)
	}
	_, e := s.PutShelf(ctx, a, 7, schemas.ShelfInput{EditionID: 3, CurrentPage: intptr(1), Status: "in_progress"})
	assertReadingCode(t, e, "EXPLICIT_CORRECTION_REQUIRED")
	_, e = s.PutShelf(ctx, a, 7, schemas.ShelfInput{EditionID: 3, CurrentPage: intptr(101), Status: "in_progress"})
	assertReadingCode(t, e, "INVALID_READING_PROGRESS")
	_, e = s.PutShelf(ctx, a, 99, schemas.ShelfInput{EditionID: 3, CurrentPage: intptr(20), Status: "in_progress"})
	assertReadingCode(t, e, "EDITION_WORK_MISMATCH")
}
func TestReadingCompletionAndExplicitEditionChange(t *testing.T) {
	f := newReadingFake()
	s := testReading(f)
	a := testReader()
	ctx := context.Background()
	v, e := s.PutShelf(ctx, a, 7, schemas.ShelfInput{EditionID: 3, CurrentPage: intptr(0), Status: "completed"})
	if e != nil {
		t.Fatal(e)
	}
	if v.CurrentPage != 100 || v.FinishedAt == nil || v.StartedAt == nil {
		t.Fatalf("conclusão inconsistente: %+v", v)
	}
	f.edition.ID = 4
	f.edition.PageCount = 200
	_, e = s.PutShelf(ctx, a, 7, schemas.ShelfInput{EditionID: 4, CurrentPage: intptr(50), Status: "in_progress"})
	assertReadingCode(t, e, "EXPLICIT_EDITION_CHANGE_REQUIRED")
	_, e = s.PutShelf(ctx, a, 7, schemas.ShelfInput{EditionID: 4, CurrentPage: intptr(50), Status: "in_progress", ChangeKind: "edition_change"})
	if e != nil {
		t.Fatal(e)
	}
	if f.events[len(f.events)-1].CreditedPages != 0 || f.record.FinishedAt != nil {
		t.Fatal("troca de edição criou crédito ou manteve conclusão")
	}
}
func TestDiaryDraftPublishOnceAndSubtypeFailureRollsBack(t *testing.T) {
	f := newReadingFake()
	s := testReading(f)
	a := testReader()
	ctx := context.Background()
	input := schemas.PostInput{EditionID: 3, WorkID: 7, Text: "Diário na página vinte", CurrentPage: intptr(20), ContainsSpoiler: boolptr(false), Status: "draft"}
	v, e := s.CreatePost(ctx, a, "diary", input)
	if e != nil {
		t.Fatal(e)
	}
	if f.record != nil || len(f.events) != 0 {
		t.Fatal("rascunho alterou estante")
	}
	published := "published"
	_, e = s.PatchPost(ctx, a, v.ID, "diary", schemas.PostPatch{Status: &published})
	if e != nil {
		t.Fatal(e)
	}
	if f.record == nil || f.record.CurrentPage != 20 || len(f.events) != 1 {
		t.Fatal("publicação não aplicou avanço")
	}
	_, e = s.PatchPost(ctx, a, v.ID, "diary", schemas.PostPatch{Status: &published})
	if e != nil {
		t.Fatal(e)
	}
	if len(f.events) != 1 {
		t.Fatal("republicação duplicou eventos")
	}
	text := "Texto corrigido"
	_, e = s.PatchPost(ctx, a, v.ID, "diary", schemas.PostPatch{Text: &text, CurrentPage: intptr(15)})
	if e != nil {
		t.Fatal(e)
	}
	if f.record.CurrentPage != 20 {
		t.Fatal("edição do snapshot reescreveu estante")
	}
	f2 := newReadingFake()
	f2.failDiary = true
	s2 := testReading(f2)
	input.Status = "published"
	_, e = s2.CreatePost(ctx, a, "diary", input)
	if e == nil {
		t.Fatal("falha no subtipo aceita")
	}
	if len(f2.posts) != 0 || f2.record != nil {
		t.Fatal("operação composta fora da unidade transacional")
	}
}
func TestDiaryRetrospectiveSnapshotDoesNotInventCurrentMonthPages(t *testing.T) {
	f := newReadingFake()
	s := testReading(f)
	yesterday := s.now().AddDate(0, 0, -1)
	_, e := s.CreatePost(context.Background(), testReader(), "diary", schemas.PostInput{EditionID: 3, Text: "Trecho retrospectivo", CurrentPage: intptr(35), ContainsSpoiler: boolptr(false), OccurredAt: &yesterday})
	if e != nil {
		t.Fatal(e)
	}
	if f.record.CurrentPage != 35 || len(f.events) != 1 || f.events[0].CreditedPages != 0 {
		t.Fatal("snapshot retrospectivo criou páginas no mês atual")
	}
}
func TestReviewRequiresProgressRatingAndAllowsRecreationOnlyAfterOwnDelete(t *testing.T) {
	f := newReadingFake()
	s := testReading(f)
	a := testReader()
	ctx := context.Background()
	input := schemas.PostInput{WorkID: 7, Text: "Boa leitura", Rating: intptr(4), ContainsSpoiler: boolptr(false)}
	_, e := s.CreatePost(ctx, a, "review", input)
	assertReadingCode(t, e, "READING_REQUIRED")
	f.record = &core.ReadingRecord{ID: 1, ReaderID: a.ID, WorkID: 7, EditionID: 3, Status: "in_progress"}
	input.Rating = intptr(6)
	_, e = s.CreatePost(ctx, a, "review", input)
	assertReadingCode(t, e, "INVALID_REVIEW")
	input.Rating = intptr(4)
	input.ContainsSpoiler = nil
	_, e = s.CreatePost(ctx, a, "review", input)
	assertReadingCode(t, e, "SPOILER_FLAG_REQUIRED")
	input.ContainsSpoiler = boolptr(false)
	v, e := s.CreatePost(ctx, a, "review", input)
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.CreatePost(ctx, a, "review", input)
	assertReadingCode(t, e, "CONFLICT")
	if e = s.DeletePost(ctx, a, v.ID, "review"); e != nil {
		t.Fatal(e)
	}
	v2, e := s.CreatePost(ctx, a, "review", input)
	if e != nil || v2.ID == v.ID {
		t.Fatalf("avaliação excluída não liberou nova identidade: %v", e)
	}
	p := f.posts[v2.ID]
	p.Status = "removed"
	f.posts[p.ID] = p
	e = s.DeletePost(ctx, a, p.ID, "review")
	assertReadingCode(t, e, "CONFLICT")
}
func TestDynamicAgeAndSpoilerPolicyCoversPostsCommentsAndLikes(t *testing.T) {
	f := newReadingFake()
	s := testReading(f)
	a := testReader()
	p := core.ReadingPost{ID: 1, ReaderID: 2, WorkID: 7, Kind: "review", Status: "published", Title: "Título protegido", Text: "Final secreto", ContainsSpoiler: true, ContainsSensitiveContent: true}
	f.posts[1] = p
	v, e := s.GetPost(context.Background(), a, 1, "", schemas.Reveal{})
	if e != nil {
		t.Fatal(e)
	}
	if v.Text != nil || v.Title != nil || !v.ContentHidden {
		t.Fatal("conteúdo protegido vazou")
	}
	v, e = s.GetPost(context.Background(), a, 1, "", schemas.Reveal{Spoilers: true, Sensitive: true})
	if e != nil || v.Text == nil {
		t.Fatal("confirmação explícita não revelou conteúdo")
	}
	f.edition.MinimumAge = 18
	youngBirth := time.Date(2012, 11, 1, 0, 0, 0, 0, time.UTC)
	a.BirthDate = &youngBirth
	_, e = s.GetPost(context.Background(), a, 1, "", schemas.Reveal{Spoilers: true, Sensitive: true})
	assertReadingCode(t, e, "NOT_FOUND")
	e = s.Like(context.Background(), a, "post", 1, true)
	assertReadingCode(t, e, "NOT_FOUND")
	_, e = s.CreateComment(context.Background(), a, 1, schemas.CommentInput{Text: "Comentário"})
	assertReadingCode(t, e, "NOT_FOUND")
}
func TestCommentParentDepthOwnershipAndIdempotentLikeNotifications(t *testing.T) {
	f := newReadingFake()
	s := testReading(f)
	a := testReader()
	ctx := context.Background()
	f.posts[100] = core.ReadingPost{ID: 100, ReaderID: 2, WorkID: 7, Kind: "diary", Status: "published", Text: "Entrada"}
	parent := uint64(20)
	f.comments[parent] = core.ReadingComment{ID: parent, PostID: 999, ReaderID: 2, Status: "published", AuthorStatus: "active"}
	_, e := s.CreateComment(ctx, a, 100, schemas.CommentInput{Text: "Resposta", ParentID: &parent})
	assertReadingCode(t, e, "INVALID_COMMENT_PARENT")
	p := f.comments[parent]
	p.PostID = 100
	p.Depth = 5
	f.comments[parent] = p
	_, e = s.CreateComment(ctx, a, 100, schemas.CommentInput{Text: "Resposta", ParentID: &parent})
	assertReadingCode(t, e, "COMMENT_DEPTH_EXCEEDED")
	p.Depth = 0
	f.comments[parent] = p
	c, e := s.CreateComment(ctx, a, 100, schemas.CommentInput{Text: "Resposta", ParentID: &parent})
	if e != nil {
		t.Fatal(e)
	}
	other := a
	other.ID = 8
	_, e = s.PatchComment(ctx, other, c.ID, schemas.CommentPatch{Text: "Invasão"})
	assertReadingCode(t, e, "FORBIDDEN")
	before := len(f.notifications)
	if e = s.Like(ctx, a, "post", 100, true); e != nil {
		t.Fatal(e)
	}
	if e = s.Like(ctx, a, "post", 100, true); e != nil {
		t.Fatal(e)
	}
	if len(f.notifications) != before+1 {
		t.Fatal("PUT repetido duplicou notificação")
	}
}
func TestReadingStatisticsUsesOwnReaderAndFillsZeroPeriods(t *testing.T) {
	f := newReadingFake()
	s := testReading(f)
	a := testReader()
	a.ID = 42
	v, e := s.Statistics(context.Background(), a, 2026)
	if e != nil {
		t.Fatal(e)
	}
	if f.statReader != 42 || len(v.Months) != 12 || v.Months[1].Pages != 0 || v.YearTotal.Pages != 35 || v.YearTotal.Works != 1 || len(v.Statuses) != 5 {
		t.Fatalf("estatísticas incorretas: %+v", v)
	}
}
func TestDeterministicContentFilter(t *testing.T) {
	s := NewReading(newReadingFake(), ReadingOptions{BlockedTerms: []string{"termo restrito"}, BlockedDomains: []string{"malicioso.test"}})
	for _, text := range []string{"Um termo restrito neste texto", "Veja https://sub.malicioso.test/arquivo", "javascript:alert(1)"} {
		if e := s.validateText(text); e == nil {
			t.Fatalf("filtro aceitou %q", text)
		}
	}
	if e := s.validateText("Texto com https://example.test/livro válido"); e != nil {
		t.Fatal(e)
	}
}
