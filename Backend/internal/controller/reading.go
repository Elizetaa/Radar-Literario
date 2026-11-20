package controller

import (
	"github.com/gin-gonic/gin"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/schemas"
	"radar-literario/backend/internal/model/services"
	"strconv"
	"time"
)

type readingHandler struct{ service *services.Reading }

func readingHandle(fn func(*gin.Context) (any, int, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, status, e := fn(c)
		if e != nil {
			Fail(c, e)
			return
		}
		Reply(c, status, v)
	}
}
func readingReveal(c *gin.Context) (schemas.Reveal, error) {
	var r schemas.Reveal
	for _, key := range []string{"reveal_spoilers", "reveal_sensitive"} {
		if value, ok := c.GetQuery(key); ok {
			b, e := strconv.ParseBool(value)
			if e != nil {
				return r, core.Invalid("INVALID_FILTER", "Os filtros de confirmação devem ser true ou false.")
			}
			if key == "reveal_spoilers" {
				r.Spoilers = b
			} else {
				r.Sensitive = b
			}
		}
	}
	return r, nil
}
func readingQueryID(c *gin.Context, key string) (uint64, error) {
	v, ok := c.GetQuery(key)
	if !ok {
		return 0, nil
	}
	id, e := strconv.ParseUint(v, 10, 64)
	if e != nil || id == 0 {
		return 0, core.Invalid("INVALID_FILTER", "Identificador de filtro inválido.")
	}
	return id, nil
}
func RegisterReading(group *gin.RouterGroup, s *services.Reading, auth gin.HandlerFunc) {
	h := &readingHandler{service: s}
	group.GET("/posts", readingHandle(h.listPosts("", false, "")))
	group.GET("/posts/:id", readingHandle(h.getPost("")))
	group.GET("/readers/:id/posts", readingHandle(h.listPosts("", false, "reader")))
	group.GET("/readers/:id/shelf", readingHandle(h.shelf(false)))
	group.GET("/works/:id/reviews", readingHandle(h.listPosts("review", false, "work")))
	group.GET("/reading-entries/:id", readingHandle(h.getPost("diary")))
	group.GET("/reviews/:id", readingHandle(h.getPost("review")))
	group.GET("/posts/:id/comments", readingHandle(h.comments))
	private := group.Group("")
	private.Use(auth)
	private.GET("/me/shelf", readingHandle(h.shelf(true)))
	private.PUT("/me/shelf/:work_id", readingHandle(h.putShelf))
	private.DELETE("/me/shelf/:work_id", readingHandle(h.deleteShelf))
	private.GET("/me/reading-entries", readingHandle(h.listPosts("diary", true, "")))
	private.GET("/me/reviews", readingHandle(h.listPosts("review", true, "")))
	private.POST("/reading-entries", readingHandle(h.createPost("diary")))
	private.POST("/reviews", readingHandle(h.createPost("review")))
	private.PATCH("/reading-entries/:id", readingHandle(h.patchPost("diary")))
	private.PATCH("/reviews/:id", readingHandle(h.patchPost("review")))
	private.DELETE("/reading-entries/:id", readingHandle(h.deletePost("diary")))
	private.DELETE("/reviews/:id", readingHandle(h.deletePost("review")))
	private.POST("/posts/:id/comments", readingHandle(h.createComment))
	private.PATCH("/comments/:id", readingHandle(h.patchComment))
	private.DELETE("/comments/:id", readingHandle(h.deleteComment))
	private.PUT("/posts/:id/like", readingHandle(h.like("post", true)))
	private.DELETE("/posts/:id/like", readingHandle(h.like("post", false)))
	private.PUT("/comments/:id/like", readingHandle(h.like("comment", true)))
	private.DELETE("/comments/:id/like", readingHandle(h.like("comment", false)))
	private.GET("/me/notifications", readingHandle(h.notifications))
	private.PUT("/me/notifications/:id/read", readingHandle(h.readNotification))
	private.GET("/me/statistics", readingHandle(h.statistics))
	private.GET("/me/reading-report", readingHandle(h.readingReport))
}
func (h *readingHandler) shelf(own bool) func(*gin.Context) (any, int, error) {
	return func(c *gin.Context) (any, int, error) {
		p, e := Page(c)
		if e != nil {
			return nil, 0, e
		}
		id := Actor(c).ID
		if !own {
			id, e = ID(c, "id")
			if e != nil {
				return nil, 0, e
			}
		}
		v, e := h.service.Shelf(c.Request.Context(), Actor(c), id, c.Query("status"), p)
		return v, 200, e
	}
}
func (h *readingHandler) putShelf(c *gin.Context) (any, int, error) {
	id, e := ID(c, "work_id")
	if e != nil {
		return nil, 0, e
	}
	var input schemas.ShelfInput
	if e = Bind(c, &input); e != nil {
		return nil, 0, e
	}
	v, e := h.service.PutShelf(c.Request.Context(), Actor(c), id, input)
	return v, 200, e
}
func (h *readingHandler) deleteShelf(c *gin.Context) (any, int, error) {
	id, e := ID(c, "work_id")
	if e != nil {
		return nil, 0, e
	}
	return nil, 204, h.service.DeleteShelf(c.Request.Context(), Actor(c), id)
}
func (h *readingHandler) getPost(kind string) func(*gin.Context) (any, int, error) {
	return func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		r, e := readingReveal(c)
		if e != nil {
			return nil, 0, e
		}
		v, e := h.service.GetPost(c.Request.Context(), Actor(c), id, kind, r)
		return v, 200, e
	}
}
func (h *readingHandler) listPosts(kind string, own bool, scope string) func(*gin.Context) (any, int, error) {
	return func(c *gin.Context) (any, int, error) {
		p, e := Page(c)
		if e != nil {
			return nil, 0, e
		}
		r, e := readingReveal(c)
		if e != nil {
			return nil, 0, e
		}
		filter := schemas.PostFilter{Kind: kind, Own: own, Page: p}
		if kind == "" {
			filter.Kind = c.Query("kind")
		}
		if scope == "reader" {
			filter.ReaderID, e = ID(c, "id")
		} else if scope == "work" {
			filter.WorkID, e = ID(c, "id")
		} else {
			filter.WorkID, e = readingQueryID(c, "work_id")
		}
		if e != nil {
			return nil, 0, e
		}
		v, e := h.service.Posts(c.Request.Context(), Actor(c), filter, r)
		return v, 200, e
	}
}
func (h *readingHandler) createPost(kind string) func(*gin.Context) (any, int, error) {
	return func(c *gin.Context) (any, int, error) {
		var input schemas.PostInput
		if e := Bind(c, &input); e != nil {
			return nil, 0, e
		}
		v, e := h.service.CreatePost(c.Request.Context(), Actor(c), kind, input)
		return v, 201, e
	}
}
func (h *readingHandler) patchPost(kind string) func(*gin.Context) (any, int, error) {
	return func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		var input schemas.PostPatch
		if e = Bind(c, &input); e != nil {
			return nil, 0, e
		}
		v, e := h.service.PatchPost(c.Request.Context(), Actor(c), id, kind, input)
		return v, 200, e
	}
}
func (h *readingHandler) deletePost(kind string) func(*gin.Context) (any, int, error) {
	return func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		return nil, 204, h.service.DeletePost(c.Request.Context(), Actor(c), id, kind)
	}
}
func (h *readingHandler) comments(c *gin.Context) (any, int, error) {
	id, e := ID(c, "id")
	if e != nil {
		return nil, 0, e
	}
	p, e := Page(c)
	if e != nil {
		return nil, 0, e
	}
	r, e := readingReveal(c)
	if e != nil {
		return nil, 0, e
	}
	parent, e := readingQueryID(c, "parent_id")
	if e != nil {
		return nil, 0, e
	}
	var ptr *uint64
	if parent != 0 {
		ptr = &parent
	}
	v, e := h.service.Comments(c.Request.Context(), Actor(c), id, ptr, p, r)
	return v, 200, e
}
func (h *readingHandler) createComment(c *gin.Context) (any, int, error) {
	id, e := ID(c, "id")
	if e != nil {
		return nil, 0, e
	}
	var input schemas.CommentInput
	if e = Bind(c, &input); e != nil {
		return nil, 0, e
	}
	v, e := h.service.CreateComment(c.Request.Context(), Actor(c), id, input)
	return v, 201, e
}
func (h *readingHandler) patchComment(c *gin.Context) (any, int, error) {
	id, e := ID(c, "id")
	if e != nil {
		return nil, 0, e
	}
	var input schemas.CommentPatch
	if e = Bind(c, &input); e != nil {
		return nil, 0, e
	}
	v, e := h.service.PatchComment(c.Request.Context(), Actor(c), id, input)
	return v, 200, e
}
func (h *readingHandler) deleteComment(c *gin.Context) (any, int, error) {
	id, e := ID(c, "id")
	if e != nil {
		return nil, 0, e
	}
	return nil, 204, h.service.DeleteComment(c.Request.Context(), Actor(c), id)
}
func (h *readingHandler) like(target string, add bool) func(*gin.Context) (any, int, error) {
	return func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		return nil, 204, h.service.Like(c.Request.Context(), Actor(c), target, id, add)
	}
}
func (h *readingHandler) notifications(c *gin.Context) (any, int, error) {
	p, e := Page(c)
	if e != nil {
		return nil, 0, e
	}
	var unread *bool
	if raw, ok := c.GetQuery("unread"); ok {
		v, e := strconv.ParseBool(raw)
		if e != nil {
			return nil, 0, core.Invalid("INVALID_FILTER", "unread deve ser true ou false.")
		}
		unread = &v
	}
	v, e := h.service.Notifications(c.Request.Context(), Actor(c), unread, p)
	return v, 200, e
}
func (h *readingHandler) readNotification(c *gin.Context) (any, int, error) {
	id, e := ID(c, "id")
	if e != nil {
		return nil, 0, e
	}
	return nil, 204, h.service.ReadNotification(c.Request.Context(), Actor(c), id)
}
func (h *readingHandler) statistics(c *gin.Context) (any, int, error) {
	year := time.Now().UTC().Year()
	if raw, ok := c.GetQuery("year"); ok {
		v, e := strconv.Atoi(raw)
		if e != nil {
			return nil, 0, core.Invalid("INVALID_YEAR", "Informe um ano válido.")
		}
		year = v
	}
	v, e := h.service.Statistics(c.Request.Context(), Actor(c), year)
	return v, 200, e
}
func (h *readingHandler) readingReport(c *gin.Context) (any, int, error) {
	p, e := Page(c)
	if e != nil {
		return nil, 0, e
	}
	v, e := h.service.ReadingReport(c.Request.Context(), Actor(c), c.DefaultQuery("group", "completed"), p)
	return v, 200, e
}
