package controller

import (
	"github.com/gin-gonic/gin"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/schemas"
	"radar-literario/backend/internal/model/services"
	"strconv"
)

type administrativeHandler func(*gin.Context) (any, int, error)

func adminHandle(op administrativeHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, status, e := op(c)
		if e != nil {
			Fail(c, e)
			return
		}
		Reply(c, status, v)
	}
}

func RegisterAdmin(g *gin.RouterGroup, s *services.Admin, auth gin.HandlerFunc) {
	g.GET("/faq", adminHandle(func(c *gin.Context) (any, int, error) {
		p, e := Page(c)
		if e != nil {
			return nil, 0, e
		}
		v, e := s.FAQs(c.Request.Context(), Actor(c), false, p)
		return v, 200, e
	}))
	g.GET("/faq/:id", adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		v, e := s.FAQ(c.Request.Context(), Actor(c), id, false)
		return v, 200, e
	}))
	private := g.Group("", auth)
	private.POST("/reports", adminHandle(func(c *gin.Context) (any, int, error) {
		var in schemas.CreateReport
		if e := Bind(c, &in); e != nil {
			return nil, 0, e
		}
		v, e := s.CreateReport(c.Request.Context(), Actor(c), in)
		return v, 201, e
	}))
	private.GET("/me/reports", adminHandle(func(c *gin.Context) (any, int, error) {
		p, e := Page(c)
		if e != nil {
			return nil, 0, e
		}
		v, e := s.Reports(c.Request.Context(), Actor(c), false, c.Query("status"), p)
		return v, 200, e
	}))
	private.GET("/me/reports/:id", adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		v, e := s.Report(c.Request.Context(), Actor(c), id, false)
		return v, 200, e
	}))
	private.POST("/me/reports/:id/cancel", adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		e = s.CancelReport(c.Request.Context(), Actor(c), id)
		return nil, 204, e
	}))
	private.GET("/me/reports/:id/messages", adminReportMessages(s, false))
	private.POST("/me/reports/:id/messages", adminAddReportMessage(s, false))
	private.POST("/support-tickets", adminHandle(func(c *gin.Context) (any, int, error) {
		var in schemas.CreateSupportTicket
		if e := Bind(c, &in); e != nil {
			return nil, 0, e
		}
		v, e := s.CreateTicket(c.Request.Context(), Actor(c), in)
		return v, 201, e
	}))
	private.GET("/me/support-tickets", adminTickets(s, false))
	private.GET("/support-tickets/:id", adminTicket(s, false))
	private.GET("/support-tickets/:id/messages", adminTicketMessages(s))
	private.POST("/support-tickets/:id/messages", adminAddTicketMessage(s, false))
	private.POST("/support-tickets/:id/close", adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		e = s.CloseTicket(c.Request.Context(), Actor(c), id)
		return nil, 204, e
	}))
	admin := private.Group("/admin")
	admin.GET("/administrators", adminHandle(func(c *gin.Context) (any, int, error) {
		p, e := Page(c)
		if e != nil {
			return nil, 0, e
		}
		v, e := s.Administrators(c.Request.Context(), Actor(c), p)
		return v, 200, e
	}))
	admin.POST("/administrators", adminHandle(func(c *gin.Context) (any, int, error) {
		var in schemas.CreateAdministrator
		if e := Bind(c, &in); e != nil {
			return nil, 0, e
		}
		v, e := s.CreateAdministrator(c.Request.Context(), Actor(c), in)
		return v, 201, e
	}))
	admin.GET("/administrators/:id", adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		v, e := s.Administrator(c.Request.Context(), Actor(c), id)
		return v, 200, e
	}))
	admin.PATCH("/administrators/:id", adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		var in schemas.PatchAdministrator
		if e = Bind(c, &in); e != nil {
			return nil, 0, e
		}
		v, e := s.UpdateAdministrator(c.Request.Context(), Actor(c), id, in)
		return v, 200, e
	}))
	admin.DELETE("/administrators/:id", adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		var in schemas.Reason
		if e = Bind(c, &in); e != nil {
			return nil, 0, e
		}
		e = s.RevokeAdministrator(c.Request.Context(), Actor(c), id, in.Reason)
		return nil, 204, e
	}))
	admin.GET("/users", adminHandle(func(c *gin.Context) (any, int, error) {
		p, e := Page(c)
		if e != nil {
			return nil, 0, e
		}
		v, e := s.Users(c.Request.Context(), Actor(c), c.Query("status"), p)
		return v, 200, e
	}))
	admin.GET("/users/:id", adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		v, e := s.User(c.Request.Context(), Actor(c), id)
		return v, 200, e
	}))
	admin.GET("/reports", adminHandle(func(c *gin.Context) (any, int, error) {
		p, e := Page(c)
		if e != nil {
			return nil, 0, e
		}
		v, e := s.Reports(c.Request.Context(), Actor(c), true, c.Query("status"), p)
		return v, 200, e
	}))
	admin.GET("/reports/:id", adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		v, e := s.Report(c.Request.Context(), Actor(c), id, true)
		return v, 200, e
	}))
	admin.PATCH("/reports/:id", adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		var in schemas.PatchReport
		if e = Bind(c, &in); e != nil {
			return nil, 0, e
		}
		v, e := s.UpdateReport(c.Request.Context(), Actor(c), id, in)
		return v, 200, e
	}))
	admin.GET("/reports/:id/messages", adminReportMessages(s, true))
	admin.POST("/reports/:id/messages", adminAddReportMessage(s, true))
	admin.POST("/posts/:id/moderation", adminModerate(s, "post"))
	admin.POST("/comments/:id/moderation", adminModerate(s, "comment"))
	admin.GET("/posts/:id", adminInspectContent(s, "post"))
	admin.GET("/comments/:id", adminInspectContent(s, "comment"))
	admin.GET("/penalties", adminHandle(func(c *gin.Context) (any, int, error) {
		p, e := Page(c)
		if e != nil {
			return nil, 0, e
		}
		var user *uint64
		if raw := c.Query("user_id"); raw != "" {
			id, e := strconv.ParseUint(raw, 10, 64)
			if e != nil || id == 0 {
				return nil, 0, core.Invalid("INVALID_ID", "Identificador de leitor inválido.")
			}
			user = &id
		}
		v, e := s.Penalties(c.Request.Context(), Actor(c), user, p)
		return v, 200, e
	}))
	admin.POST("/penalties", adminHandle(func(c *gin.Context) (any, int, error) {
		var in schemas.CreatePenalty
		if e := Bind(c, &in); e != nil {
			return nil, 0, e
		}
		v, e := s.ApplyPenalty(c.Request.Context(), Actor(c), in)
		return v, 201, e
	}))
	admin.POST("/penalties/:id/revoke", adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		var in schemas.Reason
		if e = Bind(c, &in); e != nil {
			return nil, 0, e
		}
		e = s.RevokePenalty(c.Request.Context(), Actor(c), id, in.Reason)
		return nil, 204, e
	}))
	admin.GET("/audit-logs", adminHandle(func(c *gin.Context) (any, int, error) {
		p, e := Page(c)
		if e != nil {
			return nil, 0, e
		}
		v, e := s.AuditLogs(c.Request.Context(), Actor(c), c.Query("target_type"), p)
		return v, 200, e
	}))
	admin.GET("/faq", adminHandle(func(c *gin.Context) (any, int, error) {
		p, e := Page(c)
		if e != nil {
			return nil, 0, e
		}
		v, e := s.FAQs(c.Request.Context(), Actor(c), true, p)
		return v, 200, e
	}))
	admin.POST("/faq", adminSaveFAQ(s, false))
	admin.GET("/faq/:id", adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		v, e := s.FAQ(c.Request.Context(), Actor(c), id, true)
		return v, 200, e
	}))
	admin.PATCH("/faq/:id", adminSaveFAQ(s, true))
	admin.DELETE("/faq/:id", adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		e = s.DeleteFAQ(c.Request.Context(), Actor(c), id)
		return nil, 204, e
	}))
	admin.GET("/support-tickets", adminTickets(s, true))
	admin.GET("/support-tickets/:id", adminTicket(s, true))
	admin.PATCH("/support-tickets/:id", adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		var in schemas.PatchSupportTicket
		if e = Bind(c, &in); e != nil {
			return nil, 0, e
		}
		v, e := s.UpdateTicket(c.Request.Context(), Actor(c), id, in)
		return v, 200, e
	}))
	admin.GET("/support-tickets/:id/messages", adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		p, e := Page(c)
		if e != nil {
			return nil, 0, e
		}
		v, e := s.AdminTicketMessages(c.Request.Context(), Actor(c), id, p)
		return v, 200, e
	}))
	admin.POST("/support-tickets/:id/messages", adminAddTicketMessage(s, true))
}
func adminReportMessages(s *services.Admin, admin bool) gin.HandlerFunc {
	return adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		p, e := Page(c)
		if e != nil {
			return nil, 0, e
		}
		v, e := s.ReportMessages(c.Request.Context(), Actor(c), id, admin, p)
		return v, 200, e
	})
}
func adminAddReportMessage(s *services.Admin, admin bool) gin.HandlerFunc {
	return adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		var in schemas.CreateReportMessage
		if e = Bind(c, &in); e != nil {
			return nil, 0, e
		}
		v, e := s.AddReportMessage(c.Request.Context(), Actor(c), id, admin, in)
		return v, 201, e
	})
}
func adminModerate(s *services.Admin, kind string) gin.HandlerFunc {
	return adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		var in schemas.Moderation
		if e = Bind(c, &in); e != nil {
			return nil, 0, e
		}
		e = s.Moderate(c.Request.Context(), Actor(c), kind, id, in)
		return nil, 204, e
	})
}

