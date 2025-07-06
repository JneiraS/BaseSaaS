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

// DocumentService encapsule la logique métier pour la gestion des documents.
type DocumentService struct {
	documentRepo repositories.DocumentRepository
	cfg          *config.Config
}

// NewDocumentService crée une nouvelle instance de DocumentService.
func NewDocumentService(documentRepo repositories.DocumentRepository, cfg *config.Config) *DocumentService {
	// Assurez-vous que le répertoire de stockage existe
	if _, err := os.Stat(cfg.DocumentStoragePath); os.IsNotExist(err) {
		err := os.MkdirAll(cfg.DocumentStoragePath, 0755)
		if err != nil {
			// Gérer l'erreur de création de répertoire, peut-être loguer et continuer avec un avertissement
			fmt.Printf("AVERTISSEMENT: Impossible de créer le répertoire de stockage des documents: %v\n", err)
		}
	}
	return &DocumentService{documentRepo: documentRepo, cfg: cfg}
}

// UploadDocument gère le téléchargement et l'enregistrement d'un document.
func (s *DocumentService) UploadDocument(userID uint, name string, file *multipart.FileHeader) error {
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("impossible d'ouvrir le fichier téléchargé: %w", err)
	}
	defer src.Close()

	// Générer un nom de fichier unique pour éviter les collisions
	uniqueFileName := fmt.Sprintf("%d_%s_%s", userID, time.Now().Format("20060102150405"), filepath.Base(file.Filename))
	destPath := filepath.Join(s.cfg.DocumentStoragePath, uniqueFileName)

	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("impossible de créer le fichier sur le serveur: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("impossible de copier le fichier: %w", err)
	}

	// Enregistrer les informations du document dans la base de données
	document := &models.Document{
		Name:       name,
		FilePath:   destPath,
		FileSize:   file.Size,
		MimeType:   file.Header.Get("Content-Type"),
		UploadDate: time.Now(),
		UserID:     userID,
	}

	if err := s.documentRepo.CreateDocument(document); err != nil {
		// Si l'enregistrement en DB échoue, tenter de supprimer le fichier physique
		os.Remove(destPath)
		return fmt.Errorf("impossible d'enregistrer le document en base de données: %w", err)
	}

	return nil
}

// GetDocumentByID récupère un document par son ID.
func (s *DocumentService) GetDocumentByID(id uint) (*models.Document, error) {
	return s.documentRepo.FindDocumentByID(id)
}

// GetDocumentsByUserID récupère tous les documents d'un utilisateur.
func (s *DocumentService) GetDocumentsByUserID(userID uint) ([]models.Document, error) {
	return s.documentRepo.FindDocumentsByUserID(userID)
}

// DeleteDocument supprime un document de la base de données et du système de fichiers.
func (s *DocumentService) DeleteDocument(documentID uint) error {
	return s.documentRepo.DeleteDocument(documentID)
}

// GetTotalDocumentsCount retourne le nombre total de documents pour un utilisateur donné.
func (s *DocumentService) GetTotalDocumentsCount(userID uint) (int64, error) {
	return s.documentRepo.GetTotalDocumentsCount(userID)
}

