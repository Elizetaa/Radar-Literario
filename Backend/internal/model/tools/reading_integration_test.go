//go:build integration

package tools_test

import (
	"context"
	"errors"
	"fmt"
	gomysql "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"path/filepath"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/schemas"
	"radar-literario/backend/internal/model/services"
	"radar-literario/backend/internal/model/tools"
	"strings"
	"sync"
	"testing"
	"time"
)

type readingFixture struct {
	db                           *gorm.DB
	store                        interfaces.Store
	reading                      *services.Reading
	reader, other                core.Actor
	work, edition, secondEdition uint64
}

func readingTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		values, e := godotenv.Read(filepath.Join("..", "..", "..", ".env.test"))
		if e != nil {
			t.Skip("configure TEST_DB_DSN ou Backend/.env.test para MariaDB isolado")
		}
		if !strings.Contains(values["DB_NAME"], "test") {
			t.Fatal("DB_NAME dos testes deve conter test")
		}
		cfg := gomysql.NewConfig()
		cfg.User = values["DB_USER"]
		cfg.Passwd = values["DB_PASSWORD"]
		cfg.Net = "tcp"
		cfg.Addr = values["DB_HOST"] + ":" + values["DB_PORT"]
		cfg.DBName = values["DB_NAME"]
		cfg.ParseTime = true
		cfg.Loc = time.UTC
		cfg.Params = map[string]string{"charset": "utf8mb4", "time_zone": "'+00:00'"}
		dsn = cfg.FormatDSN()
	}
	parsed, e := gomysql.ParseDSN(dsn)
	if e != nil || !strings.Contains(parsed.DBName, "test") {
		t.Fatal("use exclusivamente um banco de testes com test no nome")
	}
	db, e := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if e != nil {
		t.Fatalf("MariaDB de testes indisponível: %T", e)
	}
	pool, e := db.DB()
	if e != nil {
		t.Fatal(e)
	}
	pool.SetMaxOpenConns(12)
	t.Cleanup(func() { pool.Close() })
	return db
}
func setupReadingFixture(t *testing.T) *readingFixture {
	t.Helper()
	db := readingTestDatabase(t)
	store := tools.NewStore(db)
	ctx := context.Background()
	suffix := fmt.Sprintf("r%d", time.Now().UnixNano())
	born := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	insert := func(table string, values map[string]any) uint64 {
		id, e := store.Insert(ctx, table, values)
		if e != nil {
			t.Fatalf("fixture %s: %v", table, e)
		}
		return id
	}
	user := func(n string) uint64 {
		id := insert("users", map[string]any{"name": "Leitor de integração", "nickname": suffix + n, "email": suffix + n + "@example.test", "password_hash": "hash-apenas-fixture-nao-autenticavel", "birth_date": born, "status": "active"})
		insert("readers", map[string]any{"user_id": id, "profile_visibility": "public"})
		return id
	}
	a, b := user("a"), user("b")
	publisher := insert("publishers", map[string]any{"name": "Editora integração"})
	work := insert("works", map[string]any{"title": "Obra integração " + suffix, "synopsis": "Sinopse fictícia", "original_language": "pt", "minimum_age": 0, "status": "published"})
	edition := insert("editions", map[string]any{"work_id": work, "publisher_id": publisher, "title": "Edição integração", "isbn_absence_reason": "Material fictício de testes", "edition_number": 1, "page_count": 100, "language": "pt", "minimum_age": 0, "status": "published"})
	second := insert("editions", map[string]any{"work_id": work, "publisher_id": publisher, "title": "Segunda edição integração", "isbn_absence_reason": "Material fictício de testes", "edition_number": 2, "page_count": 120, "language": "en", "minimum_age": 0, "status": "published"})
	t.Cleanup(func() {
		queries := []struct {
			sql  string
			args []any
		}{
			{"DELETE FROM notifications WHERE recipient_id IN (?,?) OR actor_id IN (?,?)", []any{a, b, a, b}},
			{"DELETE FROM comment_likes WHERE reader_id IN (?,?)", []any{a, b}},
			{"DELETE FROM post_likes WHERE reader_id IN (?,?)", []any{a, b}},
			{"DELETE FROM comments WHERE reader_id IN (?,?) ORDER BY depth DESC,id DESC", []any{a, b}},
			{"DELETE FROM reviews WHERE reader_id IN (?,?)", []any{a, b}},
			{"DELETE FROM reading_diaries WHERE post_id IN(SELECT id FROM posts WHERE reader_id IN (?,?))", []any{a, b}},
			{"DELETE FROM posts WHERE reader_id IN (?,?)", []any{a, b}},
			{"DELETE FROM reading_records WHERE reader_id IN (?,?)", []any{a, b}},
			{"DELETE FROM editions WHERE work_id=?", []any{work}},
			{"DELETE FROM works WHERE id=?", []any{work}},
			{"DELETE FROM publishers WHERE id=?", []any{publisher}},
			{"DELETE FROM readers WHERE user_id IN (?,?)", []any{a, b}},
			{"DELETE FROM users WHERE id IN (?,?)", []any{a, b}},
		}
		for _, q := range queries {
			if e := db.Exec(q.sql, q.args...).Error; e != nil {
				t.Errorf("limpeza fixture: %v", e)
			}
		}
	})
	return &readingFixture{db: db, store: store, reading: services.NewReading(tools.NewReadingRepository(store), services.ReadingOptions{}), reader: core.Actor{ID: a, Reader: true, BirthDate: &born}, other: core.Actor{ID: b, Reader: true, BirthDate: &born}, work: work, edition: edition, secondEdition: second}
}
func readingInt(v int) *int    { return &v }
func readingBool(v bool) *bool { return &v }
func integrationCode(t *testing.T, e error, code string) {
	t.Helper()
	var de *core.Error
	if !errors.As(e, &de) || de.Code != code {
		t.Fatalf("esperado %s, recebido %v", code, e)
	}
}
func TestMariaDBReadingLifecycleConcurrencyAndPrivacy(t *testing.T) {
	f := setupReadingFixture(t)
	ctx := context.Background()
	s := f.reading
	draft, e := s.CreatePost(ctx, f.reader, "diary", schemas.PostInput{WorkID: f.work, EditionID: f.edition, Text: "Leitura de vinte páginas", CurrentPage: readingInt(20), ContainsSpoiler: readingBool(true), Status: "draft"})
	if e != nil {
		t.Fatal(e)
	}
	var count int64
	f.db.Raw("SELECT COUNT(*) FROM reading_records WHERE reader_id=?", f.reader.ID).Scan(&count)
	if count != 0 {
		t.Fatal("rascunho criou estante")
	}
	published := "published"
	if _, e = s.PatchPost(ctx, f.reader, draft.ID, "diary", schemas.PostPatch{Status: &published}); e != nil {
		t.Fatal(e)
	}
	if _, e = s.PatchPost(ctx, f.reader, draft.ID, "diary", schemas.PostPatch{Status: &published}); e != nil {
		t.Fatal(e)
	}
	f.db.Raw("SELECT COUNT(*) FROM reading_progress_events WHERE reader_id=?", f.reader.ID).Scan(&count)
	if count != 1 {
		t.Fatalf("publicação duplicou eventos: %d", count)
	}
	_, e = s.PatchPost(ctx, f.other, draft.ID, "diary", schemas.PostPatch{Text: func() *string { v := "Invasão"; return &v }()})
	integrationCode(t, e, "FORBIDDEN")
	comment, e := s.CreateComment(ctx, f.other, draft.ID, schemas.CommentInput{Text: "Comentário sobre trecho com spoiler"})
	if e != nil {
		t.Fatal(e)
	}
	if comment.Text != nil {
		t.Fatal("comentário expôs spoiler sem confirmação")
	}
	var wg sync.WaitGroup
	errs := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); errs <- s.Like(ctx, f.other, "post", draft.ID, true) }()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Errorf("curtida concorrente: %v", e)
		}
	}
	f.db.Raw("SELECT COUNT(*) FROM post_likes WHERE reader_id=? AND post_id=?", f.other.ID, draft.ID).Scan(&count)
	if count != 1 {
		t.Fatalf("curtidas duplicadas: %d", count)
	}
	f.db.Raw("SELECT COUNT(*) FROM notifications WHERE recipient_id=? AND event='post_like' AND post_id=?", f.reader.ID, draft.ID).Scan(&count)
	if count != 1 {
		t.Fatalf("notificações duplicadas: %d", count)
	}
	if e = s.Like(ctx, f.reader, "comment", comment.ID, true); e != nil {
		t.Fatal(e)
	}
	input := schemas.PostInput{WorkID: f.work, Text: "Avaliação concorrente", Rating: readingInt(5), ContainsSpoiler: readingBool(false)}
	reviews := make(chan error, 8)
	ids := make(chan uint64, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, e := s.CreatePost(ctx, f.reader, "review", input)
			reviews <- e
			if e == nil {
				ids <- v.ID
			}
		}()
	}
	wg.Wait()
	close(reviews)
	close(ids)
	success := 0
	for e := range reviews {
		if e == nil {
			success++
		} else {
			integrationCode(t, e, "CONFLICT")
		}
	}
	if success != 1 {
		t.Fatalf("avaliações concorrentes criadas: %d", success)
	}
	var reviewID uint64
	for id := range ids {
		reviewID = id
	}
	if e = s.DeletePost(ctx, f.reader, reviewID, "review"); e != nil {
		t.Fatal(e)
	}
	recreated, e := s.CreatePost(ctx, f.reader, "review", input)
	if e != nil || recreated.ID == reviewID {
		t.Fatalf("recriação: %v", e)
	}
	if e = f.db.Exec("UPDATE works SET minimum_age=18 WHERE id=?", f.work).Error; e != nil {
		t.Fatal(e)
	}
	young := f.other
	born := time.Now().UTC().AddDate(-12, 0, 0)
	young.BirthDate = &born
	_, e = s.GetPost(ctx, young, draft.ID, "diary", schemas.Reveal{Spoilers: true, Sensitive: true})
	integrationCode(t, e, "NOT_FOUND")
	_, e = s.Comments(ctx, young, draft.ID, nil, schemas.Page{Page: 1, PageSize: 20}, schemas.Reveal{Spoilers: true})
	integrationCode(t, e, "NOT_FOUND")
	e = s.Like(ctx, young, "comment", comment.ID, true)
	integrationCode(t, e, "NOT_FOUND")
	n, e := s.Notifications(ctx, young, nil, schemas.Page{Page: 1, PageSize: 20})
	if e != nil {
		t.Fatal(e)
	}
	if n.Pagination.Total == 0 {
		t.Fatal("fixture sem notificação para verificar idade")
	}
	for _, v := range n.Data.([]schemas.Notification) {
		if v.TargetAvailable || v.PostID != nil || v.CommentID != nil || v.ActorID != nil {
			t.Fatal("notificação revelou alvo acima da idade")
		}
	}
	if e = f.db.Exec("UPDATE works SET minimum_age=0 WHERE id=?", f.work).Error; e != nil {
		t.Fatal(e)
	}
	f.db.Exec("UPDATE posts SET status='removed' WHERE id=?", recreated.ID)
	e = s.DeletePost(ctx, f.reader, recreated.ID, "review")
	integrationCode(t, e, "CONFLICT")
	_, e = s.CreatePost(ctx, f.reader, "review", input)
	integrationCode(t, e, "CONFLICT")
}

