//go:build integration

package tools_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/schemas"
	"radar-literario/backend/internal/model/services"
	"radar-literario/backend/internal/model/tools"
)

// The hook commits a competing operation on another pool connection after the
// first transaction has read its snapshot and before it attempts its write.
type adminAfterTarget struct {
	interfaces.AdminRepository
	after func() error
}

func (r adminAfterTarget) Within(ctx context.Context, fn func(interfaces.AdminRepository) error) error {
	return r.AdminRepository.Within(ctx, func(tx interfaces.AdminRepository) error { return fn(adminAfterTarget{tx, r.after}) })
}
func (r adminAfterTarget) Target(ctx context.Context, kind string, id uint64) (core.ReportTarget, error) {
	v, e := r.AdminRepository.Target(ctx, kind, id)
	if e == nil {
		e = r.after()
	}
	return v, e
}
func onceError(fn func() error) func() error {
	var once sync.Once
	var err error
	return func() error { once.Do(func() { err = fn() }); return err }
}

func TestMariaDBModerationCannotBeUndoneByStaleAuthorWrite(t *testing.T) {
	f := setupReadingFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	post, e := f.reading.CreatePost(ctx, f.reader, "diary", schemas.PostInput{WorkID: f.work, EditionID: f.edition, CurrentPage: readingInt(20), Text: "Texto original protegido pela moderação", ContainsSpoiler: readingBool(false)})
	if e != nil {
		t.Fatal(e)
	}
	adminRepo := tools.NewAdminRepository(f.store)
	repo := tools.NewReadingRepository(f.store)
	e = repo.Within(ctx, func(tx interfaces.ReadingRepository) error {
		stale, err := tx.Post(ctx, post.ID, f.reader.ID)
		if err != nil {
			return err
		}
		if err = adminRepo.Moderate(ctx, "post", post.ID, "hidden", time.Now().UTC()); err != nil {
			return err
		}
		stale.Text = "Não deve republicar após ocultação"
		stale.CurrentPage = readingInt(30)
		if err = tx.SaveDiary(ctx, stale); err != nil {
			return err
		}
		return tx.SavePost(ctx, stale)
	})
	integrationCode(t, e, "CONFLICT")
	var saved struct{ Status, Text string }
	if e = f.db.Raw("SELECT status,text FROM posts WHERE id=?", post.ID).Scan(&saved).Error; e != nil {
		t.Fatal(e)
	}
	if saved.Status != "hidden" || saved.Text != "Texto original protegido pela moderação" {
		t.Fatalf("autor desfez moderação: %#v", saved)
	}
	var page int
	if e = f.db.Raw("SELECT current_page FROM reading_diaries WHERE post_id=?", post.ID).Scan(&page).Error; e != nil || page != 20 {
		t.Fatal("falha do CAS não reverteu atualização do subtipo")
	}
	if e = adminRepo.Moderate(ctx, "post", post.ID, "published", time.Now().UTC()); e != nil {
		t.Fatal(e)
	}
	comment, e := f.reading.CreateComment(ctx, f.other, post.ID, schemas.CommentInput{Text: "Comentário original"})
	if e != nil {
		t.Fatal(e)
	}
	e = repo.Within(ctx, func(tx interfaces.ReadingRepository) error {
		stale, err := tx.Comment(ctx, comment.ID, f.other.ID)
		if err != nil {
			return err
		}
		if err = adminRepo.Moderate(ctx, "comment", comment.ID, "removed", time.Now().UTC()); err != nil {
			return err
		}
		stale.Text = "Não deve restaurar comentário removido"
		return tx.SaveComment(ctx, stale)
	})
	integrationCode(t, e, "CONFLICT")
	if e = f.db.Raw("SELECT status,text FROM comments WHERE id=?", comment.ID).Scan(&saved).Error; e != nil {
		t.Fatal(e)
	}
	if saved.Status != "removed" || saved.Text != "Comentário original" {
		t.Fatal("autor desfez remoção do comentário")
	}
}

