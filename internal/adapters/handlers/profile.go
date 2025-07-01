package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/JneiraS/BaseSasS/components"
	"github.com/JneiraS/BaseSasS/internal/database"
	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
	"gorm.io/gorm"
)

// ProfileService encapsule la logique métier pour les profils utilisateur
type ProfileService struct {
	db *gorm.DB
}

// NewProfileService crée une nouvelle instance du service profil
func NewProfileService(db *gorm.DB) *ProfileService {
	return &ProfileService{db: db}
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

	var user models.User

	// Utilisation d'une transaction pour garantir la cohérence
	err := ps.db.Transaction(func(tx *gorm.DB) error {
		// Récupérer l'utilisateur existant
		if err := tx.First(&user, userID).Error; err != nil {
			return fmt.Errorf("utilisateur non trouvé: %w", err)
		}

		// Mettre à jour uniquement les champs modifiables
		user.Name = strings.TrimSpace(updatedData.Name)
		user.Email = strings.TrimSpace(updatedData.Email)

		// Sauvegarder les modifications
		result := tx.Save(&user)
		if result.Error != nil {
			return fmt.Errorf("erreur lors de la sauvegarde: %w", result.Error)
		}

		// Vérifier qu'exactement une ligne a été affectée
		if result.RowsAffected != 1 {
			return fmt.Errorf("nombre de lignes affectées inattendu: %d", result.RowsAffected)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Page profil (protégée)
func ProfileHandler(c *gin.Context) {
	session := sessions.Default(c)
	user, ok := session.Get("user").(models.User)
	if !ok {
		// Si l'utilisateur n'est pas en session ou n'est pas du bon type, rediriger vers la connexion
		c.Redirect(http.StatusFound, "/login")
		return
	}
	csrfToken := csrf.GetToken(c)
	navbar := components.NavBar(user, csrfToken)

	c.HTML(http.StatusOK, "profile.tmpl", gin.H{
		"title":      "Profil",
		"user":       user,
		"navbar":     navbar,
		"csrf_token": csrfToken,
	})
}

// UpdateProfileHandler gère la mise à jour du profil utilisateur (version améliorée)
func UpdateProfileHandler(c *gin.Context) {
	session := sessions.Default(c)
	loggedInUser, ok := session.Get("user").(models.User)
	if !ok {
		log.Printf("DEBUG: Utilisateur non trouvé en session ou typage incorrect.")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var updatedUser models.User
	if err := c.ShouldBind(&updatedUser); err != nil {
		log.Printf("ERREUR: Erreur de binding du formulaire: %v", err)
		ps := NewProfileService(database.DB)
		ps.handleError(c, session, "Erreur lors de la lecture des données du formulaire.")
		return
	}

	log.Printf("DEBUG: Données reçues du formulaire - Nom: %s, Email: %s", updatedUser.Name, updatedUser.Email)

	// Utiliser le service pour mettre à jour l'utilisateur
	profileService := NewProfileService(database.DB)
	updatedUserFromDB, err := profileService.UpdateUser(loggedInUser.ID, updatedUser)
	if err != nil {
		log.Printf("ERREUR: Erreur lors de la mise à jour du profil: %v", err)
		profileService.handleError(c, session, fmt.Sprintf("Erreur lors de la mise à jour: %s", err.Error()))
		return
	}

	log.Printf("DEBUG: Utilisateur mis à jour avec succès: %+v", updatedUserFromDB)

	// Mettre à jour la session une seule fois avec toutes les modifications
	session.Set("user", *updatedUserFromDB)
	session.AddFlash("Votre profil a été mis à jour avec succès !", "success")

	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
		// Ne pas faire échouer la requête si seule la session pose problème
	}

	c.Redirect(http.StatusFound, "/profile")
}

// handleError centralise la gestion des erreurs
func (ps *ProfileService) handleError(c *gin.Context, session sessions.Session, message string) {
	session.AddFlash(message, "error")
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de la session d'erreur: %v", err)
	}
	c.Redirect(http.StatusFound, "/profile")
}
