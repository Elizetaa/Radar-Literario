package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/schemas"
)

type Catalog struct {
	repo           interfaces.CatalogRepository
	minimumReviews int
}

func NewCatalog(repo interfaces.CatalogRepository, minimumReviews int) *Catalog {
	if minimumReviews < 1 {
		minimumReviews = 1
	}
	return &Catalog{repo: repo, minimumReviews: minimumReviews}
}

// NormalizeISBN returns the canonical ISBN-13, including conversion from ISBN-10.
func NormalizeISBN(raw string) (string, error) {
	value := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(raw), "-", ""), " ", ""))
	invalid := func() (string, error) {
		return "", core.Invalid("INVALID_ISBN", "Informe um ISBN-10 ou ISBN-13 com dígito verificador válido.")
	}
	if len(value) == 10 {
		sum := 0
		for i, c := range value {
			digit := int(c - '0')
			if i == 9 && c == 'X' {
				digit = 10
			} else if c < '0' || c > '9' {
				return invalid()
			}
			sum += (10 - i) * digit
		}
		if sum%11 != 0 {
			return invalid()
		}
		value = "978" + value[:9]
		sum = 0
		for i, c := range value {
			weight := 1
			if i%2 == 1 {
				weight = 3
			}
			sum += int(c-'0') * weight
		}
		return value + string(rune('0'+(10-sum%10)%10)), nil
	}
	if len(value) != 13 || (!strings.HasPrefix(value, "978") && !strings.HasPrefix(value, "979")) {
		return invalid()
	}
	sum := 0
	for i, c := range value {
		if c < '0' || c > '9' {
			return invalid()
		}
		weight := 1
		if i%2 == 1 {
			weight = 3
		}
		sum += int(c-'0') * weight
	}
	if sum%10 != 0 {
		return invalid()
	}
	return value, nil
}

func catalogText(raw string, min, max int, field string) (string, error) {
	v := strings.TrimSpace(raw)
	n := utf8.RuneCountInString(v)
	if n < min || n > max {
		return "", core.Invalid("INVALID_"+strings.ToUpper(field), "O campo "+field+" tem tamanho inválido.")
	}
	return v, nil
}

var catalogLanguage = regexp.MustCompile(`^[a-z]{2,3}(?:-[a-z0-9]{2,8})*$`)

func normalizeLanguage(s string) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) > 35 || !catalogLanguage.MatchString(s) {
		return "", core.Invalid("INVALID_LANGUAGE", "Informe um código de idioma, por exemplo pt-br ou en.")
	}
	return s, nil
}
func catalogDate(raw json.RawMessage, old *time.Time) (*time.Time, error) {
	if len(raw) == 0 {
		return old, nil
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return nil, core.Invalid("INVALID_DATE", "Informe data YYYY-MM-DD ou null.")
	}
	return core.Date(text)
}
func catalogIDs(ids []uint64) error {
	if len(ids) > 100 {
		return core.Invalid("INVALID_RELATION", "Use no máximo 100 relacionamentos.")
	}
	seen := map[uint64]bool{}
	for _, id := range ids {
		if id == 0 || seen[id] {
			return core.Invalid("INVALID_RELATION", "Identificadores devem ser positivos e únicos.")
		}
		seen[id] = true
	}
	return nil
}
func catalogStrict(raw json.RawMessage, dst any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return core.Invalid("INVALID_PROPOSAL", "A proposta contém campos desconhecidos ou inválidos.")
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return core.Invalid("INVALID_PROPOSAL", "Informe apenas um objeto de proposta.")
	}
	return nil
}
func catalogDateOut(value *time.Time) *string {
	if value == nil {
		return nil
	}
	s := value.Format("2006-01-02")
	return &s
}
func catalogAuthorOut(a core.Author) schemas.Author {
	return schemas.Author{ID: a.ID, Name: a.Name, Biography: a.Biography, BirthDate: catalogDateOut(a.BirthDate), DeathDate: catalogDateOut(a.DeathDate), CountryOfOrigin: a.CountryOfOrigin, Sex: a.Sex}
}
func catalogPublisherOut(p core.Publisher) schemas.Publisher {
	return schemas.Publisher{ID: p.ID, Name: p.Name, WebsiteURL: p.WebsiteURL, CountryOfOrigin: p.CountryOfOrigin}
}
func catalogGenreOut(g core.Genre) schemas.Genre {
	return schemas.Genre{ID: g.ID, Name: g.Name, Active: g.Active}
}
func catalogWorkOut(w core.Work, recommended bool) schemas.Work {
	authors := make([]schemas.Author, 0, len(w.Authors))
	genres := make([]schemas.Genre, 0, len(w.Genres))
	for _, a := range w.Authors {
		authors = append(authors, catalogAuthorOut(a))
	}
	for _, g := range w.Genres {
		genres = append(genres, catalogGenreOut(g))
	}
	result := schemas.Work{ID: w.ID, Title: w.Title, Synopsis: w.Synopsis, OriginalLanguage: w.OriginalLanguage, FirstPublicationDate: catalogDateOut(w.FirstPublicationDate), Status: w.Status, MinimumAge: w.MinimumAge, Version: w.Version, Authors: authors, Genres: genres, AverageRating: w.AverageRating, ReviewCount: w.ReviewCount, ReaderCount: w.ReaderCount, Affinity: w.Affinity, CreatedAt: w.CreatedAt, UpdatedAt: w.UpdatedAt}
	if recommended {
		result.RecommendationReason = "Classificação da comunidade como alternativa sem afinidade."
		if w.Affinity > 0 {
			result.RecommendationReason = "Gêneros compartilhados com seus interesses e avaliações de nota 4 ou 5."
		}
	}
	return result
}
func catalogEditionOut(e core.Edition) schemas.Edition {
	return schemas.Edition{ID: e.ID, WorkID: e.WorkID, PublisherID: e.PublisherID, Title: e.Title, ISBN: e.ISBN, ISBNAbsenceReason: e.ISBNAbsenceReason, EditionNumber: e.EditionNumber, PageCount: e.PageCount, PublicationDate: catalogDateOut(e.PublicationDate), Language: e.Language, ImageURL: e.ImageURL, Status: e.Status, MinimumAge: e.MinimumAge, Version: e.Version, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt}
}
func catalogSuggestionOut(s core.CatalogSuggestion) schemas.CatalogSuggestion {
	return schemas.CatalogSuggestion{ID: s.ID, ReaderID: s.ReaderID, Operation: s.Operation, Resource: s.Resource, TargetID: s.TargetID, ExpectedVersion: s.ExpectedVersion, Proposal: json.RawMessage(s.Proposal), Status: s.Status, DecisionBy: s.DecisionBy, DecisionReason: s.DecisionReason, DecidedAt: s.DecidedAt, CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt}
}

