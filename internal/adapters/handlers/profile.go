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

// ProfileHandler displays the user's profile page.
// It retrieves the authenticated user from the session and renders the "profile.tmpl" template.
func (app *App) ProfileHandler(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	user, ok := session.Get("user").(models.User)
	if !ok {
		// If the user is not in the session or is of the wrong type, redirect to login.
		c.Redirect(http.StatusFound, "/login")
		return
	}
	csrfToken := c.MustGet("csrf_token").(string)
	navbar := components.NavBar(user, csrfToken, session)

	// Render the profile page with user information, navbar, and CSRF token.
	c.HTML(http.StatusOK, "profile.tmpl", gin.H{
		"title":      "Profil",
		"user":       user,
		"navbar":     navbar,
		"csrf_token": csrfToken,
	})
	// Save session changes if any (e.g., flash messages).
	if err := session.Save(); err != nil {
		// Handle session save error if necessary
		// log.Printf("Erreur lors de la sauvegarde de session dans ProfileHandler: %v", err)
	}
}

// UpdateProfileHandler handles the submission of the user profile update form.
// It binds the form data, validates it, updates the user via the ProfileService,
// and updates the session with the new user data.
func (app *App) UpdateProfileHandler(c *gin.Context) {
	session := c.MustGet("session").(sessions.Session)
	loggedInUser, ok := session.Get("user").(models.User)
	if !ok {
		log.Printf("DEBUG: Utilisateur non trouvé en session ou typage incorrect.")
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var updatedUser models.User
	// Bind the form data to the updatedUser struct. If binding fails, log and handle the error.
	if err := c.ShouldBind(&updatedUser); err != nil {
		log.Printf("ERREUR: Erreur de binding du formulaire: %v", err)
		app.handleProfileError(c, session, "Erreur lors de la lecture des données du formulaire.")
		return
	}

	log.Printf("DEBUG: Données reçues du formulaire - Nom: %s, Email: %s", updatedUser.Name, updatedUser.Email)

	// Use the ProfileService to update the user.
	profileService := app.profileService
	updatedUserFromDB, err := profileService.UpdateUser(loggedInUser.ID, updatedUser)
	if err != nil {
		log.Printf("ERREUR: Erreur lors de la mise à jour du profil: %v", err)
		app.handleProfileError(c, session, fmt.Sprintf("Erreur lors de la mise à jour: %s", err.Error()))
		return
	}

	log.Printf("DEBUG: Utilisateur mis à jour avec succès: %+v", updatedUserFromDB)

	// Update the session with the modified user data and add a success flash message.
	session.Set("user", *updatedUserFromDB)
	session.AddFlash("Votre profil a été mis à jour avec succès !", "success")

	// Save the session. Log any errors but do not fail the request if only session saving fails.
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de la session: %v", err)
	}

	// Redirect back to the profile page.
	c.Redirect(http.StatusFound, "/profile")
}

// handleProfileError centralizes error handling for profile-related operations.
// It adds an error flash message to the session and redirects to the profile page.
func (app *App) handleProfileError(c *gin.Context, session sessions.Session, message string) {
	session.AddFlash(message, "error")
	if err := session.Save(); err != nil {
		log.Printf("ERREUR: Erreur lors de la sauvegarde de la session d'erreur: %v", err)
	}
	c.Redirect(http.StatusFound, "/profile")
}
