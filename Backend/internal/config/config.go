package config

import (
	"encoding/base64"
	"fmt"
	"github.com/joho/godotenv"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env, HTTPHost                                             string
	HTTPPort                                                  int
	DBHost, DBName, DBUser, DBPassword                        string
	DBPort, DBMaxOpen, DBMaxIdle                              int
	DBLifetime, DBTimeout                                     time.Duration
	CookieName, SameSite                                      string
	CookieSecure                                              bool
	SessionIdle, SessionAbsolute, ResetTTL, RequestTimeout    time.Duration
	CORSOrigins, CSRFOrigins                                  []string
	EncryptionKey                                             []byte
	SMTPHost, SMTPUser, SMTPPassword, SMTPFrom                string
	SMTPPort                                                  int
	SMTPTLS                                                   bool
	FrontendURL                                               string
	BodyLimit                                                 int64
	RateLimit, MinimumReviews, MaxCommentDepth, RetentionDays int
	BlockedTerms, BlockedDomains                              []string
}

func str(k, d string) string {
	if v, ok := os.LookupEnv(k); ok {
		return v
	}
	return d
}
func list(k, d string) []string {
	var r []string
	for _, v := range strings.Split(str(k, d), ",") {
		if v = strings.TrimSpace(v); v != "" {
			r = append(r, v)
		}
	}
	return r
}
func Load() (Config, error) {
	if e := godotenv.Load(".env"); e != nil && !os.IsNotExist(e) {
		return Config{}, fmt.Errorf("não foi possível ler .env: %w", e)
	}
	c := Config{Env: str("APP_ENV", "development"), HTTPHost: str("HTTP_HOST", "127.0.0.1"), DBHost: str("DB_HOST", "127.0.0.1"), DBName: str("DB_NAME", "radar_literario"), DBUser: str("DB_USER", "radar_app"), DBPassword: str("DB_PASSWORD", ""), CookieName: str("SESSION_COOKIE_NAME", "radar_session"), SameSite: str("COOKIE_SAME_SITE", "lax"), CORSOrigins: list("CORS_ALLOWED_ORIGINS", "http://localhost:5173"), CSRFOrigins: list("CSRF_ALLOWED_ORIGINS", "http://localhost:5173"), SMTPHost: str("SMTP_HOST", "localhost"), SMTPUser: str("SMTP_USER", ""), SMTPPassword: str("SMTP_PASSWORD", ""), SMTPFrom: str("SMTP_FROM", "nao-responda@example.test"), FrontendURL: str("PUBLIC_FRONTEND_URL", "http://localhost:5173"), BlockedTerms: list("CONTENT_BLOCKED_TERMS", ""), BlockedDomains: list("CONTENT_BLOCKED_DOMAINS", "")}
	ints := []struct {
		k, d string
		p    *int
	}{{"HTTP_PORT", "8080", &c.HTTPPort}, {"DB_PORT", "3306", &c.DBPort}, {"DB_MAX_OPEN", "20", &c.DBMaxOpen}, {"DB_MAX_IDLE", "5", &c.DBMaxIdle}, {"SMTP_PORT", "1025", &c.SMTPPort}, {"AUTH_RATE_LIMIT", "30", &c.RateLimit}, {"RANKING_MINIMUM_REVIEWS", "1", &c.MinimumReviews}, {"COMMENT_MAX_DEPTH", "5", &c.MaxCommentDepth}, {"RETENTION_DAYS", "180", &c.RetentionDays}}
	for _, v := range ints {
		n, e := strconv.Atoi(str(v.k, v.d))
		if e != nil || n < 1 {
			return c, fmt.Errorf("%s deve ser inteiro positivo", v.k)
		}
		*v.p = n
	}
	durations := []struct {
		k, d string
		p    *time.Duration
	}{{"DB_CONN_MAX_LIFETIME", "5m", &c.DBLifetime}, {"DB_TIMEOUT", "5s", &c.DBTimeout}, {"SESSION_IDLE_TIMEOUT", "30m", &c.SessionIdle}, {"SESSION_ABSOLUTE_TIMEOUT", "168h", &c.SessionAbsolute}, {"PASSWORD_RESET_TTL", "30m", &c.ResetTTL}, {"REQUEST_TIMEOUT", "15s", &c.RequestTimeout}}
	for _, v := range durations {
		d, e := time.ParseDuration(str(v.k, v.d))
		if e != nil || d <= 0 {
			return c, fmt.Errorf("%s deve ter duração positiva como 30m", v.k)
		}
		*v.p = d
	}
	var e error
	c.BodyLimit, e = strconv.ParseInt(str("MAX_BODY_BYTES", "1048576"), 10, 64)
	if e != nil || c.BodyLimit < 1 {
		return c, fmt.Errorf("MAX_BODY_BYTES inválido")
	}
	c.CookieSecure, e = strconv.ParseBool(str("COOKIE_SECURE", "false"))
	if e != nil {
		return c, fmt.Errorf("COOKIE_SECURE inválido")
	}
	c.SMTPTLS, e = strconv.ParseBool(str("SMTP_TLS", "false"))
	if e != nil {
		return c, fmt.Errorf("SMTP_TLS inválido")
	}
	c.EncryptionKey, e = base64.StdEncoding.DecodeString(str("OUTBOX_ENCRYPTION_KEY", ""))
	if e != nil || len(c.EncryptionKey) != 32 {
		return c, fmt.Errorf("OUTBOX_ENCRYPTION_KEY deve conter 32 bytes em base64")
	}
	if c.DBPassword == "" || c.DBPassword == "defina_localmente" || c.DBUser == "root" {
		return c, fmt.Errorf("configure DB_PASSWORD e um DB_USER próprio, diferente de root")
	}
	if c.SameSite != "lax" && c.SameSite != "strict" && c.SameSite != "none" {
		return c, fmt.Errorf("COOKIE_SAME_SITE inválido")
	}
	if c.SameSite == "none" && !c.CookieSecure {
		return c, fmt.Errorf("SameSite none exige COOKIE_SECURE")
	}
	for _, origin := range append(append([]string{}, c.CORSOrigins...), c.CSRFOrigins...) {
		u, e := url.Parse(origin)
		if e != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.Path != "" || u.RawQuery != "" || u.User != nil {
			return c, fmt.Errorf("origem HTTP inválida: %s", origin)
		}
	}
	u, e := url.Parse(c.FrontendURL)
	if e != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return c, fmt.Errorf("PUBLIC_FRONTEND_URL inválida")
	}
	if c.SessionIdle > c.SessionAbsolute || c.DBMaxIdle > c.DBMaxOpen || c.DBPort > 65535 || c.HTTPPort > 65535 {
		return c, fmt.Errorf("limites de configuração incoerentes")
	}
	if c.Env != "development" && c.Env != "test" && c.Env != "production" {
		return c, fmt.Errorf("APP_ENV inválido")
	}
	if c.Env == "production" && (!c.CookieSecure || !c.SMTPTLS || os.Getenv("SMTP_HOST") == "" || os.Getenv("SMTP_FROM") == "" || strings.HasSuffix(c.SMTPFrom, ".test") || len(c.DBPassword) < 12 || u.Scheme != "https") {
		return c, fmt.Errorf("produção exige cookies seguros, SMTP com TLS, senha forte de banco e frontend HTTPS")
	}
	return c, nil
}
func (c Config) Address() string { return net.JoinHostPort(c.HTTPHost, strconv.Itoa(c.HTTPPort)) }
