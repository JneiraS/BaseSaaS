package models

import (
	"time"

	"gorm.io/gorm"
)

// Document represents a file uploaded by a user within the application.
// It embeds gorm.Model for common fields like ID, CreatedAt, UpdatedAt, and DeletedAt.
type Document struct {
	gorm.Model
	Name       string    `json:"name" form:"name"`       // The original name of the uploaded document.
	FilePath   string    `json:"file_path"`             // The relative or absolute path to the file on the server's file system.
	FileSize   int64     `json:"file_size"`             // The size of the file in bytes.
	MimeType   string    `json:"mime_type"`             // The MIME type of the file (e.g., "application/pdf", "image/png").
	UploadDate time.Time `json:"upload_date"`           // The timestamp when the document was uploaded.

	// UserID is the ID of the application user who uploaded this document.
	// This establishes a relationship between the document and its owner.
	UserID uint `json:"user_id"`
}
