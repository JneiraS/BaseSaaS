package services

import (
	"fmt"
	"strings"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/domain/repositories"
)

// EventService encapsulates the business logic for managing events.
// It interacts with the EventRepository to perform CRUD operations and other event-related tasks.
type EventService struct {
	eventRepo repositories.EventRepository
}

// NewEventService creates a new instance of EventService.
// It takes an EventRepository as a dependency, adhering to the dependency inversion principle.
func NewEventService(eventRepo repositories.EventRepository) *EventService {
	return &EventService{eventRepo: eventRepo}
}

// CreateEvent handles the creation of a new event.
// It performs validation on the event data before persisting it via the repository.
func (s *EventService) CreateEvent(event *models.Event) error {
	if err := s.validateEvent(event); err != nil {
		return err
	}
	return s.eventRepo.CreateEvent(event)
}

// GetEventByID retrieves an event by its unique identifier.
func (s *EventService) GetEventByID(id uint) (*models.Event, error) {
	return s.eventRepo.FindEventByID(id)
}

// GetEventsByUserID retrieves all events associated with a specific user ID.
func (s *EventService) GetEventsByUserID(userID uint) ([]models.Event, error) {
	return s.eventRepo.FindEventsByUserID(userID)
}

// UpdateEvent handles the update of an existing event.
// It performs validation on the updated event data before persisting the changes.
func (s *EventService) UpdateEvent(event *models.Event) error {
	if err := s.validateEvent(event); err != nil {
		return err
	}
	return s.eventRepo.UpdateEvent(event)
}

// DeleteEvent handles the deletion of an event by its unique identifier.
func (s *EventService) DeleteEvent(id uint) error {
	return s.eventRepo.DeleteEvent(id)
}

// GetTotalEventsCount returns the total number of events for a given user ID.
func (s *EventService) GetTotalEventsCount(userID uint) (int64, error) {
	return s.eventRepo.GetTotalEventsCount(userID)
}

// validateEvent performs business logic validation on an Event model.
// It checks for required fields and logical consistency (e.g., start date before end date).
func (s *EventService) validateEvent(event *models.Event) error {
	event.Title = strings.TrimSpace(event.Title)
	event.Description = strings.TrimSpace(event.Description)

	if event.Title == "" {
		return fmt.Errorf("le titre est requis")
	}
	if event.Description == "" {
		return fmt.Errorf("la description est requise")
	}
	if event.StartDate.IsZero() {
		return fmt.Errorf("la date de début est requise")
	}
	if event.EndDate.IsZero() {
		return fmt.Errorf("la date de fin est requise")
	}
	if event.StartDate.After(event.EndDate) {
		return fmt.Errorf("la date de début ne peut pas être après la date de fin")
	}

	return nil
}
