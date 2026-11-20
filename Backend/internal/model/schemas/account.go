package schemas

import (
	"encoding/json"
	"time"
)

type Optional[T any] struct {
	Set   bool
	Value *T
}

func (o *Optional[T]) UnmarshalJSON(b []byte) error {
	o.Set = true
	if string(b) == "null" {
		o.Value = nil
		return nil
	}
	var v T
	if e := json.Unmarshal(b, &v); e != nil {
		return e
	}
	o.Value = &v
	return nil
}

type RegisterReader struct {
	Name      string `json:"name"`
	Nickname  string `json:"nickname"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	BirthDate string `json:"birth_date"`
}
type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type ForgotPassword struct {
	Email string `json:"email"`
}
type ResetPassword struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}
type ChangePassword struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
type DeleteAccount struct {
	Password string `json:"password"`
}
type PatchAccount struct {
	Name              *string          `json:"name"`
	Nickname          *string          `json:"nickname"`
	Email             *string          `json:"email"`
	BirthDate         *string          `json:"birth_date"`
	ProfileImageURL   *string          `json:"profile_image_url"`
	Biography         *string          `json:"biography"`
	ProfileVisibility *string          `json:"profile_visibility"`
	Sex               Optional[string] `json:"sex"`
	CountryOfOrigin   Optional[string] `json:"country_of_origin"`
	FavoriteWorkID    Optional[uint64] `json:"favorite_work_id"`
}
type AccountView struct {
	ID                uint64    `json:"id"`
	Name              string    `json:"name"`
	Nickname          string    `json:"nickname"`
	Email             string    `json:"email"`
	BirthDate         *string   `json:"birth_date"`
	ProfileImageURL   string    `json:"profile_image_url"`
	Biography         string    `json:"biography"`
	ProfileVisibility string    `json:"profile_visibility"`
	Sex               *string   `json:"sex"`
	CountryOfOrigin   *string   `json:"country_of_origin"`
	FavoriteWorkID    *uint64   `json:"favorite_work_id"`
	Role              string    `json:"role,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}
type ReaderView struct {
	ID              uint64  `json:"id"`
	Name            string  `json:"name"`
	Nickname        string  `json:"nickname"`
	Biography       string  `json:"biography"`
	ProfileImageURL string  `json:"profile_image_url"`
	FavoriteWorkID  *uint64 `json:"favorite_work_id"`
}
