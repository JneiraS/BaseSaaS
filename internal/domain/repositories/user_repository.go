package repositories

import (
	"time"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"gorm.io/gorm"
)

// UserDB represents the database model for a user, used for GORM persistence.
// It includes GORM's Model for common fields like ID, CreatedAt, UpdatedAt, DeletedAt.
type UserDB struct {
	gorm.Model
	OIDCID         string `gorm:"column:oidc_id;uniqueIndex"` // OpenID Connect ID, unique identifier from the OIDC provider
	Email          string                                  // User's email address
	Name           string                                  // User's full name
	Username       string                                  // User's chosen username
	LastConnection time.Time                               // Timestamp of the user's last successful connection
}

// UserRepository defines the interface for user persistence operations.
// It abstracts the underlying database implementation, allowing for different
// data storage mechanisms (e.g., GORM, SQL, NoSQL) to be used interchangeably.
type UserRepository interface {
	FindUserByOIDCID(oidcID string) (*models.User, error)
	FindUserByID(id uint) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
}

// GormUserRepository is an implementation of UserRepository that uses GORM
// for interacting with a relational database.
type GormUserRepository struct {
	db *gorm.DB // GORM database client
}

// NewGormUserRepository creates a new instance of GormUserRepository.
// It takes a GORM DB instance as a dependency.
func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

// FindUserByOIDCID retrieves a user from the database by their OIDC ID.
// It returns the user as a domain model or an error if not found.
func (r *GormUserRepository) FindUserByOIDCID(oidcID string) (*models.User, error) {
	var userDB UserDB
	result := r.db.Where("oidc_id = ?", oidcID).First(&userDB)
	if result.Error != nil {
		return nil, result.Error
	}
	// Convert UserDB to domain User model.
	return &models.User{
		ID:             userDB.ID,
		OIDCID:         userDB.OIDCID,
		Email:          userDB.Email,
		Name:           userDB.Name,
		Username:       userDB.Username,
		LastConnection: userDB.LastConnection,
		CreatedAt:      userDB.CreatedAt,
		UpdatedAt:      userDB.UpdatedAt,
		DeletedAt:      userDB.DeletedAt,
	}, nil
}

// FindUserByID retrieves a user from the database by their internal ID.
// It returns the user as a domain model or an error if not found.
func (r *GormUserRepository) FindUserByID(id uint) (*models.User, error) {
	var userDB UserDB
	result := r.db.First(&userDB, id)
	if result.Error != nil {
		return nil, result.Error
	}
	// Convert UserDB to domain User model.
	return &models.User{
		ID:             userDB.ID,
		OIDCID:         userDB.OIDCID,
		Email:          userDB.Email,
		Name:           userDB.Name,
		Username:       userDB.Username,
		LastConnection: userDB.LastConnection,
		CreatedAt:      userDB.CreatedAt,
		UpdatedAt:      userDB.UpdatedAt,
		DeletedAt:      userDB.DeletedAt,
	}, nil
}

// CreateUser persists a new user to the database.
// It converts the domain model User to a database-specific UserDB model
// before saving and then updates the domain model with the generated ID and timestamps.
func (r *GormUserRepository) CreateUser(user *models.User) error {
	userDB := UserDB{
		OIDCID:         user.OIDCID,
		Email:          user.Email,
		Name:           user.Name,
		Username:       user.Username,
		LastConnection: user.LastConnection,
	}
	if err := r.db.Create(&userDB).Error; err != nil {
		return err
	}
	// Update the original domain model with DB-generated fields (e.g., ID, CreatedAt).
	user.ID = userDB.ID
	user.CreatedAt = userDB.CreatedAt
	user.UpdatedAt = userDB.UpdatedAt
	user.DeletedAt = userDB.DeletedAt
	return nil
}

// UpdateUser updates an existing user in the database.
// It converts the domain model to a database model and saves the changes.
func (r *GormUserRepository) UpdateUser(user *models.User) error {
	userDB := UserDB{
		Model:          gorm.Model{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, DeletedAt: user.DeletedAt},
		OIDCID:         user.OIDCID,
		Email:          user.Email,
		Name:           user.Name,
		Username:       user.Username,
		LastConnection: user.LastConnection,
	}
	return r.db.Save(&userDB).Error
}
