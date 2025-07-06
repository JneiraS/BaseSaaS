package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/JneiraS/BaseSasS/components"
	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/services"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// EventHandlers encapsule les dépendances pour les handlers des événements.
type EventHandlers struct {
	eventService *services.EventService
}

// NewEventHandlers crée une nouvelle instance de EventHandlers.
func NewEventHandlers(eventService *services.EventService) *EventHandlers {
	return &EventHandlers{eventService: eventService}
}

// ListEvents affiche la liste des événements.
func (h *EventHandlers) ListEvents(c *gin.Context) {
	// Récupérer l'utilisateur connecté depuis la session
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Récupérer les événements associés à cet utilisateur
	events, err := h.eventService.GetEventsByUserID(user.ID)
	if err != nil {
		// Gérer l'erreur, par exemple, afficher un message d'erreur
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la récupération des événements"})
		return
	}

	// Récupérer le jeton CSRF pour la navbar
	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	c.HTML(http.StatusOK, "events.tmpl", gin.H{
		"title":      "Mes Événements",
		"navbar":     navbar,
		"user":       user,
		"events":     events,
		"csrf_token": csrfToken, // Ajout du jeton CSRF au contexte du template
	})
}

// ShowCreateEventForm affiche le formulaire de création d'un nouvel événement.
func (h *EventHandlers) ShowCreateEventForm(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	c.HTML(http.StatusOK, "event_form.tmpl", gin.H{
		"title":      "Créer un nouvel événement",
		"navbar":     navbar,
		"user":       user,
		"csrf_token": csrfToken,
		"event":      models.Event{StartDate: time.Now(), EndDate: time.Now().Add(time.Hour)}, // Valeurs par défaut
	})
}

// CreateEvent gère la soumission du formulaire de création d'événement.
func (h *EventHandlers) CreateEvent(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var newEvent models.Event
	if err := c.ShouldBind(&newEvent); err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "Données d'événement invalides: " + err.Error()})
		return
	}

	newEvent.UserID = user.ID

	if err := h.eventService.CreateEvent(&newEvent); err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la création de l'événement: " + err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/events")
}

// ShowEditEventForm affiche le formulaire de modification d'un événement existant.
func (h *EventHandlers) ShowEditEventForm(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID d'événement invalide"})
		return
	}

	event, err := h.eventService.GetEventByID(uint(eventID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Événement non trouvé"})
		return
	}

	// Vérifier que l'événement appartient bien à l'utilisateur connecté
	if event.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	c.HTML(http.StatusOK, "event_form.tmpl", gin.H{
		"title":      "Modifier l'événement",
		"navbar":     navbar,
		"user":       user,
		"csrf_token": csrfToken,
		"event":      event,
	})
}

// UpdateEvent gère la soumission du formulaire de modification d'événement.
func (h *EventHandlers) UpdateEvent(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID d'événement invalide"})
		return
	}

	var updatedEvent models.Event
	if err := c.ShouldBind(&updatedEvent); err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "Données d'événement invalides: " + err.Error()})
		return
	}

	// Récupérer l'événement existant pour s'assurer qu'il appartient à l'utilisateur
	existingEvent, err := h.eventService.GetEventByID(uint(eventID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Événement non trouvé"})
		return
	}

	if existingEvent.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	// Mettre à jour les champs de l'événement existant avec les données du formulaire
	existingEvent.Title = updatedEvent.Title
	existingEvent.Description = updatedEvent.Description
	existingEvent.StartDate = updatedEvent.StartDate
	existingEvent.EndDate = updatedEvent.EndDate

	if err := h.eventService.UpdateEvent(existingEvent); err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la mise à jour de l'événement: " + err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/events")
}

// DeleteEvent gère la suppression d'un événement.
func (h *EventHandlers) DeleteEvent(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID d'événement invalide"})
		return
	}

	// Vérifier que l'événement appartient bien à l'utilisateur connecté avant de supprimer
	existingEvent, err := h.eventService.GetEventByID(uint(eventID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Événement non trouvé"})
		return
	}

	if existingEvent.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	if err := h.eventService.DeleteEvent(uint(eventID)); err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la suppression de l'événement: " + err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/events")
}
