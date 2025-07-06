package models

import (
	"gorm.io/gorm"
)

// Poll represents a poll created by a user.
// It embeds gorm.Model for common fields like ID, CreatedAt, UpdatedAt, and DeletedAt.
type Poll struct {
	gorm.Model
	Question string   `json:"question" form:"question"` // The question posed in the poll.
	UserID   uint     `json:"user_id"`                   // The ID of the user who created the poll.
	Options  []Option `gorm:"foreignKey:PollID"`         // A slice of Option models associated with this poll (one-to-many relationship).
}

// Option represents a voting option for a poll.
// It embeds gorm.Model for common fields.
type Option struct {
	gorm.Model
	Text   string `json:"text" form:"text"`         // The text of the voting option.
	PollID uint   `json:"poll_id"`                 // The ID of the poll to which this option belongs (foreign key).
	Votes  []Vote `gorm:"foreignKey:OptionID"`     // A slice of Vote models associated with this option (one-to-many relationship).
}

// Vote represents a user's vote for a specific poll option.
// It embeds gorm.Model for common fields.
type Vote struct {
	gorm.Model
	OptionID uint `json:"option_id"` // The ID of the option for which the user voted (foreign key).
	UserID   uint `json:"user_id"`   // The ID of the user who cast the vote (foreign key).
}
