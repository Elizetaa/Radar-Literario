package interfaces

import (
	"context"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/schemas"
	"time"
)

type AdminRepository interface {
	Within(context.Context, func(AdminRepository) error) error
	User(context.Context, uint64, bool) (core.AdminUser, error)
	Users(context.Context, string, schemas.Page) ([]core.AdminUser, int64, error)
	CreateAdministrativeUser(context.Context, core.AdminUser) (uint64, error)
	Administrator(context.Context, uint64, bool) (core.AdministratorRecord, error)
	Administrators(context.Context, schemas.Page) ([]core.AdministratorRecord, int64, error)
	LockSystemAdministrators(context.Context) (int64, error)
	SaveAdministrator(context.Context, uint64, string, string, bool) error
	HasBlockingPenalty(context.Context, uint64, time.Time) (bool, error)
	RevokeSessions(context.Context, uint64, time.Time) error
	Target(context.Context, string, uint64) (core.ReportTarget, error)
	ContentEvidence(context.Context, string, uint64) (core.ContentEvidence, error)
	CreateReport(context.Context, core.Report) (uint64, error)
	Report(context.Context, uint64, bool) (core.Report, error)
	Reports(context.Context, *uint64, string, schemas.Page) ([]core.Report, int64, error)
	SaveReport(context.Context, core.Report) error
	ReportMessages(context.Context, uint64, bool, schemas.Page) ([]core.ReportMessage, int64, error)
	CreateReportMessage(context.Context, core.ReportMessage) (uint64, error)
	Moderate(context.Context, string, uint64, string, time.Time) error
	NotifyModeration(context.Context, uint64, uint64, string, uint64, string, time.Time) error
	CreatePenalty(context.Context, core.Penalty) (uint64, error)
	Penalty(context.Context, uint64, bool) (core.Penalty, error)
	Penalties(context.Context, *uint64, schemas.Page) ([]core.Penalty, int64, error)
	LinkPenaltyReports(context.Context, uint64, []uint64) error
	RevokePenalty(context.Context, uint64, uint64, string, time.Time) error
	FAQ(context.Context, uint64, bool) (core.FAQEntry, error)
	FAQs(context.Context, bool, schemas.Page) ([]core.FAQEntry, int64, error)
	SaveFAQ(context.Context, core.FAQEntry, bool) (uint64, error)
	DeleteFAQ(context.Context, uint64) error
	CreateTicket(context.Context, core.SupportTicket) (uint64, error)
	Ticket(context.Context, uint64, bool) (core.SupportTicket, error)
	Tickets(context.Context, *uint64, string, schemas.Page) ([]core.SupportTicket, int64, error)
	SaveTicket(context.Context, core.SupportTicket) error
	TicketMessages(context.Context, uint64, schemas.Page) ([]core.SupportMessage, int64, error)
	CreateTicketMessage(context.Context, core.SupportMessage) (uint64, error)
	Audit(context.Context, uint64, string, string, uint64, string, time.Time) error
	AuditLogs(context.Context, string, schemas.Page) ([]core.AuditLog, int64, error)
}
