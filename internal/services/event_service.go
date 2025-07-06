package services

import (
	"fmt"
	"strings"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/domain/repositories"
)

// EventService encapsule la logique métier pour la gestion des événements.
type EventService struct {
	eventRepo repositories.EventRepository
}

// NewEventService crée une nouvelle instance de EventService.
func NewEventService(eventRepo repositories.EventRepository) *EventService {
	return &EventService{eventRepo: eventRepo}
}

// CreateEvent gère la création d'un nouvel événement.
func (s *EventService) CreateEvent(event *models.Event) error {
	if err := s.validateEvent(event); err != nil {
		return err
	}
	return s.eventRepo.CreateEvent(event)
}

// GetEventByID récupère un événement par son ID.
func (s *EventService) GetEventByID(id uint) (*models.Event, error) {
	return s.eventRepo.FindEventByID(id)
}

// GetEventsByUserID récupère tous les événements d'un utilisateur.
func (s *EventService) GetEventsByUserID(userID uint) ([]models.Event, error) {
	return s.eventRepo.FindEventsByUserID(userID)
}

// UpdateEvent gère la mise à jour d'un événement.
func (s *EventService) UpdateEvent(event *models.Event) error {
	if err := s.validateEvent(event); err != nil {
		return err
	}
	return s.eventRepo.UpdateEvent(event)
}

// DeleteEvent gère la suppression d'un événement.
func (s *EventService) DeleteEvent(id uint) error {
	return s.eventRepo.DeleteEvent(id)
}

// validateEvent valide les données d'un événement.
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
