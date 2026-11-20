package core

import "time"

type User struct {
	ID                                                           uint64
	Name, Nickname, Email, PasswordHash, Status, ProfileImageURL string
	BirthDate                                                    *time.Time
	EmailVerifiedAt, DeletedAt                                   *time.Time
	CreatedAt, UpdatedAt                                         time.Time
}
type Account struct {
	User
	Biography, ProfileVisibility string
	Sex, CountryOfOrigin         *string
	FavoriteWorkID               *uint64
	Role                         string
	IsReader                     bool
}
type Session struct {
	ID, UserID                       uint64
	TokenHash, CSRFHash              string
	CreatedAt, LastSeenAt, ExpiresAt time.Time
	RevokedAt                        *time.Time
}
type Reset struct {
	ID, UserID           uint64
	TokenHash            string
	ExpiresAt, CreatedAt time.Time
	ConsumedAt           *time.Time
}
type Outbox struct {
	ID                 uint64
	Recipient, Payload string
	Attempts           int
	AvailableAt        time.Time
}