func (s *Catalog) Works(ctx context.Context, a core.Actor, f interfaces.CatalogFilter) (schemas.List, error) {
	if f.Admin {
		if err := a.Require("catalog"); err != nil {
			return schemas.List{}, err
		}
	}
	f.MaxAge = a.Age(time.Now().UTC())
	if f.Admin {
		f.MaxAge = 200
	}
	if f.Status != "" && !core.ValidCatalogStatus(f.Status) {
		return schemas.List{}, core.Invalid("INVALID_STATUS", "Estado de catálogo inválido.")
	}
	if len(f.Title) > 300 || len(f.Author) > 160 {
		return schemas.List{}, core.Invalid("INVALID_FILTER", "Filtro excede o limite de tamanho.")
	}
	if f.ISBN != "" {
		isbn, err := NormalizeISBN(f.ISBN)
		if err != nil {
			return schemas.List{}, err
		}
		f.ISBN = isbn
	}
	if f.Language != "" {
		language, err := normalizeLanguage(f.Language)
		if err != nil {
			return schemas.List{}, err
		}
		f.Language = language
	}
	switch f.Order {
	case "", "title", "newest", "popular", "top-rated", "recommended":
	default:
		return schemas.List{}, core.Invalid("INVALID_ORDER", "Ordenação inválida.")
	}
	if f.Preferred || f.Order == "recommended" {
		if err := a.RequireReader(); err != nil {
			return schemas.List{}, err
		}
		f.ReaderID = a.ID
	}
	f.MinimumReviews = s.minimumReviews
	rows, total, err := s.repo.ListWorks(ctx, f)
	if err != nil {
		return schemas.List{}, err
	}
	result := make([]schemas.Work, 0, len(rows))
	for _, row := range rows {
		result = append(result, catalogWorkOut(row, f.Order == "recommended"))
	}
	return schemas.Paginated(result, f.Page, total), nil
}
func (s *Catalog) Recommendations(ctx context.Context, a core.Actor, p schemas.Page) (schemas.List, error) {
	return s.Works(ctx, a, interfaces.CatalogFilter{Page: p, Order: "recommended"})
}
func (s *Catalog) Work(ctx context.Context, a core.Actor, id uint64, admin bool) (schemas.Work, error) {
	if admin {
		if err := a.Require("catalog"); err != nil {
			return schemas.Work{}, err
		}
	}
	age := a.Age(time.Now().UTC())
	if admin {
		age = 200
	}
	w, err := s.repo.Work(ctx, id, age, false)
	if err != nil {
		return schemas.Work{}, err
	}
	if !admin && (w.Status != "published" || w.MinimumAge > age) {
		return schemas.Work{}, core.NotFound()
	}
	return catalogWorkOut(w, false), nil
}
func (s *Catalog) Editions(ctx context.Context, a core.Actor, workID uint64, p schemas.Page, admin bool, status string) (schemas.List, error) {
	if admin {
		if err := a.Require("catalog"); err != nil {
			return schemas.List{}, err
		}
	}
	if workID > 0 {
		if _, err := s.Work(ctx, a, workID, admin); err != nil {
			return schemas.List{}, err
		}
	}
	if status != "" && !core.ValidCatalogStatus(status) {
		return schemas.List{}, core.Invalid("INVALID_STATUS", "Estado de catálogo inválido.")
	}
	rows, total, err := s.repo.ListEditions(ctx, interfaces.CatalogFilter{Page: p, WorkID: workID, Admin: admin, Status: status, MaxAge: a.Age(time.Now().UTC())})
	if err != nil {
		return schemas.List{}, err
	}
	result := make([]schemas.Edition, 0, len(rows))
	for _, row := range rows {
		result = append(result, catalogEditionOut(row))
	}
	return schemas.Paginated(result, p, total), nil
}
func (s *Catalog) editionVisible(ctx context.Context, a core.Actor, e core.Edition) error {
	if e.Status != "published" || e.MinimumAge > a.Age(time.Now().UTC()) {
		return core.NotFound()
	}
	_, err := s.Work(ctx, a, e.WorkID, false)
	return err
}
func (s *Catalog) Edition(ctx context.Context, a core.Actor, id uint64, admin bool) (schemas.Edition, error) {
	if admin {
		if err := a.Require("catalog"); err != nil {
			return schemas.Edition{}, err
		}
	}
	e, err := s.repo.Edition(ctx, id, false)
	if err != nil {
		return schemas.Edition{}, err
	}
	if !admin {
		if err = s.editionVisible(ctx, a, e); err != nil {
			return schemas.Edition{}, err
		}
	}
	return catalogEditionOut(e), nil
}
func (s *Catalog) ISBN(ctx context.Context, a core.Actor, value string) (schemas.Edition, error) {
	isbn, err := NormalizeISBN(value)
	if err != nil {
		return schemas.Edition{}, err
	}
	e, err := s.repo.EditionByISBN(ctx, isbn)
	if err != nil {
		return schemas.Edition{}, err
	}
	if err = s.editionVisible(ctx, a, e); err != nil {
		return schemas.Edition{}, err
	}
	return catalogEditionOut(e), nil
}

