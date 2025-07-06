package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/JneiraS/BaseSasS/components"
	"github.com/JneiraS/BaseSasS/internal/domain/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Page profil (protégée)
func (app *App) ProfileHandler(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		// Si l'utilisateur n'est pas en session ou n'est pas du bon type, rediriger vers la connexion
		c.Redirect(http.StatusFound, "/login")
		return
	}
	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	c.HTML(http.StatusOK, "profile.tmpl", gin.H{
		"title":      "Profil",
		"user":       user,
		"navbar":     navbar,
		"csrf_token": csrfToken,
	})
	if err := session.Save(); err != nil {
		// Gérer l'erreur de sauvegarde de session si nécessaire
		// log.Printf("Erreur lors de la sauvegarde de session dans ProfileHandler: %v", err)
	}
}

// UpdateProfileHandler gère la mise à jour du profil utilisateur (version améliorée)
func (app *App) UpdateProfileHandler(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	loggedInUser, ok := session.Get("user").(models.User)
	if !ok {
		log.Printf("DEBUG: Utilisateur non trouvé en session ou typage incorrect.")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var updatedUser models.User
	if err := c.ShouldBind(&updatedUser); err != nil {
		log.Printf("ERREUR: Erreur de binding du formulaire: %v", err)
		app.handleProfileError(c, session, "Erreur lors de la lecture des données du formulaire.")
		return
	}

	log.Printf("DEBUG: Données reçues du formulaire - Nom: %s, Email: %s", updatedUser.Name, updatedUser.Email)

	// Utiliser le service pour mettre à jour l'utilisateur
	profileService := app.profileService
	updatedUserFromDB, err := profileService.UpdateUser(loggedInUser.ID, updatedUser)
	if err != nil {
		log.Printf("ERREUR: Erreur lors de la mise à jour du profil: %v", err)
		app.handleProfileError(c, session, fmt.Sprintf("Erreur lors de la mise à jour: %s", err.Error()))
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

// handleProfileError centralise la gestion des erreurs pour le profil
func (app *App) handleProfileError(c *gin.Context, session sessions.Session, message string) {
	session.AddFlash(message, "error")
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de la session d'erreur: %v", err)
	}
	c.Redirect(http.StatusFound, "/profile")
}
