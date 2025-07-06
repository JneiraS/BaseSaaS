package repositories

import (
	"time"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"gorm.io/gorm"
)

// DocumentDB représente le modèle de document pour la persistance GORM.
type DocumentDB struct {
	gorm.Model
	Name       string
	FilePath   string
	FileSize   int64
	MimeType   string
	UploadDate time.Time
	UserID     uint
}

// TableName spécifie le nom de la table pour le modèle DocumentDB.
func (DocumentDB) TableName() string {
	return "documents"
}

// DocumentRepository définit l'interface pour les opérations de persistance des documents.
type DocumentRepository interface {
	CreateDocument(document *models.Document) error
	FindDocumentByID(id uint) (*models.Document, error)
	FindDocumentsByUserID(userID uint) ([]models.Document, error)
	DeleteDocument(id uint) error
}

// GormDocumentRepository est une implémentation de DocumentRepository utilisant GORM.
type GormDocumentRepository struct {
	db *gorm.DB
}

// NewGormDocumentRepository crée une nouvelle instance de GormDocumentRepository.
func NewGormDocumentRepository(db *gorm.DB) *GormDocumentRepository {
	return &GormDocumentRepository{db: db}
}

// CreateDocument crée un nouveau document.
func (r *GormDocumentRepository) CreateDocument(document *models.Document) error {
	documentDB := toDocumentDB(document)
	if err := r.db.Create(&documentDB).Error; err != nil {
		return err
	}
	*document = *toDocument(documentDB)
	return nil
}

// FindDocumentByID recherche un document par son ID.
func (r *GormDocumentRepository) FindDocumentByID(id uint) (*models.Document, error) {
	var documentDB DocumentDB
	result := r.db.First(&documentDB, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return toDocument(&documentDB), nil
}

// FindDocumentsByUserID recherche tous les documents pour un utilisateur donné.
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

// DeleteDocument supprime un document par son ID.
func (r *GormDocumentRepository) DeleteDocument(id uint) error {
	return r.db.Delete(&DocumentDB{}, id).Error
}

// toDocumentDB convertit un modèle de domaine Document en un modèle de base de données DocumentDB.
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

// toDocument convertit un modèle de base de données DocumentDB en un modèle de domaine Document.
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
