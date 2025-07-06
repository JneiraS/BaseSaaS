package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/JneiraS/BaseSasS/components"
	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/services"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// DocumentHandlers encapsulates the dependencies for document-related HTTP handlers.
// It holds a reference to the DocumentService, which contains the business logic for documents.
type DocumentHandlers struct {
	documentService *services.DocumentService
}

// NewDocumentHandlers creates a new instance of DocumentHandlers.
// It takes a DocumentService as a dependency, adhering to the dependency inversion principle.
func NewDocumentHandlers(documentService *services.DocumentService) *DocumentHandlers {
	return &DocumentHandlers{documentService: documentService}
}

// ListDocuments displays a list of documents for the authenticated user.
// It retrieves documents from the DocumentService and renders them using the "documents.tmpl" template.
func (h *DocumentHandlers) ListDocuments(c *gin.Context) {
	// Retrieve the authenticated user from the session.
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Retrieve documents associated with the current user.
	documents, err := h.documentService.GetDocumentsByUserID(user.ID)
	if err != nil {
		log.Printf("ERREUR: Erreur lors de la récupération des documents: %v", err)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la récupération des documents."})
		return
	}

	// Retrieve CSRF token for the navigation bar.
	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	// Render the documents list page.
	c.HTML(http.StatusOK, "documents.tmpl", gin.H{
		"title":      "Mes Documents",
		"navbar":     navbar,
		"user":       user,
		"documents":  documents,
		"csrf_token": csrfToken,
	})
	// Save session changes if any (e.g., flash messages).
	if err := session.Save(); err != nil {
		// Handle session save error if necessary
		// log.Printf("Erreur lors de la sauvegarde de session dans ListDocuments: %v", err)
	}
}

// ShowUploadForm displays the form for uploading a new document.
func (h *DocumentHandlers) ShowUploadForm(c *gin.Context) {
	// Retrieve the authenticated user from the session.
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Retrieve CSRF token for the navigation bar.
	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	// Render the document upload form page.
	c.HTML(http.StatusOK, "document_upload_form.tmpl", gin.H{
		"title":      "Télécharger un document",
		"navbar":     navbar,
		"user":       user,
		"csrf_token": csrfToken,
	})
	// Save session changes if any.
	if err := session.Save(); err != nil {
		// Handle session save error if necessary
		// log.Printf("Erreur lors de la sauvegarde de session dans ShowUploadForm: %v", err)
	}
}

// UploadDocument handles the submission of the document upload form.
// It retrieves the file and its name from the form, and calls the service to handle the upload and database record creation.
func (h *DocumentHandlers) UploadDocument(c *gin.Context) {
	// Retrieve the authenticated user from the session.
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Retrieve the uploaded file from the form.
	file, err := c.FormFile("document")
	if err != nil {
		log.Printf("ERREUR: Erreur lors de la récupération du fichier: %v", err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "Erreur lors de la récupération du fichier: " + err.Error()})
		return
	}

	// Get the document name from the form (if provided). Use the original filename if not provided.
	documentName := c.PostForm("name")
	if documentName == "" {
		documentName = file.Filename
	}

	// Call the service to handle the file upload and database record creation.
	if err := h.documentService.UploadDocument(user.ID, documentName, file); err != nil {
		log.Printf("ERREUR: Échec du téléchargement du document: %v", err)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Échec du téléchargement du document: " + err.Error()})
		return
	}

	// Add a success flash message and redirect to the documents list page.
	session.AddFlash("Document téléchargé avec succès !", "success")
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
	}
	c.Redirect(http.StatusFound, "/documents")
}

// DownloadDocument handles the download of a specific document.
// It retrieves the document by ID, ensures it belongs to the authenticated user, and serves the file.
func (h *DocumentHandlers) DownloadDocument(c *gin.Context) {
	// Retrieve the authenticated user from the session.
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Parse the document ID from the URL parameter.
	documentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de document invalide"})
		return
	}

	// Retrieve the document from the service.
	document, err := h.documentService.GetDocumentByID(uint(documentID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Document non trouvé"})
		return
	}

	// Verify that the document belongs to the authenticated user for security.
	if document.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	// Serve the file to the client.
	c.FileAttachment(document.FilePath, document.Name)
}

// DeleteDocument handles the deletion of a document.
// It retrieves the document by ID, ensures it belongs to the authenticated user, and calls the service to delete it.
func (h *DocumentHandlers) DeleteDocument(c *gin.Context) {
	// Retrieve the authenticated user from the session.
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Parse the document ID from the URL parameter.
	documentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de document invalide"})
		return
	}

	// Verify that the document belongs to the authenticated user before deletion.
	document, err := h.documentService.GetDocumentByID(uint(documentID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Document non trouvé"})
		return
	}

	if document.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	// Call the service to delete the document. Handle any errors during deletion.
	if err := h.documentService.DeleteDocument(uint(documentID)); err != nil {
		log.Printf("ERREUR: Échec de la suppression du document: %v", err)
		session.AddFlash("Échec de la suppression du document: "+err.Error(), "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Échec de la suppression du document: " + err.Error()})
		return
	}

	// Add a success flash message and redirect to the documents list page.
	session.AddFlash("Document supprimé avec succès !", "success")
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
	}
	c.Redirect(http.StatusFound, "/documents")
}
