package services

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/JneiraS/BaseSasS/internal/domain/repositories"
)

// ProfileService encapsule la logique métier pour les profils utilisateur
type ProfileService struct {
	userRepo repositories.UserRepository
}

// NewProfileService crée une nouvelle instance du service profil
func NewProfileService(userRepo repositories.UserRepository) *ProfileService {
	return &ProfileService{userRepo: userRepo}
}

// validateUserInput valide les données utilisateur
func (ps *ProfileService) validateUserInput(user models.User) error {
	if strings.TrimSpace(user.Name) == "" {
		return fmt.Errorf("le nom ne peut pas être vide")
	}

	if len(user.Name) > 100 {
		return fmt.Errorf("le nom ne peut pas dépasser 100 caractères")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(user.Email) {
		return fmt.Errorf("format d'email invalide")
	}

	return nil
}

// UpdateUser met à jour un utilisateur en base de données avec transaction
func (ps *ProfileService) UpdateUser(userID uint, updatedData models.User) (*models.User, error) {
	if err := ps.validateUserInput(updatedData); err != nil {
		return nil, err
	}

	user, err := ps.userRepo.FindUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("utilisateur non trouvé: %w", err)
	}

	// Mettre à jour uniquement les champs modifiables
	user.Username = strings.TrimSpace(updatedData.Username)

	// Sauvegarder les modifications
	if err := ps.userRepo.UpdateUser(user); err != nil {
		return nil, fmt.Errorf("erreur lors de la sauvegarde: %w", err)
	}

	return user, nil
}
