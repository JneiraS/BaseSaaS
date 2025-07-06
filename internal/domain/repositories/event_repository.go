package repositories

import (
	"time"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"gorm.io/gorm"
)

// EventDB représente le modèle d'événement pour la persistance GORM.
type EventDB struct {
	gorm.Model
	Title       string
	Description string
	StartDate   time.Time
	EndDate     time.Time
	UserID      uint
}

// TableName spécifie le nom de la table pour le modèle EventDB.
func (EventDB) TableName() string {
	return "events"
}

// EventRepository définit l'interface pour les opérations de persistance des événements.
type EventRepository interface {
	CreateEvent(event *models.Event) error
	FindEventByID(id uint) (*models.Event, error)
	FindEventsByUserID(userID uint) ([]models.Event, error)
	UpdateEvent(event *models.Event) error
	DeleteEvent(id uint) error
}

// GormEventRepository est une implémentation de EventRepository utilisant GORM.
type GormEventRepository struct {
	db *gorm.DB
}

// NewGormEventRepository crée une nouvelle instance de GormEventRepository.
func NewGormEventRepository(db *gorm.DB) *GormEventRepository {
	return &GormEventRepository{db: db}
}

// CreateEvent crée un nouvel événement.
func (r *GormEventRepository) CreateEvent(event *models.Event) error {
	eventDB := toEventDB(event)
	if err := r.db.Create(&eventDB).Error; err != nil {
		return err
	}
	*event = *toEvent(eventDB)
	return nil
}

// FindEventByID recherche un événement par son ID.
func (r *GormEventRepository) FindEventByID(id uint) (*models.Event, error) {
	var eventDB EventDB
	result := r.db.First(&eventDB, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return toEvent(&eventDB), nil
}

// FindEventsByUserID recherche tous les événements pour un utilisateur donné.
func (r *GormEventRepository) FindEventsByUserID(userID uint) ([]models.Event, error) {
	var eventsDB []EventDB
	if err := r.db.Where("user_id = ?", userID).Find(&eventsDB).Error; err != nil {
		return nil, err
	}
	var events []models.Event
	for _, edb := range eventsDB {
		events = append(events, *toEvent(&edb))
	}
	return events, nil
}

// UpdateEvent met à jour un événement existant.
func (r *GormEventRepository) UpdateEvent(event *models.Event) error {
	eventDB := toEventDB(event)
	return r.db.Save(&eventDB).Error
}

// DeleteEvent supprime un événement par son ID.
func (r *GormEventRepository) DeleteEvent(id uint) error {
	return r.db.Delete(&EventDB{}, id).Error
}

// toEventDB convertit un modèle de domaine Event en un modèle de base de données EventDB.
func toEventDB(e *models.Event) *EventDB {
	return &EventDB{
		Model:       gorm.Model{ID: e.ID, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt, DeletedAt: e.DeletedAt},
		Title:       e.Title,
		Description: e.Description,
		StartDate:   e.StartDate,
		EndDate:     e.EndDate,
		UserID:      e.UserID,
	}
}

// toEvent convertit un modèle de base de données EventDB en un modèle de domaine Event.
func toEvent(edb *EventDB) *models.Event {
	return &models.Event{
		Model:       gorm.Model{ID: edb.ID, CreatedAt: edb.CreatedAt, UpdatedAt: edb.UpdatedAt, DeletedAt: edb.DeletedAt},
		Title:       edb.Title,
		Description: edb.Description,
		StartDate:   edb.StartDate,
		EndDate:     edb.EndDate,
		UserID:      edb.UserID,
	}
}
