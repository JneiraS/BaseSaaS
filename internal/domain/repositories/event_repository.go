package repositories

import (
	"time"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"gorm.io/gorm"
)

// EventDB represents the database model for an event, used for GORM persistence.
// It includes GORM's Model for common fields like ID, CreatedAt, UpdatedAt, DeletedAt.
type EventDB struct {
	gorm.Model
	Title       string    // Title of the event
	Description string    // Description of the event
	StartDate   time.Time // Start date and time of the event
	EndDate     time.Time // End date and time of the event
	UserID      uint      // Foreign key linking to the User who created the event
}

// TableName specifies the table name for the EventDB model in the database.
// This overrides GORM's default naming convention.
func (EventDB) TableName() string {
	return "events"
}

// EventRepository defines the interface for event persistence operations.
// It abstracts the underlying database implementation, allowing for different
// data storage mechanisms (e.g., GORM, SQL, NoSQL) to be used interchangeably.
type EventRepository interface {
	CreateEvent(event *models.Event) error
	FindEventByID(id uint) (*models.Event, error)
	FindEventsByUserID(userID uint) ([]models.Event, error)
	UpdateEvent(event *models.Event) error
	DeleteEvent(id uint) error
	GetTotalEventsCount(userID uint) (int64, error)
}

// GormEventRepository is an implementation of EventRepository that uses GORM
// for interacting with a relational database.
type GormEventRepository struct {
	db *gorm.DB // GORM database client
}

// NewGormEventRepository creates a new instance of GormEventRepository.
// It takes a GORM DB instance as a dependency.
func NewGormEventRepository(db *gorm.DB) *GormEventRepository {
	return &GormEventRepository{db: db}
}

// CreateEvent persists a new event to the database.
// It converts the domain model Event to a database-specific EventDB model
// before saving and then updates the domain model with the generated ID.
func (r *GormEventRepository) CreateEvent(event *models.Event) error {
	eventDB := toEventDB(event)
	if err := r.db.Create(&eventDB).Error; err != nil {
		return err
	}
	*event = *toEvent(eventDB) // Update the original event with DB-generated fields (e.g., ID)
	return nil
}

// FindEventByID retrieves an event from the database by its ID.
// It returns the event as a domain model or an error if not found.
func (r *GormEventRepository) FindEventByID(id uint) (*models.Event, error) {
	var eventDB EventDB
	result := r.db.First(&eventDB, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return toEvent(&eventDB), nil
}

// FindEventsByUserID retrieves all events associated with a specific user ID.
// It queries the database for events where the UserID matches the provided ID.
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

// UpdateEvent updates an existing event in the database.
// It converts the domain model to a database model and saves the changes.
func (r *GormEventRepository) UpdateEvent(event *models.Event) error {
	eventDB := toEventDB(event)
	return r.db.Save(&eventDB).Error
}

// DeleteEvent deletes an event from the database by its ID.
func (r *GormEventRepository) DeleteEvent(id uint) error {
	return r.db.Delete(&EventDB{}, id).Error
}

// GetTotalEventsCount returns the total number of events for a given user ID.
// It performs a count query on the events table, filtered by user_id.
func (r *GormEventRepository) GetTotalEventsCount(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&EventDB{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// toEventDB converts a domain Event model to a database-specific EventDB model.
// This is used before persisting the event to the database.
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

// toEvent converts a database-specific EventDB model back to a domain Event model.
// This is used after retrieving data from the database.
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
