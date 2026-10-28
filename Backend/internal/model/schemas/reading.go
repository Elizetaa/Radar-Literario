package schemas

import (
	"encoding/json"
	"time"
)

type NullableText struct {
	Set   bool
	Value *string
}

func (v *NullableText) UnmarshalJSON(b []byte) error {
	v.Set = true
	return json.Unmarshal(b, &v.Value)
}

type ShelfInput struct {
	EditionID   uint64     `json:"edition_id"`
	CurrentPage *int       `json:"current_page"`
	Status      string     `json:"status"`
	ChangeKind  string     `json:"change_kind"`
	StartedAt   *time.Time `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at"`
}
type PostInput struct {
	WorkID                   uint64     `json:"work_id"`
	EditionID                uint64     `json:"edition_id"`
	Title                    string     `json:"title"`
	Text                     string     `json:"text"`
	DataURL                  *string    `json:"data_url"`
	Status                   string     `json:"status"`
	ContainsSpoiler          *bool      `json:"contains_spoiler"`
	ContainsSensitiveContent *bool      `json:"contains_sensitive_content"`
	MinimumAge               int        `json:"minimum_age"`
	CurrentPage              *int       `json:"current_page"`
	Rating                   *int       `json:"rating"`
	OccurredAt               *time.Time `json:"occurred_at"`
}
type PostPatch struct {
	Title                    *string      `json:"title"`
	Text                     *string      `json:"text"`
	DataURL                  NullableText `json:"data_url"`
	Status                   *string      `json:"status"`
	ContainsSpoiler          *bool        `json:"contains_spoiler"`
	ContainsSensitiveContent *bool        `json:"contains_sensitive_content"`
	MinimumAge               *int         `json:"minimum_age"`
	CurrentPage              *int         `json:"current_page"`
	Rating                   *int         `json:"rating"`
}
type CommentInput struct {
	Text     string  `json:"text"`
	ParentID *uint64 `json:"parent_id"`
}
type CommentPatch struct {
	Text string `json:"text"`
}
type Reveal struct{ Spoilers, Sensitive bool }
type PostFilter struct {
	Kind             string
	ReaderID, WorkID uint64
	Own              bool
	Page             Page
}

type ShelfItem struct {
	ID           uint64     `json:"id"`
	WorkID       uint64     `json:"work_id"`
	EditionID    uint64     `json:"edition_id"`
	Title        string     `json:"title"`
	EditionTitle string     `json:"edition_title"`
	Language     string     `json:"language"`
	CurrentPage  int        `json:"current_page"`
	Status       string     `json:"status"`
	StartedAt    *time.Time `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
type Post struct {
	ID                       uint64     `json:"id"`
	PostID                   uint64     `json:"post_id"`
	ReaderID                 uint64     `json:"reader_id"`
	WorkID                   uint64     `json:"work_id"`
	Kind                     string     `json:"kind"`
	Title                    *string    `json:"title"`
	Text                     *string    `json:"text"`
	DataURL                  *string    `json:"data_url"`
	Status                   string     `json:"status"`
	AuthorName               string     `json:"author_name"`
	AuthorNickname           string     `json:"author_nickname"`
	ContainsSpoiler          bool       `json:"contains_spoiler"`
	ContainsSensitiveContent bool       `json:"contains_sensitive_content"`
	ContentHidden            bool       `json:"content_hidden"`
	MinimumAge               int        `json:"minimum_age"`
	EditionID                *uint64    `json:"edition_id"`
	CurrentPage              *int       `json:"current_page"`
	Rating                   *int       `json:"rating"`
	OccurredAt               *time.Time `json:"occurred_at"`
	PublishedAt              *time.Time `json:"published_at"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
	LikeCount                int64      `json:"like_count"`
	CommentCount             int64      `json:"comment_count"`
	Liked                    bool       `json:"liked"`
}
type Comment struct {
	ID             uint64    `json:"id"`
	PostID         uint64    `json:"post_id"`
	ReaderID       uint64    `json:"reader_id"`
	ParentID       *uint64   `json:"parent_id"`
	Depth          int       `json:"depth"`
	Text           *string   `json:"text"`
	Status         string    `json:"status"`
	AuthorName     string    `json:"author_name"`
	AuthorNickname string    `json:"author_nickname"`
	ContentHidden  bool      `json:"content_hidden"`
	LikeCount      int64     `json:"like_count"`
	Liked          bool      `json:"liked"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
type Notification struct {
	ID              uint64     `json:"id"`
	Event           string     `json:"event"`
	ActorID         *uint64    `json:"actor_id"`
	PostID          *uint64    `json:"post_id"`
	CommentID       *uint64    `json:"comment_id"`
	TargetAvailable bool       `json:"target_available"`
	ReadAt          *time.Time `json:"read_at"`
	CreatedAt       time.Time  `json:"created_at"`
}
type Statistics struct {
	Year             int           `json:"year"`
	Timezone         string        `json:"timezone"`
	CurrentPages     int64         `json:"current_pages"`
	AverageRating    *float64      `json:"average_rating"`
	DiaryCount       int64         `json:"diary_count"`
	MostReadLanguage *string       `json:"most_read_language"`
	Statuses         []StatusCount `json:"statuses"`
	Months           []Period      `json:"months"`
	YearTotal        Period        `json:"year_total"`
	TopGenres        []GenreCount  `json:"top_genres"`
}
type StatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}
type Period struct {
	Month int   `json:"month,omitempty"`
	Works int64 `json:"works"`
	Pages int64 `json:"pages"`
}
type GenreCount struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}
