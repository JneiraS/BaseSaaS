package models

import (
	"time"

	"gorm.io/gorm"
)

// Event represents an event organized by an association or user.
// It embeds gorm.Model for common fields like ID, CreatedAt, UpdatedAt, and DeletedAt.
type Event struct {
	gorm.Model
	Title       string    `json:"title" form:"title"`             // The title or name of the event.
	Description string    `json:"description" form:"description"` // A detailed description of the event.
	StartDate   time.Time `json:"start_date" form:"start_date" time_format:"2006-01-02T15:04"` // The start date and time of the event.
	EndDate     time.Time `json:"end_date" form:"end_date" time_format:"2006-01-02T15:04"`     // The end date and time of the event.

	// UserID is the ID of the application user who created or manages this event.
	// This establishes a relationship between the event and its owner.
	UserID uint `json:"user_id"`
}
