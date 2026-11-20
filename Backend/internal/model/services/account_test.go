package services

import (
	"context"
	"errors"
	"radar-literario/backend/internal/config"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/schemas"
	"strings"
	"testing"
	"time"
)

type accountFake struct {
	interfaces.AccountRepository
	account             core.Account
	session             core.Session
	blocked, restricted bool
	touch               bool
	revoked             bool
	hash                string
	reset               core.Reset
	payload             string
	mailFinished        bool
}

func (f *accountFake) Within(ctx context.Context, fn func(interfaces.AccountRepository) error) error {
	return fn(f)
}
func (f *accountFake) Account(context.Context, uint64, bool) (core.Account, error) {
	return f.account, nil
}
func (f *accountFake) ByEmail(_ context.Context, email string) (core.Account, error) {
	if f.account.Email == email {
		return f.account, nil
	}
	return core.Account{}, core.NotFound()
}
func (f *accountFake) Session(context.Context, string) (core.Session, error) { return f.session, nil }
func (f *accountFake) Penalties(context.Context, uint64, time.Time) (bool, bool, error) {
	return f.blocked, f.restricted, nil
}
func (f *accountFake) TouchSession(context.Context, uint64, time.Time, time.Time) (bool, error) {
	f.touch = true
	return true, nil
}
func (f *accountFake) Revoke(context.Context, uint64, string, time.Time) error {
	f.revoked = true
	return nil
}
func (f *accountFake) SaveReset(_ context.Context, r core.Reset, _ string, p string) error {
	f.reset = r
	f.payload = p
	return nil
}
func (f *accountFake) ConsumeReset(_ context.Context, h string, _ time.Time) (uint64, error) {
	if h != f.reset.TokenHash {
		return 0, core.Invalid("INVALID_RESET_TOKEN", "inválido")
	}
	return f.reset.UserID, nil
}
func (f *accountFake) ChangeHash(_ context.Context, _ uint64, h string) error { f.hash = h; return nil }
func accountService(f *accountFake) *Accounts {
	return NewAccounts(f, config.Config{Env: "test", SessionIdle: 30 * time.Minute, SessionAbsolute: 7 * 24 * time.Hour, ResetTTL: 30 * time.Minute, EncryptionKey: make([]byte, 32), FrontendURL: "http://localhost:5173"}, nil)
}
func TestSessionsEnforceAbsoluteIdleRevocationAndPenalties(t *testing.T) {
	now := time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC)
	birth := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name    string
		session core.Session
		blocked bool
		want    bool
	}{{"valid", core.Session{ID: 1, UserID: 1, LastSeenAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Hour)}, false, true}, {"absolute", core.Session{ID: 1, UserID: 1, LastSeenAt: now, ExpiresAt: now}, false, false}, {"idle", core.Session{ID: 1, UserID: 1, LastSeenAt: now.Add(-30 * time.Minute), ExpiresAt: now.Add(time.Hour)}, false, false}, {"revoked", core.Session{ID: 1, UserID: 1, LastSeenAt: now, ExpiresAt: now.Add(time.Hour), RevokedAt: &now}, false, false}, {"penalty", core.Session{ID: 1, UserID: 1, LastSeenAt: now, ExpiresAt: now.Add(time.Hour)}, true, false}}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &accountFake{session: c.session, account: core.Account{User: core.User{ID: 1, Status: "active", BirthDate: &birth}, IsReader: true}, blocked: c.blocked}
			s := accountService(f)
			s.now = func() time.Time { return now }
			a, e := s.Authenticate(context.Background(), strings.Repeat("x", 43))
			if (e == nil) != c.want {
				t.Fatalf("a=%v err=%v", a, e)
			}
			if !c.want && f.touch {
				t.Fatal("invalid session touched")
			}
		})
	}
}
func TestRecoveryEncryptsOutboxAndHashesToken(t *testing.T) {
	f := &accountFake{account: core.Account{User: core.User{ID: 7, Email: "reader@example.test", Status: "active"}}}
	s := accountService(f)
	if e := s.Forgot(context.Background(), "reader@example.test"); e != nil {
		t.Fatal(e)
	}
	if f.reset.TokenHash == "" || len(f.reset.TokenHash) != 64 {
		t.Fatal("token was not hashed")
	}
	if strings.Contains(f.payload, "reset-password") || strings.Contains(f.payload, "reader@example.test") {
		t.Fatal("outbox leaked plaintext")
	}
	p, e := s.open(f.payload)
	if e != nil || !strings.Contains(p.Body, "/reset-password?token=") {
		t.Fatalf("payload unavailable: %v", e)
	}
	if strings.Contains(p.Body, f.reset.TokenHash) {
		t.Fatal("hash used as token")
	}
	changed := []byte(f.payload)
	changed[len(changed)/2] ^= 1
	if _, e = s.open(string(changed)); e == nil {
		t.Fatal("tampered outbox accepted")
	}
	if e = s.Forgot(context.Background(), "unknown@example.test"); e != nil {
		t.Fatal("recovery reveals account absence")
	}
}
func TestAccountValidationRejectsPrivilegeAndBadIdentityAtService(t *testing.T) {
	s := accountService(&accountFake{})
	for _, in := range []schemas.RegisterReader{{Name: "A", Nickname: "valid", Email: "a@example.test", Password: "longenough", BirthDate: "2000-01-01"}, {Name: "Valid", Nickname: "bad space", Email: "a@example.test", Password: "longenough", BirthDate: "2000-01-01"}, {Name: "Valid", Nickname: "valid", Email: "invalid", Password: "longenough", BirthDate: "2000-01-01"}, {Name: "Valid", Nickname: "valid", Email: "a@example.test", Password: "123", BirthDate: "2000-01-01"}, {Name: "Valid", Nickname: "valid", Email: "a@example.test", Password: "longenough", BirthDate: "2999-01-01"}} {
		_, e := s.Register(context.Background(), in)
		var d *core.Error
		if !errors.As(e, &d) || d.Status != 400 {
			t.Fatalf("invalid registration accepted: %v", e)
		}
	}
}
func TestAuthenticatedPasswordChangeChecksCurrentAndRevokes(t *testing.T) {
	hash, e := core.HashPassword("original-password")
	if e != nil {
		t.Fatal(e)
	}
	f := &accountFake{account: core.Account{User: core.User{ID: 1, PasswordHash: hash}}}
	s := accountService(f)
	in := schemas.ChangePassword{CurrentPassword: "wrong", NewPassword: "new-password"}
	if e = s.ChangePassword(context.Background(), core.Actor{ID: 1}, in); e == nil || f.revoked {
		t.Fatal("accepted wrong current password")
	}
	in.CurrentPassword = "original-password"
	if e = s.ChangePassword(context.Background(), core.Actor{ID: 1}, in); e != nil {
		t.Fatal(e)
	}
	if !f.revoked || !core.CheckPassword(f.hash, "new-password") {
		t.Fatal("password change not persisted/revoked")
	}
}
