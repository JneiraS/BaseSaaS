package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/JneiraS/BaseSasS/internal/config"
	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/domain/repositories"
)

// DocumentService encapsulates the business logic for managing documents.
// It interacts with the DocumentRepository for database operations and the file system for storage.
type DocumentService struct {
	documentRepo repositories.DocumentRepository
	cfg          *config.Config
}

// NewDocumentService creates a new instance of DocumentService.
// It ensures that the document storage directory exists upon initialization.
func NewDocumentService(documentRepo repositories.DocumentRepository, cfg *config.Config) *DocumentService {
	// Ensure the document storage directory exists. Create it if it doesn't.
	if _, err := os.Stat(cfg.DocumentStoragePath); os.IsNotExist(err) {
		err := os.MkdirAll(cfg.DocumentStoragePath, 0755)
		if err != nil {
			// Log the error if directory creation fails, but continue with a warning.
			fmt.Printf("AVERTISSEMENT: Impossible de créer le répertoire de stockage des documents: %v\n", err)
		}
	}
	return &DocumentService{documentRepo: documentRepo, cfg: cfg}
}

// UploadDocument handles the upload and storage of a document.
// It saves the file to the configured storage path and records its metadata in the database.
func (s *DocumentService) UploadDocument(userID uint, name string, file *multipart.FileHeader) error {
	// Open the uploaded file.
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("impossible d'ouvrir le fichier téléchargé: %w", err)
	}
	defer src.Close()

	// Generate a unique file name to prevent collisions.
	uniqueFileName := fmt.Sprintf("%d_%s_%s", userID, time.Now().Format("20060102150405"), filepath.Base(file.Filename))
	// Construct the full destination path for the file.
	destPath := filepath.Join(s.cfg.DocumentStoragePath, uniqueFileName)

	// Create the destination file on the server.
	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("impossible de créer le fichier sur le serveur: %w", err)
	}
	defer dst.Close()

	// Copy the uploaded file content to the destination file.
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("impossible de copier le fichier: %w", err)
	}

	// Prepare document metadata for database storage.
	document := &models.Document{
		Name:       name,
		FilePath:   destPath,
		FileSize:   file.Size,
		MimeType:   file.Header.Get("Content-Type"),
		UploadDate: time.Now(),
		UserID:     userID,
	}

	// Save document information to the database.
	if err := s.documentRepo.CreateDocument(document); err != nil {
		// If database record creation fails, attempt to remove the physically saved file to prevent orphans.
		os.Remove(destPath)
		return fmt.Errorf("impossible d'enregistrer le document en base de données: %w", err)
	}

	return nil
}

// GetDocumentByID retrieves a document by its unique identifier.
func (s *DocumentService) GetDocumentByID(id uint) (*models.Document, error) {
	return s.documentRepo.FindDocumentByID(id)
}

// GetDocumentsByUserID retrieves all documents associated with a specific user ID.
func (s *DocumentService) GetDocumentsByUserID(userID uint) ([]models.Document, error) {
	return s.documentRepo.FindDocumentsByUserID(userID)
}

// DeleteDocument deletes a document record from the database.
// Note: This service method only handles the database record deletion.
// The actual file deletion from the file system might be handled elsewhere or in the repository.
func (s *DocumentService) DeleteDocument(documentID uint) error {
	return s.documentRepo.DeleteDocument(documentID)
}

// GetTotalDocumentsCount returns the total number of documents for a given user ID.
func (s *DocumentService) GetTotalDocumentsCount(userID uint) (int64, error) {
	return s.documentRepo.GetTotalDocumentsCount(userID)
}
