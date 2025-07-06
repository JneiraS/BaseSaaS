package services

import (
	"fmt"
	"strings"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/domain/repositories"
)

// FinanceService encapsulates the business logic for financial management.
// It interacts with the TransactionRepository to perform CRUD operations and financial calculations.
type FinanceService struct {
	transactionRepo repositories.TransactionRepository
}

// NewFinanceService creates a new instance of FinanceService.
// It takes a TransactionRepository as a dependency, adhering to the dependency inversion principle.
func NewFinanceService(transactionRepo repositories.TransactionRepository) *FinanceService {
	return &FinanceService{transactionRepo: transactionRepo}
}

// CreateTransaction handles the creation of a new financial transaction.
// It performs validation on the transaction data before persisting it via the repository.
func (s *FinanceService) CreateTransaction(transaction *models.Transaction) error {
	if err := s.validateTransaction(transaction); err != nil {
		return err
	}
	return s.transactionRepo.CreateTransaction(transaction)
}

// GetTransactionByID retrieves a financial transaction by its unique identifier.
func (s *FinanceService) GetTransactionByID(id uint) (*models.Transaction, error) {
	return s.transactionRepo.FindTransactionByID(id)
}

// GetTransactionsByUserID retrieves all financial transactions associated with a specific user ID.
func (s *FinanceService) GetTransactionsByUserID(userID uint) ([]models.Transaction, error) {
	return s.transactionRepo.FindTransactionsByUserID(userID)
}

// UpdateTransaction handles the update of an existing financial transaction.
// It performs validation on the updated transaction data before persisting the changes.
func (s *FinanceService) UpdateTransaction(transaction *models.Transaction) error {
	if err := s.validateTransaction(transaction); err != nil {
		return err
	}
	return s.transactionRepo.UpdateTransaction(transaction)
}

// DeleteTransaction handles the deletion of a financial transaction by its unique identifier.
func (s *FinanceService) DeleteTransaction(id uint) error {
	return s.transactionRepo.DeleteTransaction(id)
}

// GetTotalIncome returns the total sum of all income transactions for a given user ID.
// It delegates the call to the underlying repository.
func (s *FinanceService) GetTotalIncome(userID uint) (float64, error) {
	return s.transactionRepo.GetTotalIncome(userID)
}

// GetTotalExpenses returns the total sum of all expense transactions for a given user ID.
// It delegates the call to the underlying repository.
func (s *FinanceService) GetTotalExpenses(userID uint) (float64, error) {
	return s.transactionRepo.GetTotalExpenses(userID)
}

// validateTransaction performs business logic validation on a Transaction model.
// It checks for valid amount, non-empty description, and a valid date.
func (s *FinanceService) validateTransaction(transaction *models.Transaction) error {
	transaction.Description = strings.TrimSpace(transaction.Description)

	if transaction.Amount <= 0 {
		return fmt.Errorf("le montant doit être supérieur à zéro")
	}
	if transaction.Description == "" {
		return fmt.Errorf("la description est requise")
	}
	if transaction.Date.IsZero() {
		return fmt.Errorf("la date est requise")
	}

	return nil
}
