//go:build integration

package tools_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"radar-literario/backend/internal/config"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/schemas"
	"radar-literario/backend/internal/model/services"
	"radar-literario/backend/internal/model/tools"
)

type capturedMail struct {
	mu                       sync.Mutex
	recipient, subject, body string
	failures                 bool
	calls                    int
}

func (m *capturedMail) Send(_ context.Context, to, subject, body string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recipient = to
	m.subject = subject
	m.body = body
	m.calls++
	if m.failures {
		return errors.New("controlled SMTP failure")
	}
	return nil
}

func accountTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		values, e := godotenv.Read(filepath.Join("..", "..", "..", ".env.test"))
		if e != nil {
			t.Skip("configure TEST_DB_DSN ou Backend/.env.test para MariaDB isolado")
		}
		dsn = values["TEST_DB_DSN"]
	}
	parsed, e := mysqlDriver.ParseDSN(dsn)
	if e != nil || !strings.HasSuffix(parsed.DBName, "_test") {
		t.Fatal("contas exigem TEST_DB_DSN em banco com sufixo _test")
	}
	db, e := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if e != nil {
		t.Fatalf("MariaDB de testes indisponível: %T", e)
	}
	pool, e := db.DB()
	if e != nil {
		t.Fatal(e)
	}
	pool.SetMaxOpenConns(12)
	t.Cleanup(func() { pool.Close() })
	return db
}
func accountConfig() config.Config {
	return config.Config{Env: "test", EncryptionKey: []byte("0123456789abcdef0123456789abcdef"), SessionIdle: 30 * time.Minute, SessionAbsolute: 24 * time.Hour, ResetTTL: 30 * time.Minute, FrontendURL: "http://localhost:5173", RetentionDays: 180}
}
func accountName() string { return fmt.Sprintf("a%x", time.Now().UnixNano()) }
func registerAccount(t *testing.T, s *services.Accounts, nickname string) schemas.AccountView {
	t.Helper()
	u, e := s.Register(context.Background(), schemas.RegisterReader{Name: "Conta fictícia", Nickname: nickname, Email: nickname + "@example.test", Password: "testing-account-password", BirthDate: "1990-01-01"})
	if e != nil {
		t.Fatal(e)
	}
	return u
}
func accountLogin(t *testing.T, s *services.Accounts, email string) string {
	t.Helper()
	token, _, e := s.Login(context.Background(), schemas.Login{Email: email, Password: "testing-account-password"}, "")
	if e != nil {
		t.Fatal(e)
	}
	return token
}
func mailToken(t *testing.T, body string) string {
	t.Helper()
	for _, part := range strings.Fields(body) {
		if strings.Contains(part, "/reset-password?token=") {
			u, e := url.Parse(part)
			if e != nil {
				t.Fatal(e)
			}
			token := u.Query().Get("token")
			if len(token) != 43 {
				t.Fatal("invalid reset link token")
			}
			return token
		}
	}
	t.Fatal("missing reset link")
	return ""
}

// Scope the outbox worker to the fixture recipient: integration tests must never
// deliver or alter messages belonging to other test fixtures.
type scopedMailRepository struct {
	interfaces.AccountRepository
	recipient string
}

type accountLookupInterleaving struct {
	interfaces.AccountRepository
	afterLookup func()
}

func (r accountLookupInterleaving) ByEmail(ctx context.Context, email string) (core.Account, error) {
	u, e := r.AccountRepository.ByEmail(ctx, email)
	if e == nil && r.afterLookup != nil {
		r.afterLookup()
	}
	return u, e
}

func (r scopedMailRepository) PendingMail(ctx context.Context, now time.Time) ([]core.Outbox, error) {
	rows, e := r.AccountRepository.PendingMail(ctx, now)
	if e != nil {
		return nil, e
	}
	var own []core.Outbox
	for _, v := range rows {
		if v.Recipient == r.recipient {
			own = append(own, v)
		}
	}
	return own, nil
}

