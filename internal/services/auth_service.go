package services

import (
	"time"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type AuthService struct {
	Provider     *oidc.Provider
	Oauth2Config *oauth2.Config
	db           *gorm.DB
}

func NewAuthService(provider *oidc.Provider, config *oauth2.Config, db *gorm.DB) *AuthService {
	return &AuthService{
		Provider:     provider,
		Oauth2Config: config,
		db:           db,
	}
}

func (s *AuthService) IsConfigured() bool {
	return s.Provider != nil && s.Oauth2Config != nil
}

// FindOrCreateUserFromClaims recherche un utilisateur par son OIDCID, le crée s'il n'existe pas,
// ou met à jour ses informations s'il existe.
func (s *AuthService) FindOrCreateUserFromClaims(claims struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Sub   string `json:"sub"`
}) (*models.User, error) {
	var user models.User
	result := s.db.Where("oidc_id = ?", claims.Sub).First(&user)

	if result.Error == gorm.ErrRecordNotFound {
		// L'utilisateur n'existe pas, le créer
		user = models.User{
			OIDCID:         claims.Sub,
			Email:          claims.Email,
			Name:           claims.Name,
			LastConnection: time.Now(),
		}
		if createResult := s.db.Create(&user); createResult.Error != nil {
			return nil, createResult.Error
		}
	} else if result.Error != nil {
		// Erreur lors de la recherche
		return nil, result.Error
	} else {
		// L'utilisateur existe, mettre à jour ses informations si nécessaire
		user.Email = claims.Email
		user.Name = claims.Name
		user.LastConnection = time.Now()

		if updateResult := s.db.Save(&user); updateResult.Error != nil {
			return nil, updateResult.Error
		}
	}
	return &user, nil
}
