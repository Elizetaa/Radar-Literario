package interfaces

import (
	"context"
	"radar-literario/backend/internal/model/core"
	"time"
)

type AccountRepository interface {
	Within(context.Context, func(AccountRepository) error) error
	Create(context.Context, core.User) (uint64, error)
	Account(context.Context, uint64, bool) (core.Account, error)
	ByEmail(context.Context, string) (core.Account, error)
	UpdateAccount(context.Context, uint64, map[string]any, map[string]any) error
	WorkEligible(context.Context, uint64, int) error
	Session(context.Context, string) (core.Session, error)
	AddSession(context.Context, core.Session) error
	TouchSession(context.Context, uint64, time.Time, time.Time) (bool, error)
	Revoke(context.Context, uint64, string, time.Time) error
	RevokeToken(context.Context, string, time.Time) error
	Penalties(context.Context, uint64, time.Time) (bool, bool, error)
	SaveReset(context.Context, core.Reset, string, string) error
	ConsumeReset(context.Context, string, time.Time) (uint64, error)
	ChangeHash(context.Context, uint64, string) error
	Anonymize(context.Context, uint64, time.Time) error
	Bootstrap(context.Context, core.User, bool) (uint64, error)
	PendingMail(context.Context, time.Time) ([]core.Outbox, error)
	ClaimMail(context.Context, uint64, time.Time) (bool, error)
	FinishMail(context.Context, uint64, bool, time.Time) error
	Purge(context.Context, time.Time) error
}
type Mailer interface {
	Send(context.Context, string, string, string) error
}