func TestMariaDBAccountsConcurrentUniqueEmailAndNickname(t *testing.T) {
	db := accountTestDB(t)
	prefix := accountName()
	repo := tools.NewAccountRepository(tools.NewStore(db))
	service := services.NewAccounts(repo, accountConfig(), nil)
	t.Cleanup(func() {
		var ids []uint64
		db.Raw("SELECT id FROM users WHERE nickname LIKE ? OR email LIKE ?", prefix+"%", prefix+"%").Scan(&ids)
		for _, id := range ids {
			db.Exec("DELETE FROM readers WHERE user_id=?", id)
			db.Exec("DELETE FROM users WHERE id=?", id)
		}
	})
	for _, shared := range []string{"email", "nickname"} {
		t.Run(shared, func(t *testing.T) {
			const n = 6
			start := make(chan struct{})
			results := make(chan error, n)
			for i := 0; i < n; i++ {
				go func(i int) {
					<-start
					nick := fmt.Sprintf("%s%s%d", prefix, shared[:1], i)
					email := fmt.Sprintf("%s%s%d@example.test", prefix, shared[:1], i)
					if shared == "email" {
						email = prefix + "same@example.test"
					} else {
						nick = prefix + "same"
					}
					_, e := service.Register(context.Background(), schemas.RegisterReader{Name: "Concorrência de contas", Nickname: nick, Email: email, Password: "testing-account-password", BirthDate: "1990-01-01"})
					results <- e
				}(i)
			}
			close(start)
			success, conflict := 0, 0
			for i := 0; i < n; i++ {
				e := <-results
				if e == nil {
					success++
					continue
				}
				var d *core.Error
				if errors.As(e, &d) && d.Status == 409 {
					conflict++
				} else {
					t.Fatalf("unexpected concurrent registration error: %v", e)
				}
			}
			if success != 1 || conflict != n-1 {
				t.Fatalf("success=%d conflict=%d", success, conflict)
			}
		})
	}
}

