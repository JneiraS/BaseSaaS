package models

import (
	"time"

	"gorm.io/gorm"
)

// MembershipStatus définit le type pour le statut d'un membre.
type MembershipStatus string

// Définition des statuts possibles pour un membre.
const (
	StatusActive   MembershipStatus = "Actif"
	StatusInactive MembershipStatus = "Inactif"
	StatusPending  MembershipStatus = "En attente"
	StatusExpired  MembershipStatus = "Expiré"
)

// Member représente un membre d'une association.
type Member struct {
	gorm.Model
	FirstName string `json:"first_name" form:"first_name"`
	LastName  string `json:"last_name" form:"last_name"`
	Email     string `json:"email" form:"email"`

	// UserID est l'ID de l'utilisateur de l'application qui gère ce membre.
	// Cela lie le membre à une association/un compte spécifique.
	UserID uint `json:"user_id"`

	MembershipStatus MembershipStatus `json:"membership_status" form:"membership_status"`
	JoinDate         time.Time        `json:"join_date" form:"join_date" time_format:"2006-01-02"`
	EndDate          *time.Time       `json:"end_date,omitempty" form:"end_date" time_format:"2006-01-02"` // Pointeur pour autoriser les valeurs nulles
}
