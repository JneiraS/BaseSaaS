package repositories

import (
	"time"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"gorm.io/gorm"
)

// TransactionDB représente le modèle de transaction pour la persistance GORM.
type TransactionDB struct {
	gorm.Model
	Amount      float64
	Type        models.TransactionType
	Description string
	Date        time.Time
	UserID      uint
}

// TableName spécifie le nom de la table pour le modèle TransactionDB.
func (TransactionDB) TableName() string {
	return "transactions"
}

// TransactionRepository définit l'interface pour les opérations de persistance des transactions.
type TransactionRepository interface {
	CreateTransaction(transaction *models.Transaction) error
	FindTransactionByID(id uint) (*models.Transaction, error)
	FindTransactionsByUserID(userID uint) ([]models.Transaction, error)
	UpdateTransaction(transaction *models.Transaction) error
	DeleteTransaction(id uint) error
}

// GormTransactionRepository est une implémentation de TransactionRepository utilisant GORM.
type GormTransactionRepository struct {
	db *gorm.DB
}

// NewGormTransactionRepository crée une nouvelle instance de GormTransactionRepository.
func NewGormTransactionRepository(db *gorm.DB) *GormTransactionRepository {
	return &GormTransactionRepository{db: db}
}

// CreateTransaction crée une nouvelle transaction.
func (r *GormTransactionRepository) CreateTransaction(transaction *models.Transaction) error {
	transactionDB := toTransactionDB(transaction)
	if err := r.db.Create(&transactionDB).Error; err != nil {
		return err
	}
	*transaction = *toTransaction(transactionDB)
	return nil
}

// FindTransactionByID recherche une transaction par son ID.
func (r *GormTransactionRepository) FindTransactionByID(id uint) (*models.Transaction, error) {
	var transactionDB TransactionDB
	result := r.db.First(&transactionDB, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return toTransaction(&transactionDB), nil
}

// FindTransactionsByUserID recherche toutes les transactions pour un utilisateur donné.
func (r *GormTransactionRepository) FindTransactionsByUserID(userID uint) ([]models.Transaction, error) {
	var transactionsDB []TransactionDB
	if err := r.db.Where("user_id = ?", userID).Find(&transactionsDB).Error; err != nil {
		return nil, err
	}
	var transactions []models.Transaction
	for _, tdb := range transactionsDB {
		transactions = append(transactions, *toTransaction(&tdb))
	}
	return transactions, nil
}

// UpdateTransaction met à jour une transaction existante.
func (r *GormTransactionRepository) UpdateTransaction(transaction *models.Transaction) error {
	transactionDB := toTransactionDB(transaction)
	return r.db.Save(&transactionDB).Error
}

// DeleteTransaction supprime une transaction par son ID.
func (r *GormTransactionRepository) DeleteTransaction(id uint) error {
	return r.db.Delete(&TransactionDB{}, id).Error
}

// toTransactionDB convertit un modèle de domaine Transaction en un modèle de base de données TransactionDB.
func toTransactionDB(t *models.Transaction) *TransactionDB {
	return &TransactionDB{
		Model:       gorm.Model{ID: t.ID, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt, DeletedAt: t.DeletedAt},
		Amount:      t.Amount,
		Type:        t.Type,
		Description: t.Description,
		Date:        t.Date,
		UserID:      t.UserID,
	}
}

// toTransaction convertit un modèle de base de données TransactionDB en un modèle de domaine Transaction.
func toTransaction(tdb *TransactionDB) *models.Transaction {
	return &models.Transaction{
		Model:       gorm.Model{ID: tdb.ID, CreatedAt: tdb.CreatedAt, UpdatedAt: tdb.UpdatedAt, DeletedAt: tdb.DeletedAt},
		Amount:      tdb.Amount,
		Type:        tdb.Type,
		Description: tdb.Description,
		Date:        tdb.Date,
		UserID:      tdb.UserID,
	}
}
