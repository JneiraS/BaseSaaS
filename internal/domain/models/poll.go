package models

import (
	"gorm.io/gorm"
)

// Poll représente un sondage créé par un utilisateur.
type Poll struct {
	gorm.Model
	Question string `json:"question" form:"question"`
	UserID   uint   `json:"user_id"` // L'utilisateur qui a créé le sondage
	Options  []Option `gorm:"foreignKey:PollID"` // Les options de vote pour ce sondage
}

// Option représente une option de vote pour un sondage.
type Option struct {
	gorm.Model
	Text   string `json:"text" form:"text"`
	PollID uint   `json:"poll_id"` // L'ID du sondage auquel cette option appartient
	Votes  []Vote `gorm:"foreignKey:OptionID"` // Les votes pour cette option
}

// Vote représente un vote d'un utilisateur pour une option de sondage.
type Vote struct {
	gorm.Model
	OptionID uint `json:"option_id"` // L'ID de l'option pour laquelle l'utilisateur a voté
	UserID   uint `json:"user_id"`   // L'ID de l'utilisateur qui a voté
}
