package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	// ID est le champ ID de GORM, mais nous le conservons pour la compatibilité JSON
	// et pour stocker l'ID OIDC si nécessaire.
	OIDCID   string `json:"id" gorm:"column:oidc_id;uniqueIndex"` // L'ID de l'utilisateur provenant d'OIDC
	Email    string `json:"email"`
	Name     string `json:"name"`
	Username string `json:"preferred_username"`
}
