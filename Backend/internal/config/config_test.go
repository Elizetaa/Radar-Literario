package config

import (
	"encoding/base64"
	"strings"
	"testing"
)

func base(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", "test")
	t.Setenv("DB_USER", "test_app")
	t.Setenv("DB_PASSWORD", "secret-local-test")
	t.Setenv("OUTBOX_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	t.Setenv("COOKIE_SECURE", "false")
	t.Setenv("COOKIE_SAME_SITE", "lax")
}
func TestConfigRejectsMissingSecretsRootAndInvalidBounds(t *testing.T) {
	for _, c := range []struct{ k, v, code string }{{"OUTBOX_ENCRYPTION_KEY", "", "OUTBOX_ENCRYPTION_KEY"}, {"DB_USER", "root", "DB_USER"}, {"DB_PASSWORD", "", "DB_PASSWORD"}, {"DB_PORT", "99999", "limites"}, {"DB_MAX_IDLE", "99", "limites"}, {"SESSION_IDLE_TIMEOUT", "999h", "limites"}, {"CORS_ALLOWED_ORIGINS", "*", "origem"}, {"COOKIE_SAME_SITE", "none", "SameSite"}, {"MAX_BODY_BYTES", "0", "MAX_BODY_BYTES"}} {
		t.Run(c.k, func(t *testing.T) {
			base(t)
			t.Setenv(c.k, c.v)
			_, e := Load()
			if e == nil || !strings.Contains(e.Error(), c.code) {
				t.Fatalf("expected %s error, got %v", c.code, e)
			}
		})
	}
}
func TestConfigProductionNeedsSecureCookiesSMTPAndFrontend(t *testing.T) {
	base(t)
	t.Setenv("APP_ENV", "production")
	if _, e := Load(); e == nil {
		t.Fatal("production accepted insecure defaults")
	}
	t.Setenv("COOKIE_SECURE", "true")
	t.Setenv("SMTP_TLS", "true")
	t.Setenv("SMTP_HOST", "smtp.example.org")
	t.Setenv("SMTP_FROM", "radar@example.org")
	t.Setenv("PUBLIC_FRONTEND_URL", "https://radar.example.org")
	if _, e := Load(); e != nil {
		t.Fatal(e)
	}
}