func TestMariaDBShelfConcurrencyCommentTreeAndUTCPeriods(t *testing.T) {
	f := setupReadingFixture(t)
	ctx := context.Background()
	s := f.reading
	var wg sync.WaitGroup
	errs := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, e := s.PutShelf(ctx, f.reader, f.work, schemas.ShelfInput{EditionID: f.edition, CurrentPage: readingInt(35), Status: "in_progress"})
			errs <- e
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	var count int64
	f.db.Raw("SELECT COUNT(*) FROM reading_records WHERE reader_id=?", f.reader.ID).Scan(&count)
	if count != 1 {
		t.Fatalf("estante concorrente duplicada: %d", count)
	}
	f.db.Raw("SELECT COUNT(*) FROM reading_progress_events WHERE reader_id=?", f.reader.ID).Scan(&count)
	if count != 1 {
		t.Fatalf("avanço repetido criou eventos: %d", count)
	}
	diary, e := s.CreatePost(ctx, f.reader, "diary", schemas.PostInput{WorkID: f.work, EditionID: f.edition, Text: "Diário para árvore", CurrentPage: readingInt(35), ContainsSpoiler: readingBool(false)})
	if e != nil {
		t.Fatal(e)
	}
	parent, e := s.CreateComment(ctx, f.other, diary.ID, schemas.CommentInput{Text: "Comentário pai"})
	if e != nil {
		t.Fatal(e)
	}
	child, e := s.CreateComment(ctx, f.reader, diary.ID, schemas.CommentInput{Text: "Resposta filha", ParentID: &parent.ID})
	if e != nil {
		t.Fatal(e)
	}
	if e = f.db.Exec("UPDATE comments SET parent_id=? WHERE id=?", child.ID, parent.ID).Error; e == nil {
		t.Fatal("trigger permitiu reparent/ciclo")
	}
	if e = s.DeleteComment(ctx, f.other, parent.ID); e != nil {
		t.Fatal(e)
	}
	list, e := s.Comments(ctx, f.reader, diary.ID, nil, schemas.Page{Page: 1, PageSize: 20}, schemas.Reveal{})
	if e != nil {
		t.Fatal(e)
	}
	rows := list.Data.([]schemas.Comment)
	if len(rows) != 1 || rows[0].Text != nil || rows[0].Status != "deleted" {
		t.Fatalf("tombstone incorreto: %+v", rows)
	}
	replies, e := s.Comments(ctx, f.reader, diary.ID, &parent.ID, schemas.Page{Page: 1, PageSize: 20}, schemas.Reveal{})
	if e != nil || replies.Pagination.Total != 1 {
		t.Fatalf("árvore perdeu resposta: %v", e)
	}
	_, e = s.PutShelf(ctx, f.reader, f.work, schemas.ShelfInput{EditionID: f.edition, CurrentPage: readingInt(100), Status: "completed"})
	if e != nil {
		t.Fatal(e)
	}
	if e = f.db.Exec("UPDATE reading_records SET started_at='2025-01-01 00:00:00',finished_at='2025-12-31 23:59:59.999999' WHERE reader_id=?", f.reader.ID).Error; e != nil {
		t.Fatal(e)
	}
	var events []struct{ ID uint64 }
	if e = f.db.Raw("SELECT id FROM reading_progress_events WHERE reader_id=? ORDER BY id ASC", f.reader.ID).Scan(&events).Error; e != nil {
		t.Fatal(e)
	}
	if len(events) != 2 {
		t.Fatalf("eventos de fixture: %d", len(events))
	}
	f.db.Exec("UPDATE reading_progress_events SET created_at='2025-01-01 00:00:00' WHERE id=?", events[0].ID)
	f.db.Exec("UPDATE reading_progress_events SET created_at='2026-01-01 00:00:00' WHERE id=?", events[1].ID)
	stats, e := s.Statistics(ctx, f.reader, 2025)
	if e != nil {
		t.Fatal(e)
	}
	if stats.YearTotal.Pages != 35 || stats.YearTotal.Works != 1 || stats.Months[0].Pages != 35 || stats.Months[11].Works != 1 || stats.MostReadLanguage == nil || *stats.MostReadLanguage != "pt" {
		t.Fatalf("intervalo UTC 2025 incorreto: %+v", stats)
	}
	next, e := s.Statistics(ctx, f.reader, 2026)
	if e != nil || next.YearTotal.Pages != 65 || next.YearTotal.Works != 0 {
		t.Fatalf("limite exclusivo anual incorreto: %v %+v", e, next)
	}
}

