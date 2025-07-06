package services

import (
	"fmt"
	"strings"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/domain/repositories"
)

// FinanceService encapsule la logique métier pour la gestion financière.
type FinanceService struct {
	transactionRepo repositories.TransactionRepository
}

// NewFinanceService crée une nouvelle instance de FinanceService.
func NewFinanceService(transactionRepo repositories.TransactionRepository) *FinanceService {
	return &FinanceService{transactionRepo: transactionRepo}
}

// CreateTransaction gère la création d'une nouvelle transaction.
func (s *FinanceService) CreateTransaction(transaction *models.Transaction) error {
	if err := s.validateTransaction(transaction); err != nil {
		return err
	}
	return s.transactionRepo.CreateTransaction(transaction)
}

// GetTransactionByID récupère une transaction par son ID.
func (s *FinanceService) GetTransactionByID(id uint) (*models.Transaction, error) {
	return s.transactionRepo.FindTransactionByID(id)
}

// GetTransactionsByUserID récupère toutes les transactions d'un utilisateur.
func (s *FinanceService) GetTransactionsByUserID(userID uint) ([]models.Transaction, error) {
	return s.transactionRepo.FindTransactionsByUserID(userID)
}

// UpdateTransaction gère la mise à jour d'une transaction.
func (s *FinanceService) UpdateTransaction(transaction *models.Transaction) error {
	if err := s.validateTransaction(transaction); err != nil {
		return err
	}
	return s.transactionRepo.UpdateTransaction(transaction)
}

// DeleteTransaction gère la suppression d'une transaction.
func (s *FinanceService) DeleteTransaction(id uint) error {
	return s.transactionRepo.DeleteTransaction(id)
}

// validateTransaction valide les données d'une transaction.
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