func TestMariaDBAccountRecoveryUseOnceRevokesAndPersists(t *testing.T) {
	db := accountTestDB(t)
	name := accountName()
	base := tools.NewAccountRepository(tools.NewStore(db))
	capture := &capturedMail{}
	repo := scopedMailRepository{AccountRepository: base, recipient: name + "@example.test"}
	cfg := accountConfig()
	s := services.NewAccounts(repo, cfg, capture)
	u := registerAccount(t, s, name)
	ctx := context.Background()
	var ownOutbox []uint64
	t.Cleanup(func() {
		for _, id := range ownOutbox {
			db.Exec("DELETE FROM mail_outbox WHERE id=?", id)
		}
		db.Exec("DELETE FROM mail_outbox WHERE recipient=?", u.Email)
		db.Exec("DELETE FROM password_resets WHERE user_id=?", u.ID)
		db.Exec("DELETE FROM sessions WHERE user_id=?", u.ID)
		db.Exec("DELETE FROM readers WHERE user_id=?", u.ID)
		db.Exec("DELETE FROM users WHERE id=?", u.ID)
	})
	oldSession := accountLogin(t, s, u.Email)
	if _, e := s.Authenticate(ctx, oldSession); e != nil {
		t.Fatal(e)
	}
	if e := s.Forgot(ctx, u.Email); e != nil {
		t.Fatal(e)
	}
	var payload string
	db.Raw("SELECT payload FROM mail_outbox WHERE recipient=? ORDER BY id DESC LIMIT 1", u.Email).Scan(&payload)
	var firstOutbox uint64
	db.Raw("SELECT id FROM mail_outbox WHERE recipient=? ORDER BY id DESC LIMIT 1", u.Email).Scan(&firstOutbox)
	ownOutbox = append(ownOutbox, firstOutbox)
	if payload == "" || strings.Contains(payload, "reset-password") {
		t.Fatal("recovery outbox stored plaintext")
	}
	if e := s.DispatchMail(ctx); e != nil {
		t.Fatal(e)
	}
	token := mailToken(t, capture.body)
	var stored string
	db.Raw("SELECT token_hash FROM password_resets WHERE user_id=? ORDER BY id DESC LIMIT 1", u.ID).Scan(&stored)
	if stored == token || stored != services.TokenHash(token) {
		t.Fatal("reset token not stored as hash")
	}
	var delivered int64
	db.Raw("SELECT COUNT(*) FROM mail_outbox WHERE id=? AND delivered_at IS NOT NULL AND payload='' AND recipient=''", firstOutbox).Scan(&delivered)
	if delivered == 0 {
		t.Fatal("delivered outbox payload not cleared")
	}
	const n = 6
	start := make(chan struct{})
	results := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() {
			<-start
			results <- s.ResetPassword(ctx, schemas.ResetPassword{Token: token, Password: "changed-account-password"})
		}()
	}
	close(start)
	success, invalid := 0, 0
	for i := 0; i < n; i++ {
		e := <-results
		if e == nil {
			success++
			continue
		}
		var d *core.Error
		if errors.As(e, &d) && d.Status == 400 && d.Code == "INVALID_RESET_TOKEN" {
			invalid++
		} else {
			t.Fatalf("unexpected concurrent reset: %v", e)
		}
	}
	if success != 1 || invalid != n-1 {
		t.Fatalf("reset success=%d invalid=%d", success, invalid)
	}
	if _, e := s.Authenticate(ctx, oldSession); e == nil {
		t.Fatal("old session survived recovery")
	}
	restarted := services.NewAccounts(repo, cfg, capture)
	if _, _, e := restarted.Login(ctx, schemas.Login{Email: u.Email, Password: "testing-account-password"}, ""); e == nil {
		t.Fatal("old password still accepted after service restart")
	}
	if _, _, e := restarted.Login(ctx, schemas.Login{Email: u.Email, Password: "changed-account-password"}, ""); e != nil {
		t.Fatal(e)
	}
	if e := restarted.Forgot(ctx, u.Email); e != nil {
		t.Fatal(e)
	}
	var secondOutbox uint64
	db.Raw("SELECT id FROM mail_outbox WHERE recipient=? ORDER BY id DESC LIMIT 1", u.Email).Scan(&secondOutbox)
	ownOutbox = append(ownOutbox, secondOutbox)
	if e := restarted.DispatchMail(ctx); e != nil {
		t.Fatal(e)
	}
	expired := mailToken(t, capture.body)
	if e := db.Exec("UPDATE password_resets SET expires_at=? WHERE token_hash=?", time.Now().UTC().Add(-time.Minute), services.TokenHash(expired)).Error; e != nil {
		t.Fatal(e)
	}
	if e := restarted.ResetPassword(ctx, schemas.ResetPassword{Token: expired, Password: "another-account-password"}); e == nil {
		t.Fatal("expired reset token accepted")
	}
	if e := restarted.Forgot(ctx, "unknown-"+u.Email); e != nil {
		t.Fatal("forgot-password reveals account absence")
	}
}

