package schemas

import (
	"encoding/json"
	"time"
)

type AdminOptionalID struct {
	Set   bool
	Value *uint64
}

func (v *AdminOptionalID) UnmarshalJSON(data []byte) error {
	v.Set = true
	if string(data) == "null" {
		v.Value = nil
		return nil
	}
	var id uint64
	if err := json.Unmarshal(data, &id); err != nil {
		return err
	}
	v.Value = &id
	return nil
}

type CreateAdministrator struct {
	UserID    *uint64 `json:"user_id"`
	Name      string  `json:"name"`
	Nickname  string  `json:"nickname"`
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	BirthDate string  `json:"birth_date"`
	Role      string  `json:"role"`
	Reason    string  `json:"reason"`
}
type PatchAdministrator struct {
	Role   *string `json:"role"`
	Status *string `json:"status"`
	Reason string  `json:"reason"`
}
type Administrator struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Nickname  string    `json:"nickname"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type AdministrativeUser struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Nickname  string    `json:"nickname"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
type CreateReport struct {
	ReportedUserID    *uint64 `json:"reported_user_id"`
	ReportedPostID    *uint64 `json:"reported_post_id"`
	ReportedCommentID *uint64 `json:"reported_comment_id"`
	Reason            string  `json:"reason"`
	Description       string  `json:"description"`
	ReportDataURL     string  `json:"report_data_url"`
}
type PatchReport struct {
	Status         *string         `json:"status"`
	AssignedTo     AdminOptionalID `json:"assigned_to"`
	ResolutionNote *string         `json:"resolution_note"`
}
type CreateReportMessage struct {
	Text               string `json:"text"`
	Visibility         string `json:"visibility"`
	RequestInformation bool   `json:"request_information"`
}
type Report struct {
	ID                uint64     `json:"id"`
	ReporterID        *uint64    `json:"reporter_id,omitempty"`
	ReportedUserID    *uint64    `json:"reported_user_id"`
	ReportedPostID    *uint64    `json:"reported_post_id"`
	ReportedCommentID *uint64    `json:"reported_comment_id"`
	AssignedTo        *uint64    `json:"assigned_to,omitempty"`
	Reason            string     `json:"reason"`
	Description       string     `json:"description"`
	ReportDataURL     string     `json:"report_data_url,omitempty"`
	Status            string     `json:"status"`
	ResolutionNote    *string    `json:"resolution_note"`
	ResolvedAt        *time.Time `json:"resolved_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}
type ReportMessage struct {
	ID         uint64    `json:"id"`
	AuthorID   uint64    `json:"author_id"`
	Text       string    `json:"text"`
	Visibility string    `json:"visibility"`
	CreatedAt  time.Time `json:"created_at"`
}
type Moderation struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type ContentEvidence struct {
	ID                       uint64    `json:"id"`
	ReaderID                 uint64    `json:"reader_id"`
	PostID                   uint64    `json:"post_id"`
	WorkID                   uint64    `json:"work_id"`
	Kind                     string    `json:"kind"`
	Title                    string    `json:"title"`
	Text                     string    `json:"text"`
	Status                   string    `json:"status"`
	ContainsSpoiler          bool      `json:"contains_spoiler"`
	ContainsSensitiveContent bool      `json:"contains_sensitive_content"`
	MinimumAge               int       `json:"minimum_age"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}
type CreatePenalty struct {
	UserID    uint64     `json:"user_id"`
	Type      string     `json:"type"`
	Reason    string     `json:"reason"`
	StartsAt  *time.Time `json:"starts_at"`
	EndsAt    *time.Time `json:"ends_at"`
	ReportIDs []uint64   `json:"report_ids"`
}
type Reason struct {
	Reason string `json:"reason"`
}
type Penalty struct {
	ID               uint64     `json:"id"`
	UserID           uint64     `json:"user_id"`
	AppliedBy        uint64     `json:"applied_by"`
	RevokedBy        *uint64    `json:"revoked_by"`
	Type             string     `json:"type"`
	Reason           string     `json:"reason"`
	StartsAt         time.Time  `json:"starts_at"`
	EndsAt           *time.Time `json:"ends_at"`
	RevokedAt        *time.Time `json:"revoked_at"`
	RevocationReason *string    `json:"revocation_reason"`
	CreatedAt        time.Time  `json:"created_at"`
}
type SaveFAQ struct {
	Question  *string `json:"question"`
	Answer    *string `json:"answer"`
	SortOrder *int    `json:"sort_order"`
	Status    *string `json:"status"`
}
type FAQ struct {
	ID        uint64    `json:"id"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	SortOrder int       `json:"sort_order"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}
type CreateSupportTicket struct {
	Subject string `json:"subject"`
	Text    string `json:"text"`
}
type PatchSupportTicket struct {
	Status     *string         `json:"status"`
	AssignedTo AdminOptionalID `json:"assigned_to"`
}
type CreateSupportMessage struct {
	Text string `json:"text"`
}
type SupportTicket struct {
	ID          uint64    `json:"id"`
	RequesterID uint64    `json:"requester_id"`
	AssignedTo  *uint64   `json:"assigned_to"`
	Subject     string    `json:"subject"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type SupportMessage struct {
	ID        uint64    `json:"id"`
	AuthorID  uint64    `json:"author_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
type AuditLog struct {
	ID         uint64    `json:"id"`
	ActorID    *uint64   `json:"actor_id"`
	TargetID   *uint64   `json:"target_id"`
	TargetType string    `json:"target_type"`
	Action     string    `json:"action"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
}
