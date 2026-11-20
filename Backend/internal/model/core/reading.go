package core

import "time"

// ReadingRecord is the current shelf relation; publication snapshots are independent.
type ReadingRecord struct {
	ID, ReaderID, WorkID, EditionID   uint64
	CurrentPage, CreditedPeak         int
	Status                            string
	StartedAt, FinishedAt             *time.Time
	CreatedAt, UpdatedAt              time.Time
	WorkTitle, EditionTitle, Language string
}

type ReadingEdition struct {
	ID, WorkID            uint64
	PageCount, MinimumAge int
	Status, WorkStatus    string
}

type ReadingPost struct {
	ID, ReaderID, WorkID                                 uint64
	Kind, Title, Text, DataURL, Status                   string
	ContainsSpoiler, ContainsSensitiveContent            bool
	MinimumAge, WorkMinimumAge                           int
	WorkStatus, AuthorStatus, AuthorName, AuthorNickname string
	EditionID                                            *uint64
	CurrentPage, Rating                                  *int
	OccurredAt, PublishedAt, EffectsAppliedAt            *time.Time
	CreatedAt, UpdatedAt                                 time.Time
	LikeCount, CommentCount                              int64
	Liked                                                bool
}

type ReadingComment struct {
	ID, PostID, ReaderID                     uint64
	ParentID                                 *uint64
	Depth                                    int
	Text, Status, AuthorName, AuthorNickname string
	AuthorStatus                             string
	CreatedAt, UpdatedAt                     time.Time
	LikeCount                                int64
	Liked                                    bool
}

type ReadingNotification struct {
	ID, RecipientID            uint64
	ActorID, PostID, CommentID *uint64
	Event, EventKey            string
	ReadAt                     *time.Time
	CreatedAt                  time.Time
	Visible                    bool
}

type ProgressEvent struct {
	RecordID, ReaderID                       uint64
	PreviousPage, CurrentPage, CreditedPages int
	ChangeKind                               string
	CreatedAt                                time.Time
}

type ReadingStatistics struct {
	CurrentPages     int64
	AverageRating    *float64
	DiaryCount       int64
	MostReadLanguage *string
	Statuses         []StatusCount
	Months           []ReadingPeriod
	Genres           []GenreCount
}
type StatusCount struct {
	Status string
	Count  int64
}
type GenreCount struct {
	ID    uint64
	Name  string
	Count int64
}
type ReadingPeriod struct {
	Month        int
	Works, Pages int64
}