func TestMariaDBAccountDeletionAndRetentionRemovePrivateData(t *testing.T) {
	db := accountTestDB(t)
	tx := db.Begin()
	defer tx.Rollback()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	repo := tools.NewAccountRepository(tools.NewStore(tx))
	capture := &capturedMail{}
	cfg := accountConfig()
	s := services.NewAccounts(repo, cfg, capture)
	u := registerAccount(t, s, accountName())
	ctx := context.Background()
	a := core.Actor{ID: u.ID, Reader: true}
	session := accountLogin(t, s, u.Email)
	if e := s.Forgot(ctx, u.Email); e != nil {
		t.Fatal(e)
	}
	now := time.Now().UTC()
	if e := tx.Exec("INSERT INTO support_tickets(requester_id,subject,status,created_at,updated_at) VALUES (?,?,'open',?,?)", u.ID, "Dados privados de suporte", now, now).Error; e != nil {
		t.Fatal(e)
	}
	var ticket uint64
	tx.Raw("SELECT id FROM support_tickets WHERE requester_id=? ORDER BY id DESC LIMIT 1", u.ID).Scan(&ticket)
	if e := tx.Exec("INSERT INTO support_messages(ticket_id,author_id,text,created_at) VALUES (?,?,?,?)", ticket, u.ID, "Mensagem pessoal para apagar", now).Error; e != nil {
		t.Fatal(e)
	}
	if e := s.Delete(ctx, a, "wrong-password"); e == nil {
		t.Fatal("account deleted without password confirmation")
	}
	if e := s.Delete(ctx, a, "testing-account-password"); e != nil {
		t.Fatal(e)
	}
	if _, e := s.Authenticate(ctx, session); e == nil {
		t.Fatal("deleted account session accepted")
	}
	var state struct {
		Name, Email, Nickname, PasswordHash, Status string
		BirthDate                                   *time.Time
	}
	tx.Raw("SELECT name,email,nickname,password_hash,status,birth_date FROM users WHERE id=?", u.ID).Scan(&state)
	if state.Status != "deactivated" || state.Email == u.Email || state.PasswordHash != "" || state.BirthDate != nil || state.Nickname == u.Nickname {
		t.Fatal("identity not anonymized")
	}
	var n int64
	tx.Raw("SELECT COUNT(*) FROM mail_outbox WHERE recipient=?", u.Email).Scan(&n)
	if n != 0 {
		t.Fatal("deleted account left pending recovery mail")
	}
	var support string
	tx.Raw("SELECT text FROM support_messages WHERE ticket_id=?", ticket).Scan(&support)
	if support != "" {
		t.Fatal("private support message remained")
	}
	// Create a removed publication older than the configured retention boundary.
	store := tools.NewStore(tx)
	work, e := store.Insert(ctx, "works", map[string]any{"title": "Obra retenção", "synopsis": "Fictícia", "original_language": "pt", "status": "published", "minimum_age": 0})
	if e != nil {
		t.Fatal(e)
	}
	before := now.AddDate(0, 0, -181)
	post, e := store.Insert(ctx, "posts", map[string]any{"reader_id": u.ID, "work_id": work, "kind": "review", "title": "Título privado", "text": "Evidência textual que deve desaparecer", "status": "removed", "contains_spoiler": false, "contains_sensitive_content": false, "minimum_age": 0, "created_at": before, "updated_at": before})
	if e != nil {
		t.Fatal(e)
	}
	if e := s.Purge(ctx); e != nil {
		t.Fatal(e)
	}
	var text string
	tx.Raw("SELECT text FROM posts WHERE id=?", post).Scan(&text)
	if strings.Contains(text, "Evidência textual") {
		t.Fatal("retention did not scrub evidence")
	}
}

func TestMariaDBRecoverySMTPFailureRemainsRetryable(t *testing.T) {
	db := accountTestDB(t)
	tx := db.Begin()
	defer tx.Rollback()
	base := tools.NewAccountRepository(tools.NewStore(tx))
	capture := &capturedMail{failures: true}
	name := accountName()
	repo := scopedMailRepository{AccountRepository: base, recipient: name + "@example.test"}
	s := services.NewAccounts(repo, accountConfig(), capture)
	u := registerAccount(t, s, name)
	if e := s.Forgot(context.Background(), u.Email); e != nil {
		t.Fatal(e)
	}
	if e := s.DispatchMail(context.Background()); e != nil {
		t.Fatal(e)
	}
	var row struct {
		Attempts    int
		Payload     string
		DeliveredAt *time.Time
		AvailableAt time.Time
	}
	tx.Raw("SELECT attempts,payload,delivered_at,available_at FROM mail_outbox WHERE recipient=? ORDER BY id DESC LIMIT 1", u.Email).Scan(&row)
	if row.Attempts != 1 || row.DeliveredAt != nil || row.Payload == "" || !row.AvailableAt.After(time.Now().UTC()) {
		t.Fatal("SMTP failure lost persistent retry")
	}
}