type failingDiaryRepository struct{ interfaces.ReadingRepository }

func (r failingDiaryRepository) Within(ctx context.Context, fn func(interfaces.ReadingRepository) error) error {
	return r.ReadingRepository.Within(ctx, func(tx interfaces.ReadingRepository) error { return fn(failingDiaryRepository{tx}) })
}
func (r failingDiaryRepository) CreateDiary(context.Context, core.ReadingPost) error {
	return errors.New("falha controlada após criar post")
}
func TestMariaDBReadingRollbackConstraintsAndStatistics(t *testing.T) {
	f := setupReadingFixture(t)
	ctx := context.Background()
	repo := tools.NewReadingRepository(f.store)
	s := services.NewReading(failingDiaryRepository{repo}, services.ReadingOptions{})
	_, e := s.CreatePost(ctx, f.reader, "diary", schemas.PostInput{WorkID: f.work, EditionID: f.edition, Text: "Operação atômica", CurrentPage: readingInt(20), ContainsSpoiler: readingBool(false)})
	if e == nil {
		t.Fatal("falha controlada não propagou")
	}
	var count int64
	f.db.Raw("SELECT COUNT(*) FROM posts WHERE reader_id=?", f.reader.ID).Scan(&count)
	if count != 0 {
		t.Fatal("post persistiu após falha do subtipo")
	}
	f.db.Raw("SELECT COUNT(*) FROM reading_records WHERE reader_id=?", f.reader.ID).Scan(&count)
	if count != 0 {
		t.Fatal("estante persistiu após rollback")
	}
	for _, step := range []struct {
		page int
		kind string
	}{{20, "advance"}, {35, "advance"}, {10, "correction"}, {35, "advance"}} {
		_, e = f.reading.PutShelf(ctx, f.reader, f.work, schemas.ShelfInput{EditionID: f.edition, CurrentPage: readingInt(step.page), Status: "in_progress", ChangeKind: step.kind})
		if e != nil {
			t.Fatal(e)
		}
	}
	_, e = f.reading.PutShelf(ctx, f.other, f.work, schemas.ShelfInput{EditionID: f.edition, CurrentPage: readingInt(80), Status: "in_progress"})
	if e != nil {
		t.Fatal(e)
	}
	stats, e := f.reading.Statistics(ctx, f.reader, time.Now().UTC().Year())
	if e != nil {
		t.Fatal(e)
	}
	if stats.YearTotal.Pages != 35 || stats.CurrentPages != 35 || len(stats.Months) != 12 {
		t.Fatalf("estatísticas misturaram contas ou duplicaram snapshots: %+v", stats)
	}
	bad := f.db.Exec("INSERT INTO reading_records(reader_id,work_id,edition_id,current_page,credited_peak,status,created_at,updated_at) VALUES(?,?,?,?,?,'in_progress',UTC_TIMESTAMP(6),UTC_TIMESTAMP(6))", f.reader.ID, f.work, f.edition, 1, 1).Error
	if bad == nil {
		t.Fatal("UNIQUE leitor/obra não aplicado")
	}
	bad = f.db.Exec("UPDATE reading_records SET current_page=-1 WHERE reader_id=?", f.reader.ID).Error
	if bad == nil {
		t.Fatal("CHECK progresso negativo não aplicado")
	}
	diary, e := f.reading.CreatePost(ctx, f.reader, "diary", schemas.PostInput{WorkID: f.work, EditionID: f.edition, Text: "Diário persistido", CurrentPage: readingInt(35), ContainsSpoiler: readingBool(false)})
	if e != nil {
		t.Fatal(e)
	}
	bad = f.db.Exec("INSERT INTO reviews(post_id,reader_id,work_id,rating) VALUES(?,?,?,5)", diary.ID, f.reader.ID, f.work).Error
	if bad == nil {
		t.Fatal("FK tipo do subtipo não aplicada")
	}
	bad = f.db.Exec("INSERT INTO reading_records(reader_id,work_id,edition_id,current_page,credited_peak,status,created_at,updated_at) VALUES(?,?,?,?,?,'planned',UTC_TIMESTAMP(6),UTC_TIMESTAMP(6))", f.reader.ID, 999999999, f.edition, 0, 0).Error
	if bad == nil {
		t.Fatal("FK composta obra/edição não aplicada")
	}
	newService := services.NewReading(tools.NewReadingRepository(tools.NewStore(f.db.Session(&gorm.Session{NewDB: true}))), services.ReadingOptions{})
	v, e := newService.GetPost(ctx, f.reader, diary.ID, "diary", schemas.Reveal{})
	if e != nil || v.ID != diary.ID {
		t.Fatalf("nova composição não recuperou persistência: %v", e)
	}
	if e = f.reading.DeleteShelf(ctx, f.reader, f.work); e != nil {
		t.Fatal(e)
	}
	stats, e = f.reading.Statistics(ctx, f.reader, time.Now().UTC().Year())
	if e != nil || stats.YearTotal.Pages != 0 || stats.CurrentPages != 0 {
		t.Fatalf("remoção de estante não removeu seus eventos: %v %+v", e, stats)
	}
	if _, e = newService.GetPost(ctx, f.reader, diary.ID, "diary", schemas.Reveal{}); e != nil {
		t.Fatal("remoção de estante removeu diário")
	}
}
