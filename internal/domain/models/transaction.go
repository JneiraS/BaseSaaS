package models

import (
	"time"

	"gorm.io/gorm"
)

// TransactionType définit le type pour le type de transaction.
type TransactionType string

// Définition des types de transaction possibles.
const (
	TypeIncome  TransactionType = "Revenu"
	TypeExpense TransactionType = "Dépense"
)

// Transaction représente une transaction financière (revenu ou dépense).
type Transaction struct {
	gorm.Model
	Amount      float64         `json:"amount" form:"amount"`
	Type        TransactionType `json:"type" form:"type"`
	Description string          `json:"description" form:"description"`
	Date        time.Time       `json:"date" form:"date" time_format:"2006-01-02"`

	// UserID est l'ID de l'utilisateur de l'application qui a enregistré cette transaction.
	UserID uint `json:"user_id"`
}
