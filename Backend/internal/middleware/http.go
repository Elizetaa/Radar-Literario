package middleware

import (
	"context"
	"crypto/subtle"
	"errors"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"radar-literario/backend/internal/config"
	"radar-literario/backend/internal/controller"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/services"
	"strings"
	"sync"
	"time"
)

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
func HTTP(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, e := services.Token()
		if e != nil {
			controller.Fail(c, e)
			return
		}
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Cache-Control", "no-store")
		start := time.Now()
		defer func() {
			if p := recover(); p != nil {
				slog.Error("pânico recuperado", "request_id", id)
				controller.Fail(c, core.E(500, "INTERNAL_ERROR", "Não foi possível concluir a operação."))
			}
			slog.Info("http", "request_id", id, "method", c.Request.Method, "route", c.FullPath(), "status", c.Writer.Status(), "duration_ms", time.Since(start).Milliseconds())
		}()
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, cfg.BodyLimit)
		ctx, cancel := context.WithTimeout(c.Request.Context(), cfg.RequestTimeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
func CORS(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Vary", "Origin")
			if !contains(cfg.CORSOrigins, origin) {
				controller.Fail(c, core.E(403, "ORIGIN_DENIED", "Origem não autorizada."))
				return
			}
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
func Session(accounts *services.Accounts, cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, e := c.Cookie(cfg.CookieName)
		if e == nil && token != "" {
			a, e := accounts.Authenticate(c.Request.Context(), token)
			if e == nil {
				c.Set("actor", a)
			} else {
				var de *core.Error
				if !errors.As(e, &de) || de.Status != 401 {
					controller.Fail(c, core.E(503, "DEPENDENCY_UNAVAILABLE", "O serviço de autenticação está temporariamente indisponível."))
					return
				}
			}
		}
		c.Next()
	}
}
func CSRF(accounts *services.Accounts, cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		origin := c.GetHeader("Origin")
		if origin != "" && !contains(cfg.CSRFOrigins, origin) {
			controller.Fail(c, core.E(403, "CSRF_ORIGIN_DENIED", "Origem não autorizada para escrita."))
			return
		}
		if strings.EqualFold(c.GetHeader("Sec-Fetch-Site"), "cross-site") && origin == "" {
			controller.Fail(c, core.E(403, "CSRF_INVALID", "Confirme a origem da solicitação."))
			return
		}
		supplied := c.GetHeader("X-CSRF-Token")
		var expected string
		if controller.Actor(c).ID != 0 {
			token, _ := c.Cookie(cfg.CookieName)
			expected = accounts.CSRF(token)
		} else {
			expected, _ = c.Cookie(cfg.CookieName + "_preauth")
		}
		if len(expected) != 43 || len(supplied) != 43 || subtle.ConstantTimeCompare([]byte(expected), []byte(supplied)) != 1 {
			controller.Fail(c, core.E(403, "CSRF_INVALID", "Obtenha um token CSRF e envie no cabeçalho X-CSRF-Token."))
			return
		}
		c.Next()
	}
}

type bucket struct {
	count int
	until time.Time
}
type Limiter struct {
	mu      sync.Mutex
	buckets map[string]bucket
	limit   int
}

func NewLimiter(limit int) *Limiter { return &Limiter{buckets: map[string]bucket{}, limit: limit} }
func (l *Limiter) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		now := time.Now()
		l.mu.Lock()
		if len(l.buckets) > 10000 {
			for k, v := range l.buckets {
				if now.After(v.until) {
					delete(l.buckets, k)
				}
			}
		}
		b := l.buckets[key]
		if now.After(b.until) {
			b = bucket{until: now.Add(time.Minute)}
		}
		b.count++
		l.buckets[key] = b
		l.mu.Unlock()
		if b.count > l.limit {
			c.Header("Retry-After", "60")
			controller.Fail(c, core.E(429, "RATE_LIMITED", "Muitas tentativas. Aguarde um minuto."))
			return
		}
		c.Next()
	}
}