func adminInspectContent(s *services.Admin, kind string) gin.HandlerFunc {
	return adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		v, e := s.InspectContent(c.Request.Context(), Actor(c), kind, id)
		return v, 200, e
	})
}
func adminSaveFAQ(s *services.Admin, patch bool) gin.HandlerFunc {
	return adminHandle(func(c *gin.Context) (any, int, error) {
		var id uint64
		var e error
		status := 201
		if patch {
			id, e = ID(c, "id")
			if e != nil {
				return nil, 0, e
			}
			status = 200
		}
		var in schemas.SaveFAQ
		if e = Bind(c, &in); e != nil {
			return nil, 0, e
		}
		v, e := s.SaveFAQ(c.Request.Context(), Actor(c), id, in)
		return v, status, e
	})
}
func adminTickets(s *services.Admin, admin bool) gin.HandlerFunc {
	return adminHandle(func(c *gin.Context) (any, int, error) {
		p, e := Page(c)
		if e != nil {
			return nil, 0, e
		}
		v, e := s.Tickets(c.Request.Context(), Actor(c), admin, c.Query("status"), p)
		return v, 200, e
	})
}
func adminTicket(s *services.Admin, admin bool) gin.HandlerFunc {
	return adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		v, e := s.Ticket(c.Request.Context(), Actor(c), id, admin)
		return v, 200, e
	})
}
func adminTicketMessages(s *services.Admin) gin.HandlerFunc {
	return adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		p, e := Page(c)
		if e != nil {
			return nil, 0, e
		}
		v, e := s.TicketMessages(c.Request.Context(), Actor(c), id, p)
		return v, 200, e
	})
}
func adminAddTicketMessage(s *services.Admin, admin bool) gin.HandlerFunc {
	return adminHandle(func(c *gin.Context) (any, int, error) {
		id, e := ID(c, "id")
		if e != nil {
			return nil, 0, e
		}
		var in schemas.CreateSupportMessage
		if e = Bind(c, &in); e != nil {
			return nil, 0, e
		}
		v, e := s.AddTicketMessage(c.Request.Context(), Actor(c), id, admin, in)
		return v, 201, e
	})
}
