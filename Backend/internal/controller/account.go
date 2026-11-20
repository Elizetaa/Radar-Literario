package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"radar-literario/backend/internal/config"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/schemas"
	"radar-literario/backend/internal/model/services"
)

func Cookie(c *gin.Context, cfg config.Config, name, value string, age int) {
	same := http.SameSiteLaxMode
	if cfg.SameSite == "strict" {
		same = http.SameSiteStrictMode
	}
	if cfg.SameSite == "none" {
		same = http.SameSiteNoneMode
	}
	http.SetCookie(c.Writer, &http.Cookie{Name: name, Value: value, Path: "/", HttpOnly: true, Secure: cfg.CookieSecure, SameSite: same, MaxAge: age})
}
func RegisterAccounts(g *gin.RouterGroup, s *services.Accounts, cfg config.Config, auth, limit gin.HandlerFunc) {
	g.GET("/auth/csrf", func(c *gin.Context) {
		raw, _ := c.Cookie(cfg.CookieName)
		if Actor(c).ID != 0 {
			Reply(c, 200, gin.H{"csrf_token": s.CSRF(raw)})
			return
		}
		token, e := services.Token()
		if e != nil {
			Fail(c, e)
			return
		}
		Cookie(c, cfg, cfg.CookieName+"_preauth", token, 1800)
		Reply(c, 200, gin.H{"csrf_token": token})
	})
	g.POST("/readers", limit, func(c *gin.Context) {
		var in schemas.RegisterReader
		if e := Bind(c, &in); e != nil {
			Fail(c, e)
			return
		}
		v, e := s.Register(c.Request.Context(), in)
		if e != nil {
			Fail(c, e)
			return
		}
		Reply(c, 201, v)
	})
	g.POST("/auth/login", limit, func(c *gin.Context) {
		var in schemas.Login
		if e := Bind(c, &in); e != nil {
			Fail(c, e)
			return
		}
		previous, _ := c.Cookie(cfg.CookieName)
		token, v, e := s.Login(c.Request.Context(), in, previous)
		if e != nil {
			Fail(c, e)
			return
		}
		Cookie(c, cfg, cfg.CookieName, token, int(cfg.SessionAbsolute.Seconds()))
		Cookie(c, cfg, cfg.CookieName+"_preauth", "", -1)
		Reply(c, 200, gin.H{"user": v, "csrf_token": s.CSRF(token)})
	})
	g.POST("/auth/logout", func(c *gin.Context) {
		raw, _ := c.Cookie(cfg.CookieName)
		if e := s.Logout(c.Request.Context(), Actor(c), raw); e != nil {
			Fail(c, e)
			return
		}
		Cookie(c, cfg, cfg.CookieName, "", -1)
		Reply(c, 204, nil)
	})
	g.POST("/auth/forgot-password", limit, func(c *gin.Context) {
		var in schemas.ForgotPassword
		if e := Bind(c, &in); e != nil {
			Fail(c, e)
			return
		}
		if e := s.Forgot(c.Request.Context(), in.Email); e != nil {
			Fail(c, e)
			return
		}
		Reply(c, 200, gin.H{"message": "Se a conta estiver disponível, enviaremos as instruções por e-mail."})
	})
	g.POST("/auth/reset-password", limit, func(c *gin.Context) {
		var in schemas.ResetPassword
		if e := Bind(c, &in); e != nil {
			Fail(c, e)
			return
		}
		if e := s.ResetPassword(c.Request.Context(), in); e != nil {
			Fail(c, e)
			return
		}
		Cookie(c, cfg, cfg.CookieName, "", -1)
		Reply(c, 204, nil)
	})
	g.GET("/me", auth, func(c *gin.Context) {
		v, e := s.Me(c.Request.Context(), Actor(c))
		if e != nil {
			Fail(c, e)
			return
		}
		Reply(c, 200, v)
	})
	g.PATCH("/me", auth, func(c *gin.Context) {
		var in schemas.PatchAccount
		if e := Bind(c, &in); e != nil {
			Fail(c, e)
			return
		}
		v, e := s.Patch(c.Request.Context(), Actor(c), in)
		if e != nil {
			Fail(c, e)
			return
		}
		Reply(c, 200, v)
	})
	g.DELETE("/me", auth, func(c *gin.Context) {
		var in schemas.DeleteAccount
		if e := Bind(c, &in); e != nil {
			Fail(c, e)
			return
		}
		if e := s.Delete(c.Request.Context(), Actor(c), in.Password); e != nil {
			Fail(c, e)
			return
		}
		Cookie(c, cfg, cfg.CookieName, "", -1)
		Reply(c, 204, nil)
	})
	g.PUT("/me/password", auth, func(c *gin.Context) {
		var in schemas.ChangePassword
		if e := Bind(c, &in); e != nil {
			Fail(c, e)
			return
		}
		if e := s.ChangePassword(c.Request.Context(), Actor(c), in); e != nil {
			Fail(c, e)
			return
		}
		Cookie(c, cfg, cfg.CookieName, "", -1)
		Reply(c, 204, nil)
	})
	g.GET("/readers/:id", func(c *gin.Context) {
		id, e := ID(c, "id")
		if e != nil {
			Fail(c, e)
			return
		}
		v, e := s.Public(c.Request.Context(), Actor(c), id)
		if e != nil {
			Fail(c, e)
			return
		}
		Reply(c, 200, v)
	})
}
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if Actor(c).ID == 0 {
			Fail(c, core.ErrUnauthenticated)
			return
		}
		c.Next()
	}
}
