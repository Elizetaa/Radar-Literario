package schemas

import (
	"encoding/json"
	"time"
)

// Pointer fields distinguish an omitted PATCH member from zero/false.
type WorkInput struct {
	Title                *string         `json:"title"`
	Synopsis             *string         `json:"synopsis"`
	OriginalLanguage     *string         `json:"original_language"`
	FirstPublicationDate json.RawMessage `json:"first_publication_date"`
	Status               *string         `json:"status"`
	MinimumAge           *int            `json:"minimum_age"`
	AuthorIDs            *[]uint64       `json:"author_ids"`
	GenreIDs             *[]uint64       `json:"genre_ids"`
	InitialEdition       *EditionInput   `json:"initial_edition"`
}

type EditionInput struct {
	WorkID            *uint64         `json:"work_id"`
	PublisherID       *uint64         `json:"publisher_id"`
	Title             *string         `json:"title"`
	ISBN              json.RawMessage `json:"isbn"`
	ISBNAbsenceReason *string         `json:"isbn_absence_reason"`
	EditionNumber     *int            `json:"edition_number"`
	PageCount         *int            `json:"page_count"`
	PublicationDate   json.RawMessage `json:"publication_date"`
	Language          *string         `json:"language"`
	ImageURL          *string         `json:"image_url"`
	Status            *string         `json:"status"`
	MinimumAge        *int            `json:"minimum_age"`
}

type AuthorInput struct {
	Name            *string         `json:"name"`
	Biography       *string         `json:"biography"`
	BirthDate       json.RawMessage `json:"birth_date"`
	DeathDate       json.RawMessage `json:"death_date"`
	CountryOfOrigin *string         `json:"country_of_origin"`
	Sex             json.RawMessage `json:"sex"`
}

type PublisherInput struct {
	Name            *string `json:"name"`
	WebsiteURL      *string `json:"website_url"`
	CountryOfOrigin *string `json:"country_of_origin"`
}

type GenreInput struct {
	Name   *string `json:"name"`
	Active *bool   `json:"active"`
}

type SuggestionInput struct {
	Operation string          `json:"operation"`
	Resource  string          `json:"resource"`
	TargetID  *uint64         `json:"target_id"`
	Proposal  json.RawMessage `json:"proposal"`
}

type CatalogDecision struct {
	Reason string `json:"reason"`
}
type InterestsInput struct {
	GenreIDs []uint64 `json:"genre_ids"`
}

type Author struct {
	ID              uint64  `json:"id"`
	Name            string  `json:"name"`
	Biography       string  `json:"biography"`
	BirthDate       *string `json:"birth_date"`
	DeathDate       *string `json:"death_date"`
	CountryOfOrigin string  `json:"country_of_origin"`
	Sex             *string `json:"sex,omitempty"`
}
type Publisher struct {
	ID              uint64 `json:"id"`
	Name            string `json:"name"`
	WebsiteURL      string `json:"website_url"`
	CountryOfOrigin string `json:"country_of_origin"`
}
type Genre struct {
	ID     uint64 `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type Work struct {
	ID                   uint64    `json:"id"`
	Title                string    `json:"title"`
	Synopsis             string    `json:"synopsis"`
	OriginalLanguage     string    `json:"original_language"`
	FirstPublicationDate *string   `json:"first_publication_date"`
	Status               string    `json:"status"`
	MinimumAge           int       `json:"minimum_age"`
	Version              uint64    `json:"version"`
	Authors              []Author  `json:"authors"`
	Genres               []Genre   `json:"genres"`
	AverageRating        *float64  `json:"average_rating"`
	ReviewCount          int64     `json:"review_count"`
	ReaderCount          int64     `json:"reader_count"`
	Affinity             int64     `json:"affinity"`
	RecommendationReason string    `json:"recommendation_reason,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type Edition struct {
	ID                uint64    `json:"id"`
	WorkID            uint64    `json:"work_id"`
	PublisherID       uint64    `json:"publisher_id"`
	Title             string    `json:"title"`
	ISBN              *string   `json:"isbn"`
	ISBNAbsenceReason string    `json:"isbn_absence_reason,omitempty"`
	EditionNumber     int       `json:"edition_number"`
	PageCount         int       `json:"page_count"`
	PublicationDate   *string   `json:"publication_date"`
	Language          string    `json:"language"`
	ImageURL          string    `json:"image_url"`
	Status            string    `json:"status"`
	MinimumAge        int       `json:"minimum_age"`
	Version           uint64    `json:"version"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type CatalogSuggestion struct {
	ID              uint64          `json:"id"`
	ReaderID        uint64          `json:"reader_id"`
	Operation       string          `json:"operation"`
	Resource        string          `json:"resource"`
	TargetID        *uint64         `json:"target_id"`
	ExpectedVersion *uint64         `json:"expected_version"`
	Proposal        json.RawMessage `json:"proposal"`
	Status          string          `json:"status"`
	DecisionBy      *uint64         `json:"decision_by"`
	DecisionReason  string          `json:"decision_reason"`
	DecidedAt       *time.Time      `json:"decided_at"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}