func catalogRelations(w core.Work) ([]uint64, []uint64) {
	authors := make([]uint64, 0, len(w.Authors))
	genres := make([]uint64, 0, len(w.Genres))
	for _, a := range w.Authors {
		authors = append(authors, a.ID)
	}
	for _, g := range w.Genres {
		genres = append(genres, g.ID)
	}
	return authors, genres
}
func (s *Catalog) prepareWork(ctx context.Context, repo interfaces.CatalogRepository, id uint64, in schemas.WorkInput) (core.Work, []uint64, []uint64, error) {
	w := core.Work{Status: "draft"}
	var err error
	if id > 0 {
		w, err = repo.Work(ctx, id, 200, true)
		if err != nil {
			return w, nil, nil, err
		}
		if in.InitialEdition != nil {
			return w, nil, nil, core.Invalid("INVALID_EDITION", "initial_edition é permitido somente ao criar a obra.")
		}
	}
	if in.Title != nil {
		w.Title = *in.Title
	}
	if in.Synopsis != nil {
		w.Synopsis = *in.Synopsis
	}
	if in.OriginalLanguage != nil {
		w.OriginalLanguage = *in.OriginalLanguage
	}
	if in.Status != nil {
		w.Status = *in.Status
	}
	if in.MinimumAge != nil {
		w.MinimumAge = *in.MinimumAge
	}
	w.Title, err = catalogText(w.Title, 1, 300, "title")
	if err != nil {
		return w, nil, nil, err
	}
	w.Synopsis, err = catalogText(w.Synopsis, 0, 30000, "synopsis")
	if err != nil {
		return w, nil, nil, err
	}
	w.OriginalLanguage, err = normalizeLanguage(w.OriginalLanguage)
	if err != nil {
		return w, nil, nil, err
	}
	w.FirstPublicationDate, err = catalogDate(in.FirstPublicationDate, w.FirstPublicationDate)
	if err != nil {
		return w, nil, nil, err
	}
	if !core.ValidCatalogStatus(w.Status) || !core.ValidMinimumAge(w.MinimumAge) {
		return w, nil, nil, core.Invalid("INVALID_CATALOG", "Estado ou classificação etária inválidos.")
	}
	authors, genres := catalogRelations(w)
	if in.AuthorIDs != nil {
		authors = *in.AuthorIDs
	}
	if in.GenreIDs != nil {
		genres = *in.GenreIDs
	}
	if err = catalogIDs(authors); err != nil {
		return w, nil, nil, err
	}
	if err = catalogIDs(genres); err != nil {
		return w, nil, nil, err
	}
	if w.Status == "published" {
		if (id == 0 && in.MinimumAge == nil) || len(authors) == 0 || len(genres) == 0 || w.Synopsis == "" {
			return w, nil, nil, core.Invalid("INCOMPLETE_WORK", "Publicação exige sinopse, autores, gêneros e classificação etária explícita.")
		}
		if w.FirstPublicationDate != nil && w.FirstPublicationDate.After(time.Now().UTC()) {
			return w, nil, nil, core.Invalid("INVALID_PUBLICATION_DATE", "Uma obra publicada não pode ter data futura.")
		}
	}
	for _, authorID := range authors {
		if _, err = repo.Author(ctx, authorID); err != nil {
			return w, nil, nil, err
		}
	}
	for _, genreID := range genres {
		var genre core.Genre
		genre, err = repo.Genre(ctx, genreID)
		if err != nil {
			return w, nil, nil, err
		}
		if w.Status == "published" && !genre.Active {
			return w, nil, nil, core.Invalid("INACTIVE_GENRE", "Obras publicadas exigem gêneros ativos.")
		}
	}
	return w, authors, genres, nil
}
func (s *Catalog) prepareEdition(ctx context.Context, repo interfaces.CatalogRepository, id uint64, in schemas.EditionInput) (core.Edition, error) {
	return s.prepareEditionWithWork(ctx, repo, id, in, nil)
}
func (s *Catalog) prepareEditionWithWork(ctx context.Context, repo interfaces.CatalogRepository, id uint64, in schemas.EditionInput, proposedWork *core.Work) (core.Edition, error) {
	e := core.Edition{Status: "draft", EditionNumber: 1}
	var err error
	if id > 0 {
		e, err = repo.Edition(ctx, id, true)
		if err != nil {
			return e, err
		}
		if in.WorkID != nil && *in.WorkID != e.WorkID {
			return e, core.Invalid("IMMUTABLE_WORK", "Não é permitido transferir uma edição para outra obra.")
		}
	}
	if in.WorkID != nil {
		e.WorkID = *in.WorkID
	}
	if in.PublisherID != nil {
		e.PublisherID = *in.PublisherID
	}
	if in.Title != nil {
		e.Title = *in.Title
	}
	if in.ISBNAbsenceReason != nil {
		e.ISBNAbsenceReason = strings.TrimSpace(*in.ISBNAbsenceReason)
	}
	if in.EditionNumber != nil {
		e.EditionNumber = *in.EditionNumber
	}
	if in.PageCount != nil {
		e.PageCount = *in.PageCount
	}
	if in.Language != nil {
		e.Language = *in.Language
	}
	if in.ImageURL != nil {
		e.ImageURL = *in.ImageURL
	}
	if in.Status != nil {
		e.Status = *in.Status
	}
	if in.MinimumAge != nil {
		e.MinimumAge = *in.MinimumAge
	}
	if len(in.ISBN) > 0 {
		if bytes.Equal(bytes.TrimSpace(in.ISBN), []byte("null")) {
			e.ISBN = nil
		} else {
			var raw string
			if json.Unmarshal(in.ISBN, &raw) != nil {
				return e, core.Invalid("INVALID_ISBN", "ISBN deve ser string ou null.")
			}
			isbn, normalizeErr := NormalizeISBN(raw)
			if normalizeErr != nil {
				return e, normalizeErr
			}
			e.ISBN = &isbn
		}
	}
	e.Title, err = catalogText(e.Title, 1, 300, "title")
	if err != nil {
		return e, err
	}
	e.Language, err = normalizeLanguage(e.Language)
	if err != nil {
		return e, err
	}
	if len(e.ISBNAbsenceReason) > 300 || e.ISBN == nil && e.ISBNAbsenceReason == "" {
		return e, core.Invalid("ISBN_REQUIRED", "Informe ISBN válido ou motivo para sua ausência.")
	}
	if e.PageCount < 1 || e.PageCount > 1000000 || e.EditionNumber < 1 || e.EditionNumber > 100000 {
		return e, core.Invalid("INVALID_EDITION", "Páginas e número da edição devem ser positivos e respeitar os limites.")
	}
	if !core.ValidCatalogStatus(e.Status) || !core.ValidMinimumAge(e.MinimumAge) {
		return e, core.Invalid("INVALID_CATALOG", "Estado ou classificação inválidos.")
	}
	if err = core.ValidateURL(e.ImageURL); err != nil {
		return e, err
	}
	e.PublicationDate, err = catalogDate(in.PublicationDate, e.PublicationDate)
	if err != nil {
		return e, err
	}
	w := core.Work{}
	if proposedWork != nil {
		w = *proposedWork
	} else {
		w, err = repo.Work(ctx, e.WorkID, 200, false)
		if err != nil {
			return e, err
		}
	}
	if _, err = repo.Publisher(ctx, e.PublisherID); err != nil {
		return e, err
	}
	if e.PublicationDate != nil && w.FirstPublicationDate != nil && e.PublicationDate.Before(*w.FirstPublicationDate) {
		return e, core.Invalid("INVALID_PUBLICATION_DATE", "A edição não pode anteceder a primeira publicação da obra.")
	}
	if e.Status == "published" {
		if id == 0 && in.MinimumAge == nil {
			return e, core.Invalid("INCOMPLETE_EDITION", "Publicação exige classificação etária explícita.")
		}
		if w.Status != "published" {
			return e, core.Invalid("UNPUBLISHED_WORK", "Publique a obra antes da edição.")
		}
		if e.PublicationDate != nil && e.PublicationDate.After(time.Now().UTC()) {
			return e, core.Invalid("INVALID_PUBLICATION_DATE", "Uma edição publicada não pode ter data futura.")
		}
	}
	if e.ISBN != nil {
		existing, lookupErr := repo.EditionByISBN(ctx, *e.ISBN)
		if lookupErr == nil && existing.ID != e.ID {
			return e, core.Conflict("Já existe uma edição com este ISBN.")
		}
		var domainErr *core.Error
		if lookupErr != nil && (!errors.As(lookupErr, &domainErr) || domainErr.Status != 404) {
			return e, lookupErr
		}
	}
	if id > 0 {
		if err = repo.ValidateEditionPages(ctx, id, e.PageCount); err != nil {
			return e, err
		}
	}
	return e, nil
}
func (s *Catalog) saveWork(ctx context.Context, repo interfaces.CatalogRepository, a core.Actor, id uint64, in schemas.WorkInput, reason string) (core.Work, error) {
	w, authors, genres, err := s.prepareWork(ctx, repo, id, in)
	if err != nil {
		return w, err
	}
	if err = repo.SaveWork(ctx, &w, authors, genres); err != nil {
		return w, err
	}
	if in.InitialEdition != nil {
		editionIn := *in.InitialEdition
		editionIn.WorkID = &w.ID
		e, prepareErr := s.prepareEdition(ctx, repo, 0, editionIn)
		if prepareErr != nil {
			return w, prepareErr
		}
		if err = repo.SaveEdition(ctx, &e); err != nil {
			return w, err
		}
		if err = repo.Audit(ctx, a, "edition.save", "edition", e.ID, reason); err != nil {
			return w, err
		}
	}
	if err = repo.Audit(ctx, a, "work.save", "work", w.ID, reason); err != nil {
		return w, err
	}
	return w, nil
}
func (s *Catalog) SaveWork(ctx context.Context, a core.Actor, id uint64, in schemas.WorkInput) (schemas.Work, error) {
	if err := a.Require("catalog"); err != nil {
		return schemas.Work{}, err
	}
	var result core.Work
	err := s.repo.Within(ctx, func(repo interfaces.CatalogRepository) error {
		var err error
		result, err = s.saveWork(ctx, repo, a, id, in, "Manutenção administrativa de catálogo.")
		return err
	})
	if err != nil {
		return schemas.Work{}, err
	}
	return s.Work(ctx, a, result.ID, true)
}
func (s *Catalog) ArchiveWork(ctx context.Context, a core.Actor, id uint64) error {
	if err := a.Require("catalog"); err != nil {
		return err
	}
	status := "archived"
	_, err := s.SaveWork(ctx, a, id, schemas.WorkInput{Status: &status})
	return err
}
func (s *Catalog) SaveEdition(ctx context.Context, a core.Actor, id uint64, in schemas.EditionInput) (schemas.Edition, error) {
	if err := a.Require("catalog"); err != nil {
		return schemas.Edition{}, err
	}
	var result core.Edition
	err := s.repo.Within(ctx, func(repo interfaces.CatalogRepository) error {
		var err error
		result, err = s.prepareEdition(ctx, repo, id, in)
		if err != nil {
			return err
		}
		if err = repo.SaveEdition(ctx, &result); err != nil {
			return err
		}
		return repo.Audit(ctx, a, "edition.save", "edition", result.ID, "Manutenção administrativa de catálogo.")
	})
	if err != nil {
		return schemas.Edition{}, err
	}
	return s.Edition(ctx, a, result.ID, true)
}
func (s *Catalog) ArchiveEdition(ctx context.Context, a core.Actor, id uint64) error {
	status := "archived"
	_, err := s.SaveEdition(ctx, a, id, schemas.EditionInput{Status: &status})
	return err
}

