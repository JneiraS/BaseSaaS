package repositories

import (
	"time"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"gorm.io/gorm"
)

// DocumentDB represents the database model for a document, used for GORM persistence.
// It includes GORM's Model for common fields like ID, CreatedAt, UpdatedAt, DeletedAt.
type DocumentDB struct {
	gorm.Model
	Name       string    // Original name of the document file.
	FilePath   string    // Path to the stored file on the server.
	FileSize   int64     // Size of the file in bytes.
	MimeType   string    // MIME type of the file (e.g., "application/pdf").
	UploadDate time.Time // Date and time when the document was uploaded.
	UserID     uint      // Foreign key linking to the User who uploaded this document.
}

// TableName specifies the table name for the DocumentDB model in the database.
// This overrides GORM's default naming convention.
func (DocumentDB) TableName() string {
	return "documents"
}

// DocumentRepository defines the interface for document persistence operations.
// It abstracts the underlying database implementation.
type DocumentRepository interface {
	CreateDocument(document *models.Document) error
	FindDocumentByID(id uint) (*models.Document, error)
	FindDocumentsByUserID(userID uint) ([]models.Document, error)
	DeleteDocument(id uint) error
	GetTotalDocumentsCount(userID uint) (int64, error)
}

// GormDocumentRepository is an implementation of DocumentRepository that uses GORM
// for interacting with a relational database.
type GormDocumentRepository struct {
	db *gorm.DB // GORM database client
}

// NewGormDocumentRepository creates a new instance of GormDocumentRepository.
// It takes a GORM DB instance as a dependency.
func NewGormDocumentRepository(db *gorm.DB) *GormDocumentRepository {
	return &GormDocumentRepository{db: db}
}

// CreateDocument persists a new document to the database.
// It converts the domain model Document to a database-specific DocumentDB model
// before saving and then updates the domain model with the generated ID.
func (r *GormDocumentRepository) CreateDocument(document *models.Document) error {
	documentDB := toDocumentDB(document)
	if err := r.db.Create(&documentDB).Error; err != nil {
		return err
	}
	*document = *toDocument(documentDB) // Update the original document with DB-generated fields (e.g., ID)
	return nil
}

// FindDocumentByID retrieves a document from the database by its ID.
// It returns the document as a domain model or an error if not found.
func (r *GormDocumentRepository) FindDocumentByID(id uint) (*models.Document, error) {
	var documentDB DocumentDB
	result := r.db.First(&documentDB, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return toDocument(&documentDB), nil
}

// FindDocumentsByUserID retrieves all documents associated with a specific user ID.
// It queries the database for documents where the UserID matches the provided ID.
func (r *GormDocumentRepository) FindDocumentsByUserID(userID uint) ([]models.Document, error) {
	var documentsDB []DocumentDB
	if err := r.db.Where("user_id = ?", userID).Find(&documentsDB).Error; err != nil {
		return nil, err
	}
	var documents []models.Document
	for _, ddb := range documentsDB {
		documents = append(documents, *toDocument(&ddb))
	}
	return documents, nil
}

// DeleteDocument deletes a document from the database by its ID.
func (r *GormDocumentRepository) DeleteDocument(id uint) error {
	return r.db.Delete(&DocumentDB{}, id).Error
}

// GetTotalDocumentsCount returns the total number of documents for a given user ID.
// It performs a count query on the documents table, filtered by user_id.
func (r *GormDocumentRepository) GetTotalDocumentsCount(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&DocumentDB{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// toDocumentDB converts a domain Document model to a database-specific DocumentDB model.
// This is used before persisting the document to the database.
func toDocumentDB(d *models.Document) *DocumentDB {
	return &DocumentDB{
		Model:      gorm.Model{ID: d.ID, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt, DeletedAt: d.DeletedAt},
		Name:       d.Name,
		FilePath:   d.FilePath,
		FileSize:   d.FileSize,
		MimeType:   d.MimeType,
		UploadDate: d.UploadDate,
		UserID:     d.UserID,
	}
}

// toDocument converts a database-specific DocumentDB model back to a domain Document model.
// This is used after retrieving data from the database.
func toDocument(ddb *DocumentDB) *models.Document {
	return &models.Document{
		Model:      gorm.Model{ID: ddb.ID, CreatedAt: ddb.CreatedAt, UpdatedAt: ddb.UpdatedAt, DeletedAt: ddb.DeletedAt},
		Name:       ddb.Name,
		FilePath:   ddb.FilePath,
		FileSize:   ddb.FileSize,
		MimeType:   ddb.MimeType,
		UploadDate: ddb.UploadDate,
		UserID:     ddb.UserID,
	}
}
