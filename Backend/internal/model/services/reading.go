package services

import (
	"context"
	"fmt"
	"net/url"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/schemas"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

type ReadingOptions struct {
	MaxCommentDepth              int
	BlockedTerms, BlockedDomains []string
	Recommendations              func(context.Context, core.Actor, schemas.Page) (schemas.List, error)
	PreferredWorks               func(context.Context, core.Actor, schemas.Page) (schemas.List, error)
}
type Reading struct {
	repo    interfaces.ReadingRepository
	options ReadingOptions
	now     func() time.Time
}

func NewReading(repo interfaces.ReadingRepository, options ReadingOptions) *Reading {
	if options.MaxCommentDepth <= 0 {
		options.MaxCommentDepth = 5
	}
	if options.MaxCommentDepth > 20 {
		options.MaxCommentDepth = 20
	}
	return &Reading{repo: repo, options: options, now: func() time.Time { return time.Now().UTC() }}
}

var contentURL = regexp.MustCompile(`https?://[^\s<>\(\)]+`)

func (s *Reading) validateText(text string) error {
	if strings.TrimSpace(text) == "" || utf8.RuneCountInString(text) > 10000 {
		return core.Invalid("INVALID_CONTENT", "Informe conteúdo entre 1 e 10.000 caracteres.")
	}
	lower := strings.ToLower(text)
	for _, term := range s.options.BlockedTerms {
		if term = strings.TrimSpace(strings.ToLower(term)); term != "" && strings.Contains(lower, term) {
			return core.Invalid("CONTENT_FILTERED", "O conteúdo contém um termo bloqueado. Revise o texto.")
		}
	}
	links := contentURL.FindAllString(text, -1)
	if len(links) > 5 || strings.Contains(lower, "javascript:") || strings.Contains(lower, "file:") || strings.Contains(lower, "data:text/") {
		return core.Invalid("CONTENT_FILTERED", "O conteúdo contém excesso de links ou um esquema não permitido.")
	}
	for _, link := range links {
		u, e := url.Parse(strings.TrimRight(link, ".,;!?"))
		if e != nil || u.Hostname() == "" || u.User != nil {
			return core.Invalid("INVALID_URL", "Há um link inválido no conteúdo.")
		}
		host := strings.ToLower(u.Hostname())
		for _, domain := range s.options.BlockedDomains {
			domain = strings.TrimSpace(strings.ToLower(domain))
			if domain != "" && (host == domain || strings.HasSuffix(host, "."+domain)) {
				return core.Invalid("CONTENT_FILTERED", "O conteúdo contém um domínio bloqueado.")
			}
		}
	}
	// Long runs of one character are a deterministic spam signal, configurable terms cover the rest.
	var previous rune
	run := 0
	for _, char := range text {
		if char == previous {
			run++
		} else {
			run = 1
			previous = char
		}
		if run >= 80 {
			return core.Invalid("CONTENT_FILTERED", "O conteúdo contém repetição excessiva de caracteres.")
		}
	}
	return nil
}
func validReadingStatus(status string) bool {
	switch status {
	case "planned", "in_progress", "paused", "completed", "abandoned":
		return true
	}
	return false
}
func shelfDTO(v core.ReadingRecord) schemas.ShelfItem {
	return schemas.ShelfItem{ID: v.ID, WorkID: v.WorkID, EditionID: v.EditionID, Title: v.WorkTitle, EditionTitle: v.EditionTitle, Language: v.Language, CurrentPage: v.CurrentPage, Status: v.Status, StartedAt: v.StartedAt, FinishedAt: v.FinishedAt, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
func (s *Reading) Shelf(ctx context.Context, a core.Actor, owner uint64, status string, page schemas.Page) (schemas.List, error) {
	if owner == 0 {
		owner = a.ID
	}
	if owner == 0 {
		return schemas.List{}, core.ErrUnauthenticated
	}
	if status != "" && status != "incomplete" && !validReadingStatus(status) {
		return schemas.List{}, core.Invalid("INVALID_STATUS", "Estado de leitura inválido.")
	}
	if owner != a.ID {
		if e := s.repo.ProfileVisible(ctx, owner, a); e != nil {
			return schemas.List{}, e
		}
	}
	rows, total, e := s.repo.Shelf(ctx, owner, a.Age(s.now()), status, page)
	if e != nil {
		return schemas.List{}, e
	}
	data := make([]schemas.ShelfItem, 0, len(rows))
	for _, v := range rows {
		data = append(data, shelfDTO(v))
	}
	return schemas.Paginated(data, page, total), nil
}
func (s *Reading) PutShelf(ctx context.Context, a core.Actor, workID uint64, input schemas.ShelfInput) (schemas.ShelfItem, error) {
	if e := a.RequireReader(); e != nil {
		return schemas.ShelfItem{}, e
	}
	if input.EditionID == 0 || input.CurrentPage == nil || !validReadingStatus(input.Status) {
		return schemas.ShelfItem{}, core.Invalid("INVALID_SHELF", "Informe edition_id, current_page e um status de leitura válido.")
	}
	if input.ChangeKind == "" {
		input.ChangeKind = "advance"
	}
	if input.ChangeKind == "snapshot" {
		return schemas.ShelfItem{}, core.Invalid("INVALID_CHANGE_KIND", "snapshot é reservado aos diários retrospectivos.")
	}
	var saved core.ReadingRecord
	e := s.repo.Within(ctx, func(repo interfaces.ReadingRepository) error {
		if e := repo.LockReader(ctx, a.ID); e != nil {
			return e
		}
		edition, e := repo.Edition(ctx, input.EditionID)
		if e != nil {
			return e
		}
		if e = s.checkEdition(a, edition, workID); e != nil {
			return e
		}
		record, e := repo.Record(ctx, a.ID, workID)
		if e != nil {
			return e
		}
		if record == nil {
			record = &core.ReadingRecord{ReaderID: a.ID, WorkID: workID, EditionID: edition.ID, CreatedAt: s.now()}
		}
		if e = s.updateProgress(ctx, repo, record, edition, input); e != nil {
			return e
		}
		saved = *record
		return nil
	})
	return shelfDTO(saved), e
}
func (s *Reading) checkEdition(a core.Actor, e core.ReadingEdition, workID uint64) error {
	if e.WorkID != workID {
		return core.Invalid("EDITION_WORK_MISMATCH", "A edição deve pertencer à obra informada.")
	}
	if e.Status != "published" || e.WorkStatus != "published" || e.MinimumAge > a.Age(s.now()) {
		return core.NotFound()
	}
	return nil
}
func (s *Reading) updateProgress(ctx context.Context, repo interfaces.ReadingRepository, record *core.ReadingRecord, edition core.ReadingEdition, input schemas.ShelfInput) error {
	if input.CurrentPage == nil {
		return core.Invalid("INVALID_READING_PROGRESS", "Informe current_page.")
	}
	page := *input.CurrentPage
	now := s.now()
	previous := record.CurrentPage
	changedEdition := record.ID != 0 && record.EditionID != edition.ID
	if page < 0 || page > edition.PageCount {
		return core.Invalid("INVALID_READING_PROGRESS", "O progresso não pode ultrapassar o total de páginas da edição.")
	}
	if input.Status == "completed" {
		page = edition.PageCount
	}
	if page < 0 || page > edition.PageCount {
		return core.Invalid("INVALID_READING_PROGRESS", "O progresso não pode ultrapassar o total de páginas da edição.")
	}
	if input.Status == "planned" && page != 0 {
		return core.Invalid("INVALID_READING_PROGRESS", "Uma leitura planejada deve ter página zero.")
	}
	if input.ChangeKind != "advance" && input.ChangeKind != "correction" && input.ChangeKind != "edition_change" && input.ChangeKind != "snapshot" {
		return core.Invalid("INVALID_CHANGE_KIND", "Use advance, correction ou edition_change.")
	}
	if changedEdition && input.ChangeKind != "edition_change" {
		return core.Invalid("EXPLICIT_EDITION_CHANGE_REQUIRED", "Para mudar de edição, informe change_kind=edition_change e o novo progresso.")
	}
	if !changedEdition && input.ChangeKind == "edition_change" {
		return core.Invalid("INVALID_CHANGE_KIND", "edition_change exige uma edição diferente.")
	}
	if page < previous && !changedEdition && input.ChangeKind != "correction" {
		return core.Invalid("EXPLICIT_CORRECTION_REQUIRED", "Para reduzir o progresso, informe change_kind=correction.")
	}
	if input.StartedAt != nil {
		if input.StartedAt.After(now) {
			return core.Invalid("INVALID_READING_DATES", "A data de início não pode estar no futuro.")
		}
		record.StartedAt = input.StartedAt
	}
	if record.StartedAt == nil && input.Status != "planned" {
		record.StartedAt = &now
	}
	if input.Status == "completed" {
		if input.FinishedAt != nil {
			record.FinishedAt = input.FinishedAt
		} else if record.FinishedAt == nil {
			record.FinishedAt = &now
		}
	} else {
		if input.FinishedAt != nil {
			return core.Invalid("INVALID_READING_DATES", "finished_at exige estado completed.")
		}
		record.FinishedAt = nil
	}
	if record.FinishedAt != nil && (record.FinishedAt.After(now) || (record.StartedAt != nil && record.FinishedAt.Before(*record.StartedAt))) {
		return core.Invalid("INVALID_READING_DATES", "A conclusão deve ocorrer entre o início e o instante atual.")
	}
	credit := 0
	if input.ChangeKind == "advance" && !changedEdition && page > record.CreditedPeak {
		credit = page - record.CreditedPeak
	}
	if page > record.CreditedPeak {
		record.CreditedPeak = page
	}
	record.EditionID = edition.ID
	record.CurrentPage = page
	record.Status = input.Status
	record.UpdatedAt = now
	if e := repo.SaveRecord(ctx, record); e != nil {
		return e
	}
	if previous != page || changedEdition || credit > 0 {
		return repo.AddProgress(ctx, core.ProgressEvent{RecordID: record.ID, ReaderID: record.ReaderID, PreviousPage: previous, CurrentPage: page, CreditedPages: credit, ChangeKind: input.ChangeKind, CreatedAt: now})
	}
	return nil
}
func (s *Reading) DeleteShelf(ctx context.Context, a core.Actor, workID uint64) error {
	if e := a.RequireReader(); e != nil {
		return e
	}
	return s.repo.Within(ctx, func(repo interfaces.ReadingRepository) error {
		if e := repo.LockReader(ctx, a.ID); e != nil {
			return e
		}
		return repo.RemoveRecord(ctx, a.ID, workID)
	})
}
func postEligible(a core.Actor, p core.ReadingPost, now time.Time, allowOwnDraft bool) error {
	if p.AuthorStatus != "active" || p.WorkStatus != "published" || max(p.MinimumAge, p.WorkMinimumAge) > a.Age(now) {
		return core.NotFound()
	}
	if p.Status == "published" {
		return nil
	}
	if allowOwnDraft && p.Status == "draft" && p.ReaderID == a.ID && a.ID != 0 {
		return nil
	}
	return core.NotFound()
}
func postDTO(p core.ReadingPost, reveal schemas.Reveal) schemas.Post {
	hidden := (p.ContainsSpoiler && !reveal.Spoilers) || (p.ContainsSensitiveContent && !reveal.Sensitive)
	var title, text, data *string
	if !hidden {
		title = &p.Title
		text = &p.Text
		if p.DataURL != "" {
			data = &p.DataURL
		}
	}
	return schemas.Post{ID: p.ID, PostID: p.ID, ReaderID: p.ReaderID, WorkID: p.WorkID, Kind: p.Kind, Title: title, Text: text, DataURL: data, Status: p.Status, AuthorName: p.AuthorName, AuthorNickname: p.AuthorNickname, ContainsSpoiler: p.ContainsSpoiler, ContainsSensitiveContent: p.ContainsSensitiveContent, ContentHidden: hidden, MinimumAge: max(p.MinimumAge, p.WorkMinimumAge), EditionID: p.EditionID, CurrentPage: p.CurrentPage, Rating: p.Rating, OccurredAt: p.OccurredAt, PublishedAt: p.PublishedAt, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt, LikeCount: p.LikeCount, CommentCount: p.CommentCount, Liked: p.Liked}
}
func (s *Reading) GetPost(ctx context.Context, a core.Actor, id uint64, kind string, reveal schemas.Reveal) (schemas.Post, error) {
	p, e := s.repo.Post(ctx, id, a.ID)
	if e != nil {
		return schemas.Post{}, e
	}
	if kind != "" && kind != p.Kind {
		return schemas.Post{}, core.NotFound()
	}
	if e = postEligible(a, p, s.now(), true); e != nil {
		return schemas.Post{}, e
	}
	return postDTO(p, reveal), nil
}
func (s *Reading) Posts(ctx context.Context, a core.Actor, filter schemas.PostFilter, reveal schemas.Reveal) (schemas.List, error) {
	if filter.Kind != "" && filter.Kind != "diary" && filter.Kind != "review" {
		return schemas.List{}, core.Invalid("INVALID_POST_KIND", "Use kind=diary ou review.")
	}
	if filter.Own && a.ID == 0 {
		return schemas.List{}, core.ErrUnauthenticated
	}
	if filter.ReaderID != 0 && filter.ReaderID != a.ID {
		if e := s.repo.ProfileVisible(ctx, filter.ReaderID, a); e != nil {
			return schemas.List{}, e
		}
	}
	rows, total, e := s.repo.Posts(ctx, a, filter)
	if e != nil {
		return schemas.List{}, e
	}
	data := make([]schemas.Post, 0, len(rows))
	for _, p := range rows {
		data = append(data, postDTO(p, reveal))
	}
	return schemas.Paginated(data, filter.Page, total), nil
}
func (s *Reading) validatePost(p core.ReadingPost) error {
	if e := s.validateText(p.Text); e != nil {
		return e
	}
	if utf8.RuneCountInString(p.Title) > 200 {
		return core.Invalid("INVALID_TITLE", "O título deve ter até 200 caracteres.")
	}
	if e := core.ValidateURL(p.DataURL); e != nil {
		return e
	}
	if !core.ValidAge(p.MinimumAge) {
		return core.Invalid("INVALID_AGE_RATING", "Classificação etária inválida.")
	}
	if p.Status != "draft" && p.Status != "published" {
		return core.Invalid("INVALID_POST_STATUS", "A publicação própria aceita draft ou published.")
	}
	return nil
}
func (s *Reading) CreatePost(ctx context.Context, a core.Actor, kind string, input schemas.PostInput) (schemas.Post, error) {
	if e := a.RequireReader(); e != nil {
		return schemas.Post{}, e
	}
	if input.ContainsSpoiler == nil {
		return schemas.Post{}, core.Invalid("SPOILER_FLAG_REQUIRED", "Informe explicitamente contains_spoiler como true ou false.")
	}
	if kind != "diary" && kind != "review" {
		return schemas.Post{}, core.Invalid("INVALID_POST_KIND", "Tipo inválido.")
	}
	now := s.now()
	status := input.Status
	if status == "" {
		status = "published"
	}
	p := core.ReadingPost{ReaderID: a.ID, WorkID: input.WorkID, Kind: kind, Title: strings.TrimSpace(input.Title), Text: strings.TrimSpace(input.Text), Status: status, ContainsSpoiler: *input.ContainsSpoiler, MinimumAge: input.MinimumAge, CreatedAt: now, UpdatedAt: now}
	if input.DataURL != nil {
		p.DataURL = *input.DataURL
	}
	if input.ContainsSensitiveContent != nil {
		p.ContainsSensitiveContent = *input.ContainsSensitiveContent
	}
	if e := s.validatePost(p); e != nil {
		return schemas.Post{}, e
	}
	if kind == "diary" {
		if input.EditionID == 0 || input.CurrentPage == nil || input.Rating != nil {
			return schemas.Post{}, core.Invalid("INVALID_DIARY", "Diário exige edition_id e current_page e não aceita rating.")
		}
		p.EditionID = &input.EditionID
		p.CurrentPage = input.CurrentPage
		p.OccurredAt = input.OccurredAt
		if p.OccurredAt != nil && p.OccurredAt.After(now) {
			return schemas.Post{}, core.Invalid("INVALID_DIARY_DATE", "O instante do diário não pode estar no futuro.")
		}
	} else {
		if input.WorkID == 0 || input.Rating == nil || *input.Rating < 1 || *input.Rating > 5 || input.EditionID != 0 || input.CurrentPage != nil || input.OccurredAt != nil {
			return schemas.Post{}, core.Invalid("INVALID_REVIEW", "Avaliação exige work_id e rating inteiro de 1 a 5; não aceita progresso de diário.")
		}
		p.Rating = input.Rating
	}
	e := s.repo.Within(ctx, func(repo interfaces.ReadingRepository) error {
		if e := repo.LockReader(ctx, a.ID); e != nil {
			return e
		}
		if kind == "diary" {
			edition, e := repo.Edition(ctx, *p.EditionID)
			if e != nil {
				return e
			}
			if p.WorkID == 0 {
				p.WorkID = edition.WorkID
			}
			if e = s.checkEdition(a, edition, p.WorkID); e != nil {
				return e
			}
			if *p.CurrentPage < 0 || *p.CurrentPage > edition.PageCount {
				return core.Invalid("INVALID_READING_PROGRESS", "Informe uma página dentro dos limites da edição.")
			}
			p.WorkMinimumAge = edition.MinimumAge
			p.WorkStatus = edition.WorkStatus
		} else {
			record, e := repo.Record(ctx, a.ID, p.WorkID)
			if e != nil {
				return e
			}
			if record == nil || (record.Status != "in_progress" && record.Status != "completed") {
				return core.Invalid("READING_REQUIRED", "Avalie uma obra em leitura ou concluída na sua estante.")
			}
			edition, e := repo.Edition(ctx, record.EditionID)
			if e != nil {
				return e
			}
			if e = s.checkEdition(a, edition, p.WorkID); e != nil {
				return e
			}
			p.WorkMinimumAge = edition.MinimumAge
			p.WorkStatus = edition.WorkStatus
		}
		if p.MinimumAge > a.Age(now) {
			return core.Forbidden()
		}
		if status == "published" {
			p.PublishedAt = &now
		}
		if e := repo.CreatePost(ctx, &p); e != nil {
			return e
		}
		if kind == "diary" {
			if e := repo.CreateDiary(ctx, p); e != nil {
				return e
			}
			if status == "published" {
				if e := s.applyDiary(ctx, repo, a, &p); e != nil {
					return e
				}
			}
		} else {
			if e := repo.CreateReview(ctx, p); e != nil {
				return e
			}
		}
		return nil
	})
	if e != nil {
		return schemas.Post{}, e
	}
	return s.GetPost(ctx, a, p.ID, kind, schemas.Reveal{})
}
func (s *Reading) applyDiary(ctx context.Context, repo interfaces.ReadingRepository, a core.Actor, p *core.ReadingPost) error {
	if p.EffectsAppliedAt != nil {
		return nil
	}
	edition, e := repo.Edition(ctx, *p.EditionID)
	if e != nil {
		return e
	}
	if e = s.checkEdition(a, edition, p.WorkID); e != nil {
		return e
	}
	record, e := repo.Record(ctx, a.ID, p.WorkID)
	if e != nil {
		return e
	}
	if record == nil {
		record = &core.ReadingRecord{ReaderID: a.ID, WorkID: p.WorkID, EditionID: edition.ID, CreatedAt: s.now()}
	}
	if record.EditionID != edition.ID {
		return core.Conflict("A edição do diário difere da estante. Altere a edição da estante explicitamente antes de publicar.")
	}
	page := max(record.CurrentPage, *p.CurrentPage)
	status := record.Status
	if status != "completed" {
		status = "in_progress"
	}
	kind := "advance"
	now := s.now()
	if p.OccurredAt != nil && p.OccurredAt.Before(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)) {
		kind = "snapshot"
	}
	if e = s.updateProgress(ctx, repo, record, edition, schemas.ShelfInput{EditionID: edition.ID, CurrentPage: &page, Status: status, ChangeKind: kind}); e != nil {
		return e
	}
	p.EffectsAppliedAt = &now
	return repo.SavePost(ctx, *p)
}
func (s *Reading) PatchPost(ctx context.Context, a core.Actor, id uint64, kind string, input schemas.PostPatch) (schemas.Post, error) {
	if e := a.RequireReader(); e != nil {
		return schemas.Post{}, e
	}
	e := s.repo.Within(ctx, func(repo interfaces.ReadingRepository) error {
		if e := repo.LockReader(ctx, a.ID); e != nil {
			return e
		}
		p, e := repo.Post(ctx, id, a.ID)
		if e != nil {
			return e
		}
		if p.Kind != kind {
			return core.NotFound()
		}
		if p.ReaderID != a.ID {
			return core.Forbidden()
		}
		if p.Status != "published" && p.Status != "draft" {
			return core.Conflict("Uma decisão de moderação precisa ser revisada antes de editar o conteúdo.")
		}
		if e = postEligible(a, p, s.now(), true); e != nil {
			return e
		}
		previousStatus := p.Status
		if input.Title != nil {
			p.Title = strings.TrimSpace(*input.Title)
		}
		if input.Text != nil {
			p.Text = strings.TrimSpace(*input.Text)
		}
		if input.DataURL.Set {
			p.DataURL = ""
			if input.DataURL.Value != nil {
				p.DataURL = *input.DataURL.Value
			}
		}
		if input.Status != nil {
			p.Status = *input.Status
		}
		if input.ContainsSpoiler != nil {
			p.ContainsSpoiler = *input.ContainsSpoiler
		}
		if input.ContainsSensitiveContent != nil {
			p.ContainsSensitiveContent = *input.ContainsSensitiveContent
		}
		if input.MinimumAge != nil {
			p.MinimumAge = *input.MinimumAge
		}
		if p.Status == "draft" && previousStatus == "published" {
			return core.Invalid("INVALID_POST_TRANSITION", "Uma publicação já publicada não retorna a rascunho.")
		}
		if e = s.validatePost(p); e != nil {
			return e
		}
		if p.MinimumAge > a.Age(s.now()) {
			return core.Forbidden()
		}
		if kind == "diary" {
			if input.Rating != nil {
				return core.Invalid("INVALID_DIARY", "Diário não possui nota.")
			}
			if input.CurrentPage != nil {
				edition, e := repo.Edition(ctx, *p.EditionID)
				if e != nil {
					return e
				}
				if *input.CurrentPage < 0 || *input.CurrentPage > edition.PageCount {
					return core.Invalid("INVALID_READING_PROGRESS", "Página inválida para a edição.")
				}
				p.CurrentPage = input.CurrentPage
				if e = repo.SaveDiary(ctx, p); e != nil {
					return e
				}
			}
		} else {
			if input.CurrentPage != nil {
				return core.Invalid("INVALID_REVIEW", "Avaliação não possui progresso.")
			}
			if input.Rating != nil {
				if *input.Rating < 1 || *input.Rating > 5 {
					return core.Invalid("INVALID_RATING", "A nota deve ser inteira entre 1 e 5.")
				}
				p.Rating = input.Rating
				if e = repo.SaveReview(ctx, p); e != nil {
					return e
				}
			}
		}
		now := s.now()
		p.UpdatedAt = now
		if previousStatus == "draft" && p.Status == "published" {
			p.PublishedAt = &now
			if kind == "diary" {
				if e = s.applyDiary(ctx, repo, a, &p); e != nil {
					return e
				}
			} else {
				record, err := repo.Record(ctx, a.ID, p.WorkID)
				if err != nil {
					return err
				}
				if record == nil || (record.Status != "in_progress" && record.Status != "completed") {
					return core.Invalid("READING_REQUIRED", "Publique a avaliação enquanto a obra estiver em leitura ou concluída na sua estante.")
				}
				edition, err := repo.Edition(ctx, record.EditionID)
				if err != nil {
					return err
				}
				if err = s.checkEdition(a, edition, p.WorkID); err != nil {
					return err
				}
			}
		}
		return repo.SavePost(ctx, p)
	})
	if e != nil {
		return schemas.Post{}, e
	}
	return s.GetPost(ctx, a, id, kind, schemas.Reveal{})
}
func (s *Reading) DeletePost(ctx context.Context, a core.Actor, id uint64, kind string) error {
	if e := a.RequireReader(); e != nil {
		return e
	}
	return s.repo.Within(ctx, func(repo interfaces.ReadingRepository) error {
		if e := repo.LockReader(ctx, a.ID); e != nil {
			return e
		}
		p, e := repo.Post(ctx, id, a.ID)
		if e != nil {
			return e
		}
		if p.Kind != kind {
			return core.NotFound()
		}
		if p.ReaderID != a.ID {
			return core.Forbidden()
		}
		if p.Status == "deleted" {
			return nil
		}
		if p.Status == "hidden" || p.Status == "removed" {
			return core.Conflict("O conteúdo está sob moderação e exige revisão administrativa.")
		}
		p.Status = "deleted"
		p.UpdatedAt = s.now()
		if e = repo.SavePost(ctx, p); e != nil {
			return e
		}
		if kind == "review" {
			return repo.ReleaseReview(ctx, p.ID)
		}
		return nil
	})
}

func (s *Reading) Comments(ctx context.Context, a core.Actor, postID uint64, parent *uint64, page schemas.Page, reveal schemas.Reveal) (schemas.List, error) {
	p, e := s.repo.Post(ctx, postID, a.ID)
	if e != nil {
		return schemas.List{}, e
	}
	if e = postEligible(a, p, s.now(), false); e != nil {
		return schemas.List{}, e
	}
	if parent != nil {
		c, e := s.repo.Comment(ctx, *parent, a.ID)
		if e != nil {
			return schemas.List{}, e
		}
		if c.PostID != postID {
			return schemas.List{}, core.NotFound()
		}
	}
	rows, total, e := s.repo.Comments(ctx, postID, a.ID, parent, page)
	if e != nil {
		return schemas.List{}, e
	}
	data := make([]schemas.Comment, 0, len(rows))
	for _, v := range rows {
		data = append(data, commentDTO(v, p, reveal))
	}
	return schemas.Paginated(data, page, total), nil
}
func commentDTO(c core.ReadingComment, p core.ReadingPost, reveal schemas.Reveal) schemas.Comment {
	hidden := c.Status != "published" || c.AuthorStatus != "active" || (p.ContainsSpoiler && !reveal.Spoilers) || (p.ContainsSensitiveContent && !reveal.Sensitive)
	var text *string
	if !hidden {
		text = &c.Text
	}
	count, liked := c.LikeCount, c.Liked
	if c.Status != "published" || c.AuthorStatus != "active" {
		count = 0
		liked = false
	}
	return schemas.Comment{ID: c.ID, PostID: c.PostID, ReaderID: c.ReaderID, ParentID: c.ParentID, Depth: c.Depth, Text: text, Status: c.Status, AuthorName: c.AuthorName, AuthorNickname: c.AuthorNickname, ContentHidden: hidden, LikeCount: count, Liked: liked, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}
}
func (s *Reading) CreateComment(ctx context.Context, a core.Actor, postID uint64, input schemas.CommentInput) (schemas.Comment, error) {
	if e := a.RequireReader(); e != nil {
		return schemas.Comment{}, e
	}
	if e := s.validateText(input.Text); e != nil {
		return schemas.Comment{}, e
	}
	var saved core.ReadingComment
	var post core.ReadingPost
	e := s.repo.Within(ctx, func(repo interfaces.ReadingRepository) error {
		if e := repo.LockReader(ctx, a.ID); e != nil {
			return e
		}
		p, e := repo.Post(ctx, postID, a.ID)
		if e != nil {
			return e
		}
		if e = postEligible(a, p, s.now(), false); e != nil {
			return e
		}
		post = p
		now := s.now()
		saved = core.ReadingComment{PostID: postID, ReaderID: a.ID, ParentID: input.ParentID, Text: strings.TrimSpace(input.Text), Status: "published", CreatedAt: now, UpdatedAt: now, AuthorStatus: "active"}
		var parent *core.ReadingComment
		if input.ParentID != nil {
			v, e := repo.Comment(ctx, *input.ParentID, a.ID)
			if e != nil {
				return e
			}
			if v.PostID != postID || v.Status != "published" || v.AuthorStatus != "active" {
				return core.Invalid("INVALID_COMMENT_PARENT", "A resposta deve referenciar um comentário disponível da mesma publicação.")
			}
			saved.Depth = v.Depth + 1
			if saved.Depth > s.options.MaxCommentDepth {
				return core.Invalid("COMMENT_DEPTH_EXCEEDED", "O limite de profundidade de respostas foi atingido.")
			}
			parent = &v
		}
		if e = repo.CreateComment(ctx, &saved); e != nil {
			return e
		}
		if p.ReaderID != a.ID {
			if e = repo.Notify(ctx, core.ReadingNotification{RecipientID: p.ReaderID, ActorID: &a.ID, Event: "comment", PostID: &postID, CommentID: &saved.ID, EventKey: fmt.Sprintf("comment:%d:%d", saved.ID, p.ReaderID), CreatedAt: now}); e != nil {
				return e
			}
		}
		if parent != nil && parent.ReaderID != a.ID && parent.ReaderID != p.ReaderID {
			return repo.Notify(ctx, core.ReadingNotification{RecipientID: parent.ReaderID, ActorID: &a.ID, Event: "reply", PostID: &postID, CommentID: &saved.ID, EventKey: fmt.Sprintf("reply:%d:%d", saved.ID, parent.ReaderID), CreatedAt: now})
		}
		return nil
	})
	if e != nil {
		return schemas.Comment{}, e
	}
	saved, e = s.repo.Comment(ctx, saved.ID, a.ID)
	if e != nil {
		return schemas.Comment{}, e
	}
	return commentDTO(saved, post, schemas.Reveal{}), nil
}
func (s *Reading) PatchComment(ctx context.Context, a core.Actor, id uint64, input schemas.CommentPatch) (schemas.Comment, error) {
	if e := a.RequireReader(); e != nil {
		return schemas.Comment{}, e
	}
	if e := s.validateText(input.Text); e != nil {
		return schemas.Comment{}, e
	}
	var saved core.ReadingComment
	var post core.ReadingPost
	e := s.repo.Within(ctx, func(repo interfaces.ReadingRepository) error {
		if e := repo.LockReader(ctx, a.ID); e != nil {
			return e
		}
		c, e := repo.Comment(ctx, id, a.ID)
		if e != nil {
			return e
		}
		if c.ReaderID != a.ID {
			return core.Forbidden()
		}
		if c.Status != "published" {
			return core.Conflict("Comentário indisponível para edição.")
		}
		p, e := repo.Post(ctx, c.PostID, a.ID)
		if e != nil {
			return e
		}
		if e = postEligible(a, p, s.now(), false); e != nil {
			return e
		}
		c.Text = strings.TrimSpace(input.Text)
		c.UpdatedAt = s.now()
		saved = c
		post = p
		return repo.SaveComment(ctx, c)
	})
	return commentDTO(saved, post, schemas.Reveal{}), e
}
func (s *Reading) DeleteComment(ctx context.Context, a core.Actor, id uint64) error {
	if e := a.RequireReader(); e != nil {
		return e
	}
	return s.repo.Within(ctx, func(repo interfaces.ReadingRepository) error {
		if e := repo.LockReader(ctx, a.ID); e != nil {
			return e
		}
		c, e := repo.Comment(ctx, id, a.ID)
		if e != nil {
			return e
		}
		if c.ReaderID != a.ID {
			return core.Forbidden()
		}
		if c.Status == "deleted" {
			return nil
		}
		if c.Status != "published" {
			return core.Conflict("O comentário está sob moderação.")
		}
		p, e := repo.Post(ctx, c.PostID, a.ID)
		if e != nil {
			return e
		}
		if e = postEligible(a, p, s.now(), false); e != nil {
			return e
		}
		c.Status = "deleted"
		c.UpdatedAt = s.now()
		return repo.SaveComment(ctx, c)
	})
}
func (s *Reading) Like(ctx context.Context, a core.Actor, target string, id uint64, add bool) error {
	if e := a.RequireReader(); e != nil {
		return e
	}
	return s.repo.Within(ctx, func(repo interfaces.ReadingRepository) error {
		if e := repo.LockReader(ctx, a.ID); e != nil {
			return e
		}
		var postID, recipientID uint64
		var commentID *uint64
		if target == "comment" {
			c, e := repo.Comment(ctx, id, a.ID)
			if e != nil {
				return e
			}
			if c.Status != "published" || c.AuthorStatus != "active" {
				return core.NotFound()
			}
			postID = c.PostID
			recipientID = c.ReaderID
			commentID = &id
		} else {
			postID = id
		}
		p, e := repo.Post(ctx, postID, a.ID)
		if e != nil {
			return e
		}
		if e = postEligible(a, p, s.now(), false); e != nil {
			return e
		}
		if target == "post" {
			recipientID = p.ReaderID
		}
		changed, e := repo.Like(ctx, target, id, a.ID, add)
		if e != nil {
			return e
		}
		if changed && add && recipientID != a.ID {
			return repo.Notify(ctx, core.ReadingNotification{RecipientID: recipientID, ActorID: &a.ID, Event: target + "_like", PostID: &postID, CommentID: commentID, EventKey: fmt.Sprintf("%s_like:%d:%d", target, id, a.ID), CreatedAt: s.now()})
		}
		return nil
	})
}
func (s *Reading) Notifications(ctx context.Context, a core.Actor, unread *bool, page schemas.Page) (schemas.List, error) {
	if a.ID == 0 || !a.Reader {
		return schemas.List{}, core.ErrUnauthenticated
	}
	rows, total, e := s.repo.Notifications(ctx, a, unread, page)
	if e != nil {
		return schemas.List{}, e
	}
	data := make([]schemas.Notification, 0, len(rows))
	for _, n := range rows {
		v := schemas.Notification{ID: n.ID, Event: n.Event, TargetAvailable: n.Visible, ReadAt: n.ReadAt, CreatedAt: n.CreatedAt}
		if n.Visible {
			v.PostID = n.PostID
			v.CommentID = n.CommentID
			v.ActorID = n.ActorID
		}
		data = append(data, v)
	}
	return schemas.Paginated(data, page, total), nil
}
func (s *Reading) ReadNotification(ctx context.Context, a core.Actor, id uint64) error {
	if a.ID == 0 || !a.Reader {
		return core.ErrUnauthenticated
	}
	return s.repo.ReadNotification(ctx, id, a.ID, s.now())
}
func (s *Reading) Statistics(ctx context.Context, a core.Actor, year int) (schemas.Statistics, error) {
	if a.ID == 0 || !a.Reader {
		return schemas.Statistics{}, core.ErrUnauthenticated
	}
	if year < 1900 || year > 9998 {
		return schemas.Statistics{}, core.Invalid("INVALID_YEAR", "Informe ano entre 1900 e 9998.")
	}
	v, e := s.repo.Statistics(ctx, a.ID, year)
	if e != nil {
		return schemas.Statistics{}, e
	}
	out := schemas.Statistics{Year: year, Timezone: "UTC", CurrentPages: v.CurrentPages, AverageRating: v.AverageRating, DiaryCount: v.DiaryCount, MostReadLanguage: v.MostReadLanguage, Statuses: []schemas.StatusCount{}, Months: make([]schemas.Period, 12), TopGenres: []schemas.GenreCount{}}
	counts := map[string]int64{}
	for _, r := range v.Statuses {
		counts[r.Status] = r.Count
	}
	for _, status := range []string{"planned", "in_progress", "paused", "completed", "abandoned"} {
		out.Statuses = append(out.Statuses, schemas.StatusCount{Status: status, Count: counts[status]})
	}
	for month := 1; month <= 12; month++ {
		out.Months[month-1] = schemas.Period{Month: month}
	}
	for _, r := range v.Months {
		if r.Month >= 1 && r.Month <= 12 {
			out.Months[r.Month-1].Works = r.Works
			out.Months[r.Month-1].Pages = r.Pages
			out.YearTotal.Works += r.Works
			out.YearTotal.Pages += r.Pages
		}
	}
	for _, g := range v.Genres {
		out.TopGenres = append(out.TopGenres, schemas.GenreCount{ID: g.ID, Name: g.Name, Count: g.Count})
	}
	return out, nil
}
func (s *Reading) ReadingReport(ctx context.Context, a core.Actor, group string, page schemas.Page) (schemas.List, error) {
	if a.ID == 0 || !a.Reader {
		return schemas.List{}, core.ErrUnauthenticated
	}
	status := ""
	switch group {
	case "completed", "abandoned":
		status = group
	case "incomplete":
		status = "incomplete"
	case "desired":
		status = "planned"
	case "recommended":
		if s.options.Recommendations != nil {
			return s.options.Recommendations(ctx, a, page)
		}
		return schemas.List{}, core.E(503, "RECOMMENDATIONS_UNAVAILABLE", "Recomendações indisponíveis.")
	case "preferred":
		if s.options.PreferredWorks != nil {
			return s.options.PreferredWorks(ctx, a, page)
		}
		return schemas.List{}, core.E(503, "PREFERENCES_UNAVAILABLE", "Preferências indisponíveis.")
	default:
		return schemas.List{}, core.Invalid("INVALID_REPORT_GROUP", "Use completed, incomplete, abandoned, desired, preferred ou recommended.")
	}
	return s.Shelf(ctx, a, a.ID, status, page)
}