func (s *Catalog) Authors(ctx context.Context, p schemas.Page) (schemas.List, error) {
	rows, n, err := s.repo.ListAuthors(ctx, p)
	if err != nil {
		return schemas.List{}, err
	}
	out := make([]schemas.Author, 0, len(rows))
	for _, a := range rows {
		out = append(out, catalogAuthorOut(a))
	}
	return schemas.Paginated(out, p, n), nil
}
func (s *Catalog) Author(ctx context.Context, id uint64) (schemas.Author, error) {
	a, err := s.repo.Author(ctx, id)
	return catalogAuthorOut(a), err
}
func (s *Catalog) SaveAuthor(ctx context.Context, a core.Actor, id uint64, in schemas.AuthorInput) (schemas.Author, error) {
	if err := a.Require("catalog"); err != nil {
		return schemas.Author{}, err
	}
	var result core.Author
	err := s.repo.Within(ctx, func(repo interfaces.CatalogRepository) error {
		var err error
		if id > 0 {
			result, err = repo.Author(ctx, id)
			if err != nil {
				return err
			}
		}
		if in.Name != nil {
			result.Name = *in.Name
		}
		if in.Biography != nil {
			result.Biography = *in.Biography
		}
		if in.CountryOfOrigin != nil {
			result.CountryOfOrigin = *in.CountryOfOrigin
		}
		result.Name, err = catalogText(result.Name, 1, 160, "name")
		if err != nil {
			return err
		}
		result.Biography, err = catalogText(result.Biography, 0, 30000, "biography")
		if err != nil {
			return err
		}
		result.CountryOfOrigin, err = catalogText(result.CountryOfOrigin, 0, 100, "country_of_origin")
		if err != nil {
			return err
		}
		result.BirthDate, err = catalogDate(in.BirthDate, result.BirthDate)
		if err != nil {
			return err
		}
		result.DeathDate, err = catalogDate(in.DeathDate, result.DeathDate)
		if err != nil {
			return err
		}
		if result.BirthDate != nil && result.BirthDate.After(time.Now().UTC()) || result.DeathDate != nil && result.DeathDate.After(time.Now().UTC()) || result.BirthDate != nil && result.DeathDate != nil && result.DeathDate.Before(*result.BirthDate) {
			return core.Invalid("INVALID_AUTHOR_DATES", "Datas do autor são incoerentes.")
		}
		if len(in.Sex) > 0 {
			if bytes.Equal(bytes.TrimSpace(in.Sex), []byte("null")) {
				result.Sex = nil
			} else {
				var sex string
				if json.Unmarshal(in.Sex, &sex) != nil {
					return core.Invalid("INVALID_SEX", "Informe texto opcional ou null.")
				}
				if _, err = catalogText(sex, 0, 80, "sex"); err != nil {
					return err
				}
				result.Sex = &sex
			}
		}
		if err = repo.SaveAuthor(ctx, &result); err != nil {
			return err
		}
		return repo.Audit(ctx, a, "author.save", "author", result.ID, "Manutenção administrativa de catálogo.")
	})
	return catalogAuthorOut(result), err
}
func (s *Catalog) DeleteAuthor(ctx context.Context, a core.Actor, id uint64) error {
	if err := a.Require("catalog"); err != nil {
		return err
	}
	return s.repo.Within(ctx, func(repo interfaces.CatalogRepository) error {
		if _, err := repo.Author(ctx, id); err != nil {
			return err
		}
		if err := repo.DeleteAuthor(ctx, id); err != nil {
			return err
		}
		return repo.Audit(ctx, a, "author.delete", "author", id, "Exclusão de autor sem obras vinculadas.")
	})
}
func (s *Catalog) Publishers(ctx context.Context, p schemas.Page) (schemas.List, error) {
	rows, n, err := s.repo.ListPublishers(ctx, p)
	if err != nil {
		return schemas.List{}, err
	}
	out := make([]schemas.Publisher, 0, len(rows))
	for _, v := range rows {
		out = append(out, catalogPublisherOut(v))
	}
	return schemas.Paginated(out, p, n), nil
}
func (s *Catalog) Publisher(ctx context.Context, id uint64) (schemas.Publisher, error) {
	p, err := s.repo.Publisher(ctx, id)
	return catalogPublisherOut(p), err
}
func (s *Catalog) SavePublisher(ctx context.Context, a core.Actor, id uint64, in schemas.PublisherInput) (schemas.Publisher, error) {
	if err := a.Require("catalog"); err != nil {
		return schemas.Publisher{}, err
	}
	var p core.Publisher
	err := s.repo.Within(ctx, func(repo interfaces.CatalogRepository) error {
		var err error
		if id > 0 {
			p, err = repo.Publisher(ctx, id)
			if err != nil {
				return err
			}
		}
		if in.Name != nil {
			p.Name = *in.Name
		}
		if in.WebsiteURL != nil {
			p.WebsiteURL = *in.WebsiteURL
		}
		if in.CountryOfOrigin != nil {
			p.CountryOfOrigin = *in.CountryOfOrigin
		}
		p.Name, err = catalogText(p.Name, 1, 160, "name")
		if err != nil {
			return err
		}
		p.CountryOfOrigin, err = catalogText(p.CountryOfOrigin, 0, 100, "country_of_origin")
		if err != nil {
			return err
		}
		if err = core.ValidateURL(p.WebsiteURL); err != nil {
			return err
		}
		if err = repo.SavePublisher(ctx, &p); err != nil {
			return err
		}
		return repo.Audit(ctx, a, "publisher.save", "publisher", p.ID, "Manutenção administrativa de catálogo.")
	})
	return catalogPublisherOut(p), err
}
func (s *Catalog) DeletePublisher(ctx context.Context, a core.Actor, id uint64) error {
	if err := a.Require("catalog"); err != nil {
		return err
	}
	return s.repo.Within(ctx, func(repo interfaces.CatalogRepository) error {
		if _, err := repo.Publisher(ctx, id); err != nil {
			return err
		}
		if err := repo.DeletePublisher(ctx, id); err != nil {
			return err
		}
		return repo.Audit(ctx, a, "publisher.delete", "publisher", id, "Exclusão de editora sem edições vinculadas.")
	})
}
func (s *Catalog) Genres(ctx context.Context, a core.Actor, p schemas.Page, admin bool) (schemas.List, error) {
	if admin {
		if err := a.Require("catalog"); err != nil {
			return schemas.List{}, err
		}
	}
	rows, n, err := s.repo.ListGenres(ctx, p, admin)
	if err != nil {
		return schemas.List{}, err
	}
	out := make([]schemas.Genre, 0, len(rows))
	for _, g := range rows {
		out = append(out, catalogGenreOut(g))
	}
	return schemas.Paginated(out, p, n), nil
}
func (s *Catalog) Genre(ctx context.Context, a core.Actor, id uint64, admin bool) (schemas.Genre, error) {
	if admin {
		if err := a.Require("catalog"); err != nil {
			return schemas.Genre{}, err
		}
	}
	g, err := s.repo.Genre(ctx, id)
	if err == nil && !admin && !g.Active {
		err = core.NotFound()
	}
	return catalogGenreOut(g), err
}
func (s *Catalog) SaveGenre(ctx context.Context, a core.Actor, id uint64, in schemas.GenreInput) (schemas.Genre, error) {
	if err := a.Require("catalog"); err != nil {
		return schemas.Genre{}, err
	}
	g := core.Genre{Active: true}
	err := s.repo.Within(ctx, func(repo interfaces.CatalogRepository) error {
		var err error
		if id > 0 {
			g, err = repo.Genre(ctx, id)
			if err != nil {
				return err
			}
		}
		if in.Name != nil {
			g.Name = *in.Name
		}
		if in.Active != nil {
			g.Active = *in.Active
		}
		g.Name, err = catalogText(g.Name, 1, 100, "name")
		if err != nil {
			return err
		}
		if err = repo.SaveGenre(ctx, &g); err != nil {
			return err
		}
		return repo.Audit(ctx, a, "genre.save", "genre", g.ID, "Manutenção administrativa de catálogo.")
	})
	return catalogGenreOut(g), err
}
func (s *Catalog) DeleteGenre(ctx context.Context, a core.Actor, id uint64) error {
	active := false
	_, err := s.SaveGenre(ctx, a, id, schemas.GenreInput{Active: &active})
	return err
}

