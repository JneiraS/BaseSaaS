package repositories

import (
	"time"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"gorm.io/gorm"
)

// TransactionDB represents the database model for a financial transaction, used for GORM persistence.
// It includes GORM's Model for common fields like ID, CreatedAt, UpdatedAt, and DeletedAt.
type TransactionDB struct {
	gorm.Model
	Amount      float64                 // The monetary amount of the transaction.
	Type        models.TransactionType  // The type of transaction (Income or Expense).
	Description string                  // A brief description of the transaction.
	Date        time.Time               // The date when the transaction occurred.
	UserID      uint                    // Foreign key linking to the User who recorded this transaction.
}

// TableName specifies the table name for the TransactionDB model in the database.
// This overrides GORM's default naming convention.
func (TransactionDB) TableName() string {
	return "transactions"
}

// TransactionRepository defines the interface for transaction persistence operations.
// It abstracts the underlying database implementation.
type TransactionRepository interface {
	CreateTransaction(transaction *models.Transaction) error
	FindTransactionByID(id uint) (*models.Transaction, error)
	FindTransactionsByUserID(userID uint) ([]models.Transaction, error)
	UpdateTransaction(transaction *models.Transaction) error
	DeleteTransaction(id uint) error
	GetTotalIncome(userID uint) (float64, error)
	GetTotalExpenses(userID uint) (float64, error)
}

// GormTransactionRepository is an implementation of TransactionRepository that uses GORM
// for interacting with a relational database.
type GormTransactionRepository struct {
	db *gorm.DB // GORM database client
}

// NewGormTransactionRepository creates a new instance of GormTransactionRepository.
// It takes a GORM DB instance as a dependency.
func NewGormTransactionRepository(db *gorm.DB) *GormTransactionRepository {
	return &GormTransactionRepository{db: db}
}

// CreateTransaction persists a new transaction to the database.
// It converts the domain model Transaction to a database-specific TransactionDB model
// before saving and then updates the domain model with the generated ID.
func (r *GormTransactionRepository) CreateTransaction(transaction *models.Transaction) error {
	transactionDB := toTransactionDB(transaction)
	if err := r.db.Create(&transactionDB).Error; err != nil {
		return err
	}
	*transaction = *toTransaction(transactionDB) // Update the original transaction with DB-generated fields (e.g., ID)
	return nil
}

// FindTransactionByID retrieves a transaction from the database by its ID.
// It returns the transaction as a domain model or an error if not found.
func (r *GormTransactionRepository) FindTransactionByID(id uint) (*models.Transaction, error) {
	var transactionDB TransactionDB
	result := r.db.First(&transactionDB, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return toTransaction(&transactionDB), nil
}

// FindTransactionsByUserID retrieves all transactions associated with a specific user ID.
// It queries the database for transactions where the UserID matches the provided ID.
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

// UpdateTransaction updates an existing transaction in the database.
// It converts the domain model to a database model and saves the changes.
func (r *GormTransactionRepository) UpdateTransaction(transaction *models.Transaction) error {
	transactionDB := toTransactionDB(transaction)
	return r.db.Save(&transactionDB).Error
}

// DeleteTransaction deletes a transaction from the database by its ID.
func (r *GormTransactionRepository) DeleteTransaction(id uint) error {
	return r.db.Delete(&TransactionDB{}, id).Error
}

// GetTotalIncome returns the sum of all income transactions for a given user ID.
// It queries the database for transactions of type TypeIncome and sums their amounts.
func (r *GormTransactionRepository) GetTotalIncome(userID uint) (float64, error) {
	var total float64
	if err := r.db.Model(&TransactionDB{}).Where("user_id = ? AND type = ?", userID, models.TypeIncome).Select("sum(amount)").Row().Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

// GetTotalExpenses returns the sum of all expense transactions for a given user ID.
// It queries the database for transactions of type TypeExpense and sums their amounts.
func (r *GormTransactionRepository) GetTotalExpenses(userID uint) (float64, error) {
	var total float64
	if err := r.db.Model(&TransactionDB{}).Where("user_id = ? AND type = ?", userID, models.TypeExpense).Select("sum(amount)").Row().Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

// toTransactionDB converts a domain Transaction model to a database-specific TransactionDB model.
// This is used before persisting the transaction to the database.
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

// toTransaction converts a database-specific TransactionDB model back to a domain Transaction model.
// This is used after retrieving data from the database.
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
