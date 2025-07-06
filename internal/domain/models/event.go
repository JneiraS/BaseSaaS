package models

import (
	"time"

	"gorm.io/gorm"
)

// Event représente un événement organisé par une association.
type Event struct {
	gorm.Model
	Title       string    `json:"title" form:"title"`
	Description string    `json:"description" form:"description"`
	StartDate   time.Time `json:"start_date" form:"start_date" time_format:"2006-01-02T15:04"`
	EndDate     time.Time `json:"end_date" form:"end_date" time_format:"2006-01-02T15:04"`

	// UserID est l'ID de l'utilisateur de l'application qui gère cet événement.
	UserID uint `json:"user_id"`
}