func (s *Catalog) Interests(ctx context.Context, a core.Actor) ([]schemas.Genre, error) {
	if err := a.RequireReader(); err != nil {
		return nil, err
	}
	rows, err := s.repo.Interests(ctx, a.ID)
	result := make([]schemas.Genre, 0, len(rows))
	for _, g := range rows {
		result = append(result, catalogGenreOut(g))
	}
	return result, err
}
func (s *Catalog) SetInterests(ctx context.Context, a core.Actor, ids []uint64) ([]schemas.Genre, error) {
	if err := a.RequireReader(); err != nil {
		return nil, err
	}
	if err := catalogIDs(ids); err != nil {
		return nil, err
	}
	err := s.repo.Within(ctx, func(repo interfaces.CatalogRepository) error {
		for _, id := range ids {
			g, err := repo.Genre(ctx, id)
			if err != nil {
				return err
			}
			if !g.Active {
				return core.Invalid("INACTIVE_GENRE", "Escolha gêneros ativos.")
			}
		}
		return repo.SetInterests(ctx, a.ID, ids)
	})
	if err != nil {
		return nil, err
	}
	return s.Interests(ctx, a)
}
func (s *Catalog) Preferred(ctx context.Context, a core.Actor, p schemas.Page) (schemas.List, error) {
	if err := a.RequireReader(); err != nil {
		return schemas.List{}, err
	}
	return s.Works(ctx, a, interfaces.CatalogFilter{Page: p, Preferred: true})
}
func (s *Catalog) SetPreferred(ctx context.Context, a core.Actor, id uint64, enabled bool) error {
	if err := a.RequireReader(); err != nil {
		return err
	}
	if enabled {
		if _, err := s.Work(ctx, a, id, false); err != nil {
			return err
		}
	}
	return s.repo.SetPreferred(ctx, a.ID, id, enabled)
}

