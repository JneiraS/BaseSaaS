package models

import (
	"time"

	"gorm.io/gorm"
)

// TransactionType defines the type for a financial transaction (income or expense).
type TransactionType string

// Constants defining the possible types of transactions.
const (
	TypeIncome  TransactionType = "Revenu"  // Represents an income transaction.
	TypeExpense TransactionType = "Dépense" // Represents an expense transaction.
)

// Transaction represents a financial transaction (either an income or an expense).
// It embeds gorm.Model for common fields like ID, CreatedAt, UpdatedAt, and DeletedAt.
type Transaction struct {
	gorm.Model
	Amount      float64         `json:"amount" form:"amount"`         // The monetary amount of the transaction.
	Type        TransactionType `json:"type" form:"type"`             // The type of transaction (Income or Expense).
	Description string          `json:"description" form:"description"` // A brief description of the transaction.
	Date        time.Time       `json:"date" form:"date" time_format:"2006-01-02"` // The date when the transaction occurred.

	// UserID is the ID of the application user who recorded this transaction.
	// This establishes a relationship between the transaction and its owner.
	UserID uint `json:"user_id"`
}
