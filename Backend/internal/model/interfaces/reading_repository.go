package interfaces

import (
	"context"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/schemas"
	"time"
)

type ReadingRepository interface {
	Within(context.Context, func(ReadingRepository) error) error
	LockReader(context.Context, uint64) error
	Edition(context.Context, uint64) (core.ReadingEdition, error)
	Record(context.Context, uint64, uint64) (*core.ReadingRecord, error)
	SaveRecord(context.Context, *core.ReadingRecord) error
	RemoveRecord(context.Context, uint64, uint64) error
	AddProgress(context.Context, core.ProgressEvent) error
	Shelf(context.Context, uint64, int, string, schemas.Page) ([]core.ReadingRecord, int64, error)
	ProfileVisible(context.Context, uint64, core.Actor) error
	Post(context.Context, uint64, uint64) (core.ReadingPost, error)
	Posts(context.Context, core.Actor, schemas.PostFilter) ([]core.ReadingPost, int64, error)
	CreatePost(context.Context, *core.ReadingPost) error
	SavePost(context.Context, core.ReadingPost) error
	CreateDiary(context.Context, core.ReadingPost) error
	SaveDiary(context.Context, core.ReadingPost) error
	CreateReview(context.Context, core.ReadingPost) error
	SaveReview(context.Context, core.ReadingPost) error
	ReleaseReview(context.Context, uint64) error
	Comment(context.Context, uint64, uint64) (core.ReadingComment, error)
	Comments(context.Context, uint64, uint64, *uint64, schemas.Page) ([]core.ReadingComment, int64, error)
	CreateComment(context.Context, *core.ReadingComment) error
	SaveComment(context.Context, core.ReadingComment) error
	Like(context.Context, string, uint64, uint64, bool) (bool, error)
	Notify(context.Context, core.ReadingNotification) error
	Notifications(context.Context, core.Actor, *bool, schemas.Page) ([]core.ReadingNotification, int64, error)
	ReadNotification(context.Context, uint64, uint64, time.Time) error
	Statistics(context.Context, uint64, int) (core.ReadingStatistics, error)
}
