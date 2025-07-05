package services

import (
	"time"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/domain/repositories"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type AuthService struct {
	Provider     *oidc.Provider
	Oauth2Config *oauth2.Config
	userRepo     repositories.UserRepository
}

func NewAuthService(provider *oidc.Provider, config *oauth2.Config, userRepo repositories.UserRepository) *AuthService {
	return &AuthService{
		Provider:     provider,
		Oauth2Config: config,
		userRepo:     userRepo,
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
	user, err := s.userRepo.FindUserByOIDCID(claims.Sub)

	if err == gorm.ErrRecordNotFound {
		// L'utilisateur n'existe pas, le créer
		newUser := &models.User{
			OIDCID:         claims.Sub,
			Email:          claims.Email,
			Name:           claims.Name,
			LastConnection: time.Now(),
		}
		if createErr := s.userRepo.CreateUser(newUser); createErr != nil {
			return nil, createErr
		}
		return newUser, nil
	} else if err != nil {
		// Erreur lors de la recherche
		return nil, err
	} else {
		// L'utilisateur existe, mettre à jour ses informations si nécessaire
		user.Email = claims.Email
		user.Name = claims.Name
		user.LastConnection = time.Now()

		if updateErr := s.userRepo.UpdateUser(user); updateErr != nil {
			return nil, updateErr
		}
		return user, nil
	}
}
