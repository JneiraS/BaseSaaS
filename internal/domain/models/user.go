package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID             uint      `json:"id"`
	OIDCID         string    `json:"oidc_id"` // L'ID de l'utilisateur provenant d'OIDC
	Email          string    `json:"email" form:"email"`
	Name           string    `json:"name" form:"name"`
	Username       string    `json:"username" form:"username"`
	LastConnection time.Time `json:"last_connection" form:"last_connection"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}
