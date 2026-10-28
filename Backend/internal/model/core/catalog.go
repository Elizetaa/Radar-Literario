package core

import "time"

// Catalog entities are persistence models; HTTP representations live in schemas.
type Author struct {
	ID              uint64
	Name            string
	Biography       string
	BirthDate       *time.Time
	DeathDate       *time.Time
	CountryOfOrigin string
	Sex             *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Publisher struct {
	ID              uint64
	Name            string
	WebsiteURL      string
	CountryOfOrigin string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Genre struct {
	ID     uint64
	Name   string
	Active bool
}

type Work struct {
	ID                   uint64
	Title                string
	Synopsis             string
	OriginalLanguage     string
	FirstPublicationDate *time.Time
	Status               string
	MinimumAge           int
	Version              uint64
	CreatedAt            time.Time
	UpdatedAt            time.Time
	Authors              []Author `gorm:"-"`
	Genres               []Genre  `gorm:"-"`
	AverageRating        *float64
	ReviewCount          int64
	ReaderCount          int64
	Affinity             int64
}

type Edition struct {
	ID                uint64
	WorkID            uint64
	PublisherID       uint64
	Title             string
	ISBN              *string
	ISBNAbsenceReason string
	EditionNumber     int
	PageCount         int
	PublicationDate   *time.Time
	Language          string
	ImageURL          string
	Status            string
	MinimumAge        int
	Version           uint64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type CatalogSuggestion struct {
	ID              uint64
	ReaderID        uint64
	Operation       string
	Resource        string
	TargetID        *uint64
	ExpectedVersion *uint64
	Proposal        string
	Status          string
	DecisionBy      *uint64
	DecisionReason  string
	DecidedAt       *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

const CatalogPublished = "published"

func ValidMinimumAge(age int) bool {
	return age == 0 || age == 10 || age == 12 || age == 14 || age == 16 || age == 18
}

func ValidCatalogStatus(status string) bool {
	switch status {
	case "draft", "pending_review", "published", "rejected", "hidden", "archived":
		return true
	}
	return false
}
