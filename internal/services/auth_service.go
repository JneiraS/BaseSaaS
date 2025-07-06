package services

import (
	"time"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/domain/repositories"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// AuthService encapsulates the business logic for authentication.
// It manages interactions with the OIDC provider and the user repository.
type AuthService struct {
	Provider     *oidc.Provider         // OIDC provider instance
	Oauth2Config *oauth2.Config         // OAuth2 configuration for the client
	userRepo     repositories.UserRepository // User repository for database operations
}

// NewAuthService creates a new instance of AuthService.
// It takes an OIDC provider, OAuth2 config, and a UserRepository as dependencies.
func NewAuthService(provider *oidc.Provider, config *oauth2.Config, userRepo repositories.UserRepository) *AuthService {
	return &AuthService{
		Provider:     provider,
		Oauth2Config: config,
		userRepo:     userRepo,
	}
}

// IsConfigured checks if the AuthService has been properly configured with an OIDC provider and OAuth2 config.
func (s *AuthService) IsConfigured() bool {
	return s.Provider != nil && s.Oauth2Config != nil
}

// FindOrCreateUserFromClaims searches for a user by their OIDC ID (subject claim).
// If the user exists, it updates their information (email, name, last connection).
// If the user does not exist, it creates a new user record in the database.
func (s *AuthService) FindOrCreateUserFromClaims(claims struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Sub   string `json:"sub"`
}) (*models.User, error) {
	// Attempt to find the user by their OIDC subject ID.
	user, err := s.userRepo.FindUserByOIDCID(claims.Sub)

	if err == gorm.ErrRecordNotFound {
		// User does not exist, create a new one.
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
		// An unexpected error occurred during the search.
		return nil, err
	} else {
		// User exists, update their information if necessary.
		user.Email = claims.Email
		user.Name = claims.Name
		user.LastConnection = time.Now()

		if updateErr := s.userRepo.UpdateUser(user); updateErr != nil {
			return nil, updateErr
		}
		return user, nil
	}
}