func TestMariaDBStaleModerationCannotResurrectDeletedContent(t *testing.T) {
	f := setupReadingFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	post, e := f.reading.CreatePost(ctx, f.reader, "diary", schemas.PostInput{WorkID: f.work, EditionID: f.edition, CurrentPage: readingInt(20), Text: "Diário a excluir", ContainsSpoiler: readingBool(false)})
	if e != nil {
		t.Fatal(e)
	}
	moderator := f.other
	moderator.Role = "user_manager"
	admin := services.NewAdmin(adminAfterTarget{tools.NewAdminRepository(f.store), onceError(func() error { return f.reading.DeletePost(ctx, f.reader, post.ID, "diary") })})
	e = admin.Moderate(ctx, moderator, "post", post.ID, schemas.Moderation{Status: "hidden", Reason: "Ação administrativa concorrente"})
	integrationCode(t, e, "CONFLICT")
	var status string
	if e = f.db.Raw("SELECT status FROM posts WHERE id=?", post.ID).Scan(&status).Error; e != nil || status != "deleted" {
		t.Fatal("moderação ressuscitou diário excluído")
	}
	other, e := f.reading.CreatePost(ctx, f.reader, "diary", schemas.PostInput{WorkID: f.work, EditionID: f.edition, CurrentPage: readingInt(20), Text: "Diário para comentário", ContainsSpoiler: readingBool(false)})
	if e != nil {
		t.Fatal(e)
	}
	comment, e := f.reading.CreateComment(ctx, f.other, other.ID, schemas.CommentInput{Text: "Comentário a excluir"})
	if e != nil {
		t.Fatal(e)
	}
	_, e = f.reading.CreateComment(ctx, f.reader, other.ID, schemas.CommentInput{Text: "Resposta que preserva o tombstone", ParentID: &comment.ID})
	if e != nil {
		t.Fatal(e)
	}
	admin = services.NewAdmin(adminAfterTarget{tools.NewAdminRepository(f.store), onceError(func() error { return f.reading.DeleteComment(ctx, f.other, comment.ID) })})
	e = admin.Moderate(ctx, moderator, "comment", comment.ID, schemas.Moderation{Status: "published", Reason: "Restauração administrativa concorrente"})
	integrationCode(t, e, "CONFLICT")
	if e = f.db.Raw("SELECT status FROM comments WHERE id=?", comment.ID).Scan(&status).Error; e != nil || status != "deleted" {
		t.Fatal("moderação restaurou comentário tombstone")
	}
}

func TestMariaDBEditionRoundTripDoesNotCreditPagesAgain(t *testing.T) {
	f := setupReadingFixture(t)
	ctx := context.Background()
	steps := []struct {
		edition uint64
		page    int
		kind    string
	}{{f.edition, 100, "advance"}, {f.secondEdition, 0, "edition_change"}, {f.edition, 0, "edition_change"}, {f.edition, 100, "advance"}, {f.secondEdition, 110, "edition_change"}, {f.secondEdition, 120, "advance"}}
	for _, step := range steps {
		if _, e := f.reading.PutShelf(ctx, f.reader, f.work, schemas.ShelfInput{EditionID: step.edition, CurrentPage: readingInt(step.page), Status: "in_progress", ChangeKind: step.kind}); e != nil {
			t.Fatal(e)
		}
	}
	var credit int
	if e := f.db.Raw("SELECT COALESCE(SUM(credited_pages),0) FROM reading_progress_events WHERE reader_id=?", f.reader.ID).Scan(&credit).Error; e != nil {
		t.Fatal(e)
	}
	if credit != 110 {
		t.Fatalf("troca de edição criou crédito artificial: %d", credit)
	}
	var peak int
	if e := f.db.Raw("SELECT credited_peak FROM reading_records WHERE reader_id=?", f.reader.ID).Scan(&peak).Error; e != nil || peak != 120 {
		t.Fatalf("máximo global do registro perdido: %d, %v", peak, e)
	}
}
