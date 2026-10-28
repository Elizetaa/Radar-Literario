package router

import (
	"context"
	"github.com/gin-gonic/gin"
	"radar-literario/backend/internal/config"
	"radar-literario/backend/internal/controller"
	"radar-literario/backend/internal/middleware"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/services"
)

type Dependencies struct {
	Accounts *services.Accounts
	Catalog  *services.Catalog
	Reading  *services.Reading
	Admin    *services.Admin
	Ready    func(context.Context) error
}

func New(cfg config.Config, d Dependencies) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.Use(middleware.HTTP(cfg), middleware.CORS(cfg))
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/ready", func(c *gin.Context) {
		if e := d.Ready(c.Request.Context()); e != nil {
			controller.Fail(c, core.E(503, "NOT_READY", "Serviço temporariamente indisponível."))
			return
		}
		c.JSON(200, gin.H{"status": "ready"})
	})
	g := r.Group("/api/v1", middleware.Session(d.Accounts, cfg), middleware.CSRF(d.Accounts, cfg))
	auth := controller.RequireAuth()
	controller.RegisterAccounts(g, d.Accounts, cfg, auth, middleware.NewLimiter(cfg.RateLimit).Handler())
	controller.RegisterCatalog(g, d.Catalog, auth)
	controller.RegisterReading(g, d.Reading, auth)
	controller.RegisterAdmin(g, d.Admin, auth)
	r.NoRoute(func(c *gin.Context) { controller.Fail(c, core.NotFound()) })
	return r
}
