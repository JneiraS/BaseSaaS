package repositories

import (
	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"gorm.io/gorm"
)

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
	var user models.User
	result := r.db.Where("oidc_id = ?", oidcID).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// FindUserByID recherche un utilisateur par son ID.
func (r *GormUserRepository) FindUserByID(id uint) (*models.User, error) {
	var user models.User
	result := r.db.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// CreateUser crée un nouvel utilisateur.
func (r *GormUserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

// UpdateUser met à jour un utilisateur existant.
func (r *GormUserRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}
