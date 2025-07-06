package models

import (
	"time"

	"gorm.io/gorm"
)

// Document représente un document téléchargé par un utilisateur.
type Document struct {
	gorm.Model
	Name       string    `json:"name" form:"name"`
	FilePath   string    `json:"file_path"` // Chemin relatif ou absolu du fichier sur le serveur
	FileSize   int64     `json:"file_size"` // Taille du fichier en octets
	MimeType   string    `json:"mime_type"` // Type MIME du fichier
	UploadDate time.Time `json:"upload_date"`

	// UserID est l'ID de l'utilisateur de l'application qui a téléchargé ce document.
	UserID uint `json:"user_id"`
}
