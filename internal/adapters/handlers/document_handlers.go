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

// DocumentHandlers encapsule les dépendances pour les handlers de documents.
type DocumentHandlers struct {
	documentService *services.DocumentService
}

// NewDocumentHandlers crée une nouvelle instance de DocumentHandlers.
func NewDocumentHandlers(documentService *services.DocumentService) *DocumentHandlers {
	return &DocumentHandlers{documentService: documentService}
}

// ListDocuments affiche la liste des documents.
func (h *DocumentHandlers) ListDocuments(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	documents, err := h.documentService.GetDocumentsByUserID(user.ID)
	if err != nil {
		log.Printf("ERREUR: Erreur lors de la récupération des documents: %v", err)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la récupération des documents."})
		return
	}

	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	c.HTML(http.StatusOK, "documents.tmpl", gin.H{
		"title":      "Mes Documents",
		"navbar":     navbar,
		"user":       user,
		"documents":  documents,
		"csrf_token": csrfToken,
	})
	if err := session.Save(); err != nil {
		// Gérer l'erreur de sauvegarde de session si nécessaire
		// log.Printf("Erreur lors de la sauvegarde de session dans ListDocuments: %v", err)
	}
}

// ShowUploadForm affiche le formulaire de téléchargement de document.
func (h *DocumentHandlers) ShowUploadForm(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	c.HTML(http.StatusOK, "document_upload_form.tmpl", gin.H{
		"title":      "Télécharger un document",
		"navbar":     navbar,
		"user":       user,
		"csrf_token": csrfToken,
	})
}

// UploadDocument gère le téléchargement de documents.
func (h *DocumentHandlers) UploadDocument(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	file, err := c.FormFile("document")
	if err != nil {
		log.Printf("ERREUR: Erreur lors de la récupération du fichier: %v", err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "Erreur lors de la récupération du fichier: " + err.Error()})
		return
	}

	// Récupérer le nom du document depuis le formulaire (si fourni)
	documentName := c.PostForm("name")
	if documentName == "" {
		documentName = file.Filename // Utiliser le nom de fichier par défaut si non fourni
	}

	// Appeler le service pour gérer le téléchargement et l'enregistrement en DB
	if err := h.documentService.UploadDocument(user.ID, documentName, file); err != nil {
		log.Printf("ERREUR: Échec du téléchargement du document: %v", err)
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Échec du téléchargement du document: " + err.Error()})
		return
	}

	session.AddFlash("Document téléchargé avec succès !", "success")
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
	}
	c.Redirect(http.StatusFound, "/documents")
}

// DownloadDocument gère le téléchargement d'un document.
func (h *DocumentHandlers) DownloadDocument(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	documentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de document invalide"})
		return
	}

	document, err := h.documentService.GetDocumentByID(uint(documentID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Document non trouvé"})
		return
	}

	// Vérifier que le document appartient bien à l'utilisateur connecté
	if document.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	// Envoyer le fichier au client
	c.FileAttachment(document.FilePath, document.Name)
}

// DeleteDocument gère la suppression d'un document.
func (h *DocumentHandlers) DeleteDocument(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	documentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID de document invalide"})
		return
	}

	// Vérifier que le document appartient bien à l'utilisateur connecté avant de supprimer
	document, err := h.documentService.GetDocumentByID(uint(documentID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Document non trouvé"})
		return
	}

	if document.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	if err := h.documentService.DeleteDocument(uint(documentID)); err != nil {
		log.Printf("ERREUR: Échec de la suppression du document: %v", err)
		session.AddFlash("Échec de la suppression du document: "+err.Error(), "error")
		if err := session.Save(); err != nil {
			log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		}
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Échec de la suppression du document: " + err.Error()})
		return
	}

	session.AddFlash("Document supprimé avec succès !", "success")
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
	}
	c.Redirect(http.StatusFound, "/documents")
}
