package models

import (
	"time"

	"gorm.io/gorm"
)

// MembershipStatus defines the type for a member's status.
type MembershipStatus string

// Constants defining the possible membership statuses.
const (
	StatusActive   MembershipStatus = "Actif"     // Member is currently active.
	StatusInactive MembershipStatus = "Inactif"   // Member is currently inactive.
	StatusPending  MembershipStatus = "En attente" // Member's status is pending (e.g., awaiting approval or first payment).
	StatusExpired  MembershipStatus = "Expiré"    // Member's membership has expired.
)

// Member represents a member of an association or organization.
// It embeds gorm.Model for common fields like ID, CreatedAt, UpdatedAt, and DeletedAt.
type Member struct {
	gorm.Model
	FirstName string `json:"first_name" form:"first_name"` // First name of the member.
	LastName  string `json:"last_name" form:"last_name"`   // Last name of the member.
	Email     string `json:"email" form:"email"`         // Email address of the member.

	// UserID is the ID of the application user who manages this member record.
	// This links the member to a specific association or user account.
	UserID uint `json:"user_id"`

	MembershipStatus MembershipStatus `json:"membership_status" form:"membership_status"` // The current membership status.
	JoinDate         time.Time        `json:"join_date" form:"join_date" time_format:"2006-01-02"`     // The date when the member joined.
	EndDate          *time.Time       `json:"end_date,omitempty" form:"end_date" time_format:"2006-01-02"`                   // Optional end date of the membership (pointer to allow null values).
	LastPaymentDate  *time.Time       `json:"last_payment_date,omitempty" form:"last_payment_date" time_format:"2006-01-02"` // Optional date of the last payment received from the member (pointer to allow null values).
}
