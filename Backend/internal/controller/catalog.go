package controller

import (
	"github.com/gin-gonic/gin"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/schemas"
	"radar-literario/backend/internal/model/services"
	"strconv"
)

func catalogHandler(status int, fn func(*gin.Context) (any, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := fn(c)
		if err != nil {
			Fail(c, err)
			return
		}
		Reply(c, status, result)
	}
}
func catalogFilter(c *gin.Context, admin bool) (interfaces.CatalogFilter, error) {
	p, err := Page(c)
	if err != nil {
		return interfaces.CatalogFilter{}, err
	}
	f := interfaces.CatalogFilter{Page: p, Title: c.Query("title"), Author: c.Query("author"), ISBN: c.Query("isbn"), Language: c.Query("language"), Status: c.Query("status"), Order: c.Query("order"), Admin: admin}
	if v, ok := c.GetQuery("genre_id"); ok {
		f.GenreID, err = strconv.ParseUint(v, 10, 64)
		if err != nil || f.GenreID == 0 {
			return f, core.Invalid("INVALID_GENRE", "Identificador de gênero inválido.")
		}
	}
	return f, nil
}
func catalogOptionalID(c *gin.Context) (uint64, error) {
	if c.Param("id") == "" {
		return 0, nil
	}
	return ID(c, "id")
}

