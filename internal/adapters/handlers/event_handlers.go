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

// EventHandlers encapsulates the dependencies for event-related HTTP handlers.
// It holds a reference to the EventService, which contains the business logic for events.
type EventHandlers struct {
	eventService *services.EventService
}

// NewEventHandlers creates a new instance of EventHandlers.
// It takes an EventService as a dependency, adhering to the dependency inversion principle.
func NewEventHandlers(eventService *services.EventService) *EventHandlers {
	return &EventHandlers{eventService: eventService}
}

// ListEvents displays a list of events for the authenticated user.
// It retrieves events from the EventService and renders them using the "events.tmpl" template.
func (h *EventHandlers) ListEvents(c *gin.Context) {
	// Retrieve the authenticated user from the session.
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Retrieve events associated with the current user.
	events, err := h.eventService.GetEventsByUserID(user.ID)
	if err != nil {
		// Handle error, e.g., display an error message to the user.
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la récupération des événements"})
		return
	}

	// Retrieve CSRF token for the navigation bar.
	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	// Render the events list page.
	c.HTML(http.StatusOK, "events.tmpl", gin.H{
		"title":      "Mes Événements",
		"navbar":     navbar,
		"user":       user,
		"events":     events,
		"csrf_token": csrfToken, // Add CSRF token to the template context
	})
	// Save session changes if any (e.g., flash messages).
	if err := session.Save(); err != nil {
		// Handle session save error if necessary
		// log.Printf("Erreur lors de la sauvegarde de session dans ListEvents: %v", err)
	}
}

// ShowCreateEventForm displays the form for creating a new event.
// It provides default values for start and end dates for convenience.
func (h *EventHandlers) ShowCreateEventForm(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	// Render the event creation form.
	c.HTML(http.StatusOK, "event_form.tmpl", gin.H{
		"title":      "Créer un nouvel événement",
		"navbar":     navbar,
		"user":       user,
		"csrf_token": csrfToken,
		"event":      models.Event{StartDate: time.Now(), EndDate: time.Now().Add(time.Hour)}, // Default values
	})
	if err := session.Save(); err != nil {
		// Handle session save error if necessary
		// log.Printf("Erreur lors de la sauvegarde de session dans ShowCreateEventForm: %v", err)
	}
}

// CreateEvent handles the submission of the new event creation form.
// It binds the form data to an Event model, sets the UserID, and calls the service to create the event.
func (h *EventHandlers) CreateEvent(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var newEvent models.Event
	// Bind form data to the newEvent struct. If binding fails, return a bad request error.
	if err := c.ShouldBind(&newEvent); err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "Données d'événement invalides: " + err.Error()})
		return
	}

	newEvent.UserID = user.ID // Assign the current user's ID to the new event.

	// Call the service to create the event. Handle any errors during creation.
	if err := h.eventService.CreateEvent(&newEvent); err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la création de l'événement: " + err.Error()})
		return
	}

	// Redirect to the events list page upon successful creation.
	c.Redirect(http.StatusFound, "/events")
}

// ShowEditEventForm displays the form for editing an existing event.
// It retrieves the event by ID, ensures it belongs to the authenticated user, and populates the form.
func (h *EventHandlers) ShowEditEventForm(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Parse the event ID from the URL parameter.
	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID d'événement invalide"})
		return
	}

	// Retrieve the event from the service.
	event, err := h.eventService.GetEventByID(uint(eventID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Événement non trouvé"})
		return
	}

	// Verify that the event belongs to the authenticated user for security.
	if event.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	// Render the event edit form.
	c.HTML(http.StatusOK, "event_form.tmpl", gin.H{
		"title":      "Modifier l'événement",
		"navbar":     navbar,
		"user":       user,
		"csrf_token": csrfToken,
		"event":      event,
	})
	if err := session.Save(); err != nil {
		// Handle session save error if necessary
		// log.Printf("Erreur lors de la sauvegarde de session dans ShowEditEventForm: %v", err)
	}
}

// UpdateEvent handles the submission of the event modification form.
// It retrieves the existing event, binds updated data, ensures ownership, and calls the service to update.
func (h *EventHandlers) UpdateEvent(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Parse the event ID from the URL parameter.
	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID d'événement invalide"})
		return
	}

	var updatedEvent models.Event
	// Bind form data to a temporary updatedEvent struct.
	if err := c.ShouldBind(&updatedEvent); err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "Données d'événement invalides: " + err.Error()})
		return
	}

	// Retrieve the existing event to ensure it belongs to the user before updating.
	existingEvent, err := h.eventService.GetEventByID(uint(eventID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Événement non trouvé"})
		return
	}

	if existingEvent.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	// Update the fields of the existing event with the new data from the form.
	existingEvent.Title = updatedEvent.Title
	existingEvent.Description = updatedEvent.Description
	existingEvent.StartDate = updatedEvent.StartDate
	existingEvent.EndDate = updatedEvent.EndDate

	// Call the service to update the event. Handle any errors during update.
	if err := h.eventService.UpdateEvent(existingEvent); err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la mise à jour de l'événement: " + err.Error()})
		return
	}

	// Redirect to the events list page upon successful update.
	c.Redirect(http.StatusFound, "/events")
}

// DeleteEvent handles the deletion of an event.
// It retrieves the event by ID, ensures it belongs to the authenticated user, and calls the service to delete it.
func (h *EventHandlers) DeleteEvent(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Parse the event ID from the URL parameter.
	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"error": "ID d'événement invalide"})
		return
	}

	// Verify that the event belongs to the authenticated user before deletion.
	existingEvent, err := h.eventService.GetEventByID(uint(eventID))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.tmpl", gin.H{"error": "Événement non trouvé"})
		return
	}

	if existingEvent.UserID != user.ID {
		c.HTML(http.StatusForbidden, "error.tmpl", gin.H{"error": "Accès non autorisé"})
		return
	}

	// Call the service to delete the event. Handle any errors during deletion.
	if err := h.eventService.DeleteEvent(uint(eventID)); err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": "Erreur lors de la suppression de l'événement: " + err.Error()})
		return
	}

	// Redirect to the events list page upon successful deletion.
	c.Redirect(http.StatusFound, "/events")
}