func (s *Catalog) validateSuggestion(ctx context.Context, repo interfaces.CatalogRepository, a core.Actor, in schemas.SuggestionInput) (*uint64, error) {
	if in.Operation != "create" && in.Operation != "update" && in.Operation != "delete" || in.Resource != "work" && in.Resource != "edition" {
		return nil, core.Invalid("INVALID_SUGGESTION", "Informe operação create/update/delete e recurso work/edition.")
	}
	if in.Operation == "create" && in.TargetID != nil || in.Operation != "create" && (in.TargetID == nil || *in.TargetID == 0) {
		return nil, core.Invalid("INVALID_TARGET", "A inclusão não tem alvo; alteração e exclusão exigem alvo.")
	}
	var expected *uint64
	id := uint64(0)
	if in.TargetID != nil {
		id = *in.TargetID
		if in.Resource == "work" {
			w, err := repo.Work(ctx, id, a.Age(time.Now().UTC()), false)
			if err != nil {
				return nil, err
			}
			if w.Status != "published" || w.MinimumAge > a.Age(time.Now().UTC()) {
				return nil, core.NotFound()
			}
			expected = &w.Version
		} else {
			e, err := repo.Edition(ctx, id, false)
			if err != nil {
				return nil, err
			}
			if err = s.editionVisible(ctx, a, e); err != nil {
				return nil, err
			}
			expected = &e.Version
		}
	}
	if in.Operation == "delete" {
		var values map[string]any
		if err := catalogStrict(in.Proposal, &values); err != nil {
			return nil, err
		}
		if len(values) > 0 {
			return nil, core.Invalid("INVALID_PROPOSAL", "Proposta de exclusão usa objeto vazio.")
		}
		return expected, nil
	}
	if in.Resource == "work" {
		var proposal schemas.WorkInput
		if err := catalogStrict(in.Proposal, &proposal); err != nil {
			return nil, err
		}
		if proposal.Status != nil && *proposal.Status != "published" {
			return nil, core.Invalid("INVALID_STATUS", "A proposta deve descrever catálogo publicado.")
		}
		status := "published"
		proposal.Status = &status
		w, _, _, err := s.prepareWork(ctx, repo, id, proposal)
		if err != nil {
			return nil, err
		}
		if in.Operation == "create" && proposal.InitialEdition != nil {
			editionIn := *proposal.InitialEdition
			if editionIn.WorkID != nil {
				return nil, core.Invalid("INVALID_WORK", "initial_edition não aceita work_id na inclusão de obra.")
			}
			editionIn.Status = &status
			if _, err = s.prepareEditionWithWork(ctx, repo, 0, editionIn, &w); err != nil {
				return nil, err
			}
		}
		return expected, nil
	}
	var proposal schemas.EditionInput
	if err := catalogStrict(in.Proposal, &proposal); err != nil {
		return nil, err
	}
	if proposal.Status != nil && *proposal.Status != "published" {
		return nil, core.Invalid("INVALID_STATUS", "A proposta deve descrever catálogo publicado.")
	}
	status := "published"
	proposal.Status = &status
	e, err := s.prepareEdition(ctx, repo, id, proposal)
	if err != nil {
		return nil, err
	}
	if _, err = s.Work(ctx, a, e.WorkID, false); err != nil {
		return nil, err
	}
	return expected, nil
}
func (s *Catalog) CreateSuggestion(ctx context.Context, a core.Actor, in schemas.SuggestionInput) (schemas.CatalogSuggestion, error) {
	if err := a.RequireReader(); err != nil {
		return schemas.CatalogSuggestion{}, err
	}
	if len(in.Proposal) == 0 {
		in.Proposal = json.RawMessage(`{}`)
	}
	if len(in.Proposal) > 65536 {
		return schemas.CatalogSuggestion{}, core.Invalid("INVALID_PROPOSAL", "A proposta excede 64 KiB.")
	}
	expected, err := s.validateSuggestion(ctx, s.repo, a, in)
	if err != nil {
		return schemas.CatalogSuggestion{}, err
	}
	proposal := core.CatalogSuggestion{ReaderID: a.ID, Operation: in.Operation, Resource: in.Resource, TargetID: in.TargetID, ExpectedVersion: expected, Proposal: string(in.Proposal), Status: "pending"}
	if err = s.repo.SaveSuggestion(ctx, &proposal); err != nil {
		return schemas.CatalogSuggestion{}, err
	}
	saved, err := s.repo.Suggestion(ctx, proposal.ID, false)
	return catalogSuggestionOut(saved), err
}
func (s *Catalog) Suggestions(ctx context.Context, a core.Actor, p schemas.Page, status string, admin bool) (schemas.List, error) {
	reader := a.ID
	if admin {
		if err := a.Require("catalog"); err != nil {
			return schemas.List{}, err
		}
		reader = 0
	} else if err := a.RequireReader(); err != nil {
		return schemas.List{}, err
	}
	if status != "" && status != "pending" && status != "approved" && status != "rejected" {
		return schemas.List{}, core.Invalid("INVALID_STATUS", "Estado de sugestão inválido.")
	}
	rows, total, err := s.repo.ListSuggestions(ctx, reader, status, p)
	if err != nil {
		return schemas.List{}, err
	}
	result := make([]schemas.CatalogSuggestion, 0, len(rows))
	for _, row := range rows {
		result = append(result, catalogSuggestionOut(row))
	}
	return schemas.Paginated(result, p, total), nil
}
func (s *Catalog) Suggestion(ctx context.Context, a core.Actor, id uint64, admin bool) (schemas.CatalogSuggestion, error) {
	if admin {
		if err := a.Require("catalog"); err != nil {
			return schemas.CatalogSuggestion{}, err
		}
	} else if err := a.RequireReader(); err != nil {
		return schemas.CatalogSuggestion{}, err
	}
	row, err := s.repo.Suggestion(ctx, id, false)
	if err != nil {
		return schemas.CatalogSuggestion{}, err
	}
	if !admin && row.ReaderID != a.ID {
		return schemas.CatalogSuggestion{}, core.NotFound()
	}
	return catalogSuggestionOut(row), nil
}
func (s *Catalog) DecideSuggestion(ctx context.Context, a core.Actor, id uint64, approve bool, reason string) (schemas.CatalogSuggestion, error) {
	if err := a.Require("catalog"); err != nil {
		return schemas.CatalogSuggestion{}, err
	}
	reason, err := catalogText(reason, 3, 1000, "reason")
	if err != nil {
		return schemas.CatalogSuggestion{}, err
	}
	var result core.CatalogSuggestion
	err = s.repo.Within(ctx, func(repo interfaces.CatalogRepository) error {
		var err error
		result, err = repo.Suggestion(ctx, id, true)
		if err != nil {
			return err
		}
		if result.Status != "pending" {
			return core.Conflict("Esta sugestão já recebeu uma decisão.")
		}
		if approve {
			target := uint64(0)
			if result.TargetID != nil {
				target = *result.TargetID
			}
			if result.Resource == "work" {
				if target > 0 {
					current, getErr := repo.Work(ctx, target, 200, true)
					if getErr != nil {
						return getErr
					}
					if result.ExpectedVersion == nil || current.Version != *result.ExpectedVersion {
						return core.Conflict("A obra foi alterada após a proposta; envie uma nova sugestão.")
					}
				}
				var in schemas.WorkInput
				if result.Operation == "delete" {
					status := "archived"
					in.Status = &status
				} else {
					if err = catalogStrict(json.RawMessage(result.Proposal), &in); err != nil {
						return err
					}
					status := "published"
					in.Status = &status
					if in.InitialEdition != nil {
						in.InitialEdition.Status = &status
					}
				}
				if _, err = s.saveWork(ctx, repo, a, target, in, reason); err != nil {
					return err
				}
			} else {
				if target > 0 {
					current, getErr := repo.Edition(ctx, target, true)
					if getErr != nil {
						return getErr
					}
					if result.ExpectedVersion == nil || current.Version != *result.ExpectedVersion {
						return core.Conflict("A edição foi alterada após a proposta; envie uma nova sugestão.")
					}
				}
				var in schemas.EditionInput
				if result.Operation == "delete" {
					status := "archived"
					in.Status = &status
				} else {
					if err = catalogStrict(json.RawMessage(result.Proposal), &in); err != nil {
						return err
					}
					status := "published"
					in.Status = &status
				}
				edition, prepareErr := s.prepareEdition(ctx, repo, target, in)
				if prepareErr != nil {
					return prepareErr
				}
				if err = repo.SaveEdition(ctx, &edition); err != nil {
					return err
				}
				if err = repo.Audit(ctx, a, "edition.save", "edition", edition.ID, reason); err != nil {
					return err
				}
			}
			result.Status = "approved"
		} else {
			result.Status = "rejected"
		}
		now := time.Now().UTC()
		result.DecisionBy = &a.ID
		result.DecisionReason = reason
		result.DecidedAt = &now
		if err = repo.SaveSuggestion(ctx, &result); err != nil {
			return err
		}
		return repo.Audit(ctx, a, "catalog_suggestion."+result.Status, "catalog_suggestion", id, reason)
	})
	if err != nil {
		return schemas.CatalogSuggestion{}, err
	}
	return catalogSuggestionOut(result), nil
}