func RegisterCatalog(g *gin.RouterGroup, s *services.Catalog, auth gin.HandlerFunc) {
	g.GET("/works", catalogHandler(200, func(c *gin.Context) (any, error) {
		f, err := catalogFilter(c, false)
		if err != nil {
			return nil, err
		}
		return s.Works(c.Request.Context(), Actor(c), f)
	}))
	g.GET("/works/:id", catalogHandler(200, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		return s.Work(c.Request.Context(), Actor(c), id, false)
	}))
	g.GET("/works/:id/editions", catalogHandler(200, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		p, err := Page(c)
		if err != nil {
			return nil, err
		}
		return s.Editions(c.Request.Context(), Actor(c), id, p, false, "")
	}))
	g.GET("/editions/:id", catalogHandler(200, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		return s.Edition(c.Request.Context(), Actor(c), id, false)
	}))
	g.GET("/catalog/isbn/:isbn", catalogHandler(200, func(c *gin.Context) (any, error) { return s.ISBN(c.Request.Context(), Actor(c), c.Param("isbn")) }))
	g.GET("/authors", catalogHandler(200, func(c *gin.Context) (any, error) {
		p, err := Page(c)
		if err != nil {
			return nil, err
		}
		return s.Authors(c.Request.Context(), p)
	}))
	g.GET("/authors/:id", catalogHandler(200, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		return s.Author(c.Request.Context(), id)
	}))
	g.GET("/publishers", catalogHandler(200, func(c *gin.Context) (any, error) {
		p, err := Page(c)
		if err != nil {
			return nil, err
		}
		return s.Publishers(c.Request.Context(), p)
	}))
	g.GET("/publishers/:id", catalogHandler(200, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		return s.Publisher(c.Request.Context(), id)
	}))
	g.GET("/genres", catalogHandler(200, func(c *gin.Context) (any, error) {
		p, err := Page(c)
		if err != nil {
			return nil, err
		}
		return s.Genres(c.Request.Context(), Actor(c), p, false)
	}))
	g.GET("/rankings/popular", catalogHandler(200, func(c *gin.Context) (any, error) {
		p, err := Page(c)
		if err != nil {
			return nil, err
		}
		return s.Works(c.Request.Context(), Actor(c), interfaces.CatalogFilter{Page: p, Order: "popular"})
	}))
	g.GET("/rankings/top-rated", catalogHandler(200, func(c *gin.Context) (any, error) {
		p, err := Page(c)
		if err != nil {
			return nil, err
		}
		return s.Works(c.Request.Context(), Actor(c), interfaces.CatalogFilter{Page: p, Order: "top-rated"})
	}))
	g.GET("/me/recommendations", auth, catalogHandler(200, func(c *gin.Context) (any, error) {
		p, err := Page(c)
		if err != nil {
			return nil, err
		}
		return s.Recommendations(c.Request.Context(), Actor(c), p)
	}))
	g.GET("/me/interests", auth, catalogHandler(200, func(c *gin.Context) (any, error) { return s.Interests(c.Request.Context(), Actor(c)) }))
	g.PUT("/me/interests", auth, catalogHandler(200, func(c *gin.Context) (any, error) {
		var in schemas.InterestsInput
		if err := Bind(c, &in); err != nil {
			return nil, err
		}
		return s.SetInterests(c.Request.Context(), Actor(c), in.GenreIDs)
	}))
	g.GET("/me/preferred-works", auth, catalogHandler(200, func(c *gin.Context) (any, error) {
		p, err := Page(c)
		if err != nil {
			return nil, err
		}
		return s.Preferred(c.Request.Context(), Actor(c), p)
	}))
	g.PUT("/me/preferred-works/:work_id", auth, catalogHandler(204, func(c *gin.Context) (any, error) {
		id, err := ID(c, "work_id")
		if err != nil {
			return nil, err
		}
		return nil, s.SetPreferred(c.Request.Context(), Actor(c), id, true)
	}))
	g.DELETE("/me/preferred-works/:work_id", auth, catalogHandler(204, func(c *gin.Context) (any, error) {
		id, err := ID(c, "work_id")
		if err != nil {
			return nil, err
		}
		return nil, s.SetPreferred(c.Request.Context(), Actor(c), id, false)
	}))
	g.POST("/catalog-suggestions", auth, catalogHandler(201, func(c *gin.Context) (any, error) {
		var in schemas.SuggestionInput
		if err := Bind(c, &in); err != nil {
			return nil, err
		}
		return s.CreateSuggestion(c.Request.Context(), Actor(c), in)
	}))
	g.GET("/me/catalog-suggestions", auth, catalogHandler(200, func(c *gin.Context) (any, error) {
		p, err := Page(c)
		if err != nil {
			return nil, err
		}
		return s.Suggestions(c.Request.Context(), Actor(c), p, c.Query("status"), false)
	}))
	g.GET("/me/catalog-suggestions/:id", auth, catalogHandler(200, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		return s.Suggestion(c.Request.Context(), Actor(c), id, false)
	}))
	admin := g.Group("/admin", auth)
	admin.GET("/works", catalogHandler(200, func(c *gin.Context) (any, error) {
		f, err := catalogFilter(c, true)
		if err != nil {
			return nil, err
		}
		return s.Works(c.Request.Context(), Actor(c), f)
	}))
	admin.GET("/works/:id", catalogHandler(200, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		return s.Work(c.Request.Context(), Actor(c), id, true)
	}))
	admin.POST("/works", catalogHandler(201, func(c *gin.Context) (any, error) {
		var in schemas.WorkInput
		if err := Bind(c, &in); err != nil {
			return nil, err
		}
		return s.SaveWork(c.Request.Context(), Actor(c), 0, in)
	}))
	admin.PATCH("/works/:id", catalogHandler(200, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		var in schemas.WorkInput
		if err = Bind(c, &in); err != nil {
			return nil, err
		}
		return s.SaveWork(c.Request.Context(), Actor(c), id, in)
	}))
	admin.DELETE("/works/:id", catalogHandler(204, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		return nil, s.ArchiveWork(c.Request.Context(), Actor(c), id)
	}))
	admin.GET("/editions", catalogHandler(200, func(c *gin.Context) (any, error) {
		p, err := Page(c)
		if err != nil {
			return nil, err
		}
		workID := uint64(0)
		if raw, ok := c.GetQuery("work_id"); ok {
			workID, err = strconv.ParseUint(raw, 10, 64)
			if err != nil || workID == 0 {
				return nil, core.Invalid("INVALID_WORK", "Identificador de obra inválido.")
			}
		}
		return s.Editions(c.Request.Context(), Actor(c), workID, p, true, c.Query("status"))
	}))
	admin.GET("/editions/:id", catalogHandler(200, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		return s.Edition(c.Request.Context(), Actor(c), id, true)
	}))
	admin.POST("/editions", catalogHandler(201, func(c *gin.Context) (any, error) {
		var in schemas.EditionInput
		if err := Bind(c, &in); err != nil {
			return nil, err
		}
		return s.SaveEdition(c.Request.Context(), Actor(c), 0, in)
	}))
	admin.PATCH("/editions/:id", catalogHandler(200, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		var in schemas.EditionInput
		if err = Bind(c, &in); err != nil {
			return nil, err
		}
		return s.SaveEdition(c.Request.Context(), Actor(c), id, in)
	}))
	admin.DELETE("/editions/:id", catalogHandler(204, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		return nil, s.ArchiveEdition(c.Request.Context(), Actor(c), id)
	}))
	admin.GET("/authors", catalogHandler(200, func(c *gin.Context) (any, error) {
		if err := Actor(c).Require("catalog"); err != nil {
			return nil, err
		}
		p, err := Page(c)
		if err != nil {
			return nil, err
		}
		return s.Authors(c.Request.Context(), p)
	}))
	admin.GET("/authors/:id", catalogHandler(200, func(c *gin.Context) (any, error) {
		if err := Actor(c).Require("catalog"); err != nil {
			return nil, err
		}
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		return s.Author(c.Request.Context(), id)
	}))
	admin.POST("/authors", catalogHandler(201, func(c *gin.Context) (any, error) {
		var in schemas.AuthorInput
		if err := Bind(c, &in); err != nil {
			return nil, err
		}
		return s.SaveAuthor(c.Request.Context(), Actor(c), 0, in)
	}))
	admin.PATCH("/authors/:id", catalogHandler(200, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		var in schemas.AuthorInput
		if err = Bind(c, &in); err != nil {
			return nil, err
		}
		return s.SaveAuthor(c.Request.Context(), Actor(c), id, in)
	}))
	admin.DELETE("/authors/:id", catalogHandler(204, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		return nil, s.DeleteAuthor(c.Request.Context(), Actor(c), id)
	}))
	admin.GET("/publishers", catalogHandler(200, func(c *gin.Context) (any, error) {
		if err := Actor(c).Require("catalog"); err != nil {
			return nil, err
		}
		p, err := Page(c)
		if err != nil {
			return nil, err
		}
		return s.Publishers(c.Request.Context(), p)
	}))
	admin.GET("/publishers/:id", catalogHandler(200, func(c *gin.Context) (any, error) {
		if err := Actor(c).Require("catalog"); err != nil {
			return nil, err
		}
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		return s.Publisher(c.Request.Context(), id)
	}))
	admin.POST("/publishers", catalogHandler(201, func(c *gin.Context) (any, error) {
		var in schemas.PublisherInput
		if err := Bind(c, &in); err != nil {
			return nil, err
		}
		return s.SavePublisher(c.Request.Context(), Actor(c), 0, in)
	}))
	admin.PATCH("/publishers/:id", catalogHandler(200, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		var in schemas.PublisherInput
		if err = Bind(c, &in); err != nil {
			return nil, err
		}
		return s.SavePublisher(c.Request.Context(), Actor(c), id, in)
	}))
	admin.DELETE("/publishers/:id", catalogHandler(204, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		return nil, s.DeletePublisher(c.Request.Context(), Actor(c), id)
	}))
	admin.GET("/genres", catalogHandler(200, func(c *gin.Context) (any, error) {
		p, err := Page(c)
		if err != nil {
			return nil, err
		}
		return s.Genres(c.Request.Context(), Actor(c), p, true)
	}))
	admin.GET("/genres/:id", catalogHandler(200, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		return s.Genre(c.Request.Context(), Actor(c), id, true)
	}))
	admin.POST("/genres", catalogHandler(201, func(c *gin.Context) (any, error) {
		var in schemas.GenreInput
		if err := Bind(c, &in); err != nil {
			return nil, err
		}
		return s.SaveGenre(c.Request.Context(), Actor(c), 0, in)
	}))
	admin.PATCH("/genres/:id", catalogHandler(200, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		var in schemas.GenreInput
		if err = Bind(c, &in); err != nil {
			return nil, err
		}
		return s.SaveGenre(c.Request.Context(), Actor(c), id, in)
	}))
	admin.DELETE("/genres/:id", catalogHandler(204, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		return nil, s.DeleteGenre(c.Request.Context(), Actor(c), id)
	}))
	admin.GET("/catalog-suggestions", catalogHandler(200, func(c *gin.Context) (any, error) {
		p, err := Page(c)
		if err != nil {
			return nil, err
		}
		return s.Suggestions(c.Request.Context(), Actor(c), p, c.Query("status"), true)
	}))
	admin.GET("/catalog-suggestions/:id", catalogHandler(200, func(c *gin.Context) (any, error) {
		id, err := ID(c, "id")
		if err != nil {
			return nil, err
		}
		return s.Suggestion(c.Request.Context(), Actor(c), id, true)
	}))
	for _, decision := range []struct {
		Path    string
		Approve bool
	}{{"approve", true}, {"reject", false}} {
		d := decision
		admin.POST("/catalog-suggestions/:id/"+d.Path, catalogHandler(200, func(c *gin.Context) (any, error) {
			id, err := ID(c, "id")
			if err != nil {
				return nil, err
			}
			var in schemas.CatalogDecision
			if err = Bind(c, &in); err != nil {
				return nil, err
			}
			return s.DecideSuggestion(c.Request.Context(), Actor(c), id, d.Approve, in.Reason)
		}))
	}
}
