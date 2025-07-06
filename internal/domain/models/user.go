package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents an application user.
// This model stores user-related information, including details from the OIDC provider.
type User struct {
	ID             uint      `json:"id"`             // Unique identifier for the user.
	OIDCID         string    `json:"oidc_id"`        // The unique identifier for the user from the OIDC provider (e.g., "sub" claim).
	Email          string    `json:"email" form:"email"` // User's email address.
	Name           string    `json:"name" form:"name"`   // User's full name or display name.
	Username       string    `json:"username" form:"username"` // User's chosen username.
	LastConnection time.Time `json:"last_connection" form:"last_connection"` // Timestamp of the user's last successful login.
	CreatedAt      time.Time                               // Timestamp when the user record was created.
	UpdatedAt      time.Time                               // Timestamp when the user record was last updated.
	DeletedAt      gorm.DeletedAt `gorm:"index"`          // GORM field for soft deletion, indexed for efficient querying.
}