func TestMariaDBRecoveryDoesNotQueueAfterInterleavedDeletion(t *testing.T) {
	db := accountTestDB(t)
	tx := db.Begin()
	defer tx.Rollback()
	base := tools.NewAccountRepository(tools.NewStore(tx))
	cfg := accountConfig()
	s := services.NewAccounts(base, cfg, &capturedMail{})
	u := registerAccount(t, s, accountName())
	raced := accountLookupInterleaving{AccountRepository: base, afterLookup: func() {
		if e := s.Delete(context.Background(), core.Actor{ID: u.ID, Reader: true}, "testing-account-password"); e != nil {
			t.Fatal(e)
		}
	}}
	recovery := services.NewAccounts(raced, cfg, &capturedMail{})
	if e := recovery.Forgot(context.Background(), u.Email); e != nil {
		t.Fatal(e)
	}
	var n int64
	tx.Raw("SELECT COUNT(*) FROM password_resets WHERE user_id=?", u.ID).Scan(&n)
	if n != 0 {
		t.Fatal("stale recovery lookup created token after account deletion")
	}
	tx.Raw("SELECT COUNT(*) FROM mail_outbox WHERE recipient=?", u.Email).Scan(&n)
	if n != 0 {
		t.Fatal("stale recovery lookup queued mail after account deletion")
	}
}

func TestMariaDBEmailChangeInvalidatesPreviousRecovery(t *testing.T) {
	db := accountTestDB(t)
	tx := db.Begin()
	defer tx.Rollback()
	repo := tools.NewAccountRepository(tools.NewStore(tx))
	s := services.NewAccounts(repo, accountConfig(), &capturedMail{})
	u := registerAccount(t, s, accountName())
	ctx := context.Background()
	if e := s.Forgot(ctx, u.Email); e != nil {
		t.Fatal(e)
	}
	newEmail := "changed-" + u.Email
	if _, e := s.Patch(ctx, core.Actor{ID: u.ID, Reader: true}, schemas.PatchAccount{Email: &newEmail}); e != nil {
		t.Fatal(e)
	}
	var n int64
	tx.Raw("SELECT COUNT(*) FROM password_resets WHERE user_id=?", u.ID).Scan(&n)
	if n != 0 {
		t.Fatal("email change kept old recovery tokens")
	}
	tx.Raw("SELECT COUNT(*) FROM mail_outbox WHERE recipient=?", u.Email).Scan(&n)
	if n != 0 {
		t.Fatal("email change kept recovery mail addressed to previous email")
	}
}

func TestMariaDBRecoveryOutboxDeliversViaSMTP(t *testing.T) {
	db := accountTestDB(t)
	tx := db.Begin()
	defer tx.Rollback()
	cfg, received := testSMTPServer(t, false)
	baseCfg := accountConfig()
	baseCfg.SMTPHost = cfg.SMTPHost
	baseCfg.SMTPPort = cfg.SMTPPort
	baseCfg.SMTPFrom = cfg.SMTPFrom
	name := accountName()
	repo := scopedMailRepository{AccountRepository: tools.NewAccountRepository(tools.NewStore(tx)), recipient: name + "@example.test"}
	s := services.NewAccounts(repo, baseCfg, tools.NewSMTPMailer(baseCfg))
	u := registerAccount(t, s, name)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if e := s.Forgot(ctx, u.Email); e != nil {
		t.Fatal(e)
	}
	if e := s.DispatchMail(ctx); e != nil {
		t.Fatal(e)
	}
	select {
	case message := <-received:
		if message.err != nil {
			t.Fatal(message.err)
		}
		token := mailToken(t, message.body)
		if e := s.ResetPassword(ctx, schemas.ResetPassword{Token: token, Password: "password-from-smtp-link"}); e != nil {
			t.Fatal(e)
		}
	case <-ctx.Done():
		t.Fatal("recovery mail did not reach local SMTP capture")
	}
}
