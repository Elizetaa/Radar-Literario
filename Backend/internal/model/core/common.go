package core

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/url"
	"strings"
	"time"
)

type Error struct {
	Status  int               `json:"-"`
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func (e *Error) Error() string { return e.Message }
func E(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}
func Invalid(code, message string) *Error {
	field := "_request"
	switch code {
	case "INVALID_NAME":
		field = "name"
	case "INVALID_NICKNAME":
		field = "nickname"
	case "INVALID_EMAIL":
		field = "email"
	case "INVALID_PASSWORD":
		field = "password"
	case "INVALID_BIRTH_DATE":
		field = "birth_date"
	case "INVALID_READING_PROGRESS", "EXPLICIT_CORRECTION_REQUIRED":
		field = "current_page"
	case "INVALID_RATING":
		field = "rating"
	case "INVALID_CHANGE_KIND", "EXPLICIT_EDITION_CHANGE_REQUIRED":
		field = "change_kind"
	case "INVALID_VISIBILITY":
		field = "profile_visibility"
	case "INVALID_BIOGRAPHY":
		field = "biography"
	case "INVALID_SEX":
		field = "sex"
	case "INVALID_COUNTRY":
		field = "country_of_origin"
	case "INVALID_YEAR":
		field = "year"
	case "INVALID_PAGE":
		field = "pagination"
	case "INVALID_ID":
		field = "id"
	case "INVALID_JSON":
		field = "body"
	}
	e := E(400, code, message)
	e.Fields = map[string]string{field: message}
	return e
}
func Forbidden() *Error {
	return E(403, "FORBIDDEN", "Você não tem permissão para esta operação.")
}
func NotFound() *Error               { return E(404, "NOT_FOUND", "Recurso não encontrado.") }
func Conflict(message string) *Error { return E(409, "CONFLICT", message) }

var ErrUnauthenticated = E(401, "UNAUTHENTICATED", "Entre na sua conta para continuar.")

type Actor struct {
	ID         uint64
	Role       string
	Reader     bool
	BirthDate  *time.Time
	Restricted bool
}

func (a Actor) Age(now time.Time) int {
	if a.BirthDate == nil {
		return 0
	}
	b := *a.BirthDate
	years := now.UTC().Year() - b.Year()
	if now.UTC().Month() < b.Month() || (now.UTC().Month() == b.Month() && now.UTC().Day() < b.Day()) {
		years--
	}
	if years < 0 {
		return 0
	}
	return years
}
func (a Actor) RequireReader() error {
	if a.ID == 0 {
		return ErrUnauthenticated
	}
	if !a.Reader || a.Restricted {
		return Forbidden()
	}
	return nil
}
func (a Actor) Require(permission string) error {
	if a.ID == 0 {
		return ErrUnauthenticated
	}
	if a.Role == "system_administrator" {
		return nil
	}
	switch permission {
	case "catalog":
		if a.Role == "catalog_manager" || a.Role == "administrator" {
			return nil
		}
	case "moderation", "support":
		if a.Role == "user_manager" || a.Role == "administrator" {
			return nil
		}
	case "audit":
		if a.Role != "" {
			return nil
		}
	}
	return Forbidden()
}
func HashPassword(password string) (string, error) {
	b, e := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(b), e
}
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
func ValidatePassword(password string) error {
	if len(password) < 8 || len(password) > 72 {
		return Invalid("INVALID_PASSWORD", "A senha deve ter entre 8 e 72 bytes.")
	}
	return nil
}
func ValidateURL(raw string) error {
	if raw == "" {
		return nil
	}
	u, e := url.Parse(raw)
	if e != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.User != nil || len(raw) > 2048 {
		return Invalid("INVALID_URL", "Informe uma URL HTTP ou HTTPS válida, sem credenciais.")
	}
	return nil
}
func ValidAge(age int) bool {
	return age == 0 || age == 10 || age == 12 || age == 14 || age == 16 || age == 18
}
func NormalizeEmail(s string) string    { return strings.ToLower(strings.TrimSpace(s)) }
func NormalizeNickname(s string) string { return strings.ToLower(strings.TrimSpace(s)) }
func Date(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, e := time.Parse("2006-01-02", s)
	if e != nil {
		return nil, Invalid("INVALID_DATE", fmt.Sprintf("Data inválida: use YYYY-MM-DD."))
	}
	return &t, nil
}
