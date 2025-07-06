package repositories

import (
	"time"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"gorm.io/gorm"
)

// UserDB représente le modèle utilisateur pour la persistance GORM.
type UserDB struct {
	gorm.Model
	OIDCID         string `gorm:"column:oidc_id;uniqueIndex"`
	Email          string
	Name           string
	Username       string
	LastConnection time.Time
}

// UserRepository définit l'interface pour les opérations de persistance des utilisateurs.
type UserRepository interface {
	FindUserByOIDCID(oidcID string) (*models.User, error)
	FindUserByID(id uint) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
}

// GormUserRepository est une implémentation de UserRepository utilisant GORM.
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository crée une nouvelle instance de GormUserRepository.
func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

// FindUserByOIDCID recherche un utilisateur par son OIDCID.
func (r *GormUserRepository) FindUserByOIDCID(oidcID string) (*models.User, error) {
	var userDB UserDB
	result := r.db.Where("oidc_id = ?", oidcID).First(&userDB)
	if result.Error != nil {
		return nil, result.Error
	}
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

// FindUserByID recherche un utilisateur par son ID.
func (r *GormUserRepository) FindUserByID(id uint) (*models.User, error) {
	var userDB UserDB
	result := r.db.First(&userDB, id)
	if result.Error != nil {
		return nil, result.Error
	}
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

// CreateUser crée un nouvel utilisateur.
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
	// Mettre à jour l'ID du modèle de domaine après la création
	user.ID = userDB.ID
	user.CreatedAt = userDB.CreatedAt
	user.UpdatedAt = userDB.UpdatedAt
	user.DeletedAt = userDB.DeletedAt
	return nil
}

// UpdateUser met à jour un utilisateur existant.
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
