package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"radar-literario/backend/internal/config"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/interfaces"
	"radar-literario/backend/internal/model/schemas"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

type Accounts struct {
	repo      interfaces.AccountRepository
	cfg       config.Config
	mailer    interfaces.Mailer
	dummyHash string
	now       func() time.Time
}

func NewAccounts(repo interfaces.AccountRepository, cfg config.Config, mailer interfaces.Mailer) *Accounts {
	dummy, _ := core.HashPassword("unavailable-account-password")
	return &Accounts{repo: repo, cfg: cfg, mailer: mailer, dummyHash: dummy, now: func() time.Time { return time.Now().UTC() }}
}
func Token() (string, error) {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		return "", e
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func TokenHash(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func (s *Accounts) CSRF(token string) string {
	h := hmac.New(sha256.New, s.cfg.EncryptionKey)
	h.Write([]byte("csrf:" + token))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

var nicknamePattern = regexp.MustCompile(`^[a-z0-9_]{3,30}$`)

func validIdentity(name, nickname, email string) error {
	a, e := mail.ParseAddress(email)
	if utf8.RuneCountInString(strings.TrimSpace(name)) < 2 || utf8.RuneCountInString(name) > 120 {
		return core.Invalid("INVALID_NAME", "Informe um nome de 2 a 120 caracteres.")
	}
	if !nicknamePattern.MatchString(nickname) {
		return core.Invalid("INVALID_NICKNAME", "Use apelido de 3 a 30 letras minúsculas, números ou sublinhado.")
	}
	if e != nil || a.Address != email || len(email) > 254 {
		return core.Invalid("INVALID_EMAIL", "Informe um e-mail válido.")
	}
	return nil
}
func (s *Accounts) birth(value string, required bool) (*time.Time, error) {
	if value == "" && required {
		return nil, core.Invalid("INVALID_BIRTH_DATE", "Informe a data de nascimento.")
	}
	d, e := core.Date(value)
	if e != nil {
		return nil, e
	}
	if d != nil && (d.After(s.now()) || d.Year() < 1900) {
		return nil, core.Invalid("INVALID_BIRTH_DATE", "Informe uma data de nascimento válida, entre 1900 e hoje.")
	}
	return d, nil
}
func (s *Accounts) Register(ctx context.Context, in schemas.RegisterReader) (schemas.AccountView, error) {
	var out schemas.AccountView
	in.Name = strings.TrimSpace(in.Name)
	in.Nickname = core.NormalizeNickname(in.Nickname)
	in.Email = core.NormalizeEmail(in.Email)
	if e := validIdentity(in.Name, in.Nickname, in.Email); e != nil {
		return out, e
	}
	if e := core.ValidatePassword(in.Password); e != nil {
		return out, e
	}
	birth, e := s.birth(in.BirthDate, true)
	if e != nil {
		return out, e
	}
	hash, e := core.HashPassword(in.Password)
	if e != nil {
		return out, e
	}
	var id uint64
	now := s.now()
	e = s.repo.Within(ctx, func(r interfaces.AccountRepository) error {
		var e error
		id, e = r.Create(ctx, core.User{Name: in.Name, Nickname: in.Nickname, Email: in.Email, PasswordHash: hash, BirthDate: birth, CreatedAt: now, UpdatedAt: now})
		return e
	})
	if e != nil {
		return out, e
	}
	return s.Me(ctx, core.Actor{ID: id})
}
func accountView(a core.Account) schemas.AccountView {
	var b *string
	if a.BirthDate != nil {
		v := a.BirthDate.Format("2006-01-02")
		b = &v
	}
	return schemas.AccountView{ID: a.ID, Name: a.Name, Nickname: a.Nickname, Email: a.Email, BirthDate: b, ProfileImageURL: a.ProfileImageURL, Biography: a.Biography, ProfileVisibility: a.ProfileVisibility, Sex: a.Sex, CountryOfOrigin: a.CountryOfOrigin, FavoriteWorkID: a.FavoriteWorkID, Role: a.Role, CreatedAt: a.CreatedAt}
}
func (s *Accounts) Me(ctx context.Context, a core.Actor) (schemas.AccountView, error) {
	if a.ID == 0 {
		return schemas.AccountView{}, core.ErrUnauthenticated
	}
	u, e := s.repo.Account(ctx, a.ID, false)
	return accountView(u), e
}
func (s *Accounts) Public(ctx context.Context, a core.Actor, id uint64) (schemas.ReaderView, error) {
	u, e := s.repo.Account(ctx, id, false)
	if e != nil {
		return schemas.ReaderView{}, e
	}
	if !u.IsReader || u.Status != "active" || (u.ProfileVisibility == "private" && a.ID != id) || (u.ProfileVisibility == "registered_users_only" && a.ID == 0) {
		return schemas.ReaderView{}, core.NotFound()
	}
	favorite := u.FavoriteWorkID
	if favorite != nil && s.repo.WorkEligible(ctx, *favorite, a.Age(s.now())) != nil {
		favorite = nil
	}
	return schemas.ReaderView{ID: u.ID, Name: u.Name, Nickname: u.Nickname, Biography: u.Biography, ProfileImageURL: u.ProfileImageURL, FavoriteWorkID: favorite}, nil
}
func (s *Accounts) Patch(ctx context.Context, a core.Actor, in schemas.PatchAccount) (schemas.AccountView, error) {
	var out schemas.AccountView
	if a.ID == 0 {
		return out, core.ErrUnauthenticated
	}
	e := s.repo.Within(ctx, func(r interfaces.AccountRepository) error {
		u, e := r.Account(ctx, a.ID, true)
		if e != nil {
			return e
		}
		user := map[string]any{}
		reader := map[string]any{}
		if in.Name != nil {
			u.Name = strings.TrimSpace(*in.Name)
			user["name"] = u.Name
		}
		if in.Nickname != nil {
			u.Nickname = core.NormalizeNickname(*in.Nickname)
			user["nickname"] = u.Nickname
		}
		if in.Email != nil {
			u.Email = core.NormalizeEmail(*in.Email)
			user["email"] = u.Email
		}
		if e = validIdentity(u.Name, u.Nickname, u.Email); e != nil {
			return e
		}
		if in.BirthDate != nil {
			d, e := s.birth(*in.BirthDate, true)
			if e != nil {
				return e
			}
			user["birth_date"] = d
		}
		if in.ProfileImageURL != nil {
			if e := core.ValidateURL(*in.ProfileImageURL); e != nil {
				return e
			}
			user["profile_image_url"] = *in.ProfileImageURL
		}
		if in.Biography != nil {
			if utf8.RuneCountInString(*in.Biography) > 2000 {
				return core.Invalid("INVALID_BIOGRAPHY", "A biografia aceita até 2000 caracteres.")
			}
			reader["biography"] = *in.Biography
		}
		if in.ProfileVisibility != nil {
			v := *in.ProfileVisibility
			if v != "public" && v != "registered_users_only" && v != "private" {
				return core.Invalid("INVALID_VISIBILITY", "Visibilidade inválida.")
			}
			reader["profile_visibility"] = v
		}
		if in.Sex.Set {
			if in.Sex.Value != nil && utf8.RuneCountInString(*in.Sex.Value) > 50 {
				return core.Invalid("INVALID_SEX", "O campo aceita até 50 caracteres.")
			}
			reader["sex"] = in.Sex.Value
		}
		if in.CountryOfOrigin.Set {
			if in.CountryOfOrigin.Value != nil && utf8.RuneCountInString(*in.CountryOfOrigin.Value) > 100 {
				return core.Invalid("INVALID_COUNTRY", "O país aceita até 100 caracteres.")
			}
			reader["country_of_origin"] = in.CountryOfOrigin.Value
		}
		if in.FavoriteWorkID.Set {
			if in.FavoriteWorkID.Value != nil {
				if e := r.WorkEligible(ctx, *in.FavoriteWorkID.Value, a.Age(s.now())); e != nil {
					return e
				}
			}
			reader["favorite_work_id"] = in.FavoriteWorkID.Value
		}
		if len(user) == 0 && len(reader) == 0 {
			return core.Invalid("EMPTY_UPDATE", "Informe ao menos um campo.")
		}
		return r.UpdateAccount(ctx, a.ID, user, reader)
	})
	if e != nil {
		return out, e
	}
	return s.Me(ctx, a)
}
func (s *Accounts) Login(ctx context.Context, in schemas.Login, previous string) (string, schemas.AccountView, error) {
	var out schemas.AccountView
	u, e := s.repo.ByEmail(ctx, core.NormalizeEmail(in.Email))
	hash := s.dummyHash
	if e == nil {
		hash = u.PasswordHash
	}
	valid := core.CheckPassword(hash, in.Password)
	if e != nil {
		var de *core.Error
		if !errors.As(e, &de) || de.Status != 404 {
			return "", out, e
		}
	}
	if e != nil || !valid || u.Status != "active" {
		return "", out, core.E(401, "INVALID_CREDENTIALS", "E-mail ou senha inválidos.")
	}
	now := s.now()
	blocked, _, e := s.repo.Penalties(ctx, u.ID, now)
	if e != nil {
		return "", out, e
	}
	if blocked {
		return "", out, core.E(401, "INVALID_CREDENTIALS", "E-mail ou senha inválidos.")
	}
	token, e := Token()
	if e != nil {
		return "", out, e
	}
	e = s.repo.Within(ctx, func(r interfaces.AccountRepository) error {
		locked, e := r.Account(ctx, u.ID, true)
		if e != nil {
			return e
		}
		if locked.Status != "active" || locked.PasswordHash != u.PasswordHash {
			return core.ErrUnauthenticated
		}
		blocked, _, e := r.Penalties(ctx, u.ID, now)
		if e != nil {
			return e
		}
		if blocked {
			return core.ErrUnauthenticated
		}
		if previous != "" {
			if e = r.RevokeToken(ctx, TokenHash(previous), now); e != nil {
				return e
			}
		}
		return r.AddSession(ctx, core.Session{UserID: u.ID, TokenHash: TokenHash(token), CSRFHash: TokenHash(s.CSRF(token)), CreatedAt: now, LastSeenAt: now, ExpiresAt: now.Add(s.cfg.SessionAbsolute)})
	})
	return token, accountView(u), e
}
func (s *Accounts) Authenticate(ctx context.Context, token string) (core.Actor, error) {
	if len(token) != 43 {
		return core.Actor{}, core.ErrUnauthenticated
	}
	session, e := s.repo.Session(ctx, TokenHash(token))
	if e != nil {
		return core.Actor{}, e
	}
	now := s.now()
	if session.RevokedAt != nil || !now.Before(session.ExpiresAt) || !now.Before(session.LastSeenAt.Add(s.cfg.SessionIdle)) {
		return core.Actor{}, core.ErrUnauthenticated
	}
	u, e := s.repo.Account(ctx, session.UserID, false)
	if e != nil {
		return core.Actor{}, e
	}
	if u.Status != "active" {
		return core.Actor{}, core.ErrUnauthenticated
	}
	blocked, restricted, e := s.repo.Penalties(ctx, u.ID, now)
	if e != nil {
		return core.Actor{}, e
	}
	if blocked {
		return core.Actor{}, core.ErrUnauthenticated
	}
	ok, e := s.repo.TouchSession(ctx, session.ID, now, now.Add(-s.cfg.SessionIdle))
	if e != nil {
		return core.Actor{}, e
	}
	if !ok {
		return core.Actor{}, core.ErrUnauthenticated
	}
	return core.Actor{ID: u.ID, Role: u.Role, Reader: u.IsReader, BirthDate: u.BirthDate, Restricted: restricted}, nil
}
func (s *Accounts) Logout(ctx context.Context, a core.Actor, token string) error {
	if a.ID == 0 {
		return nil
	}
	return s.repo.Revoke(ctx, a.ID, TokenHash(token), s.now())
}
func (s *Accounts) ChangePassword(ctx context.Context, a core.Actor, in schemas.ChangePassword) error {
	if a.ID == 0 {
		return core.ErrUnauthenticated
	}
	if e := core.ValidatePassword(in.NewPassword); e != nil {
		return e
	}
	hash, e := core.HashPassword(in.NewPassword)
	if e != nil {
		return e
	}
	return s.repo.Within(ctx, func(r interfaces.AccountRepository) error {
		u, e := r.Account(ctx, a.ID, true)
		if e != nil {
			return e
		}
		if !core.CheckPassword(u.PasswordHash, in.CurrentPassword) {
			return core.E(403, "INVALID_PASSWORD", "Senha atual inválida.")
		}
		if e = r.ChangeHash(ctx, a.ID, hash); e != nil {
			return e
		}
		return r.Revoke(ctx, a.ID, "", s.now())
	})
}
func (s *Accounts) Delete(ctx context.Context, a core.Actor, password string) error {
	if a.ID == 0 {
		return core.ErrUnauthenticated
	}
	return s.repo.Within(ctx, func(r interfaces.AccountRepository) error {
		u, e := r.Account(ctx, a.ID, true)
		if e != nil {
			return e
		}
		if !core.CheckPassword(u.PasswordHash, password) {
			return core.E(403, "INVALID_PASSWORD", "Senha atual inválida.")
		}
		return r.Anonymize(ctx, a.ID, s.now())
	})
}

type mailPayload struct{ Subject, Body string }

func (s *Accounts) seal(payload mailPayload) (string, error) {
	b, e := json.Marshal(payload)
	if e != nil {
		return "", e
	}
	block, e := aes.NewCipher(s.cfg.EncryptionKey)
	if e != nil {
		return "", e
	}
	gcm, e := cipher.NewGCM(block)
	if e != nil {
		return "", e
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, e = rand.Read(nonce); e != nil {
		return "", e
	}
	return base64.StdEncoding.EncodeToString(gcm.Seal(nonce, nonce, b, nil)), nil
}
func (s *Accounts) open(payload string) (mailPayload, error) {
	var p mailPayload
	b, e := base64.StdEncoding.DecodeString(payload)
	if e != nil {
		return p, e
	}
	block, e := aes.NewCipher(s.cfg.EncryptionKey)
	if e != nil {
		return p, e
	}
	gcm, e := cipher.NewGCM(block)
	if e != nil {
		return p, e
	}
	if len(b) < gcm.NonceSize() {
		return p, fmt.Errorf("invalid encrypted outbox")
	}
	plain, e := gcm.Open(nil, b[:gcm.NonceSize()], b[gcm.NonceSize():], nil)
	if e != nil {
		return p, e
	}
	e = json.Unmarshal(plain, &p)
	return p, e
}
func (s *Accounts) Forgot(ctx context.Context, email string) error {
	u, e := s.repo.ByEmail(ctx, core.NormalizeEmail(email))
	if e != nil {
		var d *core.Error
		if errors.As(e, &d) && d.Status == 404 {
			return nil
		}
		return e
	}
	if u.Status != "active" {
		return nil
	}
	token, e := Token()
	if e != nil {
		return e
	}
	now := s.now()
	link := strings.TrimRight(s.cfg.FrontendURL, "/") + "/reset-password?token=" + url.QueryEscape(token)
	payload, e := s.seal(mailPayload{Subject: "Recuperação de senha - Radar Literário", Body: "Use o link para definir uma nova senha. Ele expira em " + s.cfg.ResetTTL.String() + ".\n\n" + link + "\n\nSe você não solicitou, ignore esta mensagem."})
	if e != nil {
		return e
	}
	return s.repo.Within(ctx, func(r interfaces.AccountRepository) error {
		current, e := r.Account(ctx, u.ID, true)
		if e != nil {
			return e
		}
		if current.Status != "active" || current.Email != u.Email {
			return nil
		}
		return r.SaveReset(ctx, core.Reset{UserID: u.ID, TokenHash: TokenHash(token), CreatedAt: now, ExpiresAt: now.Add(s.cfg.ResetTTL)}, u.Email, payload)
	})
}
func (s *Accounts) ResetPassword(ctx context.Context, in schemas.ResetPassword) error {
	if e := core.ValidatePassword(in.Password); e != nil {
		return e
	}
	if len(in.Token) != 43 {
		return core.Invalid("INVALID_RESET_TOKEN", "O link de recuperação é inválido ou expirou.")
	}
	hash, e := core.HashPassword(in.Password)
	if e != nil {
		return e
	}
	return s.repo.Within(ctx, func(r interfaces.AccountRepository) error {
		id, e := r.ConsumeReset(ctx, TokenHash(in.Token), s.now())
		if e != nil {
			return e
		}
		if e = r.ChangeHash(ctx, id, hash); e != nil {
			return e
		}
		return r.Revoke(ctx, id, "", s.now())
	})
}
func (s *Accounts) DispatchMail(ctx context.Context) error {
	rows, e := s.repo.PendingMail(ctx, s.now())
	if e != nil {
		return e
	}
	for _, row := range rows {
		claimed, e := s.repo.ClaimMail(ctx, row.ID, s.now())
		if e != nil {
			return e
		}
		if !claimed {
			continue
		}
		p, e := s.open(row.Payload)
		if e == nil {
			e = s.mailer.Send(ctx, row.Recipient, p.Subject, p.Body)
		}
		if finish := s.repo.FinishMail(ctx, row.ID, e == nil, s.now()); finish != nil {
			return finish
		}
	}
	return nil
}
func (s *Accounts) Bootstrap(ctx context.Context, in schemas.RegisterReader, allowWeak bool) (uint64, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Email = core.NormalizeEmail(in.Email)
	in.Nickname = core.NormalizeNickname(in.Nickname)
	if e := validIdentity(in.Name, in.Nickname, in.Email); e != nil {
		return 0, e
	}
	if !allowWeak || s.cfg.Env != "development" {
		if e := core.ValidatePassword(in.Password); e != nil {
			return 0, e
		}
	} else if len(in.Password) == 0 || len(in.Password) > 72 {
		return 0, core.Invalid("INVALID_PASSWORD", "Senha inválida.")
	}
	birth, e := s.birth(in.BirthDate, false)
	if e != nil {
		return 0, e
	}
	hash, e := core.HashPassword(in.Password)
	if e != nil {
		return 0, e
	}
	var id uint64
	e = s.repo.Within(ctx, func(r interfaces.AccountRepository) error {
		var e error
		id, e = r.Bootstrap(ctx, core.User{Name: in.Name, Email: in.Email, Nickname: in.Nickname, PasswordHash: hash, BirthDate: birth, CreatedAt: s.now(), UpdatedAt: s.now()}, allowWeak)
		return e
	})
	return id, e
}
func (s *Accounts) Purge(ctx context.Context) error {
	return s.repo.Within(ctx, func(r interfaces.AccountRepository) error {
		return r.Purge(ctx, s.now().AddDate(0, 0, -s.cfg.RetentionDays))
	})
}
