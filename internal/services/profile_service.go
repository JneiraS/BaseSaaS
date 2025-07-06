package services

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/domain/repositories"
)

// ProfileService encapsulates the business logic for user profiles.
// It interacts with the UserRepository to perform operations related to user data.
type ProfileService struct {
	userRepo repositories.UserRepository
}

// NewProfileService creates a new instance of ProfileService.
// It takes a UserRepository as a dependency, adhering to the dependency inversion principle.
func NewProfileService(userRepo repositories.UserRepository) *ProfileService {
	return &ProfileService{userRepo: userRepo}
}

// validateUserInput performs validation on user input data.
// It checks for non-empty name, name length, and valid email format.
func (ps *ProfileService) validateUserInput(user models.User) error {
	if strings.TrimSpace(user.Name) == "" {
		return fmt.Errorf("le nom ne peut pas être vide")
	}

	if len(user.Name) > 100 {
		return fmt.Errorf("le nom ne peut pas dépasser 100 caractères")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	if !emailRegex.MatchString(user.Email) {
		return fmt.Errorf("format d'email invalide")
	}

	return nil
}

// UpdateUser updates a user's profile in the database.
// It first validates the input data, then retrieves the existing user,
// updates only the modifiable fields, and finally persists the changes.
func (ps *ProfileService) UpdateUser(userID uint, updatedData models.User) (*models.User, error) {
	if err := ps.validateUserInput(updatedData); err != nil {
		return nil, err
	}

	user, err := ps.userRepo.FindUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("utilisateur non trouvé: %w", err)
	}

	// Update only the modifiable fields from the provided updatedData.
	user.Username = strings.TrimSpace(updatedData.Username)

	// Save the changes to the database.
	if err := ps.userRepo.UpdateUser(user); err != nil {
		return nil, fmt.Errorf("erreur lors de la sauvegarde: %w", err)
	}

	return user, nil
}