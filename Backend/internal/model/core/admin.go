package core

import "time"

// Administrative persistence records are never serialized as public responses.
type AdminUser struct {
	ID                                                           uint64
	Name, Nickname, Email, PasswordHash, Status, ProfileImageURL string
	BirthDate                                                    *time.Time
	CreatedAt, UpdatedAt                                         time.Time
}

type AdministratorRecord struct {
	UserID                       uint64
	Name, Nickname, Role, Status string
	CreatedAt, UpdatedAt         time.Time
}

type Report struct {
	ID, ReporterID                                                uint64
	ReportedUserID, ReportedPostID, ReportedCommentID, AssignedTo *uint64
	Reason, Description, ReportDataURL, Status                    string
	ResolutionNote                                                *string
	ResolvedAt                                                    *time.Time
	CreatedAt, UpdatedAt                                          time.Time
}

type ReportMessage struct {
	ID, ReportID, AuthorID uint64
	Text, Visibility       string
	CreatedAt              time.Time
}

type ReportTarget struct {
	ID, OwnerID                                  uint64
	Status, UserStatus, WorkStatus, ParentStatus string
	MinimumAge                                   int
}

type ContentEvidence struct {
	ID, ReaderID, PostID, WorkID              uint64
	Kind, Title, Text, Status                 string
	ContainsSpoiler, ContainsSensitiveContent bool
	MinimumAge                                int
	CreatedAt, UpdatedAt                      time.Time
}

type Penalty struct {
	ID, UserID, AppliedBy uint64
	RevokedBy             *uint64
	Type, Reason          string
	StartsAt              time.Time
	EndsAt, RevokedAt     *time.Time
	RevocationReason      *string
	CreatedAt             time.Time
}

type FAQEntry struct {
	ID                       uint64
	Question, Answer, Status string
	SortOrder                int
	CreatedAt, UpdatedAt     time.Time
}

type SupportTicket struct {
	ID, RequesterID      uint64
	AssignedTo           *uint64
	Subject, Status      string
	CreatedAt, UpdatedAt time.Time
}

type SupportMessage struct {
	ID, TicketID, AuthorID uint64
	Text                   string
	CreatedAt              time.Time
}

type AuditLog struct {
	ID                         uint64
	ActorID, TargetID          *uint64
	Action, TargetType, Reason string
	CreatedAt                  time.Time
}
