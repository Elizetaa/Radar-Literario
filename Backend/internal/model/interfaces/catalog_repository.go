package interfaces

import (
	"context"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/schemas"
)

type CatalogFilter struct {
	Page                                         schemas.Page
	Title, Author, ISBN, Language, Status, Order string
	GenreID, ReaderID, WorkID                    uint64
	MaxAge                                       int
	Admin, Preferred                             bool
	MinimumReviews                               int
}

type CatalogRepository interface {
	Within(context.Context, func(CatalogRepository) error) error
	ListWorks(context.Context, CatalogFilter) ([]core.Work, int64, error)
	Work(context.Context, uint64, int, bool) (core.Work, error)
	SaveWork(context.Context, *core.Work, []uint64, []uint64) error
	ListEditions(context.Context, CatalogFilter) ([]core.Edition, int64, error)
	Edition(context.Context, uint64, bool) (core.Edition, error)
	EditionByISBN(context.Context, string) (core.Edition, error)
	SaveEdition(context.Context, *core.Edition) error
	ValidateEditionPages(context.Context, uint64, int) error
	ListAuthors(context.Context, schemas.Page) ([]core.Author, int64, error)
	Author(context.Context, uint64) (core.Author, error)
	SaveAuthor(context.Context, *core.Author) error
	DeleteAuthor(context.Context, uint64) error
	ListPublishers(context.Context, schemas.Page) ([]core.Publisher, int64, error)
	Publisher(context.Context, uint64) (core.Publisher, error)
	SavePublisher(context.Context, *core.Publisher) error
	DeletePublisher(context.Context, uint64) error
	ListGenres(context.Context, schemas.Page, bool) ([]core.Genre, int64, error)
	Genre(context.Context, uint64) (core.Genre, error)
	SaveGenre(context.Context, *core.Genre) error
	Interests(context.Context, uint64) ([]core.Genre, error)
	SetInterests(context.Context, uint64, []uint64) error
	SetPreferred(context.Context, uint64, uint64, bool) error
	ListSuggestions(context.Context, uint64, string, schemas.Page) ([]core.CatalogSuggestion, int64, error)
	Suggestion(context.Context, uint64, bool) (core.CatalogSuggestion, error)
	SaveSuggestion(context.Context, *core.CatalogSuggestion) error
	Audit(context.Context, core.Actor, string, string, uint64, string) error
}
